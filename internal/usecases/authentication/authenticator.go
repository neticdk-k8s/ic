package authentication

import (
	"context"
	"fmt"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/tokencache"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication/authcode"
	"github.com/pkg/errors"
)

// LoginInput is the input given to Login
type LoginInput struct {
	// Provider represents an OIDC provider configuration
	Provider oidc.Provider
	// TokenCache is the interface used for caching tokens
	TokenCache tokencache.Cache
	// AuthOptions are the options used for authentication
	AuthOptions AuthOptions
}

// AuthenticateInput is the input given to Authenticate
type AuthenticateInput struct {
	// Provider represents an OIDC provider configuration
	Provider oidc.Provider
	// CachedTokenSet is a TokenSet with cached credentials
	CachedTokenSet *oidc.TokenSet
	// AuthOptions are the options used for authentication
	AuthOptions AuthOptions
}

// AuthOptions is authentication options used by Authenticate
type AuthOptions struct {
	AuthCodeBrowser  *authcode.BrowserLoginInput
	AuthCodeKeyboard *authcode.KeyboardLoginInput
}

// AuthResult is the result of an authentication
type AuthResult struct {
	// UsingCachedToken is true if authentication is using a cached token
	UsingCachedToken bool
	// TokenSet is the TokenSet used for authentication
	TokenSet oidc.TokenSet
}

// LogoutInput is the input given to Logout
type LogoutInput struct {
	Provider   oidc.Provider
	TokenCache tokencache.Cache
}

// Authenticator represents an Authenticator
type Authenticator interface {
	Login(ctx context.Context, in LoginInput) error
	Logout(ctx context.Context, in LogoutInput) error
	Authenticate(ctx context.Context, in AuthenticateInput) (*AuthResult, error)
	SetLogger(logger.Logger)
}

type authenticator struct {
	logger logger.Logger
	// AuthCodeBrowser is the configuration used when authenticating using
	// authcode-browser
	authCodeBrowser *authcode.Browser
	// AuthCodeKeyboard is the configuration used when authenticating using
	// authcode-keyboard
	authCodeKeyboard *authcode.Keyboard
}

// NewAuthenticator creates a new Authenticator
func NewAuthenticator(logger logger.Logger) Authenticator {
	return &authenticator{
		authCodeBrowser: &authcode.Browser{
			Logger: logger,
		},
		authCodeKeyboard: &authcode.Keyboard{
			Logger: logger,
		},
		logger: logger,
	}
}

// Login performs OIDC login in three steps:
//
//  1. fetching a cached token
//  2. authenticating using the cached token or if not present performing OIDC
//     authentication using the grant type provided
//  3. caching the token obtained from the auth flow
func (a *authenticator) Login(ctx context.Context, in LoginInput) error {
	a.logger.Info("Fetching cached token")

	tokenCacheKey := tokencache.Key{
		IssuerURL:   in.Provider.IssuerURL,
		ClientID:    in.Provider.ClientID,
		ExtraScopes: in.Provider.ExtraScopes,
	}

	cachedTokenSet, err := in.TokenCache.Lookup(tokenCacheKey)
	if err != nil {
		if errors.Is(err, &tokencache.CacheMissError{}) {
			a.logger.Info("Cached token not found")
		} else {
			a.logger.Error("Fetching cached token", "err", err)
		}
	}

	authenticateInput := AuthenticateInput{
		Provider:       in.Provider,
		CachedTokenSet: cachedTokenSet,
		AuthOptions:    in.AuthOptions,
	}

	authResult, err := a.Authenticate(ctx, authenticateInput)
	if err != nil {
		return fmt.Errorf("authenticating: %w", err)
	}

	err = in.TokenCache.Save(tokenCacheKey, authResult.TokenSet)
	if err != nil {
		return fmt.Errorf("caching token: %w", err)
	}

	a.logger.Info("Login succeeded ✅")

	return nil
}

// Login performs OIDC logout by:
//
//  1. fetching a cached token
//  2. using the logout url to log out of the OIDC provider
//  3. removing the cached token
func (a *authenticator) Logout(ctx context.Context, in LogoutInput) error {
	a.logger.Info("Fetching cached token")

	tokenCacheKey := tokencache.Key{
		IssuerURL:   in.Provider.IssuerURL,
		ClientID:    in.Provider.ClientID,
		ExtraScopes: in.Provider.ExtraScopes,
	}

	cachedTokenSet, err := in.TokenCache.Lookup(tokenCacheKey)
	if err != nil {
		if errors.Is(err, &tokencache.CacheMissError{}) {
			a.logger.Warn("Cached token not found - cannot log out")
			return nil
		}
		return fmt.Errorf("looking up cached token: %w", err)
	}

	oidcClient, err := oidc.New(
		ctx,
		in.Provider,
		a.logger)
	if err != nil {
		return fmt.Errorf("setting up authentication: %w", err)
	}

	a.logger.Info("Found cached token")
	a.logger.Info("Logging out from OIDC provider")
	err = oidcClient.Logout(cachedTokenSet.IDToken)
	if err != nil {
		return fmt.Errorf("logging out of keycloak: %w", err)
	}

	a.logger.Info("Invalidating cached token")
	if err := in.TokenCache.Invalidate(tokenCacheKey); err != nil {
		return fmt.Errorf("invalidating cached token: %w", err)
	}

	a.logger.Info("Logout succeeded ✅")

	return nil
}

// Authenticate performs the OIDC authentication using the configuration given
// by AuthenticateInput
func (a *authenticator) Authenticate(ctx context.Context, in AuthenticateInput) (*AuthResult, error) {
	if in.CachedTokenSet != nil {
		a.logger.Debug("Found cached token")
		claims, err := in.CachedTokenSet.DecodeWithoutVerify()
		if err != nil {
			return nil, fmt.Errorf("decoding token: %w", err)
		}
		if !claims.IsExpired() {
			a.logger.Info("Found cached token", "expires", claims.Expiry)
			return &AuthResult{
				UsingCachedToken: true,
				TokenSet:         *in.CachedTokenSet,
			}, nil
		} else {
			a.logger.Info("Cached token is expired")
		}
	}

	oidcClient, err := oidc.New(
		ctx,
		in.Provider,
		a.logger)
	if err != nil {
		return nil, fmt.Errorf("setting up authentication: %w", err)
	}

	if in.CachedTokenSet != nil && in.CachedTokenSet.RefreshToken != "" {
		a.logger.Info("Refreshing token")
		tokenSet, err := oidcClient.Refresh(ctx, in.CachedTokenSet.RefreshToken)
		if err == nil {
			return &AuthResult{TokenSet: *tokenSet}, nil
		}
		a.logger.Error("Refreshing token", "err", err)
	}

	if in.AuthOptions.AuthCodeBrowser != nil {
		a.logger.Info("Authenticating using authcode-browser")
		tokenSet, err := a.authCodeBrowser.Login(ctx, in.AuthOptions.AuthCodeBrowser, oidcClient)
		if err != nil {
			return nil, fmt.Errorf("authcode-browser error: %w", err)
		}
		return &AuthResult{TokenSet: *tokenSet}, nil
	}

	if in.AuthOptions.AuthCodeKeyboard != nil {
		a.logger.Info("Authenticating using authcode-keyboard")
		tokenSet, err := a.authCodeKeyboard.Login(ctx, in.AuthOptions.AuthCodeKeyboard, oidcClient)
		if err != nil {
			return nil, fmt.Errorf("authcode-keyboard error: %w", err)
		}
		return &AuthResult{TokenSet: *tokenSet}, nil
	}

	return nil, fmt.Errorf("unknown authentication method")
}

// SetLogger sets the logger used for authentication
func (a *authenticator) SetLogger(l logger.Logger) {
	a.logger = l
}
