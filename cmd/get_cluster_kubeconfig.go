package cmd

import (
	"context"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const getClusterKubeConfigExample = `
# get the kubeconfig for a cluster
ic get cluster-kubeconfig --cluster-name my-cluster.my-provider`

func getClusterKubeconfigCmd(ac *ic.Context) *cobra.Command {
	o := &getClusterKubeConfigOptions{}
	c := cmd.NewSubCommand("cluster-kubeconfig", o, ac).
		WithShortDesc("Get a cluster kubeconfig").
		WithGroupID(groupCluster).
		WithExample(getClusterKubeConfigExample).
		Build()
	c.Aliases = []string{"kubeconfig"}

	o.bindFlags(c.Flags())
	c.MarkFlagRequired("cluster-name") //nolint:errcheck
	return c
}

type getClusterKubeConfigOptions struct {
	clusterName string
}

func (o *getClusterKubeConfigOptions) bindFlags(f *pflag.FlagSet) {
	f.StringVar(&o.clusterName, "cluster-name", "", "The name of the cluster")
}

func (o *getClusterKubeConfigOptions) Complete(_ context.Context, _ *ic.Context) error { return nil }
func (o *getClusterKubeConfigOptions) Validate(_ context.Context, _ *ic.Context) error { return nil }

func (o *getClusterKubeConfigOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("ClusterKubeConfig")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	var result *cluster.GetClusterKubeConfigResult
	if err := ui.Spin(ac.EC.Spinner, "Getting kubeconfig", func(_ ui.Spinner) error {
		in := cluster.GetClusterKubeConfigInput{
			Logger:      logger,
			APIClient:   ac.APIClient,
			ClusterName: o.clusterName,
		}
		result, err = cluster.GetClusterKubeConfig(ctx, in)
		return err
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Getting kubeconfig",
			"See details for more information",
			err,
			0,
		)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "getting cluster kubeconfig",
			Problem: result.Problem,
		}
	}

	if ac.EC.PFlags.OutputFormat != "json" {
		ac.EC.PFlags.OutputFormat = "yaml"
	}
	r := cluster.NewClusterKubeConfigRenderer(result.Response, ac.EC.Stdout)
	if err := r.Render(ac.EC.PFlags.OutputFormat); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Failed to render output",
			"See details for more information",
			err,
			0,
		)
	}

	return nil
}
