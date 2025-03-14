package cmd

import (
	"fmt"
	"strings"

	"github.com/neticdk-k8s/ic/internal/usecases/region"
	"github.com/neticdk/go-common/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// New creates a new "get regions" command
func NewGetRegionsCmd(ec *ExecutionContext) *cobra.Command {
	o := getRegionsOptions{}
	c := &cobra.Command{
		Use:     "regions",
		Short:   "List regions",
		GroupID: groupOther,
		RunE: func(_ *cobra.Command, _ []string) error {
			return o.run(ec)
		},
	}

	o.bindFlags(c.Flags())
	return c
}

type getRegionsOptions struct {
	Partition string
}

func (o *getRegionsOptions) bindFlags(f *pflag.FlagSet) {
	f.StringVar(&o.Partition, "partition", "", fmt.Sprintf("Partition. One of (%s)", strings.Join(types.AllPartitionsString(), "|")))
}

func (o *getRegionsOptions) run(ec *ExecutionContext) error {
	var (
		regions []string
		err     error
	)
	if o.Partition != "" {
		regions, err = region.ListRegionsForPartition(o.Partition)
		if err != nil {
			return err
		}
	} else {
		regions = region.ListRegions()
	}

	r := region.NewRegionsRenderer(regions, ec.Stdout, ec.NoHeaders)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
