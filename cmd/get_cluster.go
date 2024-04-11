package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication/authcode"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/cluster"
	"github.com/spf13/cobra"
)

// New creates a new "get cluster" command
func NewGetClusterCmd(ec *ExecutionContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "cluster [cluster-id]",
		Short: "Get a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := ec.Logger.WithPrefix("Clusters")
			ec.Authenticator.SetLogger(logger)

			loginInput := authentication.LoginInput{
				Provider:    ec.OIDCProvider,
				TokenCache:  ec.TokenCache,
				AuthOptions: authentication.AuthOptions{},
				Silent:      true,
			}
			if ec.OIDC.GrantType == "authcode-browser" {
				ec.Spin("Logging in")
				loginInput.AuthOptions.AuthCodeBrowser = &authcode.BrowserLoginInput{
					BindAddress:         ec.OIDC.AuthBindAddr,
					RedirectURLHostname: ec.OIDC.RedirectURLHostname,
				}
			} else if ec.OIDC.GrantType == "authcode-keyboard" {
				loginInput.AuthOptions.AuthCodeKeyboard = &authcode.KeyboardLoginInput{
					RedirectURI: ec.OIDC.RedirectURIAuthCodeKeyboard,
				}
			}

			tokenSet, err := ec.Authenticator.Login(cmd.Context(), loginInput)
			if err != nil {
				return fmt.Errorf("logging in: %w", err)
			}

			ec.Spin("Gettings cluster")

			if err := ec.SetupAPIClient(tokenSet.IDToken); err != nil {
				return fmt.Errorf("setting up API client: %w", err)
			}

			if err := cluster.GetCluster(cmd.Context(), args[0], cluster.GetClusterInput{
				Logger:       logger,
				APIClient:    ec.APIClient,
				OutputFormat: ec.OutputFormat,
				Spinner:      ec.Spinner,
			}); err != nil {
				return fmt.Errorf("getting cluster: %w", err)
			}

			return nil
		},
	}
	return command
}
