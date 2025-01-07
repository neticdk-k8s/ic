package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/usecases/component"
	"github.com/spf13/cobra"
)

// New creates a new "get component" command
func NewGetComponentCmd(ec *ExecutionContext) *cobra.Command {
	o := getComponentOptions{}
	c := &cobra.Command{
		Use:   "component NAMESPACE NAME",
		Short: "Get a component",
		Long: `Get a component

The component-id has the format namespace/name
`,
		GroupID: groupComponent,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(ec, args)
		},
	}
	return c
}

type getComponentOptions struct{}

func (o *getComponentOptions) run(ec *ExecutionContext, args []string) error {
	logger := ec.Logger.WithPrefix("Components")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ec.Spin("Getting component")

	in := component.GetComponentInput{
		Logger:    logger,
		APIClient: ec.APIClient,
	}
	result, err := component.GetComponent(ec.Command.Context(), args[0], args[1], in)
	if err != nil {
		return fmt.Errorf("getting component: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "getting component",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	r := component.NewComponentRenderer(result.ComponentResponse, result.JSONResponse, ec.Stdout)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
