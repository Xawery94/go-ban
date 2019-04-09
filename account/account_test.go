// +build integration

package account_test

// import (
// 	"context"
// 	"fmt"
// 	"math/rand"
// 	"testing"
// 	"time"

// 	"github.com/Xawery/auth-service/account"
// 	"github.com/Xawery/auth-service/authn"
// 	st "github.com/Xawery/auth-service/internal/servicetest"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	grpc "google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// func TestMain(m *testing.M) {
// 	st.TestMain(m)
// }

// func TestEmailAccount(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	if _, ok := st.Ready(ctx, t); !ok {
// 		return
// 	}

// 	event := st.EventSink(ctx, t, "auth-service")

// 	username := uuid.New().String() + "@smartace.online"
// 	password := uuid.New().String()

// 	rr := &account.RegisterEmailRequest{
// 		GivenName:  "Phillip",
// 		FamilyName: "Fry",
// 		Email:      username,
// 		Password:   password,
// 	}
// 	_, err := st.Accounts.RegisterEmail(ctx, rr)
// 	if !assert.NoError(t, err, "Failed to register account") {
// 		return
// 	}

// 	cr := account.Created{}
// 	if !assert.NoError(t, event(&cr, st.UsernameEquals(username))) || !assert.NotZero(t, cr.Id) {
// 		return
// 	}

// 	id := cr.Id

// 	ear := account.EmailActivationRequired{}
// 	if !assert.NoError(t, event(&ear, st.IDEquals(id))) {
// 		return
// 	}
// 	assert.Equal(t, username, ear.Email)
// 	assert.NotZero(t, ear.Code)

// 	_, err = st.Authn.Login(ctx, &authn.LoginRequest{
// 		Username: username,
// 		Password: password,
// 	})
// 	assert.Equal(t, codes.Unauthenticated.String(), status.Code(err).String(), "Login shouldn't succeed for unactivated account")

// 	_, err = st.Accounts.Activate(ctx, &account.ActivateRequest{
// 		Username: cr.Username,
// 		Code:     ear.Code,
// 	})
// 	if !assert.NoError(t, err, "Account activation failed") {
// 		return
// 	}

// 	ac := account.Activated{}
// 	if !assert.NoError(t, event(&ac, st.IDEquals(id))) {
// 		return
// 	}
// 	assert.Equal(t, username, ac.Username)

// 	// creds := grpc.PerRPCCredentials(authn.NewCredentials(st.Authn, username, password))

// 	acct, err := st.Accounts.Retrieve(ctx, &account.RetrieveRequest{Id: id}, creds)
// 	if assert.NoError(t, err, "Failed to retrieve account") {
// 		assert.Equal(t, username, acct.Email)
// 		assert.Equal(t, rr.FamilyName, acct.FamilyName)
// 		assert.Equal(t, rr.GivenName, acct.GivenName)
// 		assert.Equal(t, account.DefaultLocale, acct.Locale)
// 	}

// 	newUsername := uuid.New().String() + "@smartace.online"

// 	uer := &account.UpdateEmailRequest{
// 		Email:    newUsername,
// 		Password: rr.Password,
// 	}
// 	_, err = st.Accounts.UpdateEmail(ctx, uer, creds)
// 	if !assert.NoError(t, err, "Failed to update email") {
// 		return
// 	}

// 	eu := account.EmailUpdated{}
// 	if !assert.NoError(t, event(&eu, st.IDEquals(id))) {
// 		return
// 	}
// 	assert.Equal(t, username, eu.OldEmail)
// 	assert.Equal(t, newUsername, eu.NewEmail)

// 	ear = account.EmailActivationRequired{}
// 	if !assert.NoError(t, event(&ear, st.IDEquals(id))) || !assert.NotZero(t, ear.Code) {
// 		return
// 	}
// 	assert.Equal(t, newUsername, ear.Email)

// 	_, err = st.Accounts.Activate(ctx, &account.ActivateRequest{
// 		Username: newUsername,
// 		Code:     ear.Code,
// 	})
// 	if !assert.NoError(t, err, "Failed to activate updated email") {
// 		return
// 	}

// 	assert.NoError(t, event(&account.Activated{}, st.IDEquals(id)))

// 	// creds = grpc.PerRPCCredentials(authn.NewCredentials(st.Authn, newUsername, password))

// 	newPassword := uuid.New().String()

// 	upr := &account.UpdatePasswordRequest{
// 		OldPassword: password,
// 		NewPassword: newPassword,
// 	}
// 	_, err = st.Accounts.UpdatePassword(ctx, upr, creds)
// 	if !assert.NoError(t, err, "Failed to update password") {
// 		return
// 	}

// 	pu := account.PasswordUpdated{}
// 	if !assert.NoError(t, event(&pu, st.IDEquals(id))) {
// 		return
// 	}
// 	assert.Equal(t, newUsername, pu.Username)

// 	// creds = grpc.PerRPCCredentials(authn.NewCredentials(st.Authn, newUsername, newPassword))

// 	_, err = st.Accounts.Delete(ctx, &account.DeleteRequest{Password: password}, creds)
// 	assert.Equal(t, codes.Unauthenticated.String(), status.Code(err).String(), "Delete shouldn't succeed if password is invalid")

// 	_, err = st.Accounts.Delete(ctx, &account.DeleteRequest{Password: newPassword}, creds)
// 	if !assert.NoError(t, err) {
// 		return
// 	}

// 	dl := account.Deleted{}
// 	if !assert.NoError(t, event(&dl, st.IDEquals(id))) {
// 		return
// 	}
// 	assert.Equal(t, newUsername, dl.Username)

// 	_, err = st.Accounts.Retrieve(ctx, &account.RetrieveRequest{Id: id}, creds)
// 	assert.Equal(t, codes.NotFound.String(), status.Code(err).String(), "Retrieve shouldn't succeed for deleted account")
// }

// func TestPhoneAccount(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	if _, ok := st.Ready(ctx, t); !ok {
// 		return
// 	}

// 	event := st.EventSink(ctx, t, "auth-service")

// 	username := uuid.New().String()
// 	password := uuid.New().String()
// 	phone := fmt.Sprintf("+48660%d", rand.Intn(1000000))

// 	rm := &account.RegisterMobileRequest{
// 		GivenName:  "Phillip",
// 		FamilyName: "Fry",
// 		Phone:      phone,
// 		MobileId:   username,
// 		Password:   password,
// 		Company:    "C&C Tech",
// 	}
// 	_, err := st.Accounts.RegisterMobile(ctx, rm)
// 	if !assert.NoError(t, err, "Failed to register account") {
// 		return
// 	}

// 	cr := account.Created{}
// 	if !assert.NoError(t, event(&cr, st.UsernameEquals(username))) || !assert.NotZero(t, cr.Id) {
// 		return
// 	}

// 	id := cr.Id

// 	pac := account.PhoneActivationRequired{}
// 	if assert.NoError(t, event(&pac)) {
// 		assert.Equal(t, id, pac.Id)
// 		assert.Equal(t, phone, pac.Phone)
// 		assert.NotZero(t, pac.Code)
// 	}

// 	_, err = st.Accounts.RegisterMobile(ctx, rm)
// 	assert.Equal(t, codes.AlreadyExists.String(), status.Code(err).String(), "Duplicate registration should fail")

// 	ar := &account.ActivateRequest{
// 		Username: username,
// 		Code:     pac.Code,
// 	}
// 	_, err = st.Accounts.Activate(ctx, ar)
// 	if !assert.NoError(t, err, "Failed to activate account") {
// 		return
// 	}

// 	if !assert.NoError(t, event(&account.Activated{}, st.IDEquals(id))) {
// 		return
// 	}

// 	// creds := grpc.PerRPCCredentials(authn.NewCredentials(st.Authn, username, password))

// 	acct, err := st.Accounts.Retrieve(ctx, &account.RetrieveRequest{Id: id}, creds)
// 	if assert.NoError(t, err, "Failed to retrieve account") {
// 		assert.Equal(t, phone, acct.Phone)
// 		assert.Equal(t, rm.FamilyName, acct.FamilyName)
// 		assert.Equal(t, rm.GivenName, acct.GivenName)
// 		assert.Equal(t, account.DefaultLocale, acct.Locale)
// 	}

// 	newUsername := uuid.New().String()

// 	amr := &account.AddMobileRequest{
// 		Phone:    phone,
// 		MobileId: newUsername,
// 		Password: password,
// 	}
// 	_, err = st.Accounts.AddMobile(ctx, amr)
// 	if !assert.NoError(t, err, "Failed to add mobile") {
// 		return
// 	}

// 	pac = account.PhoneActivationRequired{}
// 	if !assert.NoError(t, event(&pac, st.IDEquals(id))) {
// 		return
// 	}
// 	assert.Equal(t, phone, pac.Phone)
// 	assert.NotZero(t, pac.Code)

// 	ar = &account.ActivateRequest{
// 		Username: newUsername,
// 		Code:     pac.Code,
// 	}
// 	_, err = st.Accounts.Activate(ctx, ar)
// 	if !assert.NoError(t, err, "Failed to activate mobile") {
// 		return
// 	}

// 	if !assert.NoError(t, event(&account.Activated{}, st.IDEquals(id))) {
// 		return
// 	}

// 	creds = grpc.PerRPCCredentials(authn.NewCredentials(st.Authn, newUsername, password))

// 	_, err = st.Accounts.Retrieve(ctx, &account.RetrieveRequest{Id: id}, creds)
// 	assert.NoError(t, err, "Failed to retrieve account")

// 	newPassword := uuid.New().String()

// 	upr := &account.UpdatePasswordRequest{
// 		OldPassword: password,
// 		NewPassword: newPassword,
// 	}
// 	_, err = st.Accounts.UpdatePassword(ctx, upr, creds)
// 	if !assert.NoError(t, err, "Failed to update password") {
// 		return
// 	}

// 	pu := account.PasswordUpdated{}
// 	if !assert.NoError(t, event(&pu, st.IDEquals(id))) {
// 		return
// 	}
// 	assert.Equal(t, newUsername, pu.Username)

// 	creds = grpc.PerRPCCredentials(authn.NewCredentials(st.Authn, newUsername, newPassword))
// 	_, err = st.Accounts.Retrieve(ctx, &account.RetrieveRequest{Id: id}, creds)
// 	assert.NoError(t, err, "Failed to retrieve account")
// }
