package authentication

import (
	"context"
	"fmt"

	"github.com/neticdk-k8s/ic/internal/logger"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk-k8s/ic/internal/tokencache"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
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

// AuthenticateLogoutInput is the input given to Logout
type AuthenticateLogoutInput struct {
	// Provider represents an OIDC provider configuration
	Provider oidc.Provider
	// CachedTokenSet is a TokenSet with cached credentials
	CachedTokenSet *oidc.TokenSet
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
	// Provider represents an OIDC provider configuration
	Provider oidc.Provider
	// TokenCache is the interface used for caching tokens
	TokenCache tokencache.Cache
}

// Authenticator represents an Authenticator
type Authenticator interface {
	// Login performs OIDC login in three steps:
	//
	//  1. fetching a cached token
	//  2. authenticating using the cached token or if not present performing OIDC authentication using the grant type provided
	//  3. caching the token obtained from the auth flow
	Login(ctx context.Context, in LoginInput) (*oidc.TokenSet, error)

	// Logout performs OIDC logout by:
	//
	//  1. fetching a cached token
	//  2. using the logout url to log out of the OIDC provider
	//  3. removing the cached token
	Logout(ctx context.Context, in LogoutInput) error

	// SetLogger sets the logger used for authentication
	SetLogger(logger.Logger)
}

type authenticator struct {
	logger         logger.Logger
	authentication Authentication
}

// NewAuthenticator creates a new Authenticator
func NewAuthenticator(logger logger.Logger, authentication Authentication) *authenticator {
	return &authenticator{
		logger:         logger,
		authentication: authentication,
	}
}

// Login performs OIDC login in three steps:
//
//  1. fetching a cached token
//  2. authenticating using the cached token or if not present performing OIDC authentication using the grant type provided
//  3. caching the token obtained from the auth flow
func (a *authenticator) Login(ctx context.Context, in LoginInput) (*oidc.TokenSet, error) {
	a.logger.Debug("Fetching cached token")

	tokenCacheKey := tokencache.Key{
		IssuerURL:   in.Provider.IssuerURL,
		ClientID:    in.Provider.ClientID,
		ExtraScopes: in.Provider.ExtraScopes,
	}

	cachedTokenSet, err := in.TokenCache.Lookup(tokenCacheKey)
	if err != nil {
		if errors.Is(err, &tokencache.CacheMissError{}) {
			a.logger.Debug("Cached token not found")
		} else {
			a.logger.Error("Fetching cached token", "err", err)
		}
	}

	authenticateInput := AuthenticateInput{
		Provider:       in.Provider,
		CachedTokenSet: cachedTokenSet,
		AuthOptions:    in.AuthOptions,
	}

	authResult, err := a.authentication.Authenticate(ctx, authenticateInput)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	tokenClaims, err := authResult.TokenSet.DecodeWithoutVerify()
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	a.logger.Debug("Token", "token", tokenClaims.Pretty)

	if authResult.UsingCachedToken {
		a.logger.Debug("Using cached token", "expires", tokenClaims.Expiry)
	} else {
		a.logger.Debug("Using new token", "expires", tokenClaims.Expiry)
		err = in.TokenCache.Save(tokenCacheKey, authResult.TokenSet)
		if err != nil {
			return nil, fmt.Errorf("caching token: %w", err)
		}
	}

	return &authResult.TokenSet, nil
}

// Logout performs OIDC logout by:
//
//  1. fetching a cached token
//  2. using the logout url to log out of the OIDC provider
//  3. removing the cached token
func (a *authenticator) Logout(ctx context.Context, in LogoutInput) error {
	a.logger.Debug("Fetching cached token")

	tokenCacheKey := tokencache.Key{
		IssuerURL:   in.Provider.IssuerURL,
		ClientID:    in.Provider.ClientID,
		ExtraScopes: in.Provider.ExtraScopes,
	}

	cachedTokenSet, err := in.TokenCache.Lookup(tokenCacheKey)
	if err != nil {
		if errors.Is(err, &tokencache.CacheMissError{}) {
			a.logger.Warn("Cached token not found - cannot log out")
			return err
		}
		return fmt.Errorf("looking up cached token: %w", err)
	}

	logoutInput := AuthenticateLogoutInput{
		Provider:       in.Provider,
		CachedTokenSet: cachedTokenSet,
	}

	err = a.authentication.Logout(ctx, logoutInput)
	if err != nil {
		return fmt.Errorf("logging out: %w", err)
	}

	a.logger.Debug("Invalidating cached token")
	if err := in.TokenCache.Invalidate(tokenCacheKey); err != nil {
		return fmt.Errorf("invalidating cached token: %w", err)
	}

	return nil
}

// SetLogger sets the logger used for authentication
func (a *authenticator) SetLogger(l logger.Logger) {
	a.logger = l
	a.authentication.SetLogger(l)
}

type Authentication interface {
	// Authenticate performs the OIDC authentication using the configuration given by AuthenticateInput
	Authenticate(ctx context.Context, in AuthenticateInput) (*AuthResult, error)

	// Logout logs out of the OIDC provider
	Logout(ctx context.Context, in AuthenticateLogoutInput) error

	// SetLogger sets the logger used for authentication
	SetLogger(logger.Logger)
}

type authentication struct {
	oidcClientFactory oidc.FactoryClient
	logger            logger.Logger
	// AuthCodeBrowser is the configuration used when authenticating using authcode-browser
	authCodeBrowser *authcode.Browser
	// AuthCodeKeyboard is the configuration used when authenticating using authcode-keyboard
	authCodeKeyboard *authcode.Keyboard
}

// NewAuthentication creates a new authentication
func NewAuthentication(logger logger.Logger, clientFactory oidc.FactoryClient, authCodeBrowser *authcode.Browser, authCodeKeyboard *authcode.Keyboard) *authentication {
	authn := &authentication{
		oidcClientFactory: &oidc.Factory{
			Logger: logger,
		},
		logger:           logger,
		authCodeBrowser:  authCodeBrowser,
		authCodeKeyboard: authCodeKeyboard,
	}
	if clientFactory != nil {
		authn.oidcClientFactory = clientFactory
	}
	return authn
}

// Authenticate performs the OIDC authentication using the configuration given by AuthenticateInput
func (a *authentication) Authenticate(ctx context.Context, in AuthenticateInput) (*AuthResult, error) {
	if in.CachedTokenSet != nil {
		a.logger.Debug("Found cached token")
		claims, err := in.CachedTokenSet.DecodeWithoutVerify()
		if err != nil {
			return nil, fmt.Errorf("decoding token: %w", err)
		}
		if !claims.IsExpired() {
			a.logger.Debug("Found cached token", "expires", claims.Expiry)
			return &AuthResult{
				UsingCachedToken: true,
				TokenSet:         *in.CachedTokenSet,
			}, nil
		} else {
			a.logger.Debug("Cached token is expired")
		}
	}

	oidcClient, err := a.oidcClientFactory.New(ctx, in.Provider)
	if err != nil {
		return nil, fmt.Errorf("creating OIDC client: %w", err)
	}

	if in.CachedTokenSet != nil && in.CachedTokenSet.RefreshToken != "" {
		a.logger.Debug("Refreshing token")
		tokenSet, err := oidcClient.Refresh(ctx, in.CachedTokenSet.RefreshToken)
		if err == nil {
			return &AuthResult{TokenSet: *tokenSet}, nil
		}
		a.logger.Error("Refreshing token", "err", err)
	}

	if in.AuthOptions.AuthCodeBrowser != nil {
		a.logger.Debug("Authenticating using authcode-browser")
		tokenSet, err := a.authCodeBrowser.Login(ctx, in.AuthOptions.AuthCodeBrowser, oidcClient)
		if err != nil {
			return nil, fmt.Errorf("authcode-browser error: %w", err)
		}
		return &AuthResult{TokenSet: *tokenSet}, nil
	}

	if in.AuthOptions.AuthCodeKeyboard != nil {
		a.logger.Debug("Authenticating using authcode-keyboard")
		tokenSet, err := a.authCodeKeyboard.Login(ctx, in.AuthOptions.AuthCodeKeyboard, oidcClient)
		if err != nil {
			return nil, fmt.Errorf("authcode-keyboard error: %w", err)
		}
		return &AuthResult{TokenSet: *tokenSet}, nil
	}

	return nil, fmt.Errorf("unknown authentication method")
}

// Logout logs out of the OIDC provider
func (a *authentication) Logout(ctx context.Context, in AuthenticateLogoutInput) error {
	oidcClient, err := a.oidcClientFactory.New(ctx, in.Provider)
	if err != nil {
		return fmt.Errorf("creating OIDC client: %w", err)
	}

	a.logger.Debug("Found cached token")
	a.logger.Debug("Logging out from OIDC provider")
	err = oidcClient.Logout(in.CachedTokenSet.IDToken)
	if err != nil {
		return fmt.Errorf("logging out of keycloak: %w", err)
	}
	return nil
}

// SetLogger sets the logger used for authentication
func (a *authentication) SetLogger(l logger.Logger) {
	a.logger = l
	a.oidcClientFactory.SetLogger(l)
	a.authCodeBrowser.Logger = l
	a.authCodeKeyboard.Logger = l
}
