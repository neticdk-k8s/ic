package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
	"github.com/neticdk/go-common/pkg/types"
	"github.com/spf13/cobra"
)

type createClusterOptions struct {
	Name                     string
	ProviderName             string
	EnvironmentName          string
	Partition                string
	Region                   string
	SubscriptionID           string
	InfrastructureProvider   string
	ResilienceZone           string
	APIEndpoint              string
	HasTechnicalOperations   bool
	HasTechnicalManagement   bool
	HasApplicationOperations bool
	HasApplicationManagement bool
	HasCustomOperations      bool
	CustomOperationsURL      string
}

// New creates a new "create cluster" command
func NewCreateClusterCmd(ec *ExecutionContext) *cobra.Command {
	opts := createClusterOptions{}

	var partitions []string
	for _, p := range types.AllPartitions() {
		partitions = append(partitions, p.String())
	}

	command := &cobra.Command{
		Use:     "cluster",
		Short:   "Create a cluster",
		GroupID: groupCluster,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			p, ok := types.ParsePartition(opts.Partition)
			if !ok {
				return fmt.Errorf(`invalid argument "%s" for "--partition" flag: must be one of: %s`, opts.Partition, strings.Join(partitions, "|"))
			}
			r, ok := types.ParseRegion(opts.Region)
			if !ok {
				return fmt.Errorf(`invalid argument "%s" for "--region" flag: see the "get regions" command`, opts.Region)
			}
			if !types.HasRegion(p, r) {
				fmt.Fprintf(ec.Stderr, `invalid argument "%s" for "--region" flag: must be a region in the partition "%s"`, opts.Region, opts.Partition)
				fmt.Fprintln(ec.Stderr)
				fmt.Fprintln(ec.Stderr)
				fmt.Fprintln(ec.Stderr, "Use one of the following regions:")
				for _, r := range types.PartitionRegions(p) {
					fmt.Fprintln(ec.Stderr, r)
				}
				return errors.New("")
			}
			return nil
		},
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

			ec.Spin("Creating cluster")

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

			ec.Logger.Info("Cluster created ✅")
			//
			// r := cluster.NewClusterRenderer(c, jsonData, ec.Stdout)
			// if err := r.Render(ec.OutputFormat); err != nil {
			// 	return fmt.Errorf("rendering output: %w", err)
			// }

			return nil
		},
	}

	f := command.Flags()
	f.StringVar(&opts.Name, "name", "", "Cluster name")
	f.StringVar(&opts.ProviderName, "provider", "", "Provider Name")
	f.StringVar(&opts.EnvironmentName, "environment", "", "Environment Name")
	f.StringVar(&opts.Partition, "partition", "netic", fmt.Sprintf("Partition. One of (%s)", strings.Join(partitions, "|")))
	f.StringVar(&opts.Region, "region", "dk-north", "Region. Depends on the partition.")
	f.StringVar(&opts.SubscriptionID, "subscription", "", "Subscription ID")
	f.StringVar(&opts.InfrastructureProvider, "infrastructure-provider", "netic", "Infrastructure Provider")
	f.StringVar(&opts.ResilienceZone, "resilience-zone", "netic", "Resilience Zone")
	f.StringVar(&opts.APIEndpoint, "api-endpoint", "", "Cluster API Server endpoint (url)")
	f.BoolVar(&opts.HasTechnicalOperations, "has-to", true, "Technical Operations")
	f.BoolVar(&opts.HasTechnicalManagement, "has-tm", true, "Technical Management")
	f.BoolVar(&opts.HasApplicationOperations, "has-ao", false, "Application Operations")
	f.BoolVar(&opts.HasApplicationManagement, "has-am", false, "Application Management")
	f.BoolVar(&opts.HasCustomOperations, "has-co", false, "Custom Operations")
	f.StringVar(&opts.CustomOperationsURL, "co-url", "", "Custom Operations URL")
	command.Flags().SortFlags = false
	command.MarkFlagRequired("name")                    //nolint:errcheck
	command.MarkFlagRequired("provider")                //nolint:errcheck
	command.MarkFlagRequired("environment")             //nolint:errcheck
	command.MarkFlagRequired("partition")               //nolint:errcheck
	command.MarkFlagRequired("region")                  //nolint:errcheck
	command.MarkFlagRequired("subscription")            //nolint:errcheck
	command.MarkFlagRequired("infrastructure-provider") //nolint:errcheck
	command.MarkFlagRequired("resilience-zone")         //nolint:errcheck
	return command
}
