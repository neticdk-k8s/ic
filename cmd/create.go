package cmd

import "github.com/spf13/cobra"

const createCommandExample = `  # Create a new Netic on-prem cluster with default SLA (TO, TM)
  ic create cluster
	--name my-cluster
	--provider my-provider
	--environment myenv
	--partition netic
	--region dk-north
	--subscription 123456
	--infrastructure-provider netic

  # Create a new Azure cluster with Application Operations
  ic create cluster
	--name my-azure-cluster
	--provider my-azure-provider
	--environment myenv
	--partition azure
	--region swedencentral
	--subscription 654321
	--infrastructure-provider azure
	--has-application-operations`

// New creates a new create command
func NewCreateCmd(ec *ExecutionContext) *cobra.Command {
	c := &cobra.Command{
		Use:     "create",
		Short:   "Create a resources",
		GroupID: groupBase,
		Args:    cobra.NoArgs,
		Example: createCommandExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	c.AddCommand(
		NewCreateClusterCmd(ec),
	)

	c.AddGroup(
		&cobra.Group{
			ID:    groupCluster,
			Title: "Cluster Commands:",
		},
	)
	return c
}
