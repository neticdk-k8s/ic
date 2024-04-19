package cmd

import (
	"fmt"
	"strings"

	"github.com/neticdk-k8s/ic/internal/ui"
	"github.com/neticdk/go-common/pkg/types"
	"github.com/spf13/cobra"
)

// New creates a new "get regions" command
func NewGetRegionsCmd(ec *ExecutionContext) *cobra.Command {
	var partition string
	command := &cobra.Command{
		Use:     "regions",
		Short:   "List regions",
		GroupID: groupOther,
		RunE: func(cmd *cobra.Command, args []string) error {
			var headers []string
			if !ec.NoHeaders {
				headers = []string{"region"}
			}
			table := ui.NewTable(ec.Stdout, headers)

			var regions types.Regions
			if partition != "" {
				part, ok := types.ParsePartition(partition)
				if !ok {
					return fmt.Errorf(`invalid partition: %s`, partition)
				}
				regions = types.PartitionRegions(part)
			} else {
				regions = types.AllRegions()
			}
			for _, p := range regions {
				table.Append([]string{p.String()})
			}
			table.Render()

			return nil
		},
	}
	var partitions []string
	for _, p := range types.AllPartitions() {
		partitions = append(partitions, p.String())
	}

	f := command.Flags()
	f.StringVar(&partition, "partition", "", fmt.Sprintf("Partition. One of (%s)", strings.Join(partitions, "|")))
	return command
}
