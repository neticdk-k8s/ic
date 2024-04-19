package cmd

import (
	"github.com/neticdk-k8s/ic/internal/ui"
	"github.com/neticdk/go-common/pkg/types"
	"github.com/spf13/cobra"
)

// New creates a new "get partitions" command
func NewGetPartitionsCmd(ec *ExecutionContext) *cobra.Command {
	command := &cobra.Command{
		Use:     "partitions",
		Short:   "List partitions",
		GroupID: groupOther,
		RunE: func(cmd *cobra.Command, args []string) error {
			var headers []string
			if !ec.NoHeaders {
				headers = []string{"partition"}
			}
			table := ui.NewTable(ec.Stdout, headers)
			for _, p := range types.AllPartitions() {
				table.Append([]string{p.String()})
			}
			table.Render()

			return nil
		},
	}
	return command
}
