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
	command := &cobra.Command{
		Use:     "get",
		Short:   "Get one or many resources",
		GroupID: groupBase,
		Example: getCommandExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	command.AddCommand(
		NewGetClustersCmd(ec),
		NewGetClusterCmd(ec),
	)

	command.AddGroup(
		&cobra.Group{
			ID:    groupCluster,
			Title: "Cluster Commands:",
		},
	)
	return command
}
