package cmd

import (
	"errors"
	"fmt"

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

// Logout represents the logout command
type Logout struct {
	Authenticator authentication.Authenticator
	TokenCache    tokencache.Cache
	Logger        logger.Logger
}

// New creates a new logout command
func (c *Logout) New() *cobra.Command {
	var o logoutOptions
	command := &cobra.Command{
		Use:   "logout",
		Short: "Log out",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := c.Logger.WithPrefix("Logout")
			c.Authenticator.SetLogger(logger)

			var err error
			if c.TokenCache, err = tokencache.NewFSCache(o.authenticationOptions.OIDCTokenCacheDir); err != nil {
				return fmt.Errorf("creating token cache: %w", err)
			}

			logoutInput := authentication.LogoutInput{
				Provider: oidc.Provider{
					IssuerURL:   o.authenticationOptions.OIDCIssuerURL,
					ClientID:    o.authenticationOptions.OIDCClientID,
					ExtraScopes: []string{"profile", "email", "roles", "offline_access"},
				},
				TokenCache: c.TokenCache,
			}

			err = c.Authenticator.Logout(cmd.Context(), logoutInput)
			if err != nil && !errors.Is(err, &tokencache.CacheMissError{}) {
				return fmt.Errorf("logging out: %w", err)
			}
			return nil
		},
	}
	command.Flags().SortFlags = false
	o.addFlags(command.Flags())
	return command
}
