package component

// import (
// 	"context"
// 	"net/http"
// 	"testing"

// 	"github.com/neticdk-k8s/ic/internal/apiclient"
// 	"github.com/neticdk-k8s/ic/internal/logger"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// func TestComponentList_ToResponse(t *testing.T) {
// 	cl := ComponentList{
// 		Components: make([]string, 0),
// 		Included:   make([]map[string]interface{}, 0),
// 	}
// 	t.Run("Valid input", func(t *testing.T) {
// 		cl.Components = []string{"my-component"}
// 		cl.Included = []map[string]interface{}{
// 			{
// 				"@id":   "my-provider-id",
// 				"@type": "Provider",
// 				"name":  "my-provider",
// 			},
// 			{
// 				"@id":   "my-resilience-zone-id",
// 				"@type": "ResilienceZone",
// 				"name":  "my-resilience-zone",
// 			},
// 			{
// 				"@id":             "my-component-id",
// 				"@type":           "Component",
// 				"name":            "my-component",
// 				"componentType":   "my-component-type",
// 				"environmentName": "my-environment",
// 				"provider":        "my-provider-id",
// 				"resilienceZone":  "my-resilience-zone-id",
// 				"kubernetesVersion": map[string]interface{}{
// 					"version": "1.2.3",
// 				},
// 			},
// 		}
// 		want := &componentListResponse{
// 			Components: []componentResponse{
// 				{
// 					Name:              "my-component",
// 					ID:                "my-component.my-provider",
// 					ComponentType:     "my-component-type",
// 					EnvironmentName:   "my-environment",
// 					ProviderName:      "my-provider",
// 					ResilienceZone:    "my-resilience-zone",
// 					KubernetesVersion: "1.2.3",
// 				},
// 			},
// 		}
// 		got := cl.ToResponse()
// 		assert.Equal(t, want, got)
// 	})

// 	t.Run("Empty input", func(t *testing.T) {
// 		cl.Components = []string{""}
// 		cl.Included = []map[string]interface{}{}
// 		want := &componentListResponse{[]componentResponse{}}
// 		got := cl.ToResponse()
// 		assert.Equal(t, want, got)
// 	})
// }

// func TestComponentList_MarshalJSON(t *testing.T) {
// 	cl := ComponentList{
// 		Components: make([]string, 0),
// 		Included:   make([]map[string]interface{}, 0),
// 	}
// 	cl.Components = []string{"my-component"}
// 	cl.Included = []map[string]interface{}{
// 		{
// 			"@id":   "my-provider-id",
// 			"@type": "Provider",
// 			"name":  "my-provider",
// 		},
// 		{
// 			"@id":   "my-resilience-zone-id",
// 			"@type": "ResilienceZone",
// 			"name":  "my-resilience-zone",
// 		},
// 		{
// 			"@id":             "my-component-id",
// 			"@type":           "Component",
// 			"name":            "my-component",
// 			"componentType":   "my-component-type",
// 			"environmentName": "my-environment",
// 			"provider":        "my-provider-id",
// 			"resilienceZone":  "my-resilience-zone-id",
// 			"kubernetesVersion": map[string]interface{}{
// 				"version": "1.2.3",
// 			},
// 		},
// 	}
// 	want := []byte(`{"components":[{"id":"my-component.my-provider","name":"my-component","provider_name":"my-provider","component_type":"my-component-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"}]}`)
// 	got, err := cl.MarshalJSON()
// 	assert.NoError(t, err)
// 	assert.Equal(t, want, got)
// }

// func TestListComponents(t *testing.T) {
// 	logger := logger.NewTestLogger(t)

// 	components := []string{"my-component", "my-component-2"}
// 	included := []map[string]interface{}{
// 		{
// 			"@id":   "my-provider-id",
// 			"@type": "Provider",
// 			"name":  "my-provider",
// 		},
// 		{
// 			"@id":   "my-resilience-zone-id",
// 			"@type": "ResilienceZone",
// 			"name":  "my-resilience-zone",
// 		},
// 		{
// 			"@id":             "my-component-id",
// 			"@type":           "Component",
// 			"name":            "my-component",
// 			"componentType":   "my-component-type",
// 			"environmentName": "my-environment",
// 			"provider":        "my-provider-id",
// 			"resilienceZone":  "my-resilience-zone-id",
// 			"kubernetesVersion": map[string]interface{}{
// 				"version": "1.2.3",
// 			},
// 		},
// 		{
// 			"@id":             "my-component-id-2",
// 			"@type":           "Component",
// 			"name":            "my-component-2",
// 			"componentType":   "my-component-type",
// 			"environmentName": "my-environment",
// 			"provider":        "my-provider-id",
// 			"resilienceZone":  "my-resilience-zone-id",
// 			"kubernetesVersion": map[string]interface{}{
// 				"version": "1.2.3",
// 			},
// 		},
// 	}

// 	mockClient := apiclient.NewMockClientWithResponsesInterface(t)
// 	mockClient.EXPECT().
// 		ListComponentsWithResponse(mock.Anything, mock.Anything).
// 		Return(
// 			&apiclient.ListComponentsResponse{
// 				Body: make([]byte, 0),
// 				HTTPResponse: &http.Response{
// 					Status:     "200 OK",
// 					StatusCode: 200,
// 				},
// 				ApplicationldJSONDefault: &apiclient.Components{
// 					Components: &components,
// 					Included:   &included,
// 					Pagination: &apiclient.Pagination{},
// 				},
// 			}, nil)
// 	in := ListComponentsInput{
// 		Logger:    logger,
// 		APIClient: mockClient,
// 	}

// 	want := &componentListResponse{
// 		Components: []componentResponse{
// 			{
// 				Name:              "my-component",
// 				ID:                "my-component.my-provider",
// 				ComponentType:     "my-component-type",
// 				EnvironmentName:   "my-environment",
// 				ProviderName:      "my-provider",
// 				ResilienceZone:    "my-resilience-zone",
// 				KubernetesVersion: "1.2.3",
// 			},
// 			{
// 				Name:              "my-component-2",
// 				ID:                "my-component-2.my-provider",
// 				ComponentType:     "my-component-type",
// 				EnvironmentName:   "my-environment",
// 				ProviderName:      "my-provider",
// 				ResilienceZone:    "my-resilience-zone",
// 				KubernetesVersion: "1.2.3",
// 			},
// 		},
// 	}

// 	got, err := ListComponents(context.TODO(), in)
// 	assert.NoError(t, err)
// 	assert.Equal(t, want, got.ComponentListResponse)

// 	wantJSON := []byte(`{"components":[{"id":"my-component.my-provider","name":"my-component","provider_name":"my-provider","component_type":"my-component-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"},{"id":"my-component-2.my-provider","name":"my-component-2","provider_name":"my-provider","component_type":"my-component-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"}]}`)
// 	assert.Equal(t, wantJSON, got.JSONResponse)
// }

// func TestGetComponent(t *testing.T) {
// 	logger := logger.NewTestLogger(t)
// 	mockClient := apiclient.NewMockClientWithResponsesInterface(t)
// 	name := "my-component"
// 	included := []map[string]interface{}{
// 		{
// 			"@id":   "my-provider-id",
// 			"@type": "Provider",
// 			"name":  "my-provider",
// 		},
// 		{
// 			"@id":   "my-resilience-zone-id",
// 			"@type": "ResilienceZone",
// 			"name":  "my-resilience-zone",
// 		},
// 		{
// 			"@id":             "my-component-id",
// 			"@type":           "Component",
// 			"name":            "my-component",
// 			"componentType":   "my-component-type",
// 			"environmentName": "my-environment",
// 			"provider":        "my-provider-id",
// 			"resilienceZone":  "my-resilience-zone-id",
// 			"kubernetesVersion": map[string]interface{}{
// 				"version": "1.2.3",
// 			},
// 		},
// 	}
// 	providerId := "my-provider-id"
// 	mockClient.EXPECT().
// 		GetComponentWithResponse(mock.Anything, mock.Anything).
// 		Return(
// 			&apiclient.GetComponentResponse{
// 				Body: make([]byte, 0),
// 				HTTPResponse: &http.Response{
// 					Status:     "200 OK",
// 					StatusCode: 200,
// 				},
// 				ApplicationldJSONDefault: &apiclient.Component{
// 					Name:     &name,
// 					Provider: &providerId,
// 					Included: &included,
// 				},
// 			}, nil)

// 	want := &componentResponse{
// 		Name:         "my-component",
// 		ProviderName: "my-provider",
// 	}

// 	in := GetComponentInput{
// 		Logger:    logger,
// 		APIClient: mockClient,
// 	}
// 	got, err := GetComponent(context.TODO(), "my-component.my-provider", in)
// 	assert.NoError(t, err)
// 	assert.Equal(t, want, got.ComponentResponse)

// 	wantJSON := []byte(`{"name":"my-component","provider_name":"my-provider"}`)
// 	assert.Equal(t, wantJSON, got.JSONResponse)
// }

// func TestNodeList_ToResponse(t *testing.T) {
// 	cl := ComponentList{
// 		Components: make([]string, 0),
// 		Included:   make([]map[string]interface{}, 0),
// 	}
// 	t.Run("Valid input", func(t *testing.T) {
// 		cl.Components = []string{"my-component"}
// 		cl.Included = []map[string]interface{}{
// 			{
// 				"@id":   "my-provider-id",
// 				"@type": "Provider",
// 				"name":  "my-provider",
// 			},
// 			{
// 				"@id":   "my-resilience-zone-id",
// 				"@type": "ResilienceZone",
// 				"name":  "my-resilience-zone",
// 			},
// 			{
// 				"@id":             "my-component-id",
// 				"@type":           "Component",
// 				"name":            "my-component",
// 				"componentType":   "my-component-type",
// 				"environmentName": "my-environment",
// 				"provider":        "my-provider-id",
// 				"resilienceZone":  "my-resilience-zone-id",
// 				"kubernetesVersion": map[string]interface{}{
// 					"version": "1.2.3",
// 				},
// 			},
// 		}
// 		want := &componentListResponse{
// 			Components: []componentResponse{
// 				{
// 					Name:              "my-component",
// 					ID:                "my-component.my-provider",
// 					ComponentType:     "my-component-type",
// 					EnvironmentName:   "my-environment",
// 					ProviderName:      "my-provider",
// 					ResilienceZone:    "my-resilience-zone",
// 					KubernetesVersion: "1.2.3",
// 				},
// 			},
// 		}
// 		got := cl.ToResponse()
// 		assert.Equal(t, want, got)
// 	})

// 	t.Run("Empty input", func(t *testing.T) {
// 		cl.Components = []string{""}
// 		cl.Included = []map[string]interface{}{}
// 		want := &componentListResponse{[]componentResponse{}}
// 		got := cl.ToResponse()
// 		assert.Equal(t, want, got)
// 	})
// }

// func TestNodeList_MarshalJSON(t *testing.T) {
// 	cl := ComponentList{
// 		Components: make([]string, 0),
// 		Included:   make([]map[string]interface{}, 0),
// 	}
// 	cl.Components = []string{"my-component"}
// 	cl.Included = []map[string]interface{}{
// 		{
// 			"@id":   "my-provider-id",
// 			"@type": "Provider",
// 			"name":  "my-provider",
// 		},
// 		{
// 			"@id":   "my-resilience-zone-id",
// 			"@type": "ResilienceZone",
// 			"name":  "my-resilience-zone",
// 		},
// 		{
// 			"@id":             "my-component-id",
// 			"@type":           "Component",
// 			"name":            "my-component",
// 			"componentType":   "my-component-type",
// 			"environmentName": "my-environment",
// 			"provider":        "my-provider-id",
// 			"resilienceZone":  "my-resilience-zone-id",
// 			"kubernetesVersion": map[string]interface{}{
// 				"version": "1.2.3",
// 			},
// 		},
// 	}
// 	want := []byte(`{"components":[{"id":"my-component.my-provider","name":"my-component","provider_name":"my-provider","component_type":"my-component-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"}]}`)
// 	got, err := cl.MarshalJSON()
// 	assert.NoError(t, err)
// 	assert.Equal(t, want, got)
// }

// func TestListComponentNodes(t *testing.T) {
// 	logger := logger.NewTestLogger(t)

// 	components := []string{"my-component", "my-component-2"}
// 	included := []map[string]interface{}{
// 		{
// 			"@id":   "my-provider-id",
// 			"@type": "Provider",
// 			"name":  "my-provider",
// 		},
// 		{
// 			"@id":   "my-resilience-zone-id",
// 			"@type": "ResilienceZone",
// 			"name":  "my-resilience-zone",
// 		},
// 		{
// 			"@id":             "my-component-id",
// 			"@type":           "Component",
// 			"name":            "my-component",
// 			"componentType":   "my-component-type",
// 			"environmentName": "my-environment",
// 			"provider":        "my-provider-id",
// 			"resilienceZone":  "my-resilience-zone-id",
// 			"kubernetesVersion": map[string]interface{}{
// 				"version": "1.2.3",
// 			},
// 		},
// 		{
// 			"@id":             "my-component-id-2",
// 			"@type":           "Component",
// 			"name":            "my-component-2",
// 			"componentType":   "my-component-type",
// 			"environmentName": "my-environment",
// 			"provider":        "my-provider-id",
// 			"resilienceZone":  "my-resilience-zone-id",
// 			"kubernetesVersion": map[string]interface{}{
// 				"version": "1.2.3",
// 			},
// 		},
// 	}

// 	mockClient := apiclient.NewMockClientWithResponsesInterface(t)
// 	mockClient.EXPECT().
// 		ListComponentsWithResponse(mock.Anything, mock.Anything).
// 		Return(
// 			&apiclient.ListComponentsResponse{
// 				Body: make([]byte, 0),
// 				HTTPResponse: &http.Response{
// 					Status:     "200 OK",
// 					StatusCode: 200,
// 				},
// 				ApplicationldJSONDefault: &apiclient.Components{
// 					Components: &components,
// 					Included:   &included,
// 					Pagination: &apiclient.Pagination{},
// 				},
// 			}, nil)
// 	in := ListComponentsInput{
// 		Logger:    logger,
// 		APIClient: mockClient,
// 	}

// 	want := &componentListResponse{
// 		Components: []componentResponse{
// 			{
// 				Name:              "my-component",
// 				ID:                "my-component.my-provider",
// 				ComponentType:     "my-component-type",
// 				EnvironmentName:   "my-environment",
// 				ProviderName:      "my-provider",
// 				ResilienceZone:    "my-resilience-zone",
// 				KubernetesVersion: "1.2.3",
// 			},
// 			{
// 				Name:              "my-component-2",
// 				ID:                "my-component-2.my-provider",
// 				ComponentType:     "my-component-type",
// 				EnvironmentName:   "my-environment",
// 				ProviderName:      "my-provider",
// 				ResilienceZone:    "my-resilience-zone",
// 				KubernetesVersion: "1.2.3",
// 			},
// 		},
// 	}

// 	got, err := ListComponents(context.TODO(), in)
// 	assert.NoError(t, err)
// 	assert.Equal(t, want, got.ComponentListResponse)

// 	wantJSON := []byte(`{"components":[{"id":"my-component.my-provider","name":"my-component","provider_name":"my-provider","component_type":"my-component-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"},{"id":"my-component-2.my-provider","name":"my-component-2","provider_name":"my-provider","component_type":"my-component-type","environment_name":"my-environment","resilience_zone":"my-resilience-zone","kubernetes_version":"1.2.3"}]}`)
// 	assert.Equal(t, wantJSON, got.JSONResponse)
// }

// func TestGetComponentNode(t *testing.T) {
// 	logger := logger.NewTestLogger(t)
// 	mockClient := apiclient.NewMockClientWithResponsesInterface(t)
// 	name := "my-component"
// 	included := []map[string]interface{}{
// 		{
// 			"@id":   "my-provider-id",
// 			"@type": "Provider",
// 			"name":  "my-provider",
// 		},
// 		{
// 			"@id":   "my-resilience-zone-id",
// 			"@type": "ResilienceZone",
// 			"name":  "my-resilience-zone",
// 		},
// 		{
// 			"@id":             "my-component-id",
// 			"@type":           "Component",
// 			"name":            "my-component",
// 			"componentType":   "my-component-type",
// 			"environmentName": "my-environment",
// 			"provider":        "my-provider-id",
// 			"resilienceZone":  "my-resilience-zone-id",
// 			"kubernetesVersion": map[string]interface{}{
// 				"version": "1.2.3",
// 			},
// 		},
// 	}
// 	providerId := "my-provider-id"
// 	mockClient.EXPECT().
// 		GetComponentWithResponse(mock.Anything, mock.Anything).
// 		Return(
// 			&apiclient.GetComponentResponse{
// 				Body: make([]byte, 0),
// 				HTTPResponse: &http.Response{
// 					Status:     "200 OK",
// 					StatusCode: 200,
// 				},
// 				ApplicationldJSONDefault: &apiclient.Component{
// 					Name:     &name,
// 					Provider: &providerId,
// 					Included: &included,
// 				},
// 			}, nil)

// 	want := &componentResponse{
// 		Name:         "my-component",
// 		ProviderName: "my-provider",
// 	}

// 	in := GetComponentInput{
// 		Logger:    logger,
// 		APIClient: mockClient,
// 	}
// 	got, err := GetComponent(context.TODO(), "my-component.my-provider", in)
// 	assert.NoError(t, err)
// 	assert.Equal(t, want, got.ComponentResponse)

// 	wantJSON := []byte(`{"name":"my-component","provider_name":"my-provider"}`)
// 	assert.Equal(t, wantJSON, got.JSONResponse)
// }

// func TestGetComponentKubeConfig(t *testing.T) {
// 	logger := logger.NewTestLogger(t)

// 	want := []byte(`apiVersion: v1
// components:
// - component:
//     certificate-authority-data: test
//     server: https://test.test.dedicated.k8s.netic.dk:6443
//   name: test
// contexts:
// - context:
//     component: test
//     user: test
//   name: test
// current-context: test
// kind: Config
// preferences: {}
// users:
// - name: test
//   user:
//     password: REDACTED
//     username: test
// `)

// 	mockClient := apiclient.NewMockClientWithResponsesInterface(t)
// 	mockClient.EXPECT().
// 		GetComponentKubeConfigWithResponse(mock.Anything, mock.Anything).
// 		Return(
// 			&apiclient.GetComponentKubeConfigResponse{
// 				Body: want,
// 				HTTPResponse: &http.Response{
// 					Status:     "200 OK",
// 					StatusCode: 200,
// 				},
// 			}, nil)

// 	in := GetComponentKubeConfigInput{
// 		Logger:        logger,
// 		APIClient:     mockClient,
// 		ComponentName: "my-component.my-provider",
// 	}
// 	got, err := GetComponentKubeConfig(context.TODO(), in)
// 	assert.NoError(t, err)
// 	assert.Equal(t, want, got.Response)
// }
