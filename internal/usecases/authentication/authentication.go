package authentication

import (
	"context"
	"fmt"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication/authcode"
)

type AuthenticateInput struct {
	Provider       oidc.Provider
	CachedTokenSet *oidc.TokenSet
	AuthOptions    AuthOptions
}

type AuthOptions struct {
	AuthCodeBrowser  *authcode.BrowserInput
	AuthCodeKeyboard *authcode.KeyboardInput
}

type AuthResult struct {
	UsingCachedToken bool
	TokenSet         oidc.TokenSet
}

type Authentication struct {
	Logger           logger.Logger
	AuthCodeBrowser  *authcode.Browser
	AuthCodeKeyboard *authcode.Keyboard
}

func (a *Authentication) Authenticate(ctx context.Context, in AuthenticateInput) (*AuthResult, error) {
	if in.CachedTokenSet != nil {
		a.Logger.Debug("Found cached token")
		claims, err := in.CachedTokenSet.DecodeWithoutVerify()
		if err != nil {
			return nil, fmt.Errorf("decoding token: %w", err)
		}
		if !claims.IsExpired() {
			a.Logger.Info("Found cached token", "expires", claims.Expiry)
			return &AuthResult{
				UsingCachedToken: true,
				TokenSet:         *in.CachedTokenSet,
			}, nil
		} else {
			a.Logger.Info("Cached token is expired")
		}
	}

	oidcClient, err := oidc.New(
		ctx,
		in.Provider,
		a.Logger)
	if err != nil {
		return nil, fmt.Errorf("setting up authentication: %w", err)
	}

	if in.CachedTokenSet != nil && in.CachedTokenSet.RefreshToken != "" {
		a.Logger.Info("Refreshing token")
		tokenSet, err := oidcClient.Refresh(ctx, in.CachedTokenSet.RefreshToken)
		if err == nil {
			return &AuthResult{TokenSet: *tokenSet}, nil
		}
		a.Logger.Error("Refreshing token", "err", err)
	}

	if in.AuthOptions.AuthCodeBrowser != nil {
		a.Logger.Info("Authenticating using authcode-browser")
		tokenSet, err := a.AuthCodeBrowser.Login(ctx, in.AuthOptions.AuthCodeBrowser, oidcClient)
		if err != nil {
			return nil, fmt.Errorf("authcode-browser error: %w", err)
		}
		return &AuthResult{TokenSet: *tokenSet}, nil
	}

	if in.AuthOptions.AuthCodeKeyboard != nil {
		a.Logger.Info("Authenticating using authcode-keyboard")
		tokenSet, err := a.AuthCodeKeyboard.Login(ctx, in.AuthOptions.AuthCodeKeyboard, oidcClient)
		if err != nil {
			return nil, fmt.Errorf("authcode-keyboard error: %w", err)
		}
		return &AuthResult{TokenSet: *tokenSet}, nil
	}

	return nil, fmt.Errorf("unknown authentication method")
}
