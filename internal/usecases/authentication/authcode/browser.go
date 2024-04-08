package authcode

import (
	"context"
	"fmt"
	"time"

	"github.com/int128/oauth2cli/oauth2params"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type BrowserInput struct {
	BindAddress         string
	RedirectURLHostname string
}

type Browser struct {
	Logger logger.Logger
}

func (b *Browser) Login(ctx context.Context, in *BrowserInput, oidcClient oidc.Interface) (*oidc.TokenSet, error) {
	state, err := oauth2params.NewState()
	if err != nil {
		return nil, fmt.Errorf("could not generate a state: %w", err)
	}

	nonce, err := oauth2params.NewState()
	if err != nil {
		return nil, fmt.Errorf("could not generate a nonce: %w", err)
	}

	pkce, err := oauth2params.NewPKCE()
	if err != nil {
		return nil, err
	}

	authCodeInput := oidc.GetTokenByAuthCodeInput{
		BindAddress:         in.BindAddress,
		RedirectURLHostname: in.RedirectURLHostname,
		PKCEParams:          pkce,
		State:               state,
		Nonce:               nonce,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	ready := make(chan string, 1)
	var out *oidc.TokenSet
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		select {
		case url := <-ready:
			b.Logger.Debug("Open url", "url", url)
			if err := browser.OpenURL(url); err != nil {
				b.Logger.Error("could not open the browser", "err", err)
			}
			return nil
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for authorization: %w", ctx.Err())
		}
	})
	eg.Go(func() error {
		defer close(ready)
		tokenSet, err := oidcClient.GetTokenByAuthCode(ctx, authCodeInput, ready)
		if err != nil {
			return errors.Wrap(err, "getting token")
		}
		out = tokenSet
		b.Logger.Debug("Got a valid token set")
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}

	return out, nil
}
