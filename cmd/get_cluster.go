package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/spf13/cobra"
)

// New creates a new "get cluster" command
func NewGetClusterCmd(ec *ExecutionContext) *cobra.Command {
	o := getClusterOptions{}
	c := &cobra.Command{
		Use:     "cluster cluster-id",
		Short:   "Get a cluster",
		GroupID: groupCluster,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(ec, args)
		},
	}
	return c
}

type getClusterOptions struct{}

func (o *getClusterOptions) run(ec *ExecutionContext, args []string) error {
	logger := ec.Logger.WithPrefix("Clusters")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ec.Spin("Getting cluster")

	in := cluster.GetClusterInput{
		Logger:    logger,
		APIClient: ec.APIClient,
	}
	result, err := cluster.GetCluster(ec.Command.Context(), args[0], in)
	if err != nil {
		return fmt.Errorf("getting cluster: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "creating cluster",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	r := cluster.NewClusterRenderer(result.ClusterResponse, result.JSONResponse, ec.Stdout)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
