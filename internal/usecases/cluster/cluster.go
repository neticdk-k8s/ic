package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/apiclient"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/ui"
)

type ListClustersInput struct {
	Logger       logger.Logger
	APIClient    *apiclient.ClientWithResponses
	OutputFormat string
	Spinner      *ui.Spinner
}

func ListClusters(ctx context.Context, in ListClustersInput) error {
	clusters, err := in.APIClient.ListClustersWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("reading clusters: %w", err)
	}
	in.Logger.Debug("apiclient", "status", clusters.StatusCode(), "content-type", clusters.HTTPResponse.Header.Get("Content-Type"))

	in.Spinner.Stop()

	if in.OutputFormat == "json" {
		return prettyPrintJSON(clusters.Body)
	}

	if clusters.StatusCode() == http.StatusOK {
		fmt.Printf("%-20s %-50s %-20s %-10s\n", "PROVIDER", "NAME", "RZ", "VERSION")
		for _, i := range *clusters.ApplicationldJSONDefault.Included {
			rzParts := strings.Split(i["resilienceZone"].(string), "/")
			fmt.Printf("%-20s %-50s %-20s %-10s\n", i["provider"], i["name"], rzParts[len(rzParts)-1], i["kubernetesVersion"].(map[string]interface{})["version"])
		}
	} else {
		return fmt.Errorf("error requesting resource: %w", err)
	}
	return nil
}

type GetClusterInput struct {
	Logger       logger.Logger
	APIClient    *apiclient.ClientWithResponses
	OutputFormat string
	Spinner      *ui.Spinner
}

func GetCluster(ctx context.Context, clusterID string, in GetClusterInput) error {
	cluster, err := in.APIClient.GetClusterWithResponse(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("getting cluster: %w", err)
	}
	in.Logger.Debug("apiclient", "status", cluster.StatusCode(), "content-type", cluster.HTTPResponse.Header.Get("Content-Type"))

	in.Spinner.Stop()

	if in.OutputFormat == "json" {
		return prettyPrintJSON(cluster.Body)
	}

	if cluster.StatusCode() == http.StatusOK {
		rzParts := strings.Split(*cluster.ApplicationldJSONDefault.ResilienceZone, "/")
		fmt.Printf("%-20s : %-s\n", "NAME", *cluster.ApplicationldJSONDefault.Name)
		fmt.Printf("%-20s : %-s\n", "PROVIDER", *cluster.ApplicationldJSONDefault.Provider)
		fmt.Printf("%-20s : %-s\n", "DESCRIPTION", *cluster.ApplicationldJSONDefault.Description)
		fmt.Printf("%-20s : %-s\n", "TYPE", *cluster.ApplicationldJSONDefault.ClusterType)
		fmt.Printf("%-20s : %-s\n", "ENVIRONMENT", *cluster.ApplicationldJSONDefault.EnvironmentName)
		fmt.Printf("%-20s : %-s\n", "RZ", rzParts[len(rzParts)-1])
		fmt.Printf("%-20s : %-s\n", "INFRASTRUCTURE", *cluster.ApplicationldJSONDefault.InfrastructureProvider)
		fmt.Printf("%-20s : %-s\n", "KUBERNETES PROVIDER", *cluster.ApplicationldJSONDefault.KubernetesProvider)
		fmt.Printf("%-20s : %-s\n", "KUBERNETES VERSION", *cluster.ApplicationldJSONDefault.KubernetesVersion.Version)
	} else {
		return fmt.Errorf("error requesting resource: %w", err)
	}
	return nil
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
