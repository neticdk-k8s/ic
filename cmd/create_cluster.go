package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk-k8s/ic/internal/validation"
	"github.com/neticdk/go-common/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// New creates a new "create cluster" command
func NewCreateClusterCmd(ec *ExecutionContext) *cobra.Command {
	o := createClusterOptions{}
	c := &cobra.Command{
		Use:     "cluster",
		Short:   "Create a cluster",
		GroupID: groupCluster,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.complete()
			if err := o.validate(); err != nil {
				return err
			}
			return o.run(ec)
		},
	}

	o.bindFlags(c.Flags())
	c.Flags().SortFlags = false
	c.MarkFlagRequired("name")            //nolint:errcheck
	c.MarkFlagRequired("provider")        //nolint:errcheck
	c.MarkFlagRequired("environment")     //nolint:errcheck
	c.MarkFlagRequired("subscription")    //nolint:errcheck
	c.MarkFlagRequired("resilience-zone") //nolint:errcheck
	c.MarkFlagsRequiredTogether("has-co", "co-url")
	return c
}

type createClusterOptions struct {
	Name                     string
	Description              string
	ProviderName             string
	EnvironmentName          string
	Partition                string
	Region                   string
	SubscriptionID           string
	InfrastructureProvider   string
	ResilienceZone           string
	HasTechnicalOperations   bool
	HasTechnicalManagement   bool
	HasApplicationOperations bool
	HasApplicationManagement bool
	HasCustomOperations      bool
	CustomOperationsURL      string
}

func (o *createClusterOptions) bindFlags(f *pflag.FlagSet) {
	f.StringVar(&o.Name, "name", "", "Cluster name")
	f.StringVar(&o.ProviderName, "provider", "", "Provider Name")
	f.StringVar(&o.Description, "description", "", "Cluster Description")
	f.StringVar(&o.EnvironmentName, "environment", "", "Environment Name")
	f.StringVar(&o.Partition, "partition", "netic", fmt.Sprintf("Partition. One of (%s)", strings.Join(types.AllPartitionsString(), "|")))
	f.StringVar(&o.Region, "region", "dk-north", "Region. Depends on the partition.")
	f.StringVar(&o.SubscriptionID, "subscription", "", "Subscription ID")
	f.StringVar(&o.InfrastructureProvider, "infrastructure-provider", "netic", fmt.Sprintf("Infrastructure Provider. One of (%s)", strings.Join(types.AllInfrastructureProvidersString(), "|")))
	f.StringVar(&o.ResilienceZone, "resilience-zone", "netic", fmt.Sprintf("Resilience Zone. One of (%s)", strings.Join(types.AllResilienceZonesString(), "|")))
	f.BoolVar(&o.HasTechnicalOperations, "has-to", true, "Technical Operations")
	f.BoolVar(&o.HasTechnicalManagement, "has-tm", true, "Technical Management")
	f.BoolVar(&o.HasApplicationOperations, "has-ao", false, "Application Operations")
	f.BoolVar(&o.HasApplicationManagement, "has-am", false, "Application Management")
	f.BoolVar(&o.HasCustomOperations, "has-co", false, "Custom Operations")
	f.StringVar(&o.CustomOperationsURL, "co-url", "", "Custom Operations URL")
}

func (o *createClusterOptions) validate() error {
	p, ok := types.ParsePartition(o.Partition)
	if !ok {
		return &InvalidArgumentError{
			Flag:  "partition",
			Val:   o.Partition,
			OneOf: types.AllPartitionsString(),
		}
	}
	r, ok := types.ParseRegion(o.Region)
	if !ok {
		return &InvalidArgumentError{
			Flag:     "region",
			Val:      o.Region,
			SeeOther: "get regions",
		}
	}
	if !types.HasRegion(p, r) {
		return &InvalidArgumentError{
			Flag:     "region",
			Val:      o.Region,
			SeeOther: fmt.Sprintf("get regions --partition %s", o.Partition),
		}
	}
	if o.HasCustomOperations && !validation.IsWebURL(o.CustomOperationsURL) {
		return &InvalidArgumentError{
			Flag:    "co-url",
			Val:     o.CustomOperationsURL,
			Context: "must be a URL using a http(s) scheme",
		}
	}
	rfc1035FieldFlags := []struct {
		Flag string
		Val  string
	}{
		{
			Flag: "name",
			Val:  o.Name,
		},
		{
			Flag: "provider",
			Val:  o.ProviderName,
		},
		{
			Flag: "resilience-zone",
			Val:  o.ResilienceZone,
		},
		{
			Flag: "environment",
			Val:  o.EnvironmentName,
		},
	}
	for _, v := range rfc1035FieldFlags {
		if !validation.IsDNSRFC1035Label(v.Val) {
			return &InvalidArgumentError{
				Flag:    v.Flag,
				Val:     v.Val,
				Context: "must be an RFC1035 DNS label",
			}
		}
	}
	if !slices.Contains(types.AllInfrastructureProvidersString(), o.InfrastructureProvider) {
		return &InvalidArgumentError{
			Flag:  "infrastructure-provider",
			Val:   o.InfrastructureProvider,
			OneOf: types.AllInfrastructureProvidersString(),
		}
	}
	if !slices.Contains(types.AllResilienceZonesString(), o.ResilienceZone) {
		return &InvalidArgumentError{
			Flag:  "resilience-zone",
			Val:   o.ResilienceZone,
			OneOf: types.AllResilienceZonesString(),
		}
	}
	if !validation.IsPrintableASCII(o.SubscriptionID) || len(o.SubscriptionID) < 5 {
		return &InvalidArgumentError{
			Flag:    "subscription",
			Val:     o.SubscriptionID,
			Context: "must be an ASCII string of minimum 5 characters length",
		}
	}
	return nil
}

func (o *createClusterOptions) complete() {
	if o.HasCustomOperations {
		o.HasTechnicalOperations = false
		o.HasTechnicalManagement = false
		o.HasApplicationOperations = false
		o.HasApplicationManagement = false
	}
	if o.HasTechnicalManagement {
		o.HasTechnicalOperations = true
	}
	if o.HasApplicationOperations {
		o.HasTechnicalManagement = true
	}
	if o.HasApplicationManagement {
		o.HasApplicationOperations = true
	}
}

func (o *createClusterOptions) run(ec *ExecutionContext) error {
	logger := ec.Logger.WithPrefix("Clusters")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ec.Spin("Creating cluster")

	in := cluster.CreateClusterInput{
		Logger:                   logger,
		APIClient:                ec.APIClient,
		Name:                     o.Name,
		Description:              o.Description,
		EnvironmentName:          o.EnvironmentName,
		Provider:                 o.ProviderName,
		Partition:                o.Partition,
		Region:                   o.Region,
		ResilienceZone:           o.ResilienceZone,
		SubscriptionID:           o.SubscriptionID,
		InfrastructureProvider:   o.InfrastructureProvider,
		HasTechnicalOperations:   o.HasTechnicalOperations,
		HasTechnicalManagement:   o.HasTechnicalManagement,
		HasApplicationOperations: o.HasApplicationOperations,
		HasApplicationManagement: o.HasApplicationManagement,
		HasCustomOperations:      o.HasCustomOperations,
		CustomOperationsURL:      o.CustomOperationsURL,
	}
	result, err := cluster.CreateCluster(ec.Command.Context(), in)
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "creating cluster",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	ec.Logger.Info("Cluster created ✅")

	r := cluster.NewClusterRenderer(result.ClusterResponse, result.JSONResponse, ec.Stdout)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
