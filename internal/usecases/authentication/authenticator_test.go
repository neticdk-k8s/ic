package authentication

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/neticdk-k8s/ic/internal/logger"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk-k8s/ic/internal/reader"
	testingJWT "github.com/neticdk-k8s/ic/internal/testing/jwt"
	"github.com/neticdk-k8s/ic/internal/tokencache"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestAuthenticator_NewAuthenticator(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		logger := logger.NewTestLogger(t)
		authn := NewAuthentication(logger, nil, &authcode.Browser{Logger: logger}, &authcode.Keyboard{Reader: reader.NewReader(), Logger: logger})
		want := &authenticator{
			authentication: authn,
			logger:         logger,
		}
		got := NewAuthenticator(logger, authn)
		assert.Equal(t, want, got)
	})
}

func TestAuthenticator_Login(t *testing.T) {
	logger := logger.NewTestLogger(t)
	testProvider := oidc.Provider{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "YOUR_CLIENT_ID",
	}
	issuedIDTokenExpiration := time.Now().Add(1 * time.Hour).Round(time.Second)
	issuedIDToken := testingJWT.EncodeF(t, func(claims *testingJWT.Claims) {
		claims.Issuer = "https://issuer.example.com"
		claims.Subject = "YOUR_SUBJECT"
		claims.ExpiresAt = jwt.NewNumericDate(issuedIDTokenExpiration)
	})
	issuedTokenSet := oidc.TokenSet{
		IDToken:      issuedIDToken,
		RefreshToken: "YOUR_REFRESH_TOKEN",
	}
	testAuthOptions := AuthOptions{
		AuthCodeBrowser: &authcode.BrowserLoginInput{
			BindAddress:         "127.0.0.1",
			RedirectURLHostname: "localhost",
		},
	}

	t.Run("NoTokenCache", func(t *testing.T) {
		tokenCacheKey := tokencache.Key{
			IssuerURL: "https://issuer.example.com",
			ClientID:  "YOUR_CLIENT_ID",
		}
		ctx := context.TODO()
		mockAuthentication := NewMockAuthentication((t))
		mockAuthentication.EXPECT().
			Authenticate(ctx, AuthenticateInput{
				Provider:    testProvider,
				AuthOptions: testAuthOptions,
			}).
			Return(&AuthResult{TokenSet: issuedTokenSet}, nil)
		mockTokenCache := tokencache.NewMockCache(t)
		mockTokenCache.EXPECT().
			Lookup(tokenCacheKey).
			Return(nil, &tokencache.CacheMissError{})
		mockTokenCache.EXPECT().
			Save(tokenCacheKey, issuedTokenSet).
			Return(nil)
		in := LoginInput{
			Provider:    testProvider,
			TokenCache:  mockTokenCache,
			AuthOptions: testAuthOptions,
		}
		a := NewAuthenticator(logger, mockAuthentication)
		_, err := a.Login(ctx, in)
		assert.NoError(t, err)
	})

	t.Run("HasValidIDToken", func(t *testing.T) {
		ctx := context.TODO()
		mockAuthentication := NewMockAuthentication(t)
		mockAuthentication.EXPECT().
			Authenticate(ctx, AuthenticateInput{
				Provider:       testProvider,
				CachedTokenSet: &issuedTokenSet,
				AuthOptions:    testAuthOptions,
			}).
			Return(&AuthResult{
				UsingCachedToken: true,
				TokenSet:         issuedTokenSet,
			}, nil)
		mockTokenCache := tokencache.NewMockCache(t)
		mockTokenCache.EXPECT().
			Lookup(tokencache.Key{
				IssuerURL: "https://issuer.example.com",
				ClientID:  "YOUR_CLIENT_ID",
			}).
			Return(&issuedTokenSet, nil)
		in := LoginInput{
			Provider:    testProvider,
			TokenCache:  mockTokenCache,
			AuthOptions: testAuthOptions,
		}
		a := NewAuthenticator(logger, mockAuthentication)
		_, err := a.Login(ctx, in)
		assert.NoError(t, err)
	})

	t.Run("AuthenticationError", func(t *testing.T) {
		ctx := context.TODO()
		mockAuthentication := NewMockAuthentication(t)
		mockAuthentication.EXPECT().
			Authenticate(ctx, AuthenticateInput{
				Provider:    testProvider,
				AuthOptions: testAuthOptions,
			}).
			Return(nil, errors.New("authentication error"))
		mockTokenCache := tokencache.NewMockCache(t)
		mockTokenCache.EXPECT().
			Lookup(tokencache.Key{
				IssuerURL: "https://issuer.example.com",
				ClientID:  "YOUR_CLIENT_ID",
			}).
			Return(nil, &tokencache.CacheMissError{})
		in := LoginInput{
			Provider:    testProvider,
			TokenCache:  mockTokenCache,
			AuthOptions: testAuthOptions,
		}
		a := NewAuthenticator(logger, mockAuthentication)
		_, err := a.Login(ctx, in)
		assert.Error(t, err)
	})
}

func TestAuthenticator_Logout(t *testing.T) {
	logger := logger.NewTestLogger(t)
	testProvider := oidc.Provider{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "YOUR_CLIENT_ID",
	}
	issuedIDTokenExpiration := time.Now().Add(1 * time.Hour).Round(time.Second)
	issuedIDToken := testingJWT.EncodeF(t, func(claims *testingJWT.Claims) {
		claims.Issuer = "https://issuer.example.com"
		claims.Subject = "YOUR_SUBJECT"
		claims.ExpiresAt = jwt.NewNumericDate(issuedIDTokenExpiration)
	})
	issuedTokenSet := oidc.TokenSet{
		IDToken:      issuedIDToken,
		RefreshToken: "YOUR_REFRESH_TOKEN",
	}

	t.Run("Success", func(t *testing.T) {
		ctx := context.TODO()
		mockAuthentication := NewMockAuthentication(t)
		mockAuthentication.EXPECT().
			Logout(ctx, AuthenticateLogoutInput{
				Provider:       testProvider,
				CachedTokenSet: &issuedTokenSet,
			}).
			Return(nil)

		mockTokenCache := tokencache.NewMockCache(t)
		mockTokenCache.EXPECT().
			Lookup(tokencache.Key{
				IssuerURL: "https://issuer.example.com",
				ClientID:  "YOUR_CLIENT_ID",
			}).
			Return(&issuedTokenSet, nil)
		mockTokenCache.EXPECT().
			Invalidate(tokencache.Key{
				IssuerURL: "https://issuer.example.com",
				ClientID:  "YOUR_CLIENT_ID",
			}).
			Return(nil)
		in := LogoutInput{
			Provider:   testProvider,
			TokenCache: mockTokenCache,
		}
		a := NewAuthenticator(logger, mockAuthentication)
		err := a.Logout(ctx, in)
		assert.NoError(t, err)
	})

	t.Run("CacheMissError", func(t *testing.T) {
		ctx := context.TODO()
		mockAuthentication := NewMockAuthentication(t)
		mockTokenCache := tokencache.NewMockCache(t)
		mockTokenCache.EXPECT().
			Lookup(tokencache.Key{
				IssuerURL: "https://issuer.example.com",
				ClientID:  "YOUR_CLIENT_ID",
			}).
			Return(nil, &tokencache.CacheMissError{})
		in := LogoutInput{
			Provider:   testProvider,
			TokenCache: mockTokenCache,
		}
		a := NewAuthenticator(logger, mockAuthentication)
		err := a.Logout(ctx, in)
		assert.Error(t, err)
		assert.ErrorIs(t, err, &tokencache.CacheMissError{})
	})
}

func TestAuthentication_NewAuthentication(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		logger := logger.NewTestLogger(t)

		want := &authentication{
			oidcClientFactory: &oidc.Factory{
				Logger: logger,
			},
			logger: logger,
			authCodeBrowser: &authcode.Browser{
				Logger: logger,
			},
			authCodeKeyboard: &authcode.Keyboard{
				Reader: reader.NewReader(),
				Logger: logger,
			},
		}

		got := NewAuthentication(logger, nil, &authcode.Browser{Logger: logger}, &authcode.Keyboard{Reader: reader.NewReader(), Logger: logger})
		assert.Equal(t, want, got)
	})
}

func TestAuthentication_Authenticate(t *testing.T) {
	logger := logger.NewTestLogger(t)

	timeout := 5 * time.Second
	expiryTime := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	goodExpiryTime := time.Now().Add(time.Hour)
	testProvider := oidc.Provider{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "YOUR_CLIENT_ID",
	}
	issuedIDToken := testingJWT.EncodeF(t, func(claims *testingJWT.Claims) {
		claims.Issuer = "https://issuer.example.com"
		claims.Subject = "YOUR_SUBJECT"
		claims.ExpiresAt = jwt.NewNumericDate(expiryTime)
	})

	goodIssuedIDToken := testingJWT.EncodeF(t, func(claims *testingJWT.Claims) {
		claims.Issuer = "https://issuer.example.com"
		claims.Subject = "YOUR_SUBJECT"
		claims.ExpiresAt = jwt.NewNumericDate(goodExpiryTime)
	})

	t.Run("HasValidIDToken", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)

		authentication := NewAuthentication(logger, nil, &authcode.Browser{Logger: logger}, &authcode.Keyboard{Reader: reader.NewReader(), Logger: logger})

		defer cancel()
		in := AuthenticateInput{
			Provider: testProvider,
			CachedTokenSet: &oidc.TokenSet{
				IDToken: goodIssuedIDToken,
			},
		}
		got, err := authentication.Authenticate(ctx, in)
		assert.NoError(t, err)
		want := &AuthResult{
			UsingCachedToken: true,
			TokenSet: oidc.TokenSet{
				IDToken: goodIssuedIDToken,
			},
		}
		assert.Equal(t, want, got)
	})

	t.Run("HasValidRefreshToken", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()
		in := AuthenticateInput{
			Provider: testProvider,
			CachedTokenSet: &oidc.TokenSet{
				IDToken:      issuedIDToken,
				RefreshToken: "VALID_REFRESH_TOKEN",
			},
		}
		mockClient := oidc.NewMockClient(t)
		mockClient.EXPECT().
			Refresh(ctx, "VALID_REFRESH_TOKEN").
			Return(&oidc.TokenSet{
				IDToken:      "NEW_ID_TOKEN",
				RefreshToken: "NEW_REFRESH_TOKEN",
			}, nil)
		mockClientFactory := oidc.NewMockFactoryClient(t)
		mockClientFactory.EXPECT().
			New(ctx, testProvider).
			Return(mockClient, nil)
		authentication := NewAuthentication(logger, mockClientFactory, &authcode.Browser{Logger: logger}, &authcode.Keyboard{Reader: reader.NewReader(), Logger: logger})
		got, err := authentication.Authenticate(ctx, in)
		if err != nil {
			t.Errorf("Do returned error: %+v", err)
		}
		want := &AuthResult{
			TokenSet: oidc.TokenSet{
				IDToken:      "NEW_ID_TOKEN",
				RefreshToken: "NEW_REFRESH_TOKEN",
			},
		}
		assert.Equal(t, want, got)
	})

	t.Run("HasExpiredRefreshToken/Browser", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()
		in := AuthenticateInput{
			Provider: testProvider,
			AuthOptions: AuthOptions{
				AuthCodeBrowser: &authcode.BrowserLoginInput{
					BindAddress:         "127.0.0.1",
					RedirectURLHostname: "localhost",
				},
			},
			CachedTokenSet: &oidc.TokenSet{
				IDToken:      issuedIDToken,
				RefreshToken: "EXPIRED_REFRESH_TOKEN",
			},
		}
		mockClient := oidc.NewMockClient(t)
		mockClient.EXPECT().
			Refresh(ctx, "EXPIRED_REFRESH_TOKEN").
			Return(nil, errors.New("token has expired"))
		mockClient.EXPECT().
			GetTokenByAuthCode(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, _ oidc.GetTokenByAuthCodeInput, readyChan chan<- string) {
				readyChan <- "LOCAL_SERVER_URL"
			}).
			Return(&oidc.TokenSet{
				IDToken:      "NEW_ID_TOKEN",
				RefreshToken: "NEW_REFRESH_TOKEN",
			}, nil)
		mockClientFactory := oidc.NewMockFactoryClient(t)
		mockClientFactory.EXPECT().
			New(ctx, testProvider).
			Return(mockClient, nil)
		authentication := NewAuthentication(logger, mockClientFactory, &authcode.Browser{Logger: logger}, &authcode.Keyboard{Reader: reader.NewReader(), Logger: logger})
		got, err := authentication.Authenticate(ctx, in)
		if err != nil {
			t.Errorf("Do returned error: %+v", err)
		}
		want := &AuthResult{
			TokenSet: oidc.TokenSet{
				IDToken:      "NEW_ID_TOKEN",
				RefreshToken: "NEW_REFRESH_TOKEN",
			},
		}
		assert.Equal(t, want, got)
	})
}
