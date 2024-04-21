package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/usecases/resiliencezone"
	"github.com/spf13/cobra"
)

// New creates a new "get resilience-zones" command
func NewGetResilienceZonesCmd(ec *ExecutionContext) *cobra.Command {
	o := getResilienceZonesOptions{}
	c := &cobra.Command{
		Use:     "resilience-zones",
		Short:   "List resilience zones",
		Aliases: []string{"rzs"},
		GroupID: groupOther,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(ec)
		},
	}
	return c
}

type getResilienceZonesOptions struct{}

func (o *getResilienceZonesOptions) run(ec *ExecutionContext) error {
	rzs := resiliencezone.ListResilienceZones()
	r := resiliencezone.NewResilienceZonesRenderer(rzs, ec.Stdout, ec.NoHeaders)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
