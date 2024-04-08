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

type LoginInput struct {
	Provider       oidc.Provider
	TokenCache     tokencache.Interface
	Authentication Authentication
	AuthOptions    AuthOptions
}

type LogoutInput struct {
	Provider   oidc.Provider
	TokenCache tokencache.Interface
}

type Interface interface {
	Login(ctx context.Context, in LoginInput) error
	Logout(ctx context.Context, in LogoutInput) error
	SetLogger(logger.Logger)
}

type authenticator struct {
	authentication Authentication
	logger         logger.Logger
}

func NewAuthenticator(logger logger.Logger) Interface {
	return &authenticator{
		authentication: Authentication{
			Logger: logger,
			AuthCodeBrowser: &authcode.Browser{
				Logger: logger,
			},
			AuthCodeKeyboard: &authcode.Keyboard{
				Logger: logger,
			},
		},
		logger: logger,
	}
}

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

	authResult, err := a.authentication.Authenticate(ctx, authenticateInput)
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

func (a *authenticator) SetLogger(l logger.Logger) {
	a.logger = l
	a.authentication.Logger = l
	a.authentication.AuthCodeBrowser.Logger = l
	a.authentication.AuthCodeKeyboard.Logger = l
}
