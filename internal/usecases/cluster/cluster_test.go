package cluster

import (
	"context"
	"net/http"
	"testing"

	"github.com/neticdk-k8s/ic/internal/apiclient"
	"github.com/neticdk-k8s/ic/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClusterList_ToResponse(t *testing.T) {
	cl := ClusterList{
		Clusters: make([]string, 0),
		Included: make([]map[string]interface{}, 0),
	}
	t.Run("Valid input", func(t *testing.T) {
		cl.Clusters = []string{"my-cluster"}
		cl.Included = []map[string]interface{}{
			{
				"@id":   "my-provider-id",
				"@type": "Provider",
				"name":  "my-provider",
			},
			{
				"@id":   "my-resilience-zone-id",
				"@type": "ResilienceZone",
				"name":  "my-resilience-zone",
			},
			{
				"@id":             "my-cluster-id",
				"@type":           "Cluster",
				"name":            "my-cluster",
				"clusterType":     "my-cluster-type",
				"environmentName": "my-environment",
				"provider":        "my-provider-id",
				"resilienceZone":  "my-resilience-zone-id",
				"kubernetesVersion": map[string]interface{}{
					"version": "1.2.3",
				},
			},
		}
		want := &clusterListResponse{
			Clusters: []clusterResponse{
				{
					Name:              "my-cluster",
					ID:                "my-cluster.my-provider",
					ClusterType:       "my-cluster-type",
					EnvironmentName:   "my-environment",
					ProviderName:      "my-provider",
					ResilienceZone:    "my-resilience-zone",
					KubernetesVersion: "1.2.3",
				},
			},
		}
		got := cl.ToResponse()
		assert.Equal(t, want, got)
	})

	t.Run("Empty input", func(t *testing.T) {
		cl.Clusters = []string{""}
		cl.Included = []map[string]interface{}{}
		want := &clusterListResponse{[]clusterResponse{}}
		got := cl.ToResponse()
		assert.Equal(t, want, got)
	})
}

func TestClusterList_MarshalJSON(t *testing.T) {
	cl := ClusterList{
		Clusters: make([]string, 0),
		Included: make([]map[string]interface{}, 0),
	}
	cl.Clusters = []string{"my-cluster"}
	cl.Included = []map[string]interface{}{
		{
			"@id":   "my-provider-id",
			"@type": "Provider",
			"name":  "my-provider",
		},
		{
			"@id":   "my-resilience-zone-id",
			"@type": "ResilienceZone",
			"name":  "my-resilience-zone",
		},
		{
			"@id":             "my-cluster-id",
			"@type":           "Cluster",
			"name":            "my-cluster",
			"clusterType":     "my-cluster-type",
			"environmentName": "my-environment",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-resilience-zone-id",
			"kubernetesVersion": map[string]interface{}{
				"version": "1.2.3",
			},
		},
	}
	want := []byte(`{"clusters":[{"id":"my-cluster.my-provider","name":"my-cluster","provider_name":"my-provider","cluster_type":"my-cluster-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"}]}`)
	got, err := cl.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestListClusters(t *testing.T) {
	logger := logger.NewTestLogger(t)

	clusters := []string{"my-cluster", "my-cluster-2"}
	included := []map[string]interface{}{
		{
			"@id":   "my-provider-id",
			"@type": "Provider",
			"name":  "my-provider",
		},
		{
			"@id":   "my-resilience-zone-id",
			"@type": "ResilienceZone",
			"name":  "my-resilience-zone",
		},
		{
			"@id":             "my-cluster-id",
			"@type":           "Cluster",
			"name":            "my-cluster",
			"clusterType":     "my-cluster-type",
			"environmentName": "my-environment",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-resilience-zone-id",
			"kubernetesVersion": map[string]interface{}{
				"version": "1.2.3",
			},
		},
		{
			"@id":             "my-cluster-id-2",
			"@type":           "Cluster",
			"name":            "my-cluster-2",
			"clusterType":     "my-cluster-type",
			"environmentName": "my-environment",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-resilience-zone-id",
			"kubernetesVersion": map[string]interface{}{
				"version": "1.2.3",
			},
		},
	}

	mockClient := apiclient.NewMockClientWithResponsesInterface(t)
	mockClient.EXPECT().
		ListClustersWithResponse(mock.Anything, mock.Anything).
		Return(
			&apiclient.ListClustersResponse{
				Body: make([]byte, 0),
				HTTPResponse: &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
				},
				ApplicationldJSONDefault: &apiclient.Clusters{
					Clusters:   &clusters,
					Included:   &included,
					Pagination: &apiclient.Pagination{},
				},
			}, nil)
	in := ListClustersInput{
		Logger:    logger,
		APIClient: mockClient,
	}

	want := &clusterListResponse{
		Clusters: []clusterResponse{
			{
				Name:              "my-cluster",
				ID:                "my-cluster.my-provider",
				ClusterType:       "my-cluster-type",
				EnvironmentName:   "my-environment",
				ProviderName:      "my-provider",
				ResilienceZone:    "my-resilience-zone",
				KubernetesVersion: "1.2.3",
			},
			{
				Name:              "my-cluster-2",
				ID:                "my-cluster-2.my-provider",
				ClusterType:       "my-cluster-type",
				EnvironmentName:   "my-environment",
				ProviderName:      "my-provider",
				ResilienceZone:    "my-resilience-zone",
				KubernetesVersion: "1.2.3",
			},
		},
	}

	got, err := ListClusters(context.TODO(), in)
	assert.NoError(t, err)
	assert.Equal(t, want, got.ClusterListResponse)

	wantJSON := []byte(`{"clusters":[{"id":"my-cluster.my-provider","name":"my-cluster","provider_name":"my-provider","cluster_type":"my-cluster-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"},{"id":"my-cluster-2.my-provider","name":"my-cluster-2","provider_name":"my-provider","cluster_type":"my-cluster-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"}]}`)
	assert.Equal(t, wantJSON, got.JSONResponse)
}

func TestGetCluster(t *testing.T) {
	logger := logger.NewTestLogger(t)
	mockClient := apiclient.NewMockClientWithResponsesInterface(t)
	name := "my-cluster"
	included := []map[string]interface{}{
		{
			"@id":   "my-provider-id",
			"@type": "Provider",
			"name":  "my-provider",
		},
		{
			"@id":   "my-resilience-zone-id",
			"@type": "ResilienceZone",
			"name":  "my-resilience-zone",
		},
		{
			"@id":             "my-cluster-id",
			"@type":           "Cluster",
			"name":            "my-cluster",
			"clusterType":     "my-cluster-type",
			"environmentName": "my-environment",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-resilience-zone-id",
			"kubernetesVersion": map[string]interface{}{
				"version": "1.2.3",
			},
		},
	}
	providerId := "my-provider-id"
	mockClient.EXPECT().
		GetClusterWithResponse(mock.Anything, mock.Anything).
		Return(
			&apiclient.GetClusterResponse{
				Body: make([]byte, 0),
				HTTPResponse: &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
				},
				ApplicationldJSONDefault: &apiclient.Cluster{
					Name:     &name,
					Provider: &providerId,
					Included: &included,
				},
			}, nil)

	want := &clusterResponse{
		Name:         "my-cluster",
		ProviderName: "my-provider",
	}

	in := GetClusterInput{
		Logger:    logger,
		APIClient: mockClient,
	}
	got, err := GetCluster(context.TODO(), "my-cluster.my-provider", in)
	assert.NoError(t, err)
	assert.Equal(t, want, got.ClusterResponse)

	wantJSON := []byte(`{"name":"my-cluster","provider_name":"my-provider"}`)
	assert.Equal(t, wantJSON, got.JSONResponse)
}

func TestNodeList_ToResponse(t *testing.T) {
	cl := ClusterList{
		Clusters: make([]string, 0),
		Included: make([]map[string]interface{}, 0),
	}
	t.Run("Valid input", func(t *testing.T) {
		cl.Clusters = []string{"my-cluster"}
		cl.Included = []map[string]interface{}{
			{
				"@id":   "my-provider-id",
				"@type": "Provider",
				"name":  "my-provider",
			},
			{
				"@id":   "my-resilience-zone-id",
				"@type": "ResilienceZone",
				"name":  "my-resilience-zone",
			},
			{
				"@id":             "my-cluster-id",
				"@type":           "Cluster",
				"name":            "my-cluster",
				"clusterType":     "my-cluster-type",
				"environmentName": "my-environment",
				"provider":        "my-provider-id",
				"resilienceZone":  "my-resilience-zone-id",
				"kubernetesVersion": map[string]interface{}{
					"version": "1.2.3",
				},
			},
		}
		want := &clusterListResponse{
			Clusters: []clusterResponse{
				{
					Name:              "my-cluster",
					ID:                "my-cluster.my-provider",
					ClusterType:       "my-cluster-type",
					EnvironmentName:   "my-environment",
					ProviderName:      "my-provider",
					ResilienceZone:    "my-resilience-zone",
					KubernetesVersion: "1.2.3",
				},
			},
		}
		got := cl.ToResponse()
		assert.Equal(t, want, got)
	})

	t.Run("Empty input", func(t *testing.T) {
		cl.Clusters = []string{""}
		cl.Included = []map[string]interface{}{}
		want := &clusterListResponse{[]clusterResponse{}}
		got := cl.ToResponse()
		assert.Equal(t, want, got)
	})
}

func TestNodeList_MarshalJSON(t *testing.T) {
	cl := ClusterList{
		Clusters: make([]string, 0),
		Included: make([]map[string]interface{}, 0),
	}
	cl.Clusters = []string{"my-cluster"}
	cl.Included = []map[string]interface{}{
		{
			"@id":   "my-provider-id",
			"@type": "Provider",
			"name":  "my-provider",
		},
		{
			"@id":   "my-resilience-zone-id",
			"@type": "ResilienceZone",
			"name":  "my-resilience-zone",
		},
		{
			"@id":             "my-cluster-id",
			"@type":           "Cluster",
			"name":            "my-cluster",
			"clusterType":     "my-cluster-type",
			"environmentName": "my-environment",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-resilience-zone-id",
			"kubernetesVersion": map[string]interface{}{
				"version": "1.2.3",
			},
		},
	}
	want := []byte(`{"clusters":[{"id":"my-cluster.my-provider","name":"my-cluster","provider_name":"my-provider","cluster_type":"my-cluster-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"}]}`)
	got, err := cl.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestListClusterNodes(t *testing.T) {
	logger := logger.NewTestLogger(t)

	clusters := []string{"my-cluster", "my-cluster-2"}
	included := []map[string]interface{}{
		{
			"@id":   "my-provider-id",
			"@type": "Provider",
			"name":  "my-provider",
		},
		{
			"@id":   "my-resilience-zone-id",
			"@type": "ResilienceZone",
			"name":  "my-resilience-zone",
		},
		{
			"@id":             "my-cluster-id",
			"@type":           "Cluster",
			"name":            "my-cluster",
			"clusterType":     "my-cluster-type",
			"environmentName": "my-environment",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-resilience-zone-id",
			"kubernetesVersion": map[string]interface{}{
				"version": "1.2.3",
			},
		},
		{
			"@id":             "my-cluster-id-2",
			"@type":           "Cluster",
			"name":            "my-cluster-2",
			"clusterType":     "my-cluster-type",
			"environmentName": "my-environment",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-resilience-zone-id",
			"kubernetesVersion": map[string]interface{}{
				"version": "1.2.3",
			},
		},
	}

	mockClient := apiclient.NewMockClientWithResponsesInterface(t)
	mockClient.EXPECT().
		ListClustersWithResponse(mock.Anything, mock.Anything).
		Return(
			&apiclient.ListClustersResponse{
				Body: make([]byte, 0),
				HTTPResponse: &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
				},
				ApplicationldJSONDefault: &apiclient.Clusters{
					Clusters:   &clusters,
					Included:   &included,
					Pagination: &apiclient.Pagination{},
				},
			}, nil)
	in := ListClustersInput{
		Logger:    logger,
		APIClient: mockClient,
	}

	want := &clusterListResponse{
		Clusters: []clusterResponse{
			{
				Name:              "my-cluster",
				ID:                "my-cluster.my-provider",
				ClusterType:       "my-cluster-type",
				EnvironmentName:   "my-environment",
				ProviderName:      "my-provider",
				ResilienceZone:    "my-resilience-zone",
				KubernetesVersion: "1.2.3",
			},
			{
				Name:              "my-cluster-2",
				ID:                "my-cluster-2.my-provider",
				ClusterType:       "my-cluster-type",
				EnvironmentName:   "my-environment",
				ProviderName:      "my-provider",
				ResilienceZone:    "my-resilience-zone",
				KubernetesVersion: "1.2.3",
			},
		},
	}

	got, err := ListClusters(context.TODO(), in)
	assert.NoError(t, err)
	assert.Equal(t, want, got.ClusterListResponse)

	wantJSON := []byte(`{"clusters":[{"id":"my-cluster.my-provider","name":"my-cluster","provider_name":"my-provider","cluster_type":"my-cluster-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"},{"id":"my-cluster-2.my-provider","name":"my-cluster-2","provider_name":"my-provider","cluster_type":"my-cluster-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"}]}`)
	assert.Equal(t, wantJSON, got.JSONResponse)
}

func TestGetClusterNode(t *testing.T) {
	logger := logger.NewTestLogger(t)
	mockClient := apiclient.NewMockClientWithResponsesInterface(t)
	name := "my-cluster"
	included := []map[string]interface{}{
		{
			"@id":   "my-provider-id",
			"@type": "Provider",
			"name":  "my-provider",
		},
		{
			"@id":   "my-resilience-zone-id",
			"@type": "ResilienceZone",
			"name":  "my-resilience-zone",
		},
		{
			"@id":             "my-cluster-id",
			"@type":           "Cluster",
			"name":            "my-cluster",
			"clusterType":     "my-cluster-type",
			"environmentName": "my-environment",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-resilience-zone-id",
			"kubernetesVersion": map[string]interface{}{
				"version": "1.2.3",
			},
		},
	}
	providerId := "my-provider-id"
	mockClient.EXPECT().
		GetClusterWithResponse(mock.Anything, mock.Anything).
		Return(
			&apiclient.GetClusterResponse{
				Body: make([]byte, 0),
				HTTPResponse: &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
				},
				ApplicationldJSONDefault: &apiclient.Cluster{
					Name:     &name,
					Provider: &providerId,
					Included: &included,
				},
			}, nil)

	want := &clusterResponse{
		Name:         "my-cluster",
		ProviderName: "my-provider",
	}

	in := GetClusterInput{
		Logger:    logger,
		APIClient: mockClient,
	}
	got, err := GetCluster(context.TODO(), "my-cluster.my-provider", in)
	assert.NoError(t, err)
	assert.Equal(t, want, got.ClusterResponse)

	wantJSON := []byte(`{"name":"my-cluster","provider_name":"my-provider"}`)
	assert.Equal(t, wantJSON, got.JSONResponse)
}

func TestGetClusterKubeConfig(t *testing.T) {
	logger := logger.NewTestLogger(t)

	want := []byte(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: test
    server: https://test.test.dedicated.k8s.netic.dk:6443
  name: test
contexts:
- context:
    cluster: test
    user: test
  name: test
current-context: test
kind: Config
preferences: {}
users:
- name: test
  user:
    password: REDACTED
    username: test
`)

	mockClient := apiclient.NewMockClientWithResponsesInterface(t)
	mockClient.EXPECT().
		GetClusterKubeConfigWithResponse(mock.Anything, mock.Anything).
		Return(
			&apiclient.GetClusterKubeConfigResponse{
				Body: want,
				HTTPResponse: &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
				},
			}, nil)

	in := GetClusterKubeConfigInput{
		Logger:      logger,
		APIClient:   mockClient,
		ClusterName: "my-cluster.my-provider",
	}
	got, err := GetClusterKubeConfig(context.TODO(), in)
	assert.NoError(t, err)
	assert.Equal(t, want, got.Response)
}
