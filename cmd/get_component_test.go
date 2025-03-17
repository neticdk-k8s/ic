package cmd

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/neticdk-k8s/ic/internal/apiclient"
	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_GetComponentCommand(t *testing.T) {
	got := new(bytes.Buffer)
	ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
	ec.Stderr = got
	ec.Stdout = got
	ui.SetDefaultOutput(got)
	ac := ic.NewContext()
	ac.EC = ec
	mockAuthenticator := authentication.NewMockAuthenticator(t)
	mockAuthenticator.EXPECT().
		SetLogger(mock.Anything).
		Run(func(_ *slog.Logger) {}).
		Return()
	mockAuthenticator.EXPECT().
		Login(mock.Anything, mock.Anything).
		Run(func(_ context.Context, in authentication.LoginInput) {}).
		Return(&oidc.TokenSet{
			AccessToken:  "YOUR_ACCESS_TOKEN",
			IDToken:      "YOUR_ID_TOKEN",
			RefreshToken: "YOUR_REFRESH_TOKEN",
		}, nil)
	ac.Authenticator = mockAuthenticator
	included := []map[string]any{
		{
			"@id":   "my-provider-id",
			"@type": "Provider",
			"name":  "my-provider",
		},
		{
			"@id":   "my-rz-id",
			"@type": "ResilienceZone",
			"name":  "my-resilience-zone",
		},
		{
			"@id":             "my-component-id",
			"@type":           "Component",
			"name":            "my-component",
			"componentType":   "dedicated",
			"environmentName": "testing",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-rz-id",
			"kubernetesVersion": map[string]any{
				"version": "v1.2.3",
			},
		},
	}
	name := "my-component"
	namespace := "my-namespace"
	mockClientWithResponsesInterface := apiclient.NewMockClientWithResponsesInterface(t)
	mockClientWithResponsesInterface.EXPECT().
		GetComponentWithResponse(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(
			&apiclient.GetComponentResponse{
				Body: make([]byte, 0),
				HTTPResponse: &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
				},
				ApplicationldJSONDefault: &apiclient.Component{
					Name:      &name,
					Namespace: &namespace,
					Included:  &included,
				},
			}, nil)
	apiClient := mockClientWithResponsesInterface
	ac.APIClient = apiClient

	cmd := newRootCmd(ac)

	t.Run("get component my-namespace my-component.my-provider", func(t *testing.T) {
		cmd.SetArgs([]string{"get", "component", "my-namespace", "my-component.my-provider"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Logging in")
		assert.Contains(t, got.String(), "Getting component")
		assert.Contains(t, got.String(), "my-component")
		assert.Contains(t, got.String(), "my-namespace")
	})

	t.Run("get component my-namespace my-component -o json", func(t *testing.T) {
		cmd.SetArgs([]string{"get", "component", "my-namespace", "my-component", "-o", "json"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "\"name\": \"my-component\"")
		assert.Contains(t, got.String(), "\"namespace\": \"my-namespace\"")
	})
}
