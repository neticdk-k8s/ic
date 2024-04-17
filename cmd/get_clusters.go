package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/spf13/cobra"
)

// New creates a new "get clusters" command
func NewGetClustersCmd(ec *ExecutionContext) *cobra.Command {
	command := &cobra.Command{
		Use:     "clusters",
		Short:   "Get list of clusters",
		GroupID: groupCluster,
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

			ec.Spin("Getting clusters")

			if err := ec.SetupDefaultAPIClient(tokenSet.AccessToken); err != nil {
				return fmt.Errorf("setting up API client: %w", err)
			}

			in := cluster.ListClustersInput{
				Logger:    logger,
				APIClient: ec.APIClient,
				PerPage:   50,
			}
			cs, jsonData, err := cluster.ListClusters(cmd.Context(), in)
			if err != nil {
				return fmt.Errorf("listing clusters: %w", err)
			}

			ec.Spinner.Stop()

			r := cluster.NewClustersRenderer(cs, jsonData, ec.Stdout, ec.NoHeaders)
			if err := r.Render(ec.OutputFormat); err != nil {
				return fmt.Errorf("rendering output: %w", err)
			}

			return nil
		},
	}
	return command
}
