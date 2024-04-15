package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/neticdk-k8s/ic/internal/apiclient"
	"github.com/neticdk-k8s/ic/internal/logger"
)

// ListClustersInput is the input given to ListClusters()
type ListClustersInput struct {
	// Logger is a logger
	Logger logger.Logger
	// APIClient is the inventory server API client used to make requests
	APIClient *apiclient.ClientWithResponses
	// Page is the initial page (0-based index)
	Page int
	// PerPage is the number of items requested for each page
	PerPage int
}

type clusterResponse struct {
	Name                   string `json:"name,omitempty"`
	Description            string `json:"description,omitempty"`
	ProviderName           string `json:"provider_name,omitempty"`
	ClusterType            string `json:"cluster_type,omitempty"`
	EnvironmentName        string `json:"environment_name,omitempty"`
	ResilienceZone         string `json:"resilience_zone,omitempty"`
	KubernetesProvider     string `json:"kubernetes_provider,omitempty"`
	InfrastructureProvider string `json:"infrastructure_provider,omitempty"`
	KubernetesVersion      string `json:"kubernetes_version,omitempty"`
	SubscriptionName       string `json:"subscription_name,omitempty"`
	CustomerName           string `json:"customer_name,omitempty"`
	ClientVersion          string `json:"client_version,omitempty"`
}

type clusterListResponse struct {
	Clusters []clusterResponse `json:"clusters,omitempty"`
}

type ClusterList struct {
	Included []map[string]interface{}
	Clusters []string
}

func (cl *ClusterList) ToResponse() *clusterListResponse {
	clr := &clusterListResponse{
		Clusters: make([]clusterResponse, 0),
	}
	includeMap := make(map[string]interface{})
	for _, i := range cl.Included {
		includeMap[i["@id"].(string)] = i
	}
	for _, i := range cl.Included {
		if i["@type"].(string) != "Cluster" {
			continue
		}
		cr := clusterResponse{}
		cr.Name = i["name"].(string)
		cr.ClusterType = i["clusterType"].(string)
		cr.EnvironmentName = i["environmentName"].(string)
		if provider, ok := includeMap[i["provider"].(string)]; ok {
			if p, ok := provider.(map[string]interface{})["name"]; ok {
				cr.ProviderName = p.(string)
			}
		}
		if provider, ok := includeMap[i["resilienceZone"].(string)]; ok {
			if p, ok := provider.(map[string]interface{})["name"]; ok {
				cr.ResilienceZone = p.(string)
			}
		}
		if kubernetesVersion, ok := i["kubernetesVersion"].(map[string]interface{}); ok {
			cr.KubernetesVersion = kubernetesVersion["version"].(string)
		}
		clr.Clusters = append(clr.Clusters, cr)
	}
	return clr
}

func (cl *ClusterList) MarshalJSON() ([]byte, error) {
	return json.Marshal(cl.ToResponse())
}

// ListClusters returns a non-paginated list of clusters
func ListClusters(ctx context.Context, in ListClustersInput) (*clusterListResponse, []byte, error) {
	cl := &ClusterList{}
	err := listClusters(ctx, &in, cl)
	if err != nil {
		return nil, nil, fmt.Errorf("apiclient: %w", err)
	}
	jsonData, err := cl.MarshalJSON()
	if err != nil {
		return nil, nil, fmt.Errorf("marshaling cluster list: %w", err)
	}
	return cl.ToResponse(), jsonData, nil
}

func listClusters(ctx context.Context, in *ListClustersInput, clusterList *ClusterList) error {
	nextPage := func(ctx context.Context, req *http.Request) error {
		q := req.URL.Query()
		q.Add("per_page", fmt.Sprintf("%d", in.PerPage))
		q.Add("page", fmt.Sprintf("%d", in.Page))
		req.URL.RawQuery = q.Encode()
		return nil
	}
	clusters, err := in.APIClient.ListClustersWithResponse(ctx, nextPage)
	if err != nil {
		return fmt.Errorf("reading clusters: %w", err)
	}
	in.Logger.Debug("apiclient",
		"status", clusters.StatusCode(),
		"content-type", clusters.HTTPResponse.Header.Get("Content-Type"))
	if clusters.StatusCode() != http.StatusOK {
		return fmt.Errorf("bad status code: %d", clusters.StatusCode())
	}
	clusterList.Clusters = append(clusterList.Clusters, *clusters.ApplicationldJSONDefault.Clusters...)
	clusterList.Included = append(clusterList.Included, *clusters.ApplicationldJSONDefault.Included...)
	if clusters.ApplicationldJSONDefault.Pagination.Next != nil {
		in.Page += 1
		return listClusters(ctx, in, clusterList)
	}
	return nil
}

// GetClusterInput is the input used by GetCluster()
type GetClusterInput struct {
	Logger    logger.Logger
	APIClient *apiclient.ClientWithResponses
}

// GetCluster returns information abuot a cluster
func GetCluster(ctx context.Context, clusterID string, in GetClusterInput) (*clusterResponse, []byte, error) {
	cluster, err := in.APIClient.GetClusterWithResponse(ctx, clusterID)
	if err != nil {
		return nil, nil, fmt.Errorf("apiclient: %w", err)
	}
	in.Logger.Debug("apiclient", "status", cluster.StatusCode(), "content-type", cluster.HTTPResponse.Header.Get("Content-Type"))

	includeMap := make(map[string]interface{})
	for _, i := range *cluster.ApplicationldJSONDefault.Included {
		includeMap[i["@id"].(string)] = i
	}
	cl := &clusterResponse{
		Name:               *cluster.ApplicationldJSONDefault.Name,
		Description:        *cluster.ApplicationldJSONDefault.Description,
		EnvironmentName:    *cluster.ApplicationldJSONDefault.EnvironmentName,
		ClusterType:        *cluster.ApplicationldJSONDefault.ClusterType,
		KubernetesProvider: *cluster.ApplicationldJSONDefault.KubernetesProvider,
		KubernetesVersion:  *cluster.ApplicationldJSONDefault.KubernetesVersion.Version,
	}
	if provider, ok := includeMap[*cluster.ApplicationldJSONDefault.Provider]; ok {
		if p, ok := provider.(map[string]interface{})["name"]; ok {
			cl.ProviderName = p.(string)
		}
	}
	if provider, ok := includeMap[*cluster.ApplicationldJSONDefault.ResilienceZone]; ok {
		if p, ok := provider.(map[string]interface{})["name"]; ok {
			cl.ResilienceZone = p.(string)
		}
	}

	jsonData, err := json.Marshal(cl)
	if err != nil {
		return nil, nil, fmt.Errorf("marshaling cluster: %w", err)
	}

	if cluster.StatusCode() != http.StatusOK {
		return nil, nil, fmt.Errorf("bad status code: %d", cluster.StatusCode())
	}
	return cl, jsonData, nil
}
