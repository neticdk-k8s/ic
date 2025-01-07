package cmd

import (
	"github.com/spf13/cobra"
)

const getCommandExample = `  # Get a list of clusters
  ic get clusters

  # Get information about a single cluster
  ic get cluster mycluster.myprovider

  # Get information about a single cluster in json format
  ic -o json get cluster mycluster.myprovider`

// New creates a new get command
func NewGetCmd(ec *ExecutionContext) *cobra.Command {
	c := &cobra.Command{
		Use:     "get",
		Short:   "Get one or many resources",
		Args:    cobra.NoArgs,
		GroupID: groupBase,
		Example: getCommandExample,
	}
	c.AddCommand(
		NewGetClustersCmd(ec),
		NewGetClusterCmd(ec),
		NewGetClusterNodesCmd(ec),
		NewGetClusterNodeCmd(ec),
		NewGetClusterKubeConfigCmd(ec),
		NewGetPartitionsCmd(ec),
		NewGetRegionsCmd(ec),
		NewGetResilienceZonesCmd(ec),
		NewGetComponentsCmd(ec),
		NewGetComponentCmd(ec),
	)

	c.AddGroup(
		&cobra.Group{
			ID:    groupCluster,
			Title: "Cluster Commands:",
		},
		&cobra.Group{
			ID:    groupComponent,
			Title: "Component Commands:",
		},
		&cobra.Group{
			ID:    groupOther,
			Title: "Other Commands:",
		},
	)
	return c
}
