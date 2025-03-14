package cmd

import (
	"fmt"

	goerr "errors"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/ui"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// New creates a new "delete cluster" command
func NewDeleteClusterCmd(ec *ExecutionContext) *cobra.Command {
	o := deleteClusterOptions{}
	c := &cobra.Command{
		Use:     "cluster cluster-id",
		Short:   "Delete a cluster",
		GroupID: groupCluster,
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return o.run(ec, args)
		},
	}

	o.bindFlags(c.Flags())

	return c
}

type deleteClusterOptions struct {
	Yes bool
}

func (o *deleteClusterOptions) bindFlags(f *pflag.FlagSet) {
	f.BoolVarP(&o.Yes, "yes", "y", false, "Automatic yes to prompts")
}

func (o *deleteClusterOptions) run(ec *ExecutionContext, args []string) error {
	logger := ec.Logger.WithPrefix("Clusters")
	ec.Authenticator.SetLogger(logger)

	if !o.Yes {
		if err := ui.Confirm("delete", args[0]); err != nil {
			if goerr.Is(err, ui.ExitConfirmError) {
				ec.Spinner.Stop()
				ec.Logger.Info("User aborted")
				return nil
			}
			return fmt.Errorf("confirming deletion: %w", err)
		}
	}

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ec.Spin("Deleting cluster")

	in := cluster.DeleteClusterInput{
		Logger:    logger,
		APIClient: ec.APIClient,
	}
	result, err := cluster.DeleteCluster(ec.Command.Context(), args[0], in)
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "deleting cluster",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	ec.Logger.Info("Cluster deleted ✅")

	return nil
}
