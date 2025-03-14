package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/tokencache"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/spf13/cobra"
)

// New creates a new logout command
func logoutCmd(ac *ic.Context) *cobra.Command {
	o := &logoutOptions{}
	c := cmd.NewSubCommand("logout", o, ac).
		WithShortDesc("Log out of Inventory Server").
		WithGroupID(groupAuth).
		Build()

	c.Flags().SortFlags = false
	return c
}

type logoutOptions struct{}

func (o *logoutOptions) Complete(_ context.Context, _ *ic.Context) error { return nil }
func (o *logoutOptions) Validate(_ context.Context, _ *ic.Context) error { return nil }
func (o *logoutOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Logout")
	ac.Authenticator.SetLogger(logger)

	logoutInput := authentication.LogoutInput{
		Provider:   *ac.OIDCProvider,
		TokenCache: ac.TokenCache,
	}

	ac.EC.Spin("Logging out")
	err := ac.Authenticator.Logout(ctx, logoutInput)
	if err != nil && !errors.Is(err, &tokencache.CacheMissError{}) {
		return fmt.Errorf("logging out: %w", err)
	}

	ac.EC.Logger.Info("Logout succeeded ✅")

	return nil
}
