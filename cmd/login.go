package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
	"github.com/spf13/cobra"
)

// New creates a new login command
func NewLoginCmd(ec *ExecutionContext) *cobra.Command {
	o := loginOptions{}
	c := &cobra.Command{
		Use:     "login",
		Short:   "Login to Inventory Server",
		GroupID: groupAuth,
		RunE: func(_ *cobra.Command, _ []string) error {
			return o.run(ec)
		},
	}
	return c
}

type loginOptions struct{}

func (o *loginOptions) run(ec *ExecutionContext) error {
	logger := ec.Logger.WithPrefix("Login")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ec.Logger.Info("Login succeeded ✅")

	return nil
}

func doLogin(ec *ExecutionContext) (*oidc.TokenSet, error) {
	loginInput := authentication.LoginInput{
		Provider:    *ec.OIDCProvider,
		TokenCache:  ec.TokenCache,
		AuthOptions: authentication.AuthOptions{},
	}
	if ec.OIDC.GrantType == "authcode-browser" {
		ec.Spin("Logging in")
		loginInput.AuthOptions.AuthCodeBrowser = &authcode.BrowserLoginInput{
			BindAddress:         ec.OIDC.AuthBindAddr,
			RedirectURLHostname: ec.OIDC.RedirectURLHostname,
		}
	} else if ec.OIDC.GrantType == "authcode-keyboard" {
		loginInput.AuthOptions.AuthCodeKeyboard = &authcode.KeyboardLoginInput{
			RedirectURI: ec.OIDC.RedirectURIAuthCodeKeyboard,
		}
	}

	tokenSet, err := ec.Authenticator.Login(ec.Command.Context(), loginInput)
	if err != nil {
		return nil, fmt.Errorf("logging in: %w", err)
	}

	if err := ec.SetupDefaultAPIClient(tokenSet.AccessToken); err != nil {
		return nil, fmt.Errorf("setting up API client: %w", err)
	}

	return tokenSet, nil
}
