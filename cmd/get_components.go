package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/usecases/component"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// New creates a new "get components" command
func NewGetComponentsCmd(ec *ExecutionContext) *cobra.Command {
	o := getComponentsOptions{}
	c := &cobra.Command{
		Use:     "components",
		Short:   "Get list of components",
		GroupID: groupComponent,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(ec)
		},
	}
	o.bindFlags(c.Flags())
	return c
}

type getComponentsOptions struct{}

func (o *getComponentsOptions) bindFlags(f *pflag.FlagSet) {}

func (o *getComponentsOptions) run(ec *ExecutionContext) error {
	logger := ec.Logger.WithPrefix("Components")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ec.Spin("Getting components")

	in := component.ListComponentsInput{
		Logger:    logger,
		APIClient: ec.APIClient,
	}
	result, err := component.ListComponents(ec.Command.Context(), in)
	if err != nil {
		return fmt.Errorf("listing components: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "listing components",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	r := component.NewComponentsRenderer(result.ComponentListResponse, result.JSONResponse, ec.Stdout, ec.NoHeaders)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
