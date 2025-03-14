package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/usecases/partition"
	"github.com/spf13/cobra"
)

// New creates a new "get partitions" command
func NewGetPartitionsCmd(ec *ExecutionContext) *cobra.Command {
	o := getPartitionsOptions{}
	c := &cobra.Command{
		Use:     "partitions",
		Short:   "List partitions",
		GroupID: groupOther,
		RunE: func(_ *cobra.Command, _ []string) error {
			return o.run(ec)
		},
	}
	return c
}

type getPartitionsOptions struct{}

func (o *getPartitionsOptions) run(ec *ExecutionContext) error {
	partitions := partition.ListPartitions()
	r := partition.NewPartitionsRenderer(partitions, ec.Stdout, ec.NoHeaders)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
