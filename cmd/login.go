package cmd

import (
	"os"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/tokencache"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication/authcode"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type loginOptions struct {
	authenticationOptions authenticationOptions
}

func (o *loginOptions) addFlags(f *pflag.FlagSet) {
	o.authenticationOptions.addFlags(f)
}

type Login struct {
	Authenticator authentication.Interface
	TokenCache    tokencache.Interface
	Logger        logger.Logger
}

func (c *Login) New() *cobra.Command {
	var o loginOptions
	command := &cobra.Command{
		Use:   "login",
		Short: "Login to Inventory Server",
		Run: func(cmd *cobra.Command, args []string) {
			logger := c.Logger.WithPrefix("Login")
			c.Authenticator.SetLogger(logger)

			loginInput := authentication.LoginInput{
				Provider: oidc.Provider{
					IssuerURL:   o.authenticationOptions.OIDCIssuerURL,
					ClientID:    o.authenticationOptions.OIDCClientID,
					ExtraScopes: []string{"profile", "email", "roles", "offline_access"},
				},
				TokenCache:  c.TokenCache,
				AuthOptions: authentication.AuthOptions{},
			}
			if o.authenticationOptions.OIDCGrantType == "authcode-browser" {
				loginInput.AuthOptions.AuthCodeBrowser = &authcode.BrowserInput{
					BindAddress:         o.authenticationOptions.OIDCAuthBindAddr,
					RedirectURLHostname: o.authenticationOptions.OIDCRedirectURLHostname,
				}
			} else if o.authenticationOptions.OIDCGrantType == "authcode-keyboard" {
				loginInput.AuthOptions.AuthCodeKeyboard = &authcode.KeyboardInput{
					RedirectURL: o.authenticationOptions.OIDCRedirectURLAuthCodeKeyboard,
				}
			}

			err := c.Authenticator.Login(cmd.Context(), loginInput)
			if err != nil {
				logger.Error("Login failed", "err", err)
				os.Exit(1)
			}
		},
	}
	command.Flags().SortFlags = false
	o.addFlags(command.Flags())
	return command
}
