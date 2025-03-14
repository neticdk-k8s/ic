package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk/go-common/pkg/qsparser"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// New creates a new "get cluster-nodes" command
func NewGetClusterNodesCmd(ec *ExecutionContext) *cobra.Command {
	o := getClusterNodesOptions{}
	c := &cobra.Command{
		Use:     "cluster-nodes",
		Aliases: []string{"nodes"},
		Short:   "Get list of nodes for a cluster",
		Long: `Get list of nodes for a nodes.

Supported field names for filters:

name, role, criName, criVersion, controlPlane, topologyRegion,
topologyZone, memoryAllocatableBytes, cpuAllocatableMillis,
memoryCapacityBytes, cpuCapacityMillis
`,

		GroupID: groupCluster,
		Example: `
# get nodes for my-cluster.my-provider
ic get cluster-nodes --cluster-name my-cluster.my-provider

use: 'ic help filters' for more information on using filters`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return o.run(ec)
		},
	}
	o.bindFlags(c.Flags())
	c.MarkFlagRequired("cluster-name") //nolint:errcheck
	return c
}

type getClusterNodesOptions struct {
	ClusterName string
	// A filter has the form: fieldName operator value (e.g. name=Peter)
	//
	// Supported operators:
	// == (or =) - equals
	// != (or !) - not equals
	// >         - greater than
	// <         - less than
	// >=        - greater than or equals
	// <=        - less than or equals
	// =~ (or ~) - matches (case insensitive regular expression)
	// !~        - does not match (case insensitive expression)
	Filters []string
}

func (o *getClusterNodesOptions) bindFlags(f *pflag.FlagSet) {
	f.StringVar(&o.ClusterName, "cluster-name", "", "The name of the cluster")
	f.StringArrayVarP(&o.Filters, "filter", "f", []string{}, "Filter output based on conditions")
}

func (o *getClusterNodesOptions) run(ec *ExecutionContext) error {
	logger := ec.Logger.WithPrefix("ClusterNodes")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	searchFields := make(map[string]*qsparser.SearchField)
	for _, f := range o.Filters {
		out, err := parseFilter(f)
		if err != nil {
			return err
		}
		searchFields[out.FieldName] = out.SearchField
	}

	ec.Spin("Getting cluster nodes")

	in := cluster.ListClusterNodesInput{
		Logger:      logger,
		APIClient:   ec.APIClient,
		PerPage:     PerPage,
		Filters:     searchFields,
		ClusterName: o.ClusterName,
	}
	result, err := cluster.ListClusterNodes(ec.Command.Context(), in)
	if err != nil {
		return fmt.Errorf("listing cluster nodes: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "listing cluster nodes",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	r := cluster.NewClusterNodesRenderer(result.ClusterNodeListResponse, result.JSONResponse, ec.Stdout, ec.NoHeaders)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
