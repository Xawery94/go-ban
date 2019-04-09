package authn

// import (
// 	"context"
// 	"crypto/rand"
// 	"crypto/sha256"
// 	"database/sql"
// 	"encoding/base64"
// 	"strconv"
// 	"time"

// 	sq "github.com/elgris/sqrl"
// 	"github.com/golang/protobuf/ptypes"
// 	"github.com/pkg/errors"
// 	"github.com/Xawery/auth-service/grpcerror"
// 	"github.com/Xawery/auth-service/model"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/peer"
// 	"google.golang.org/grpc/status"
// 	"gopkg.in/square/go-jose.v2"
// 	"gopkg.in/square/go-jose.v2/jwt"
// )

// type server struct {
// 	conf             config
// 	db               *sql.DB
// 	stmt             sq.StatementBuilderType
// 	creds            model.CredentialsRepo
// 	accts            model.AccountsRepo
// 	refresh          model.RefreshTokensRepo
// 	accessSig, idSig jose.Signer
// }

// type profile struct {
// 	ID       int64
// 	Scopes   []string
// 	Username string
// 	Account  *model.Account
// }

// // NewPg creates Postgres-backed instance of auth service
// func NewPg(db *sql.DB, opts ...Opt) (AuthnServer, error) {
// 	conf := config{
// 		accessKey:       DefaultKey,
// 		idKey:           DefaultKey,
// 		accessLifespan:  DefaultAccessLifespan,
// 		refreshLifespan: DefaultRefreshLifespan,
// 		cost:            DefaultCost,
// 	}

// 	for _, opt := range opts {
// 		opt(&conf)
// 	}

// 	so := (&jose.SignerOptions{}).WithType("JWT")

// 	accessSig, err := jose.NewSigner(conf.accessKey, so)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "Failed to create access token signer")
// 	}

// 	idSig, err := jose.NewSigner(conf.idKey, so)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "Failed to create id token signer")
// 	}

// 	return &server{
// 		conf:      conf,
// 		db:        db,
// 		stmt:      sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db),
// 		creds:     model.NewPgCredentialsRepo(),
// 		accts:     model.NewPgAccountsRepo(),
// 		refresh:   model.NewPgRefreshTokensRepo(),
// 		accessSig: accessSig,
// 		idSig:     idSig,
// 	}, nil
// }

// func (s *server) Login(ctx context.Context, req *LoginRequest) (*Auth, error) {
// 	peer, ok := peer.FromContext(ctx)
// 	if !ok {
// 		return nil, MissingPeer.Err()
// 	}

// 	rt := randomToken()
// 	p := profile{
// 		Username: req.Username,
// 	}

// 	err := s.tx(ctx, func(stmt sq.StatementBuilderType) (err error) {
// 		p.ID, p.Scopes, err = s.creds.Verify(ctx, stmt, req.Username, req.Password)
// 		if err != nil {
// 			return err
// 		}

// 		ts := time.Now()
// 		err = s.refresh.Create(ctx, stmt, &model.RefreshToken{
// 			Token:      hash(rt),
// 			AccountID:  p.ID,
// 			ClientID:   req.ClientId,
// 			Address:    peer.Addr.String(),
// 			ExpiresAt:  ts.Add(s.conf.refreshLifespan),
// 			LastAccess: ts,
// 		})
// 		if err != nil {
// 			return err
// 		}
// 		if len(req.PushToken) > 0 {
// 			err = s.accts.UpdatePushToken(ctx, stmt, req.PushToken, p.ID)
// 		}
// 		p.Account, err = s.accts.Retrieve(ctx, stmt, p.ID)
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return s.auth(rt, req.ClientId, &p)
// }

// func (s *server) Refresh(ctx context.Context, req *RefreshRequest) (*Auth, error) {
// 	peer, ok := peer.FromContext(ctx)
// 	if !ok {
// 		return nil, MissingPeer.Err()
// 	}

// 	p := profile{}

// 	var err error
// 	p.ID, p.Username, err = s.parseAccess(req.AccessToken)
// 	if err != nil {
// 		return nil, err
// 	}

// 	rt := randomToken()

// 	var clientID string
// 	err = s.tx(ctx, func(stmt sq.StatementBuilderType) error {
// 		ts := time.Now()
// 		clientID, err = s.refresh.Update(ctx, stmt, &model.RefreshTokenUpdate{
// 			OldToken:   hash(req.RefreshToken),
// 			NewToken:   hash(rt),
// 			AccountID:  p.ID,
// 			Address:    peer.Addr.String(),
// 			ExpiresAt:  ts.Add(s.conf.refreshLifespan),
// 			LastAccess: ts,
// 		})
// 		switch {
// 		case status.Code(err) == codes.NotFound:
// 			return Unauthenticated.Err()
// 		case err != nil:
// 			return err
// 		}

// 		_, p.Scopes, err = s.creds.Retrieve(ctx, stmt, p.Username)
// 		switch {
// 		case status.Code(err) == codes.NotFound:
// 			return Unauthenticated.Err()
// 		case err != nil:
// 			return err
// 		}

// 		p.Account, err = s.accts.Retrieve(ctx, stmt, p.ID)
// 		return err
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return s.auth(rt, clientID, &p)
// }

// func (s *server) tx(ctx context.Context, f func(sq.StatementBuilderType) error) error {
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	tx, err := s.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return grpcerror.New(Unavailable, errors.Wrap(err, "Failed to start transaction"))
// 	}

// 	stmt := s.stmt.RunWith(tx)

// 	if err := f(stmt); err != nil {
// 		return err
// 	}

// 	if err := tx.Commit(); err != nil {
// 		return grpcerror.New(Unavailable, errors.Wrap(err, "Failed to commit transaction"))
// 	}

// 	return nil
// }

// func (s *server) auth(rt string, clientID string, p *profile) (*Auth, error) {
// 	at, err := jwt.Signed(s.accessSig).
// 		Claims(s.claims(clientID, p.ID)).
// 		Claims(AccessToken{
// 			Scopes:   p.Scopes,
// 			Username: p.Username,
// 		}).
// 		CompactSerialize()
// 	if err != nil {
// 		return nil, grpcerror.New(Internal, errors.Wrap(err, "Failed to generate access token"))
// 	}

// 	acct := p.Account
// 	it, err := jwt.Signed(s.accessSig).
// 		Claims(s.claims(clientID, p.ID)).
// 		Claims(IDToken{
// 			Email:      acct.Email,
// 			Phone:      acct.Phone,
// 			FamilyName: acct.FamilyName,
// 			GivenName:  acct.GivenName,
// 			Locale:     acct.Locale,
// 		}).
// 		CompactSerialize()
// 	if err != nil {
// 		return nil, grpcerror.New(Internal, errors.Wrap(err, "Failed to generate id token"))
// 	}

// 	return &Auth{
// 		AccessToken:  at,
// 		IdToken:      it,
// 		RefreshToken: rt,
// 		ExpiresIn:    ptypes.DurationProto(s.conf.accessLifespan),
// 	}, nil
// }

// func (s *server) claims(clientID string, id int64) jwt.Claims {
// 	ts := time.Now()
// 	return jwt.Claims{
// 		ID:       randomToken(),
// 		Issuer:   s.conf.issuer,
// 		IssuedAt: jwt.NewNumericDate(ts),
// 		Expiry:   jwt.NewNumericDate(ts.Add(s.conf.accessLifespan)),
// 		Audience: jwt.Audience([]string{clientID}),
// 		Subject:  strconv.FormatInt(id, 10),
// 	}
// }

// func (s *server) parseAccess(raw string) (int64, string, error) {
// 	tok, err := jwt.ParseSigned(raw)
// 	if err != nil {
// 		return 0, "", grpcerror.New(InvalidToken, errors.Wrap(err, "Token validation failed"))
// 	}

// 	cl := struct {
// 		jwt.Claims
// 		ID       int64  `json:"sub,string"`
// 		Username string `json:"username"`
// 	}{}
// 	if err := tok.Claims(s.conf.accessKey.Key, &cl); err != nil {
// 		return 0, "", grpcerror.New(InvalidToken, errors.Wrap(err, "Failed to unmarshal claims"))
// 	}

// 	if cl.Issuer != s.conf.issuer {
// 		return 0, "", Unauthenticated.Err()
// 	}

// 	return cl.ID, cl.Username, nil
// }

// func randomToken() string {
// 	tok := make([]byte, 16)
// 	rand.Read(tok)

// 	return base64.RawURLEncoding.EncodeToString(tok)
// }

// func hash(src string) []byte {
// 	res := sha256.Sum256([]byte(src))
// 	return res[:]
// }
