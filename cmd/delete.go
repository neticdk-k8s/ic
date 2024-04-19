package cmd

import "github.com/spf13/cobra"

const deleteCommandExample = `  # Delete a cluster
  ic delete cluster my-cluster.my-provider`

// New creates a new delete command
func NewDeleteCmd(ec *ExecutionContext) *cobra.Command {
	command := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a resources",
		GroupID: groupBase,
		Example: deleteCommandExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	command.AddCommand(
		NewDeleteClusterCmd(ec),
	)

	command.AddGroup(
		&cobra.Group{
			ID:    groupCluster,
			Title: "Cluster Commands:",
		},
	)
	return command
}
