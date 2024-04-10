package cmd

import (
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/spf13/cobra"
)

// Get represents the get command
type Get struct {
	Logger logger.Logger
}

// New creates a new get command
func (c *Get) New() *cobra.Command {
	command := &cobra.Command{
		Use:   "get",
		Short: "Get one or many resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	command.Flags().SortFlags = false
	return command
}
