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

// New creates a new "update cluster" command
func NewUpdateClusterCmd(ec *ExecutionContext) *cobra.Command {
	o := updateClusterOptions{}
	c := &cobra.Command{
		Use:     "cluster",
		Short:   "Update a cluster's metadata",
		GroupID: groupCluster,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.complete()
			if err := o.validate(cmd); err != nil {
				return err
			}
			return o.run(ec, args)
		},
	}

	o.bindFlags(c.Flags())
	c.Flags().SortFlags = false
	c.MarkFlagsRequiredTogether("has-co", "co-url")
	c.MarkFlagsOneRequired("description", "environment", "subscription", "infrastructure-provider", "resilience-zone", "has-to", "has-tm", "has-ao", "has-am", "has-co")
	return c
}

type updateClusterOptions struct {
	Description              string
	EnvironmentName          string
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

func (o *updateClusterOptions) bindFlags(f *pflag.FlagSet) {
	f.StringVar(&o.Description, "description", "", "Cluster Description")
	f.StringVar(&o.EnvironmentName, "environment", "", "Environment Name")
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

func (o *updateClusterOptions) validate(cmd *cobra.Command) error {
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
			Flag: "resilience-zone",
			Val:  o.ResilienceZone,
		},
		{
			Flag: "environment",
			Val:  o.EnvironmentName,
		},
	}
	for _, v := range rfc1035FieldFlags {
		if cmd.Flags().Changed(v.Flag) {
			if !validation.IsDNSRFC1035Label(v.Val) {
				return &InvalidArgumentError{
					Flag:    v.Flag,
					Val:     v.Val,
					Context: "must be an RFC1035 DNS label",
				}
			}
		}
	}
	if cmd.Flags().Changed("infrastructure-provider") {
		if !slices.Contains(types.AllInfrastructureProvidersString(), o.InfrastructureProvider) {
			return &InvalidArgumentError{
				Flag:  "infrastructure-provider",
				Val:   o.InfrastructureProvider,
				OneOf: types.AllInfrastructureProvidersString(),
			}
		}
	}
	if cmd.Flags().Changed("resilience-zone") {
		if !slices.Contains(types.AllResilienceZonesString(), o.ResilienceZone) {
			return &InvalidArgumentError{
				Flag:  "resilience-zone",
				Val:   o.ResilienceZone,
				OneOf: types.AllResilienceZonesString(),
			}
		}
	}
	if cmd.Flags().Changed("subscription") {
		if !validation.IsPrintableASCII(o.SubscriptionID) || len(o.SubscriptionID) < 5 {
			return &InvalidArgumentError{
				Flag:    "subscription",
				Val:     o.SubscriptionID,
				Context: "must be an ASCII string of minimum 5 characters length",
			}
		}
	}
	return nil
}

func (o *updateClusterOptions) complete() {
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

func (o *updateClusterOptions) run(ec *ExecutionContext, args []string) error {
	logger := ec.Logger.WithPrefix("Clusters")
	ec.Authenticator.SetLogger(logger)

	_, err := doLogin(ec)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ec.Spin("Updating cluster metadata")

	in := cluster.UpdateClusterInput{
		Logger:    logger,
		APIClient: ec.APIClient,
	}
	if ec.Command.Flags().Changed("description") {
		in.Description = &o.Description
	}
	if ec.Command.Flags().Changed("environment") {
		in.EnvironmentName = &o.EnvironmentName
	}
	if ec.Command.Flags().Changed("resilience-zone") {
		in.ResilienceZone = &o.ResilienceZone
	}
	if ec.Command.Flags().Changed("subscription") {
		in.SubscriptionID = &o.SubscriptionID
	}
	if ec.Command.Flags().Changed("infrastructure-provider") {
		in.InfrastructureProvider = &o.InfrastructureProvider
	}
	if ec.Command.Flags().Changed("has-to") ||
		ec.Command.Flags().Changed("has-tm") ||
		ec.Command.Flags().Changed("has-ao") ||
		ec.Command.Flags().Changed("has-am") ||
		ec.Command.Flags().Changed("has-co") {
		in.HasTechnicalOperations = &o.HasTechnicalOperations
		in.HasTechnicalManagement = &o.HasTechnicalManagement
		in.HasApplicationOperations = &o.HasApplicationOperations
		in.HasApplicationManagement = &o.HasApplicationManagement
		in.HasCustomOperations = &o.HasCustomOperations
	}
	if ec.Command.Flags().Changed("co-url") {
		in.CustomOperationsURL = &o.CustomOperationsURL
	}
	result, err := cluster.UpdateCluster(ec.Command.Context(), args[0], in)
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "updating cluster",
			Problem: result.Problem,
		}
	}

	ec.Spinner.Stop()

	ec.Logger.Info("Cluster metadata updated ✅")

	r := cluster.NewClusterRenderer(result.ClusterResponse, result.JSONResponse, ec.Stdout)
	if err := r.Render(ec.OutputFormat); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	return nil
}
