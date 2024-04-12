package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/spf13/cobra"
)

// New creates a new "get clusters" command
func NewGetClustersCmd(ec *ExecutionContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "clusters",
		Short: "Get list of clusters",
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

			ec.Spin("Gettings clusters")

			if err := ec.SetupAPIClient(tokenSet.IDToken); err != nil {
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

			r := cluster.NewClustersRenderer(cs, jsonData, ec.Stdout)
			if err := r.Render(ec.OutputFormat); err != nil {
				return fmt.Errorf("rendering output: %w", err)
			}

			return nil
		},
	}
	return command
}

func prettyPrintJSON(body []byte) error {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(prettyJSON.String())
	return nil
}
