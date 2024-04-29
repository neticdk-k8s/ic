package cmd

import (
	"fmt"
	"regexp"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk/go-common/pkg/qsparser"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// New creates a new "get clusters" command
func NewGetClustersCmd(ec *ExecutionContext) *cobra.Command {
	o := getClustersOptions{}
	c := &cobra.Command{
		Use:     "clusters",
		Short:   "Get list of clusters",
		GroupID: groupCluster,
		Example: `
# get all cluster
ic get clusters

# get clusters in the resilience zone 'platform'
ic get clusters --filter resilienceZone=platform

use: 'ic help filters' for more information on using filters`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(ec)
		},
	}
	o.bindFlags(c.Flags())
	return c
}

type getClustersOptions struct {
	// A filter has the form: fieldName operator value (e.g. name=Peter)
	//
	// Supported operators:
	// == (or =) - equals
	// != (or !) - not equals
	// >         - greater than
	// <         - less than
	// >=        - greater than or equals
	// <=        - less than or equals
	// =~ (or ~) - matches (case insensitive regular expression)
	// !~        - does not match (case insensitive expression)
	Filters []string
}

func (o *getClustersOptions) bindFlags(f *pflag.FlagSet) {
	f.StringArrayVarP(&o.Filters, "filter", "f", []string{}, "Filter output based on conditions")
}

func (o *getClustersOptions) run(ec *ExecutionContext) error {
	logger := ec.Logger.WithPrefix("Clusters")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	searchFields := make(map[string]*qsparser.SearchField)
	for _, f := range o.Filters {
		out, err := parseFilter(f)
		if err != nil {
			return err
		}
		searchFields[out.FieldName] = out.SearchField
	}

	ec.Spin("Getting clusters")

	in := cluster.ListClustersInput{
		Logger:    logger,
		APIClient: ec.APIClient,
		PerPage:   50,
		Filters:   searchFields,
	}
	result, err := cluster.ListClusters(ec.Command.Context(), in)
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "listing clusters",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	r := cluster.NewClustersRenderer(result.ClusterListResponse, result.JSONResponse, ec.Stdout, ec.NoHeaders)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}

type parseFilterOut struct {
	FieldName   string
	SearchField *qsparser.SearchField
}

func parseFilter(filterArg string) (*parseFilterOut, error) {
	r := regexp.MustCompile(`^([a-zA-Z0-9]+)(==|!=|>=|<=|=~|!~|=|!|<|>|~)(.*)$`)
	m := r.FindStringSubmatch(filterArg)
	if m == nil {
		return nil, fmt.Errorf("syntax error in filter: %v", filterArg)
	}
	fieldName := m[1]
	searchOp := m[2]
	searchVal := m[3]
	ops := map[string]string{
		"=":  "eq",
		"==": "eq",
		"!=": "ne",
		"!":  "ne",
		">":  "gt",
		"<":  "lt",
		">=": "ge",
		"<=": "le",
		"=~": "ire",
		"~":  "ire",
		"!~": "nire",
	}
	field := &qsparser.SearchField{
		SearchVal: &searchVal,
	}
	if op, ok := ops[searchOp]; ok {
		field.SearchOp = &op
	} else {
		return nil, fmt.Errorf("unknown search operator: %s in %s", searchOp, filterArg)
	}
	return &parseFilterOut{FieldName: fieldName, SearchField: field}, nil
}
