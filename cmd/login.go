package cmd

import (
	"fmt"

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

// Login represents the login command
type Login struct {
	Authenticator authentication.Authenticator
	TokenCache    tokencache.Cache
	Logger        logger.Logger
}

// New creates a new login command
func (c *Login) New() *cobra.Command {
	var o loginOptions
	command := &cobra.Command{
		Use:   "login",
		Short: "Login to Inventory Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := c.Logger.WithPrefix("Login")
			c.Authenticator.SetLogger(logger)

			var err error
			if c.TokenCache, err = tokencache.NewFSCache(o.authenticationOptions.OIDCTokenCacheDir); err != nil {
				return fmt.Errorf("creating token cache: %w", err)
			}

			provider := oidc.Provider{
				IssuerURL:   o.authenticationOptions.OIDCIssuerURL,
				ClientID:    o.authenticationOptions.OIDCClientID,
				ExtraScopes: []string{"profile", "email", "roles", "offline_access"},
			}

			loginInput := authentication.LoginInput{
				Provider:    provider,
				TokenCache:  c.TokenCache,
				AuthOptions: authentication.AuthOptions{},
			}
			if o.authenticationOptions.OIDCGrantType == "authcode-browser" {
				loginInput.AuthOptions.AuthCodeBrowser = &authcode.BrowserLoginInput{
					BindAddress:         o.authenticationOptions.OIDCAuthBindAddr,
					RedirectURLHostname: o.authenticationOptions.OIDCRedirectURLHostname,
				}
			} else if o.authenticationOptions.OIDCGrantType == "authcode-keyboard" {
				loginInput.AuthOptions.AuthCodeKeyboard = &authcode.KeyboardLoginInput{
					RedirectURI: o.authenticationOptions.OIDCRedirectURIAuthCodeKeyboard,
				}
			}

			_, err = c.Authenticator.Login(cmd.Context(), loginInput)
			if err != nil {
				return fmt.Errorf("logging in: %w", err)
			}
			return nil
		},
	}
	command.Flags().SortFlags = false
	o.addFlags(command.Flags())
	return command
}
