package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/ui"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication/authcode"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/cluster"
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
			}
			clusters, err := cluster.ListClusters(cmd.Context(), in)
			if err != nil {
				return fmt.Errorf("listing clusters: %w", err)
			}

			ec.Spinner.Stop()

			if ec.OutputFormat == "json" {
				return prettyPrintJSON(clusters.Body)
			}

			table := ui.NewTable(ec.Stdout, []string{"provider", "name", "rz", "version"})
			for _, i := range *clusters.ApplicationldJSONDefault.Included {
				rzParts := strings.Split(i["resilienceZone"].(string), "/")
				table.Append(
					[]string{
						i["provider"].(string),
						i["name"].(string),
						rzParts[len(rzParts)-1],
						i["kubernetesVersion"].(map[string]interface{})["version"].(string),
					},
				)
			}
			table.Render()

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
