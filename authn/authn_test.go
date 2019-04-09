// +build integration

package authn_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/Xawery/auth-service/account"
	"github.com/Xawery/auth-service/authn"
	st "github.com/Xawery/auth-service/internal/servicetest"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	st.TestMain(m)
}

func TestAuthn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, ok := st.Ready(ctx, t); !ok {
		return
	}

	event := st.EventSink(ctx, t, "auth-service")

	username := uuid.New().String() + "@smartace.online"
	password := uuid.New().String()

	rr := &account.RegisterEmailRequest{
		GivenName:  "Phillip",
		FamilyName: "Fry",
		Email:      username,
		Password:   password,
	}
	_, err := st.Accounts.RegisterEmail(ctx, rr)
	if !assert.NoError(t, err, "Failed to register account") {
		return
	}

	cr := account.Created{}
	if !assert.NoError(t, event(&cr, st.UsernameEquals(username))) {
		return
	}

	id := cr.Id

	ar := account.EmailActivationRequired{}
	if !assert.NoError(t, event(&ar, st.IDEquals(id))) {
		return
	}

	_, err = st.Accounts.Activate(ctx, &account.ActivateRequest{
		Username: ar.Email,
		Code:     ar.Code,
	})
	if !assert.NoError(t, err, "Failed to activate account") {
		return
	}

	if !assert.NoError(t, event(&account.Activated{}, st.IDEquals(id))) {
		return
	}

	res, err := st.Authn.Login(ctx, &authn.LoginRequest{
		Username: ar.Email,
		Password: password,
	})
	if !assert.NoError(t, err, "Failed to login") {
		return
	}
	at, it, ok := checkAuth(t, res)
	if !ok {
		return
	}
	assert.Equal(t, username, at.Username)
	assert.Equal(t, rr.Email, it.Email)
	assert.Equal(t, rr.GivenName, it.GivenName)
	assert.Equal(t, rr.FamilyName, it.FamilyName)

	res, err = st.Authn.Refresh(ctx, &authn.RefreshRequest{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
	if assert.NoError(t, err, "Failed to refresh") {
		checkAuth(t, res)
	}
}

func checkAuth(t *testing.T, auth *authn.Auth) (*authn.AccessToken, *authn.IDToken, bool) {
	at := authn.AccessToken{}
	if !assert.NoError(t, st.ParseToken(auth.AccessToken, &at), "Invalid access token") {
		return nil, nil, false
	}
	assert.NotEmpty(t, at.Username)
	assert.NotNil(t, at.Scopes)

	it := authn.IDToken{}
	if !assert.NoError(t, st.ParseToken(auth.IdToken, &it), "Invalid id token") {
		return nil, nil, false
	}

	if !assert.NotEmpty(t, auth.RefreshToken, "Refresh token should be set") {
		return nil, nil, false
	}

	return &at, &it, true
}
