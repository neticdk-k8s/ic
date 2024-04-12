package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
	"github.com/spf13/cobra"
)

// New creates a new login command
func NewLoginCmd(ec *ExecutionContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "login",
		Short: "Login to Inventory Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := ec.Logger.WithPrefix("Login")
			ec.Authenticator.SetLogger(logger)

			loginInput := authentication.LoginInput{
				Provider:    ec.OIDCProvider,
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

			_, err := ec.Authenticator.Login(cmd.Context(), loginInput)
			if err != nil {
				return fmt.Errorf("logging in: %w", err)
			}
			return nil
		},
	}
	return command
}
