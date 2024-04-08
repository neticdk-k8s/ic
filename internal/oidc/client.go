package oidc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/int128/oauth2cli"
	"github.com/int128/oauth2cli/oauth2params"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Interface interface {
	Refresh(ctx context.Context, refreshToken string) (*TokenSet, error)
	Logout(idToken string) error
	GetTokenByAuthCode(ctx context.Context, in GetTokenByAuthCodeInput, localServerReadyChan chan<- string) (*TokenSet, error)
	GetAuthCodeURL(ctx context.Context, in AuthCodeURLInput) (string, error)
	ExchangeAuthCode(ctx context.Context, in ExchangeAuthCodeInput) (*TokenSet, error)
}

type GetTokenByAuthCodeInput struct {
	BindAddress         string
	RedirectURLHostname string
	PKCEParams          *oauth2params.PKCE
	State               string
	Nonce               string
}

type AuthCodeURLInput struct {
	RedirectURL string
	PKCEParams  *oauth2params.PKCE
	State       string
	Nonce       string
}

type ExchangeAuthCodeInput struct {
	Code        string
	PKCEParams  *oauth2params.PKCE
	Nonce       string
	RedirectURL string
}

type client struct {
	provider          *gooidc.Provider
	oauth2config      oauth2.Config
	providerLogoutURL string
	logger            logger.Logger
}

func New(ctx context.Context, p Provider, logger logger.Logger) (*client, error) {
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
		logger,
	}, nil
}

func (c *client) Refresh(ctx context.Context, refreshToken string) (*TokenSet, error) {
	currentToken := &oauth2.Token{
		Expiry:       time.Now(),
		RefreshToken: refreshToken,
	}
	source := c.oauth2config.TokenSource(ctx, currentToken)
	token, err := source.Token()
	if err != nil {
		return nil, errors.Wrap(err, "refreshing token")
	}
	return c.verifyToken(token, "")
}

// Logout deletes the session from the OIDC provider
func (c *client) Logout(idToken string) error {
	if c.providerLogoutURL == "" {
		return errors.New("logout URL not set")
	}

	logoutURL, err := url.Parse(c.providerLogoutURL)
	if err != nil {
		return fmt.Errorf("parsing logout URL: %w", err)
	}

	query := logoutURL.Query()
	if idToken != "" {
		query.Set("id_token_hint", idToken)
	}
	logoutURL.RawQuery = query.Encode()

	res, err := c.logoutWithRetries(logoutURL.String())
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("logout failed with status code: %d", res.StatusCode)
	}

	return nil
}

func (c *client) logoutWithRetries(logoutURL string) (*http.Response, error) {
	client := retryablehttp.NewClient()
	client.HTTPClient.Timeout = time.Duration(2) * time.Second
	client.Logger = OIDCSlogAdapter{Logger: c.logger}
	client.RetryWaitMin = time.Duration(2) * time.Second
	client.RetryWaitMax = time.Duration(30) * time.Second
	client.RetryMax = 5

	defer client.HTTPClient.CloseIdleConnections()
	return client.Get(logoutURL)
}

func (c *client) GetTokenByAuthCode(ctx context.Context, in GetTokenByAuthCodeInput, localServerReadyChan chan<- string) (*TokenSet, error) {
	authCodeOptions := append(
		in.PKCEParams.AuthCodeOptions(),
		oauth2.AccessTypeOffline,
		gooidc.Nonce(in.Nonce))

	cfg := oauth2cli.Config{
		OAuth2Config:           c.oauth2config,
		State:                  in.State,
		AuthCodeOptions:        authCodeOptions,
		TokenRequestOptions:    in.PKCEParams.TokenRequestOptions(),
		LocalServerReadyChan:   localServerReadyChan,
		RedirectURLHostname:    in.RedirectURLHostname,
		LocalServerBindAddress: []string{in.BindAddress},
		Logf:                   c.logger.Debugf,
	}

	token, err := oauth2cli.GetToken(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("oauth2 error: %w", err)
	}

	return c.verifyToken(token, in.Nonce)
}

func (c *client) GetAuthCodeURL(ctx context.Context, in AuthCodeURLInput) (string, error) {
	cfg := c.oauth2config
	cfg.RedirectURL = in.RedirectURL

	requestOptions := append(
		in.PKCEParams.AuthCodeOptions(),
		oauth2.AccessTypeOffline,
		gooidc.Nonce(in.Nonce))

	return cfg.AuthCodeURL(in.State, requestOptions...), nil
}

func (c *client) ExchangeAuthCode(ctx context.Context, in ExchangeAuthCodeInput) (*TokenSet, error) {
	cfg := c.oauth2config
	cfg.RedirectURL = in.RedirectURL

	token, err := cfg.Exchange(ctx, in.Code, in.PKCEParams.TokenRequestOptions()...)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	return c.verifyToken(token, in.Nonce)
}

func (c *client) verifyToken(token *oauth2.Token, nonce string) (*TokenSet, error) {
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("getting id token")
	}

	verifier := c.provider.Verifier(&gooidc.Config{ClientID: c.oauth2config.ClientID})
	verifiedIDToken, err := verifier.Verify(context.Background(), idToken)
	if err != nil {
		return nil, fmt.Errorf("verifying id token: %w", err)
	}

	if err = verifiedIDToken.VerifyAccessToken(token.AccessToken); err != nil {
		return nil, fmt.Errorf("verifying access token: %w", err)
	}

	if nonce != "" && nonce != verifiedIDToken.Nonce {
		return nil, fmt.Errorf("verifying nonce (wants %s but got %s)", nonce, verifiedIDToken.Nonce)
	}

	return &TokenSet{
		IDToken:      idToken,
		RefreshToken: token.RefreshToken,
	}, nil
}
