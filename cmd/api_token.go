package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// New creates a new api-token command
func NewAPITokenCmd(ec *ExecutionContext) *cobra.Command {
	o := apiTokenOptions{}
	c := &cobra.Command{
		Use:     "api-token",
		Short:   "Get access token for the API",
		Long:    "Get access token for the API. It uses the cached token if it is valid and performs login otherwise.",
		GroupID: groupAuth,
		RunE: func(_ *cobra.Command, _ []string) error {
			return o.run(ec)
		},
	}
	return c
}

type apiTokenOptions struct{}

func (o *apiTokenOptions) run(ec *ExecutionContext) error {
	logger := ec.Logger.WithPrefix("Login")
	ec.Authenticator.SetLogger(logger)

	tokenSet, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	fmt.Println(tokenSet.AccessToken)
	return nil
}
