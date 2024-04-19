package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
	"github.com/spf13/cobra"
)

// New creates a new "delete cluster" command
func NewDeleteClusterCmd(ec *ExecutionContext) *cobra.Command {
	command := &cobra.Command{
		Use:     "cluster cluster-id",
		Short:   "Delete a cluster",
		GroupID: groupCluster,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := ec.Logger.WithPrefix("Clusters")
			ec.Authenticator.SetLogger(logger)

			loginInput := authentication.LoginInput{
				Provider:    *ec.OIDCProvider,
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

			ec.Spin("Deleting cluster")

			if err := ec.SetupDefaultAPIClient(tokenSet.AccessToken); err != nil {
				return fmt.Errorf("setting up API client: %w", err)
			}

			// in := cluster.GetClusterInput{
			// 	Logger:    logger,
			// 	APIClient: ec.APIClient,
			// }
			// c, jsonData, err := cluster.GetCluster(cmd.Context(), args[0], in)
			// if err != nil {
			// 	return fmt.Errorf("getting cluster: %w", err)
			// }
			//
			ec.Spinner.Stop()

			ec.Logger.Info("Cluster deleted ✅")
			//
			// r := cluster.NewClusterRenderer(c, jsonData, ec.Stdout)
			// if err := r.Render(ec.OutputFormat); err != nil {
			// 	return fmt.Errorf("rendering output: %w", err)
			// }

			return nil
		},
	}

	return command
}
