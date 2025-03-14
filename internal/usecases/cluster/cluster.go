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

// ClusterList is a list of clusters
type ClusterList struct { //nolint
	// Included is a list of included items
	Included []map[string]any
	// Clusters is a list of clusters
	Clusters []string
}

func (cl *ClusterList) ToResponse() *clusterListResponse {
	clr := &clusterListResponse{
		Clusters: make([]clusterResponse, 0),
	}
	includeMap := make(map[string]any)
	for _, i := range cl.Included {
		includeMap[i["@id"].(string)] = i
		if v, ok := mapValAs[string](i, "@id"); ok {
			includeMap[v] = i
		}
	}
	for _, i := range cl.Included {
		if v, ok := mapValAs[string](i, "@type"); ok {
			if v != "Cluster" {
				continue
			}
		}
		cr := clusterResponse{}
		cr.Name, _ = mapValAs[string](i, "name")
		cr.ClusterType, _ = mapValAs[string](i, "clusterType")
		cr.EnvironmentName, _ = mapValAs[string](i, "environmentName")
		if providerName, ok := mapValAs[string](i, "provider"); ok {
			if provider, ok := includeMap[providerName]; ok {
				if p, ok := provider.(map[string]any); ok {
					cr.ProviderName, _ = mapValAs[string](p, "name")
				}
			}
		}
		cr.ID = fmt.Sprintf("%s.%s", cr.Name, cr.ProviderName)
		if rzName, ok := mapValAs[string](i, "resilienceZone"); ok {
			if rz, ok := includeMap[rzName]; ok {
				if p, ok := rz.(map[string]any); ok {
					cr.ResilienceZone, _ = mapValAs[string](p, "name")
				}
			}
		}
		if kubernetesVersion, ok := i["kubernetesVersion"].(map[string]any); ok {
			cr.KubernetesVersion, _ = mapValAs[string](kubernetesVersion, "version")
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
		return nil, fmt.Errorf("listing clusters: %w", err)
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
	in.Logger.Debug("listClusters", logStatus(response.HTTPResponse))
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
		return nil, fmt.Errorf("getting cluster: %w", err)
	}
	in.Logger.Debug("getCluster", logStatus(response.HTTPResponse))
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
	}
	response, err := in.APIClient.CreateClusterWithResponse(ctx, createCluster)
	if err != nil {
		return nil, fmt.Errorf("creating cluster: %w", err)
	}
	in.Logger.Debug("createCluster", logStatus(response.HTTPResponse))
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
	}
	response, err := in.APIClient.UpdateClusterWithResponse(ctx, clusterID, updateCluster)
	if err != nil {
		return nil, fmt.Errorf("updating cluster: %w", err)
	}
	in.Logger.Debug("updateCluster", logStatus(response.HTTPResponse))
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
		return nil, fmt.Errorf("deleting cluster: %w", err)
	}
	in.Logger.Debug("deleteCluster", logStatus(response.HTTPResponse))
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

type clusterNodeResponse struct {
	Name                    string  `json:"name,omitempty"`
	Role                    string  `json:"role,omitempty"`
	KubeProxyVersion        string  `json:"kube_proxy_version,omitempty"`
	KubeletVersion          string  `json:"kubelet_version,omitempty"`
	KernelVersion           string  `json:"kernel_version,omitempty"`
	CRIName                 string  `json:"cri_name,omitempty"`
	CRIVersion              string  `json:"cri_version,omitempty"`
	ContainerRuntimeVersion string  `json:"container_runtime_version,omitempty"`
	IsControlPlane          bool    `json:"is_control_plane"`
	Provider                string  `json:"provider,omitempty"`
	TopologyRegion          string  `json:"topology_region,omitempty"`
	TopologyZone            string  `json:"topology_zone,omitempty"`
	AllocatableCPUMillis    float64 `json:"allocatable_cpu_millis,omitempty"`
	AllocatableMemoryBytes  float64 `json:"allocatable_memory_bytes,omitempty"`
	CapacityCPUMillis       float64 `json:"capacity_cpu_millis,omitempty"`
	CapacityMemoryBytes     float64 `json:"capacity_memory_bytes,omitempty"`
}

type clusterNodesListResponse struct {
	Nodes []clusterNodeResponse `json:"nodes,omitempty"`
}

type ClusterNodesList struct { //nolint
	Included []map[string]any
	Nodes    []string
}

func (cl *ClusterNodesList) ToResponse() *clusterNodesListResponse {
	cnlr := &clusterNodesListResponse{
		Nodes: make([]clusterNodeResponse, 0),
	}
	includeMap := make(map[string]any)
	for _, i := range cl.Included {
		if v, ok := mapValAs[string](i, "@id"); ok {
			includeMap[v] = i
		}
	}
	for _, i := range cl.Included {
		if v, ok := mapValAs[string](i, "@type"); ok {
			if v != "Node" {
				continue
			}
		}
		cr := clusterNodeResponse{}
		cr.Name, _ = mapValAs[string](i, "name")
		cr.IsControlPlane, _ = mapValAs[bool](i, "isControlPlane")
		cr.KubeletVersion, _ = mapValAs[string](i, "kubeletVersion")
		cr.AllocatableCPUMillis, _ = mapValAs[float64](i, "allocatableCoresMillis")
		cr.AllocatableMemoryBytes, _ = mapValAs[float64](i, "allocatableMemoryBytes")
		cr.CapacityCPUMillis, _ = mapValAs[float64](i, "capacityCoresMillis")
		cr.CapacityMemoryBytes, _ = mapValAs[float64](i, "capacityMemoryBytes")

		cnlr.Nodes = append(cnlr.Nodes, cr)
	}
	return cnlr
}

func (cl *ClusterNodesList) MarshalJSON() ([]byte, error) {
	return json.Marshal(cl.ToResponse())
}

// ListClusterNodesInput is the input given to ListClusterNodes()
type ListClusterNodesInput struct {
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
	// ClusterName is the name of the cluster
	ClusterName string
}

// ListClusterNodesResults is the result of ListClusterNodes()
type ListClusterNodesResults struct {
	ClusterNodeListResponse *clusterNodesListResponse
	JSONResponse            []byte
	Problem                 *apiclient.Problem
}

// ListClusterNodes returns a non-paginated list of cluster nodes
func ListClusterNodes(ctx context.Context, in ListClusterNodesInput) (*ListClusterNodesResults, error) {
	nl := &ClusterNodesList{}
	problem, err := listClusterNodes(ctx, &in, nl)
	if err != nil {
		return nil, fmt.Errorf("listing cluster nodes: %w", err)
	}
	if problem != nil {
		return &ListClusterNodesResults{nil, nil, problem}, nil
	}
	jsonData, err := nl.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshaling cluster list: %w", err)
	}
	return &ListClusterNodesResults{nl.ToResponse(), jsonData, nil}, nil
}

func listClusterNodes(ctx context.Context, in *ListClusterNodesInput, nodeList *ClusterNodesList) (*apiclient.Problem, error) {
	nextPage := func(ctx context.Context, req *http.Request) error {
		sp := qsparser.SearchParams{
			Page:    &in.Page,
			PerPage: &in.PerPage,
			Fields:  in.Filters,
		}
		sp.SetRawQuery(req)
		return nil
	}
	response, err := in.APIClient.ListNodesWithResponse(ctx, in.ClusterName, nextPage)
	if err != nil {
		return nil, fmt.Errorf("reading cluster node list: %w", err)
	}
	in.Logger.Debug("listNodes", logStatus(response.HTTPResponse))
	switch response.StatusCode() {
	case http.StatusOK:
	case http.StatusBadRequest:
		return response.ApplicationproblemJSON400, nil
	case http.StatusInternalServerError:
		return response.ApplicationproblemJSON500, nil
	default:
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode())
	}
	if response.ApplicationldJSONDefault.Nodes != nil {
		nodeList.Nodes = append(nodeList.Nodes, *response.ApplicationldJSONDefault.Nodes...)
	}
	if response.ApplicationldJSONDefault.Included != nil {
		nodeList.Included = append(nodeList.Included, *response.ApplicationldJSONDefault.Included...)
	}
	if response.ApplicationldJSONDefault.Pagination.Next != nil {
		in.Page += 1
		return listClusterNodes(ctx, in, nodeList)
	}
	return nil, nil
}

// GetClusterNodeInput is the input used by GetClusterNode()
type GetClusterNodeInput struct {
	Logger      logger.Logger
	APIClient   apiclient.ClientWithResponsesInterface
	ClusterName string
	NodeName    string
}

// GetClusterNodeResult is the result of GetClusterNode()
type GetClusterNodeResult struct {
	ClusterNodeResponse *clusterNodeResponse
	JSONResponse        []byte
	Problem             *apiclient.Problem
}

// GetClusterNode returns information abuot a cluster node
func GetClusterNode(ctx context.Context, in GetClusterNodeInput) (*GetClusterNodeResult, error) {
	response, err := in.APIClient.GetNodeWithResponse(ctx, in.ClusterName, in.NodeName)
	if err != nil {
		return nil, fmt.Errorf("getting node: %w", err)
	}
	in.Logger.Debug("getNode", logStatus(response.HTTPResponse))
	switch response.StatusCode() {
	case http.StatusOK:
	case http.StatusNotFound:
		return &GetClusterNodeResult{nil, nil, response.ApplicationproblemJSON404}, nil
	case http.StatusInternalServerError:
		return &GetClusterNodeResult{nil, nil, response.ApplicationproblemJSON500}, nil
	default:
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode())
	}

	node := toClusterNodeResponse(response.ApplicationldJSONDefault)

	jsonData, err := json.Marshal(node)
	if err != nil {
		return nil, fmt.Errorf("marshaling cluste node: %w", err)
	}

	return &GetClusterNodeResult{node, jsonData, nil}, nil
}

// GetClusterKubeConfigInput is the input used by GetClusterKubeConfig()
type GetClusterKubeConfigInput struct {
	Logger      logger.Logger
	APIClient   apiclient.ClientWithResponsesInterface
	ClusterName string
}

// GetClusterKubeConfigResult is the result of GetClusterKubeConfig()
type GetClusterKubeConfigResult struct {
	Response []byte
	Problem  *apiclient.Problem
}

// GetClusterKubeConfig returns kubeconfig for a cluster
func GetClusterKubeConfig(ctx context.Context, in GetClusterKubeConfigInput) (*GetClusterKubeConfigResult, error) {
	response, err := in.APIClient.GetClusterKubeConfigWithResponse(ctx, in.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("getting cluster kubeconfig: %w", err)
	}
	in.Logger.Debug("getClusterKubeconfig", logStatus(response.HTTPResponse))
	switch response.StatusCode() {
	case http.StatusOK:
	case http.StatusNotFound:
		return &GetClusterKubeConfigResult{nil, response.ApplicationproblemJSON404}, nil
	case http.StatusInternalServerError:
		return &GetClusterKubeConfigResult{nil, response.ApplicationproblemJSON500}, nil
	default:
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode())
	}

	return &GetClusterKubeConfigResult{response.Body, nil}, nil
}

func toClusterResponse(cluster *apiclient.Cluster) *clusterResponse {
	includeMap := make(map[string]any)
	for _, i := range *cluster.Included {
		if v, ok := mapValAs[string](i, "@id"); ok {
			includeMap[v] = i
		}
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
			CoresMillis: *cpct["control-plane"].CoresMillis,
			MemoryBytes: *cpct["control-plane"].MemoryBytes,
		}
		cr.WorkerNodesCapacity = &capacity{
			NodeCount:   *cpct["worker"].Nodes,
			CoresMillis: *cpct["worker"].CoresMillis,
			MemoryBytes: *cpct["worker"].MemoryBytes,
		}
	}
	if cluster.Provider != nil {
		if provider, ok := includeMap[*cluster.Provider]; ok {
			if p, ok := provider.(map[string]any); ok {
				cr.ProviderName, _ = mapValAs[string](p, "name")
			}
		}
	}
	if cluster.ResilienceZone != nil {
		if provider, ok := includeMap[*cluster.ResilienceZone]; ok {
			if p, ok := provider.(map[string]any); ok {
				cr.ResilienceZone, _ = mapValAs[string](p, "name")
			}
		}
	}
	return cr
}

func toClusterNodeResponse(node *apiclient.Node) *clusterNodeResponse {
	cn := &clusterNodeResponse{}
	cn.Name = nilStr(node.Name)
	cn.Role = nilStr(node.Role)
	cn.KubeProxyVersion = nilStr(node.KubeProxyVersion)
	cn.KubeletVersion = nilStr(node.KubeletVersion)
	cn.KernelVersion = nilStr(node.KernelVersion)
	cn.CRIName = nilStr(node.CriName)
	cn.CRIVersion = nilStr(node.CriVersion)
	cn.ContainerRuntimeVersion = nilStr(node.ContainerRuntimeVersion)
	cn.IsControlPlane = nilBool(node.IsControlPlane)
	cn.Provider = nilStr(node.Provider)
	cn.TopologyRegion = nilStr(node.TopologyRegion)
	cn.TopologyZone = nilStr(node.TopologyZone)
	cn.AllocatableCPUMillis = float64(nilInt64(node.AllocatableCoresMillis))
	cn.AllocatableMemoryBytes = float64(nilInt64(node.AllocatableMemoryBytes))
	cn.CapacityCPUMillis = float64(nilInt64(node.CapacityCoresMillis))
	cn.CapacityMemoryBytes = float64(nilInt64(node.CapacityMemoryBytes))

	return cn
}

func nilStr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func nilBool(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}

func nilInt64(i *int64) int64 {
	if i != nil {
		return *i
	}
	return 0
}

func logStatus(r *http.Response) []any {
	return []any{"status", r.StatusCode, "content-type", r.Header.Get("Content-Type")}
}

func mapValAs[T any](haystak map[string]any, needle string) (T, bool) {
	if v, ok := haystak[needle]; ok {
		if v2, ok := v.(T); ok {
			return v2, true
		}
	}
	var zero T
	return zero, false
}
