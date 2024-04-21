package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/spf13/cobra"
)

// New creates a new "get clusters" command
func NewGetClustersCmd(ec *ExecutionContext) *cobra.Command {
	o := getClustersOptions{}
	c := &cobra.Command{
		Use:     "clusters",
		Short:   "Get list of clusters",
		GroupID: groupCluster,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(ec)
		},
	}
	return c
}

type getClustersOptions struct{}

func (o *getClustersOptions) run(ec *ExecutionContext) error {
	logger := ec.Logger.WithPrefix("Clusters")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ec.Spin("Getting clusters")

	in := cluster.ListClustersInput{
		Logger:    logger,
		APIClient: ec.APIClient,
		PerPage:   50,
	}
	result, err := cluster.ListClusters(ec.Command.Context(), in)
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "listing clusters",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	r := cluster.NewClustersRenderer(result.ClusterListResponse, result.JSONResponse, ec.Stdout, ec.NoHeaders)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
