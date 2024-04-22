package cmd

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/neticdk-k8s/ic/internal/apiclient"
	"github.com/neticdk-k8s/ic/internal/logger"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_UpdateClusterCommand(t *testing.T) {
	t.Parallel()
	ec, got := newMockedUpdateClusterEC(t)
	cmd := NewRootCmd(ec)

	cmd.SetArgs([]string{"update", "cluster", "my-cluster.my-provider", "--resilience-zone", "platform"})
	err := cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, got.String(), "Logging in")
	assert.Contains(t, got.String(), "Updating cluster")
	assert.Contains(t, got.String(), "my-cluster")
	assert.Contains(t, got.String(), "my-provider")
	assert.Contains(t, got.String(), "Cluster metadata updated")
}

func Test_UpdateClusterCommandWithJSONOutput(t *testing.T) {
	t.Parallel()
	ec, got := newMockedUpdateClusterEC(t)
	cmd := NewRootCmd(ec)

	cmd.SetArgs([]string{"update", "cluster", "my-cluster.my-provider", "--resilience-zone", "platform", "-o", "json"})
	err := cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, got.String(), "\"name\": \"my-cluster\"")
	assert.Contains(t, got.String(), "\"provider_name\": \"my-provider\"")
}

func Test_UpdateClusterCommandInvalidParameters(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		testName     string
		args         []string
		expErr       error
		expErrString string
	}{
		{
			testName:     "no cluster given",
			args:         []string{},
			expErrString: "accepts 1 arg(s), received 0",
		},
		{
			testName:     "no parameters given",
			args:         []string{"my-cluster.my-provider"},
			expErrString: "at least one of the flags",
		},
		{
			testName:     "custom operations without valid url",
			args:         []string{"my-cluster.my-provider", "--has-co", "--co-url", "invalid://host"},
			expErrString: "must be a URL",
		},
		{
			testName:     "resilience zone is invalid rfc1035 label",
			args:         []string{"my-cluster.my-provider", "--resilience-zone", "my platform"},
			expErrString: "must be an RFC1035",
		},
		{
			testName:     "environment is invalid rfc1035 label",
			args:         []string{"my-cluster.my-provider", "--environment", "invalid environment"},
			expErrString: "must be an RFC1035",
		},
		{
			testName:     "invalid infrastructure provider",
			args:         []string{"my-cluster.my-provider", "--infrastructure-provider", "invalid"},
			expErrString: "invalid",
		},
		{
			testName:     "invalid resilience zone",
			args:         []string{"my-cluster.my-provider", "--resilience-zone", "myplatform"},
			expErrString: "invalid",
		},
		{
			testName:     "invalid api endpoint url",
			args:         []string{"my-cluster.my-provider", "--api-endpoint", "invalid://host"},
			expErrString: "must be a URL",
		},
		{
			testName:     "invalid subscription length",
			args:         []string{"my-cluster.my-provider", "--subscription", "446"},
			expErrString: "minimum 5 characters",
		},
		{
			testName:     "invalid subscription",
			args:         []string{"my-cluster.my-provider", "--subscription", "ΩΩΩΩΩ"},
			expErrString: "must be an ASCII string",
		},
		{
			testName:     "has-co required with co-url",
			args:         []string{"my-cluster.my-provider", "--has-co"},
			expErrString: "they must all be set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			got := new(bytes.Buffer)
			in := ExecutionContextInput{
				Stdout: got,
				Stderr: got,
			}
			ec := NewExecutionContext(in)
			cmd := NewRootCmd(ec)
			args := append([]string{"update", "cluster"}, tc.args...)
			cmd.SetArgs(args)
			err := cmd.Execute()
			assert.Error(t, err)
			if err != nil {
				assert.Contains(t, err.Error(), tc.expErrString)
			}
		})
	}
}

func newMockedUpdateClusterEC(t *testing.T) (*ExecutionContext, *bytes.Buffer) {
	got := new(bytes.Buffer)
	in := ExecutionContextInput{
		Stdout: got,
		Stderr: got,
	}
	ec := NewExecutionContext(in)
	mockAuthenticator := authentication.NewMockAuthenticator(t)
	mockAuthenticator.EXPECT().
		SetLogger(mock.Anything).
		Run(func(_ logger.Logger) {}).
		Return()
	mockAuthenticator.EXPECT().
		Login(mock.Anything, mock.Anything).
		Run(func(_ context.Context, in authentication.LoginInput) {}).
		Return(&oidc.TokenSet{
			AccessToken:  "YOUR_ACCESS_TOKEN",
			IDToken:      "YOUR_ID_TOKEN",
			RefreshToken: "YOUR_REFRESH_TOKEN",
		}, nil)
	ec.Authenticator = mockAuthenticator
	included := []map[string]interface{}{
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
			"@id":             "my-cluster-id",
			"@type":           "Cluster",
			"name":            "my-cluster",
			"clusterType":     "dedicated",
			"environmentName": "testing",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-rz-id",
			"kubernetesVersion": map[string]interface{}{
				"version": "v1.2.3",
			},
		},
	}
	name := "my-cluster"
	providerId := "my-provider-id"
	mockClientWithResponsesInterface := apiclient.NewMockClientWithResponsesInterface(t)
	mockClientWithResponsesInterface.EXPECT().
		UpdateClusterWithResponse(mock.Anything, mock.Anything, mock.Anything).
		Return(
			&apiclient.UpdateClusterResponse{
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
	apiClient := mockClientWithResponsesInterface
	ec.APIClient = apiClient

	return ec, got
}
