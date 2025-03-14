package cmd

import "github.com/spf13/cobra"

const updateCommandExample = `  # Update a cluster, enabling Application Operations
  ic update cluster mycluster.my-provider --has-ao

  # Update a cluster, settings a new description and resilience zone
  ic update cluster mycluster.my-provider
	--description "new description"
	--resilience-zone new-resilience-zone`

// New creates a new update command
func NewUpdateCmd(ec *ExecutionContext) *cobra.Command {
	c := &cobra.Command{
		Use:     "update",
		Short:   "update a resources",
		GroupID: groupBase,
		Args:    cobra.NoArgs,
		Example: updateCommandExample,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Usage()
		},
	}
	c.AddCommand(
		NewUpdateClusterCmd(ec),
	)

	c.AddGroup(
		&cobra.Group{
			ID:    groupCluster,
			Title: "Cluster Commands:",
		},
	)
	return c
}
