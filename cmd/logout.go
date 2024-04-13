package cmd

import (
	"errors"
	"fmt"

	"github.com/neticdk-k8s/ic/internal/tokencache"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/spf13/cobra"
)

// New creates a new logout command
func NewLogoutCmd(ec *ExecutionContext) *cobra.Command {
	command := &cobra.Command{
		Use:     "logout",
		Short:   "Log out",
		GroupID: groupAuth,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := ec.Logger.WithPrefix("Logout")
			ec.Authenticator.SetLogger(logger)

			logoutInput := authentication.LogoutInput{
				Provider:   ec.OIDCProvider,
				TokenCache: ec.TokenCache,
			}

			ec.Spin("Logging out")
			err := ec.Authenticator.Logout(cmd.Context(), logoutInput)
			if err != nil && !errors.Is(err, &tokencache.CacheMissError{}) {
				return fmt.Errorf("logging out: %w", err)
			}
			return nil
		},
	}
	return command
}
