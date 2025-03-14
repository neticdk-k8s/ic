package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// New creates a new "get cluster-node" command
func NewGetClusterNodeCmd(ec *ExecutionContext) *cobra.Command {
	o := getClusterNodeOptions{}
	c := &cobra.Command{
		Use:     "cluster-node",
		Aliases: []string{"node"},
		Short:   "Get a cluster node",
		GroupID: groupCluster,
		Example: `
# get node information for my-cluster.my-provider
ic get cluster-node --cluster-name my-cluster.my-provider --node-name my-node`,
		RunE: func(_ *cobra.Command, args []string) error {
			return o.run(ec, args)
		},
	}
	o.bindFlags(c.Flags())
	c.MarkFlagRequired("cluster-name") //nolint:errcheck
	c.MarkFlagRequired("node-name")    //nolint:errcheck
	return c
}

type getClusterNodeOptions struct {
	ClusterName string
	NodeName    string
}

func (o *getClusterNodeOptions) bindFlags(f *pflag.FlagSet) {
	f.StringVar(&o.ClusterName, "cluster-name", "", "The name of the cluster")
	f.StringVar(&o.NodeName, "node-name", "", "The name of the node")
}

func (o *getClusterNodeOptions) run(ec *ExecutionContext, _ []string) error {
	logger := ec.Logger.WithPrefix("ClusterNode")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ec.Spin("Getting cluster node")

	in := cluster.GetClusterNodeInput{
		Logger:      logger,
		APIClient:   ec.APIClient,
		ClusterName: o.ClusterName,
		NodeName:    o.NodeName,
	}
	result, err := cluster.GetClusterNode(ec.Command.Context(), in)
	if err != nil {
		return fmt.Errorf("getting cluster node: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "getting cluster node",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	r := cluster.NewClusterNodeRenderer(result.ClusterNodeResponse, result.JSONResponse, ec.Stdout)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
