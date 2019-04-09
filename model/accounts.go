package model

// type AccountsRepo interface {
// 	Create(ctx context.Context, stmt sq.StatementBuilderType, acct *Account) (int64, error)
// 	Retrieve(ctx context.Context, stmt sq.StatementBuilderType, id int64) (*Account, error)
// 	RetrieveByPhone(ctx context.Context, stmt sq.StatementBuilderType, phone string) (*Account, error)
// 	RetrieveByUsername(ctx context.Context, stmt sq.StatementBuilderType, username string) (*Account, error)
// 	Search(ctx context.Context, stmt sq.StatementBuilderType, search string, consumer func(*Account) error) error
// 	Update(ctx context.Context, stmt sq.StatementBuilderType, id int64, data *AccountUpdate) error
// 	Delete(ctx context.Context, stmt sq.StatementBuilderType, id int64) error
// 	CreateDefaultScope(ctx context.Context, stmt sq.StatementBuilderType, accountID int64, operatorID int64) error
// 	UpdatePushToken(ctx context.Context, stmt sq.StatementBuilderType, token string, accountID int64) error
// }

type Account struct {
	ID         int64
	Email      string
	Phone      string
	FamilyName string
	GivenName  string
	Locale     string
	Company    string
}

type AccountUpdate struct {
	Email      string
	Phone      string
	FamilyName string
	GivenName  string
	Locale     string
	Company    string
}

var accountFields = []string{
	"id",
	"COALESCE(email, '') AS email",
	"COALESCE(phone, '') AS phone",
	"family_name",
	"given_name",
	"locale",
	"COALESCE(company, '') AS company",
}

// func NewPgAccountsRepo() AccountsRepo {
// 	return &pgAccts{}
// }

type pgAccts struct{}
