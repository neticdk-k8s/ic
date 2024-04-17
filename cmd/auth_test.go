package cmd

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/neticdk-k8s/ic/internal/logger"
	"github.com/neticdk-k8s/ic/internal/oidc"
	testingJWT "github.com/neticdk-k8s/ic/internal/testing/jwt"
	"github.com/neticdk-k8s/ic/internal/tokencache"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_AuthCommands(t *testing.T) {
	logger := logger.NewTestLogger(t)
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
		in := ExecutionContextInput{
			Stdout: got,
			Stderr: got,
		}
		ec := NewExecutionContext(in)
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
		ec.TokenCache = mockTokenCache
		ec.Authenticator = authentication.NewAuthenticator(logger, mockAuthentication)

		cmd := NewRootCmd(ec)

		cmd.SetArgs([]string{"login"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Logging in")
		assert.Contains(t, got.String(), "Login succeeded")
	})

	t.Run("logout", func(t *testing.T) {
		got := new(bytes.Buffer)
		in := ExecutionContextInput{
			Stdout: got,
			Stderr: got,
		}
		ec := NewExecutionContext(in)
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
		ec.TokenCache = mockTokenCache
		ec.Authenticator = authentication.NewAuthenticator(logger, mockAuthentication)

		cmd := NewRootCmd(ec)

		cmd.SetArgs([]string{"logout"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Logging out")
		assert.Contains(t, got.String(), "Logout succeeded")
	})
}
