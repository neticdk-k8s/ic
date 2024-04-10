package cmd

import (
	"fmt"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/tokencache"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication/authcode"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/cluster"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type getClustersOptions struct {
	authenticationOptions authenticationOptions
}

func (o *getClustersOptions) addFlags(f *pflag.FlagSet) {
	o.authenticationOptions.addFlags(f)
}

// GetClusters represents the "get clusters" command
type GetClusters struct {
	Authenticator authentication.Authenticator
	TokenCache    tokencache.Cache
	Logger        logger.Logger
}

// New creates a new "get clusters" command
func (c *GetClusters) New() *cobra.Command {
	var o getClustersOptions
	command := &cobra.Command{
		Use:   "clusters",
		Short: "Get list of clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := c.Logger.WithPrefix("Clusters")
			c.Authenticator.SetLogger(logger)

			var err error
			if c.TokenCache, err = tokencache.NewFSCache(o.authenticationOptions.OIDCTokenCacheDir); err != nil {
				return fmt.Errorf("creating token cache: %w", err)
			}

			provider := oidc.Provider{
				IssuerURL:   o.authenticationOptions.OIDCIssuerURL,
				ClientID:    o.authenticationOptions.OIDCClientID,
				ExtraScopes: []string{"profile", "email", "roles", "offline_access"},
			}

			loginInput := authentication.LoginInput{
				Provider:    provider,
				TokenCache:  c.TokenCache,
				AuthOptions: authentication.AuthOptions{},
				Silent:      true,
			}
			if o.authenticationOptions.OIDCGrantType == "authcode-browser" {
				loginInput.AuthOptions.AuthCodeBrowser = &authcode.BrowserLoginInput{
					BindAddress:         o.authenticationOptions.OIDCAuthBindAddr,
					RedirectURLHostname: o.authenticationOptions.OIDCRedirectURLHostname,
				}
			} else if o.authenticationOptions.OIDCGrantType == "authcode-keyboard" {
				loginInput.AuthOptions.AuthCodeKeyboard = &authcode.KeyboardLoginInput{
					RedirectURI: o.authenticationOptions.OIDCRedirectURIAuthCodeKeyboard,
				}
			}

			tokenSet, err := c.Authenticator.Login(cmd.Context(), loginInput)
			if err != nil {
				return fmt.Errorf("logging in: %w", err)
			}

			c := cluster.NewClient(logger, "http://localhost:8087", tokenSet.IDToken)
			clusters, err := c.GetClusters(cmd.Context())
			if err != nil {
				return fmt.Errorf("reading clusters: %w", err)
			}

			for _, i := range clusters.Included {
				fmt.Println(i.(map[string]interface{})["name"])
			}

			return nil
		},
	}
	command.Flags().SortFlags = false
	o.addFlags(command.Flags())
	return command
}
