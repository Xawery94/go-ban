package account

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/Xawery/auth-service/authz"
	"github.com/Xawery/auth-service/model"
	sq "github.com/elgris/sqrl"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
)

type server struct {
	conf    config
	accts   model.AccountsRepo
	creds   model.CredentialsRepo
	act     model.ActivationRepo
	db      *sql.DB
	stmt    sq.StatementBuilderType
	emitter event.Emitter
}

// NewPg constructs an Postgres-backed implementation of account service
func NewPg(db *sql.DB, emitter event.Emitter, opts ...Opt) AccountsServer {
	conf := config{
		locales: map[string]struct{}{
			DefaultLocale: {},
		},
	}

	for _, opt := range opts {
		opt(&conf)
	}

	return &server{
		conf:    conf,
		db:      db,
		stmt:    sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db),
		emitter: emitter,
		accts:   model.NewPgAccountsRepo(),
		creds:   model.NewPgCredentialsRepo(),
		act:     model.NewPgActivationRepo(),
	}
}

func (s *server) RegisterEmail(ctx context.Context, req *RegisterEmailRequest) (*empty.Empty, error) {
	if req.Locale == "" {
		req.Locale = DefaultLocale
	}

	if _, ok := s.conf.locales[req.Locale]; !ok {
		req.Locale = DefaultLocale
	}

	var (
		id   int64
		code int64
	)
	err := s.tx(ctx, func(stmt sq.StatementBuilderType) (err error) {
		id, err = s.accts.Create(ctx, stmt, &model.Account{
			Email:      req.Email,
			FamilyName: req.FamilyName,
			GivenName:  req.GivenName,
			Locale:     req.Locale,
		})
		if err != nil {
			return err
		}

		if err := s.creds.Create(ctx, stmt, id, req.Email, req.Password); err != nil {
			log.Printf("Create account error: %v", err)
			return err
		}

		if err := s.accts.CreateDefaultScope(ctx, stmt, id, 1); err != nil {
			log.Printf("Create scope error: %v", err)
			return err
		}

		code, err = s.act.Create(ctx, stmt, id, req.Email)
		return err
	})
	if err != nil {
		return nil, err
	}

	ts := time.Now()
	s.emit(ts, &Created{
		Id:       id,
		Username: req.Email,
	},
		req.Locale)

	s.emit(ts, &EmailActivationRequired{
		Id:    id,
		Email: req.Email,
		Code:  code,
	},
		req.Locale)

	log.Printf("Success created account id: %d", id)

	return &empty.Empty{}, nil
}

func (s *server) RegisterMobile(ctx context.Context, req *RegisterMobileRequest) (*empty.Empty, error) {
	if req.Locale == "" {
		req.Locale = DefaultLocale
	}

	if _, ok := s.conf.locales[req.Locale]; !ok {
		req.Locale = DefaultLocale
	}

	var (
		id   int64
		code int64
	)
	err := s.tx(ctx, func(stmt sq.StatementBuilderType) (err error) {
		id, err = s.accts.Create(ctx, stmt, &model.Account{
			Phone:      req.Phone,
			FamilyName: req.FamilyName,
			GivenName:  req.GivenName,
			Locale:     req.Locale,
			Company:    req.Company,
		})
		if err != nil {
			return err
		}

		if err := s.creds.Create(ctx, stmt, id, req.MobileId, req.Password); err != nil {
			log.Printf("Create account error: %v", err)
			return err
		}

		code, err = s.act.Create(ctx, stmt, id, req.MobileId)
		if err != nil {
			log.Printf("Creating activation code error: %v", err)
		}

		if err := s.accts.CreateDefaultScope(ctx, stmt, id, 1); err != nil {
			log.Printf("Create scope error: %v", err)
			return err
		}

		return err
	})
	if err != nil {
		return nil, err
	}

	ts := time.Now()
	s.emit(ts, &Created{
		Id:       id,
		Username: req.MobileId,
	},
		req.Locale)

	s.emit(ts, &PhoneActivationRequired{
		Id:    id,
		Phone: req.Phone,
		Code:  code,
	},
		req.Locale)

	log.Printf("Success created account id: %d", id)

	return &empty.Empty{}, nil
}

func (s *server) AddMobile(ctx context.Context, req *AddMobileRequest) (*empty.Empty, error) {
	var (
		id     int64
		code   int64
		locale string
	)

	err := s.tx(ctx, func(stmt sq.StatementBuilderType) (err error) {
		acct, err := s.accts.RetrieveByPhone(ctx, stmt, req.Phone)
		if err != nil {
			return err
		}
		id = acct.ID
		locale = acct.Locale

		if err := s.creds.Create(ctx, stmt, id, req.MobileId, req.Password); err != nil {
			return err
		}

		code, err = s.act.Create(ctx, stmt, id, req.MobileId)
		return err
	})
	if err != nil {
		log.Printf("Add mobile error: %v", err)
		return nil, err
	}

	s.emit(time.Now(), &PhoneActivationRequired{
		Id:    id,
		Phone: req.Phone,
		Code:  code,
	},
		locale)

	return &empty.Empty{}, nil
}

func (s *server) Activate(ctx context.Context, req *ActivateRequest) (*empty.Empty, error) {
	var (
		id     int64
		locale string
	)
	err := s.tx(ctx, func(stmt sq.StatementBuilderType) (err error) {
		id, err = s.act.Verify(ctx, stmt, req.Username, req.Code)
		if err != nil {
			return err
		}

		return s.creds.Update(ctx, stmt, req.Username, &model.CredentialsUpdate{Active: true})
	})
	if err != nil {
		return nil, err
	}

	err = s.tx(ctx, func(stmt sq.StatementBuilderType) (err error) {
		acct, err := s.accts.RetrieveByUsername(ctx, stmt, req.Username)
		if err != nil {
			return err
		}
		locale = acct.Locale
		return err
	})
	if err != nil {
		return nil, err
	}

	s.emit(time.Now(), &Activated{
		Id:       id,
		Username: req.Username,
	}, locale)

	return &empty.Empty{}, nil
}

func (s *server) Retrieve(ctx context.Context, req *RetrieveRequest) (*Account, error) {
	authz := authz.FromContext(ctx)
	switch {
	case authz.ID == 0:
		return nil, Unauthenticated.Err()
	case req.Id == 0:
		req.Id = authz.ID
	case req.Id != authz.ID && !authz.HasScope("accounts:read"):
		return nil, Unauthenticated.Err()
	}

	acct, err := s.accts.Retrieve(ctx, s.stmt, req.Id)
	if err != nil {
		log.Printf("Retrive code error: %v", err)
		return nil, err
	}

	return s.account(acct), nil
}

func (s *server) Search(req *SearchRequest, stream Accounts_SearchServer) error {
	ctx := stream.Context()
	return s.accts.Search(ctx, s.stmt, req.Search, func(acct *model.Account) error {
		return stream.Send(s.account(acct))
	})
}

func (s *server) UpdateEmail(ctx context.Context, req *UpdateEmailRequest) (*empty.Empty, error) {
	authz := authz.FromContext(ctx)
	if authz.ID == 0 {
		return nil, Unauthenticated.Err()
	}

	var (
		code   int64
		locale string
	)
	err := s.tx(ctx, func(stmt sq.StatementBuilderType) (err error) {
		if _, _, err := s.creds.Verify(ctx, stmt, authz.Username, req.Password); err != nil {
			return err
		}
		acct, err := s.accts.RetrieveByUsername(ctx, stmt, authz.Username)
		if err != nil {
			return err
		}
		locale = acct.Locale

		if err = s.accts.Update(ctx, stmt, authz.ID, &model.AccountUpdate{Email: req.Email}); err != nil {
			return err
		}

		if err := s.creds.Update(ctx, stmt, authz.Username, &model.CredentialsUpdate{Username: req.Email}); err != nil {
			return err
		}

		code, err = s.act.Create(ctx, stmt, authz.ID, req.Email)
		return err
	})
	if err != nil {
		return nil, err
	}

	ts := time.Now()
	s.emit(ts, &EmailUpdated{
		Id:       authz.ID,
		OldEmail: authz.Username,
		NewEmail: req.Email,
	}, locale)

	s.emit(ts, &EmailActivationRequired{
		Id:    authz.ID,
		Email: req.Email,
		Code:  code,
	}, locale)

	return &empty.Empty{}, nil
}

func (s *server) VerifyOldPassword(ctx context.Context, req *VerifyOldPasswordRequest) (*empty.Empty, error) {
	authz := authz.FromContext(ctx)
	if authz.ID == 0 {
		return nil, Unauthenticated.Err()
	}

	err := s.tx(ctx, func(stmt sq.StatementBuilderType) (err error) {
		if _, _, err := s.creds.Verify(ctx, stmt, authz.Username, req.OldPassword); err != nil {
			return err
		}

		return err
	})
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *server) UpdatePassword(ctx context.Context, req *UpdatePasswordRequest) (*empty.Empty, error) {
	authz := authz.FromContext(ctx)
	if authz.ID == 0 {
		return nil, Unauthenticated.Err()
	}

	var locale string

	err := s.tx(ctx, func(stmt sq.StatementBuilderType) (err error) {
		if _, _, err := s.creds.Verify(ctx, stmt, authz.Username, req.OldPassword); err != nil {
			return err
		}

		acct, err := s.accts.RetrieveByUsername(ctx, stmt, authz.Username)
		if err != nil {
			return err
		}
		locale = acct.Locale

		return s.creds.Update(ctx, stmt, authz.Username, &model.CredentialsUpdate{
			Password: req.NewPassword,
			Active:   true,
		})
	})
	if err != nil {
		return nil, err
	}

	s.emit(time.Now(), &PasswordUpdated{
		Id:       authz.ID,
		Username: authz.Username,
	}, locale)

	return &empty.Empty{}, nil
}

func (s *server) Update(ctx context.Context, req *UpdateRequest) (*empty.Empty, error) {
	authz := authz.FromContext(ctx)

	switch {
	case authz.ID == 0:
		return nil, Unauthenticated.Err()
	case req.Id == 0:
		req.Id = authz.ID
	case req.Id != authz.ID && !authz.HasScope("accounts:read"):
		return nil, Unauthenticated.Err()
	}

	if req.Locale != "" {
		if _, ok := s.conf.locales[req.Locale]; !ok {
			return nil, InvalidLocale.Err()
		}
	}

	err := s.accts.Update(ctx, s.stmt, req.Id, &model.AccountUpdate{
		FamilyName: req.FamilyName,
		GivenName:  req.GivenName,
		Locale:     req.Locale,
		Company:    req.Company,
	})
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *server) Delete(ctx context.Context, req *DeleteRequest) (*empty.Empty, error) {
	authz := authz.FromContext(ctx)
	if authz.ID == 0 {
		return nil, Unauthenticated.Err()
	}

	var locale string

	err := s.tx(ctx, func(stmt sq.StatementBuilderType) error {
		if _, _, err := s.creds.Verify(ctx, stmt, authz.Username, req.Password); err != nil {
			return err
		}

		acct, err := s.accts.RetrieveByUsername(ctx, stmt, authz.Username)
		if err != nil {
			return err
		}
		locale = acct.Locale

		if err := s.accts.Delete(ctx, stmt, authz.ID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.emit(time.Now(), &Deleted{
		Id:       authz.ID,
		Username: authz.Username,
	}, locale)

	return &empty.Empty{}, nil
}

func (s *server) InitResetMobilePassword(ctx context.Context, req *InitResetMobilePasswordRequest) (*empty.Empty, error) {
	var (
		id     int64
		phone  string
		code   int64
		locale string
	)
	err := s.tx(ctx, func(stmt sq.StatementBuilderType) error {
		acct, err := s.accts.RetrieveByUsername(ctx, stmt, req.MobileId)
		if err != nil {
			return err
		}
		id, phone, locale = acct.ID, acct.Phone, acct.Locale
		code, err = s.act.Create(ctx, stmt, id, req.MobileId)
		return err
	})
	if err != nil {
		log.Printf("Init mobile pass reset error: %v", err)
		return nil, err
	}

	if id != 0 {
		s.emit(time.Now(), &ResetMobilePasswordRequested{
			Id:    id,
			Phone: phone,
			Code:  code,
		}, locale)
	}

	return &empty.Empty{}, nil
}

func (s *server) ResetPassword(ctx context.Context, req *ResetPasswordRequest) (*empty.Empty, error) {
	var (
		id     int64
		locale string
	)
	err := s.tx(ctx, func(stmt sq.StatementBuilderType) error {
		var err error
		id, err = s.act.Verify(ctx, stmt, req.Username, req.Code)
		if err != nil {
			return err
		}

		acct, err := s.accts.RetrieveByUsername(ctx, stmt, req.Username)
		if err != nil {
			return err
		}
		locale = acct.Locale

		return s.creds.Update(ctx, stmt, req.Username, &model.CredentialsUpdate{
			Password: req.Password,
			Active:   true,
		})
	})
	if err != nil {
		log.Printf("Mobile pass reset error: %v", err)
		return nil, err
	}

	s.emit(time.Now(), &PasswordUpdated{
		Id:       id,
		Username: req.Username,
	}, locale)

	return &empty.Empty{}, nil
}

func (s *server) ResendCode(ctx context.Context, rcr *ResendCodeRequest) (*empty.Empty, error) {
	var (
		id     int64
		phone  string
		code   int64
		locale string
	)

	err := s.tx(ctx, func(stmt sq.StatementBuilderType) error {
		creds, err := s.creds.RetrieveNotActive(ctx, stmt, rcr.MobileId)
		if err != nil {
			return err
		}
		id = creds
		return err
	})

	if err != nil {
		log.Printf("ResendCode retrieve id error: %v", err)
		return nil, err
	}
	err = s.tx(ctx, func(stmt sq.StatementBuilderType) error {
		acct, err := s.accts.Retrieve(ctx, stmt, id)
		if err != nil {
			return err
		}
		phone, locale = acct.Phone, acct.Locale
		return err
	})

	if err != nil {
		log.Printf("ResendCode retrieve account error: %v", err)
		return nil, err
	}

	if id != 0 && phone != "" {
		err = s.tx(ctx, func(stmt sq.StatementBuilderType) error {
			acct, err := s.act.Retrieve(ctx, stmt, id, rcr.MobileId)
			if err != nil {
				return err
			}
			code = acct
			return err
		})
	}
	if err != nil {
		log.Printf("ResendCode retrieve code error: %v", err)
		return nil, err
	}
	if code != 0 {
		if rcr.Sender == 0 {
			s.emit(time.Now(), &PhoneActivationRequired{
				Id:    id,
				Phone: phone,
				Code:  code,
			}, locale)
		} else {
			s.emit(time.Now(), &ResetMobilePasswordRequested{
				Id:    id,
				Phone: phone,
				Code:  code,
			}, locale)
		}
	}
	return &empty.Empty{}, nil
}

func (s *server) account(src *model.Account) *Account {
	return &Account{
		Id:         src.ID,
		Phone:      src.Phone,
		Email:      src.Email,
		FamilyName: src.FamilyName,
		GivenName:  src.GivenName,
		Locale:     src.Locale,
	}
}

func (s *server) tx(ctx context.Context, f func(sq.StatementBuilderType) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return null
		// return grpcerror.New(Unavailable, errors.Wrap(err, "Failed to start transaction"))
	}

	stmt := s.stmt.RunWith(tx)

	if err := f(stmt); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return null
		// return grpcerror.New(Unavailable, errors.Wrap(err, "Failed to commit transaction"))
	}

	return nil
}

func (s *server) emit(ts time.Time, payload proto.Message, locale string) {
	if err := s.emitter.Emit(ts, payload, locale); err != nil {
		log.Printf("%+v", err)
	}
}
