package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// New creates a new "get cluster-kubeconfig" command
func NewGetClusterKubeConfigCmd(ec *ExecutionContext) *cobra.Command {
	o := getClusterKubeConfigOptions{}
	c := &cobra.Command{
		Use:     "cluster-kubeconfig",
		Aliases: []string{"kubeconfig"},
		Short:   "Get a cluster kubeconfig",
		GroupID: groupCluster,
		Example: `
# get the kubeconfig for a cluster
ic get cluster-kubeconfig --cluster-name my-cluster.my-provider`,
		RunE: func(_ *cobra.Command, args []string) error {
			return o.run(ec, args)
		},
	}
	o.bindFlags(c.Flags())
	c.MarkFlagRequired("cluster-name") //nolint:errcheck
	return c
}

type getClusterKubeConfigOptions struct {
	ClusterName string
}

func (o *getClusterKubeConfigOptions) bindFlags(f *pflag.FlagSet) {
	f.StringVar(&o.ClusterName, "cluster-name", "", "The name of the cluster")
}

func (o *getClusterKubeConfigOptions) run(ec *ExecutionContext, _ []string) error {
	logger := ec.Logger.WithPrefix("ClusterKubeConfig")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ec.Spin("Getting cluster kubeconfig")

	in := cluster.GetClusterKubeConfigInput{
		Logger:      logger,
		APIClient:   ec.APIClient,
		ClusterName: o.ClusterName,
	}
	result, err := cluster.GetClusterKubeConfig(ec.Command.Context(), in)
	if err != nil {
		return fmt.Errorf("getting cluster kubeconfig: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "getting cluster kubeconfig",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	if ec.OutputFormat != "json" {
		ec.OutputFormat = "yaml"
	}
	r := cluster.NewClusterKubeConfigRenderer(result.Response, ec.Stdout)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
