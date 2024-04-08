package authcode

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/int128/oauth2cli/oauth2params"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
)

type KeyboardInput struct {
	RedirectURL string
}

type Keyboard struct {
	Logger logger.Logger
}

func (k *Keyboard) Login(ctx context.Context, in *KeyboardInput, oidcClient oidc.Interface) (*oidc.TokenSet, error) {
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

	authCodeURL, err := oidcClient.GetAuthCodeURL(ctx, oidc.AuthCodeURLInput{
		State:       state,
		Nonce:       nonce,
		PKCEParams:  pkce,
		RedirectURL: in.RedirectURL,
	})
	if err != nil {
		return nil, err
	}

	fmt.Printf("Please visit the following URL in your browser: %s\n", authCodeURL)
	if _, err := fmt.Fprint(os.Stderr, "Enter code: "); err != nil {
		return nil, fmt.Errorf("writing to stderr: %w", err)
	}
	r := bufio.NewReader(os.Stdin)
	code, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("reading authorization code: %w", err)
	}
	code = strings.TrimRight(code, "\r\n")

	k.Logger.Debug("Exchanging code and token")
	tokenSet, err := oidcClient.ExchangeAuthCode(ctx, oidc.ExchangeAuthCodeInput{
		Code:        code,
		PKCEParams:  pkce,
		Nonce:       nonce,
		RedirectURL: in.RedirectURL,
	})
	if err != nil {
		return nil, fmt.Errorf("exchanging authorization code: %w", err)
	}

	return tokenSet, nil
}
