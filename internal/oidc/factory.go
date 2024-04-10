package oidc

import (
	"context"
	"fmt"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"golang.org/x/oauth2"
)

type FactoryClient interface {
	New(ctx context.Context, p Provider) (Client, error)
	// SetLogger sets the logger used for authentication
	SetLogger(logger.Logger)
}

type Factory struct {
	Logger logger.Logger
}

func (f *Factory) New(ctx context.Context, p Provider) (Client, error) {
	provider, err := gooidc.NewProvider(ctx, p.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("setting up provider: %w", err)
	}

	var providerLogoutURL string

	claims := make(map[string]any)
	if err := provider.Claims(&claims); err == nil {
		endSessionEndPoint, ok := claims["end_session_endpoint"]
		if ok {
			if val, ok := endSessionEndPoint.(string); ok {
				providerLogoutURL = val
			}
		}
	}

	config := oauth2.Config{
		ClientID: p.ClientID,
		Endpoint: provider.Endpoint(),
		Scopes:   append(p.ExtraScopes, gooidc.ScopeOpenID),
	}

	return &client{
		provider,
		config,
		providerLogoutURL,
		f.Logger,
	}, nil
}

// SetLogger sets the logger used for the factory
func (f *Factory) SetLogger(l logger.Logger) {
	f.Logger = l
}
