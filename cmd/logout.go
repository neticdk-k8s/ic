package cmd

import (
	"os"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/tokencache"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type logoutOptions struct {
	authenticationOptions authenticationOptions
}

func (o *logoutOptions) addFlags(f *pflag.FlagSet) {
	o.authenticationOptions.addFlags(f)
}

type Logout struct {
	Authenticator authentication.Interface
	TokenCache    tokencache.Interface
	Logger        logger.Logger
}

func (c *Logout) New() *cobra.Command {
	var o logoutOptions
	command := &cobra.Command{
		Use:   "logout",
		Short: "Log out",
		Run: func(cmd *cobra.Command, args []string) {
			logger := c.Logger.WithPrefix("Logout")
			c.Authenticator.SetLogger(logger)

			logoutInput := authentication.LogoutInput{
				Provider: oidc.Provider{
					IssuerURL:   o.authenticationOptions.OIDCIssuerURL,
					ClientID:    o.authenticationOptions.OIDCClientID,
					ExtraScopes: []string{"profile", "email", "roles", "offline_access"},
				},
				TokenCache: c.TokenCache,
			}

			err := c.Authenticator.Logout(cmd.Context(), logoutInput)
			if err != nil {
				logger.Error("Logout failed", "err", err)
				os.Exit(1)
			}
		},
	}
	command.Flags().SortFlags = false
	o.addFlags(command.Flags())
	return command
}
