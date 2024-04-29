package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// New creates a new filters command
func NewFiltersHelpCmd(ec *ExecutionContext) *cobra.Command {
	c := &cobra.Command{
		Use:   "filters",
		Short: "About filters",
		Args:  cobra.NoArgs,
	}
	c.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "About filters")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Filters can be used with various commands, typically those that return lists (e.g. get clusters). They are provided by using the --filter (-f) flag. This flag can be specified multiple times in which case the filters are joined using AND.")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Examples:")
		fmt.Fprintln(cmd.OutOrStdout(), "A filter consists of 1) a field name, 2) a search operator, and 3) a search value. It looks like this:")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "# add a filter on the field 'name' using the equals(=) operator and the search value 'peter':")
		fmt.Fprintln(cmd.OutOrStdout(), "--filter name=peter")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "# add a filter on the field 'resilienceZone' using the not equals(!=) operator and the search value 'platform':")
		fmt.Fprintln(cmd.OutOrStdout(), "--filter resilienceZone!=platform")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "# add a filter on the field 'age' using the greater than(>) operator and the search value '10':")
		fmt.Fprintln(cmd.OutOrStdout(), "--filter 'age>10'")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "# add a filter on the field 'region' using the matches(~) operator and the search value 'eu-west':")
		fmt.Fprintln(cmd.OutOrStdout(), "--filter region~eu-west")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Supported Fields:")
		fmt.Fprintln(cmd.OutOrStdout(), "Supported fields depeneds on the command. Check help for that command for a list of its supported fields.")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Supported Operators:")
		fmt.Fprintln(cmd.OutOrStdout(), "=   - is equals to")
		fmt.Fprintln(cmd.OutOrStdout(), "==  - is equals to")
		fmt.Fprintln(cmd.OutOrStdout(), "!=  - is not equals to")
		fmt.Fprintln(cmd.OutOrStdout(), "!   - is not equals to")
		fmt.Fprintln(cmd.OutOrStdout(), ">   - is greater than")
		fmt.Fprintln(cmd.OutOrStdout(), "<   - is smaller than")
		fmt.Fprintln(cmd.OutOrStdout(), ">=  - is greater than or equals to")
		fmt.Fprintln(cmd.OutOrStdout(), "<=  - is smaller than or equals to")
		fmt.Fprintln(cmd.OutOrStdout(), "=~  - matches case insensitive using regular expressions")
		fmt.Fprintln(cmd.OutOrStdout(), "~   - matches case insensitive using regular expressions")
		fmt.Fprintln(cmd.OutOrStdout(), "!~  - does not match case insensitive using regular expressions")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Note that not all fields may support all operators.")
		fmt.Fprintln(cmd.OutOrStdout())
	})
	return c
}
