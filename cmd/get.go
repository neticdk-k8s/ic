package cmd

import (
	"github.com/spf13/cobra"
)

// New creates a new get command
func NewGetCmd(ec *ExecutionContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "get",
		Short: "Get one or many resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	command.AddCommand(
		NewGetClustersCmd(ec),
		NewGetClusterCmd(ec),
	)
	return command
}
