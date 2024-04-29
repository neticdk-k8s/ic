package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/neticdk-k8s/ic/internal/apiclient"
	"github.com/neticdk-k8s/ic/internal/logger"
	"github.com/neticdk/go-common/pkg/qsparser"
)

type capacity struct {
	NodeCount   int64 `json:"node_count,omitempty"`
	CoresMillis int64 `json:"cores_millis,omitempty"`
	MemoryBytes int64 `json:"memory_bytes,omitempty"`
}

type clusterResponse struct {
	ID                     string    `json:"id,omitempty"`
	Name                   string    `json:"name,omitempty"`
	ProviderName           string    `json:"provider_name,omitempty"`
	NRN                    string    `json:"nrn,omitempty"`
	Description            string    `json:"description,omitempty"`
	ClusterType            string    `json:"cluster_type,omitempty"`
	Partition              string    `json:"partition,omitempty"`
	Region                 string    `json:"region,omitempty"`
	EnvironmentName        string    `json:"environment_name,omitempty"`
	ResilienceZone         string    `json:"resilience_zone,omitempty"`
	KubernetesProvider     string    `json:"kubernetes_provider,omitempty"`
	InfrastructureProvider string    `json:"infrastructure_provider,omitempty"`
	KubernetesVersion      string    `json:"kubernetes_version,omitempty"`
	SubscriptionName       string    `json:"subscription_name,omitempty"`
	CustomerName           string    `json:"customer_name,omitempty"`
	ClientVersion          string    `json:"client_version,omitempty"`
	ControlPlaneCapacity   *capacity `json:"control_plane_capacity,omitempty"`
	WorkerNodesCapacity    *capacity `json:"worker_nodes_capacity,omitempty"`
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
		cr.ID = fmt.Sprintf("%s.%s", cr.Name, cr.ProviderName)
		if rz, ok := includeMap[i["resilienceZone"].(string)]; ok {
			if p, ok := rz.(map[string]interface{})["name"]; ok {
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

// ListClustersInput is the input given to ListClusters()
type ListClustersInput struct {
	// Logger is a logger
	Logger logger.Logger
	// APIClient is the inventory server API client used to make requests
	APIClient apiclient.ClientWithResponsesInterface
	// Page is the initial page (0-based index)
	Page int
	// PerPage is the number of items requested for each page
	PerPage int
	// Filters is a list of search filters to apply
	Filters map[string]*qsparser.SearchField
}

// ListClusterResults is the result of ListClusters
type ListClusterResults struct {
	ClusterListResponse *clusterListResponse
	JSONResponse        []byte
	Problem             *apiclient.Problem
}

// ListClusters returns a non-paginated list of clusters
func ListClusters(ctx context.Context, in ListClustersInput) (*ListClusterResults, error) {
	cl := &ClusterList{}
	problem, err := listClusters(ctx, &in, cl)
	if err != nil {
		return nil, fmt.Errorf("apiclient: %w", err)
	}
	if problem != nil {
		return &ListClusterResults{nil, nil, problem}, nil
	}
	jsonData, err := cl.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshaling cluster list: %w", err)
	}
	return &ListClusterResults{cl.ToResponse(), jsonData, nil}, nil
}

func listClusters(ctx context.Context, in *ListClustersInput, clusterList *ClusterList) (*apiclient.Problem, error) {
	nextPage := func(ctx context.Context, req *http.Request) error {
		sp := qsparser.SearchParams{
			Page:    &in.Page,
			PerPage: &in.PerPage,
			Fields:  in.Filters,
		}
		sp.SetRawQuery(req)
		return nil
	}
	response, err := in.APIClient.ListClustersWithResponse(ctx, nextPage)
	if err != nil {
		return nil, fmt.Errorf("reading clusters: %w", err)
	}
	in.Logger.Debug("apiclient",
		"status", response.StatusCode(),
		"content-type", response.HTTPResponse.Header.Get("Content-Type"))
	switch response.StatusCode() {
	case http.StatusOK:
	case http.StatusBadRequest:
		return response.ApplicationproblemJSON400, nil
	case http.StatusInternalServerError:
		return response.ApplicationproblemJSON500, nil
	default:
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode())
	}
	if response.ApplicationldJSONDefault.Clusters != nil {
		clusterList.Clusters = append(clusterList.Clusters, *response.ApplicationldJSONDefault.Clusters...)
	}
	if response.ApplicationldJSONDefault.Included != nil {
		clusterList.Included = append(clusterList.Included, *response.ApplicationldJSONDefault.Included...)
	}
	if response.ApplicationldJSONDefault.Pagination.Next != nil {
		in.Page += 1
		return listClusters(ctx, in, clusterList)
	}
	return nil, nil
}

// GetClusterInput is the input used by GetCluster()
type GetClusterInput struct {
	Logger    logger.Logger
	APIClient apiclient.ClientWithResponsesInterface
}

// GetClusterResult is the result of GetCluster
type GetClusterResult struct {
	ClusterResponse *clusterResponse
	JSONResponse    []byte
	Problem         *apiclient.Problem
}

// GetCluster returns information abuot a cluster
func GetCluster(ctx context.Context, clusterID string, in GetClusterInput) (*GetClusterResult, error) {
	response, err := in.APIClient.GetClusterWithResponse(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("apiclient: %w", err)
	}
	in.Logger.Debug("apiclient",
		"status", response.StatusCode(),
		"content-type", response.HTTPResponse.Header.Get("Content-Type"))
	switch response.StatusCode() {
	case http.StatusOK:
	case http.StatusNotFound:
		return &GetClusterResult{nil, nil, response.ApplicationproblemJSON404}, nil
	case http.StatusInternalServerError:
		return &GetClusterResult{nil, nil, response.ApplicationproblemJSON500}, nil
	default:
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode())
	}

	cluster := toClusterResponse(response.ApplicationldJSONDefault)

	jsonData, err := json.Marshal(cluster)
	if err != nil {
		return nil, fmt.Errorf("marshaling cluster: %w", err)
	}

	return &GetClusterResult{cluster, jsonData, nil}, nil
}

// CreateClusterInput is the input used by CreateCluster()
type CreateClusterInput struct {
	Logger                   logger.Logger
	APIClient                apiclient.ClientWithResponsesInterface
	Name                     string
	Description              string
	EnvironmentName          string
	Provider                 string
	Partition                string
	Region                   string
	ResilienceZone           string
	SubscriptionID           string
	InfrastructureProvider   string
	HasTechnicalOperations   bool
	HasTechnicalManagement   bool
	HasApplicationOperations bool
	HasApplicationManagement bool
	HasCustomOperations      bool
	CustomOperationsURL      string
	APIEndpoint              string
}

// CreateClusterResult is the result of CreateCluster
type CreateClusterResult struct {
	ClusterResponse *clusterResponse
	JSONResponse    []byte
	Problem         *apiclient.Problem
}

// CreateCluster creates a cluster
func CreateCluster(ctx context.Context, in CreateClusterInput) (*CreateClusterResult, error) {
	createCluster := apiclient.CreateCluster{
		Name:                     &in.Name,
		Description:              &in.Description,
		EnvironmentName:          &in.EnvironmentName,
		Provider:                 &in.Provider,
		Partition:                &in.Partition,
		Region:                   &in.Region,
		ResilienceZone:           &in.ResilienceZone,
		SubscriptionID:           &in.SubscriptionID,
		InfrastructureProvider:   &in.InfrastructureProvider,
		HasTechnicalOperations:   &in.HasTechnicalOperations,
		HasTechnicalManagement:   &in.HasTechnicalManagement,
		HasApplicationOperations: &in.HasApplicationOperations,
		HasApplicationManagement: &in.HasApplicationManagement,
		HasCustomOperations:      &in.HasCustomOperations,
		CustomOperationsURL:      &in.CustomOperationsURL,
		ApiEndpoint:              &in.APIEndpoint,
	}
	response, err := in.APIClient.CreateClusterWithResponse(ctx, createCluster)
	if err != nil {
		return nil, fmt.Errorf("apiclient: %w", err)
	}
	in.Logger.Debug("apiclient",
		"status", response.StatusCode(),
		"content-type", response.HTTPResponse.Header.Get("Content-Type"))
	switch response.StatusCode() {
	case http.StatusCreated:
	case http.StatusBadRequest:
		return &CreateClusterResult{nil, nil, response.ApplicationproblemJSON400}, nil
	case http.StatusConflict:
		return &CreateClusterResult{nil, nil, response.ApplicationproblemJSON409}, nil
	case http.StatusInternalServerError:
		return &CreateClusterResult{nil, nil, response.ApplicationproblemJSON500}, nil
	default:
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode())
	}

	cluster := toClusterResponse(response.ApplicationldJSON201)

	jsonData, err := json.Marshal(cluster)
	if err != nil {
		return nil, fmt.Errorf("marshaling cluster: %w", err)
	}

	return &CreateClusterResult{cluster, jsonData, nil}, nil
}

// UpdateClusterInput is the input used by UpdateCluster()
type UpdateClusterInput struct {
	Logger                   logger.Logger
	APIClient                apiclient.ClientWithResponsesInterface
	Description              *string
	EnvironmentName          *string
	ResilienceZone           *string
	SubscriptionID           *string
	InfrastructureProvider   *string
	HasTechnicalOperations   *bool
	HasTechnicalManagement   *bool
	HasApplicationOperations *bool
	HasApplicationManagement *bool
	HasCustomOperations      *bool
	CustomOperationsURL      *string
	APIEndpoint              *string
}

// UpdateClusterResult is the result of UpdateCluster
type UpdateClusterResult struct {
	ClusterResponse *clusterResponse
	JSONResponse    []byte
	Problem         *apiclient.Problem
}

// UpdateCluster creates a cluster
func UpdateCluster(ctx context.Context, clusterID string, in UpdateClusterInput) (*UpdateClusterResult, error) {
	updateCluster := apiclient.UpdateCluster{
		Description:              in.Description,
		EnvironmentName:          in.EnvironmentName,
		ResilienceZone:           in.ResilienceZone,
		SubscriptionID:           in.SubscriptionID,
		InfrastructureProvider:   in.InfrastructureProvider,
		HasTechnicalOperations:   in.HasTechnicalOperations,
		HasTechnicalManagement:   in.HasTechnicalManagement,
		HasApplicationOperations: in.HasApplicationOperations,
		HasApplicationManagement: in.HasApplicationManagement,
		HasCustomOperations:      in.HasCustomOperations,
		CustomOperationsURL:      in.CustomOperationsURL,
		ApiEndpoint:              in.APIEndpoint,
	}
	response, err := in.APIClient.UpdateClusterWithResponse(ctx, clusterID, updateCluster)
	if err != nil {
		return nil, fmt.Errorf("apiclient: %w", err)
	}
	in.Logger.Debug("apiclient",
		"status", response.StatusCode(),
		"content-type", response.HTTPResponse.Header.Get("Content-Type"))
	switch response.StatusCode() {
	case http.StatusOK:
	case http.StatusBadRequest:
		return &UpdateClusterResult{nil, nil, response.ApplicationproblemJSON400}, nil
	case http.StatusNotFound:
		return &UpdateClusterResult{nil, nil, response.ApplicationproblemJSON404}, nil
	case http.StatusInternalServerError:
		return &UpdateClusterResult{nil, nil, response.ApplicationproblemJSON500}, nil
	default:
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode())
	}

	cluster := toClusterResponse(response.ApplicationldJSONDefault)

	jsonData, err := json.Marshal(cluster)
	if err != nil {
		return nil, fmt.Errorf("marshaling cluster: %w", err)
	}

	return &UpdateClusterResult{cluster, jsonData, nil}, nil
}

// DeleteClusterInput is the input used by DeleteCluster()
type DeleteClusterInput struct {
	Logger    logger.Logger
	APIClient apiclient.ClientWithResponsesInterface
}

// DeleteClusterResult is the result of DeleteCluster
type DeleteClusterResult struct {
	Problem *apiclient.Problem
}

// DeleteCluster deletes a cluster
func DeleteCluster(ctx context.Context, clusterID string, in DeleteClusterInput) (*DeleteClusterResult, error) {
	response, err := in.APIClient.DeleteClusterWithResponse(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("apiclient: %w", err)
	}
	in.Logger.Debug("apiclient",
		"status", response.StatusCode(),
		"content-type", response.HTTPResponse.Header.Get("Content-Type"))
	switch response.StatusCode() {
	case http.StatusNoContent:
	case http.StatusNotFound:
		return &DeleteClusterResult{response.ApplicationproblemJSON404}, nil
	case http.StatusInternalServerError:
		return &DeleteClusterResult{response.ApplicationproblemJSON500}, nil
	default:
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode())
	}
	return &DeleteClusterResult{nil}, nil
}

func toClusterResponse(cluster *apiclient.Cluster) *clusterResponse {
	includeMap := make(map[string]interface{})
	for _, i := range *cluster.Included {
		includeMap[i["@id"].(string)] = i
	}
	cr := &clusterResponse{}
	cr.Name = nilStr(cluster.Name)
	cr.NRN = nilStr(cluster.Nrn)
	cr.Description = nilStr(cluster.Description)
	cr.Partition = nilStr(cluster.Partition)
	cr.Region = nilStr(cluster.Region)
	cr.EnvironmentName = nilStr(cluster.EnvironmentName)
	cr.InfrastructureProvider = nilStr(cluster.InfrastructureProvider)
	cr.ClusterType = nilStr(cluster.ClusterType)
	cr.KubernetesProvider = nilStr(cluster.KubernetesProvider)
	if cluster.KubernetesVersion != nil {
		cr.KubernetesVersion = *cluster.KubernetesVersion.Version
	}
	if cluster.ClientVersion != nil {
		cr.ClientVersion = *cluster.ClientVersion.Version
	}
	if cluster.Capacity != nil {
		cpct := *cluster.Capacity
		cr.ControlPlaneCapacity = &capacity{
			NodeCount:   *cpct["control-plane"].Nodes,
			CoresMillis: *cpct["control-plane"].Cores,
			MemoryBytes: *cpct["control-plane"].Memory,
		}
		cr.WorkerNodesCapacity = &capacity{
			NodeCount:   *cpct["worker"].Nodes,
			CoresMillis: *cpct["worker"].Cores,
			MemoryBytes: *cpct["worker"].Memory,
		}
	}
	if cluster.Provider != nil {
		if provider, ok := includeMap[*cluster.Provider]; ok {
			if p, ok := provider.(map[string]interface{})["name"]; ok {
				cr.ProviderName = p.(string)
			}
		}
	}
	if cluster.ResilienceZone != nil {
		if provider, ok := includeMap[*cluster.ResilienceZone]; ok {
			if p, ok := provider.(map[string]interface{})["name"]; ok {
				cr.ResilienceZone = p.(string)
			}
		}
	}
	return cr
}

func nilStr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
