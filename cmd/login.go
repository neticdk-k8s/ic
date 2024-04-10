package cmd

import (
	"os"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/reader"
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
		Run: func(cmd *cobra.Command, args []string) {
			logger := c.Logger.WithPrefix("Login")
			c.Authenticator.SetLogger(logger)

			var err error
			if c.TokenCache, err = tokencache.NewFSCache(o.authenticationOptions.OIDCTokenCacheDir); err != nil {
				logger.Error("Creating token cache", "err", err)
				os.Exit(1)
			}

			provider := oidc.Provider{
				IssuerURL:   o.authenticationOptions.OIDCIssuerURL,
				ClientID:    o.authenticationOptions.OIDCClientID,
				ExtraScopes: []string{"profile", "email", "roles", "offline_access"},
			}

			oidcClient, err := oidc.New(cmd.Context(), provider, logger)
			if err != nil {
				logger.Error("Failed creating OIDC client", "err", err)
				os.Exit(1)
			}

			loginInput := authentication.LoginInput{
				Authentication: authentication.NewAuthentication(logger, oidcClient, &authcode.Browser{Logger: logger}, &authcode.Keyboard{Reader: reader.NewReader(), Logger: logger}),
				Provider:       provider,
				TokenCache:     c.TokenCache,
				AuthOptions:    authentication.AuthOptions{},
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

			err = c.Authenticator.Login(cmd.Context(), loginInput)
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
