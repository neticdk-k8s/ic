package cmd

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/oidc"
	testingJWT "github.com/neticdk-k8s/ic/internal/testing/jwt"
	"github.com/neticdk-k8s/ic/internal/tokencache"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_AuthCommands(t *testing.T) {
	logger := slog.Default()
	issuedIDTokenExpiration := time.Now().Add(1 * time.Hour).Round(time.Second)
	issuedIDToken := testingJWT.EncodeF(t, func(claims *testingJWT.Claims) {
		claims.Issuer = "https://issuer.example.com"
		claims.Subject = "YOUR_SUBJECT"
		claims.ExpiresAt = jwt.NewNumericDate(issuedIDTokenExpiration)
	})
	issuedTokenSet := oidc.TokenSet{
		AccessToken:  issuedIDToken,
		IDToken:      issuedIDToken,
		RefreshToken: "YOUR_REFRESH_TOKEN",
	}

	t.Run("login", func(t *testing.T) {
		got := new(bytes.Buffer)
		ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
		ec.Stderr = got
		ec.Stdout = got
		ui.SetDefaultOutput(got)
		ac := ic.NewContext()
		ac.EC = ec
		mockAuthentication := authentication.NewMockAuthentication(t)
		mockAuthentication.EXPECT().
			Authenticate(mock.Anything, mock.Anything).
			Return(&authentication.AuthResult{
				UsingCachedToken: true,
				TokenSet:         issuedTokenSet,
			}, nil)
		mockAuthentication.EXPECT().
			SetLogger(mock.Anything).
			Return()
		mockTokenCache := tokencache.NewMockCache(t)
		mockTokenCache.EXPECT().
			Lookup(mock.Anything).
			Return(&issuedTokenSet, nil)
		ac.TokenCache = mockTokenCache
		ac.Authenticator = authentication.NewAuthenticator(logger, mockAuthentication)

		cmd := newRootCmd(ac)

		cmd.SetArgs([]string{"login"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Logging in")
		assert.Contains(t, got.String(), "Logged in")
	})

	t.Run("logout", func(t *testing.T) {
		got := new(bytes.Buffer)
		ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
		ec.Stderr = got
		ec.Stdout = got
		ui.SetDefaultOutput(got)
		ac := ic.NewContext()
		ac.EC = ec
		mockAuthentication := authentication.NewMockAuthentication(t)
		mockAuthentication.EXPECT().
			Logout(mock.Anything, mock.Anything).
			Return(nil)
		mockAuthentication.EXPECT().
			SetLogger(mock.Anything).
			Return()
		mockTokenCache := tokencache.NewMockCache(t)
		mockTokenCache.EXPECT().
			Lookup(mock.Anything).
			Return(&issuedTokenSet, nil)
		mockTokenCache.EXPECT().
			Invalidate(mock.Anything).
			Return(nil)
		ac.TokenCache = mockTokenCache
		ac.Authenticator = authentication.NewAuthenticator(logger, mockAuthentication)

		cmd := newRootCmd(ac)

		cmd.SetArgs([]string{"logout"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Logging out")
		assert.Contains(t, got.String(), "Logged out")
	})
}
