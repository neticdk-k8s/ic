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

func Test_DeleteClusterCommand(t *testing.T) {
	got := new(bytes.Buffer)
	ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
	ec.Stdin = nil
	ec.Stderr = got
	ec.Stdout = got
	ec.PFlags.NoInputEnabled = true
	ec.PFlags.ForceEnabled = true
	ec.PFlags.NoHeadersEnabled = true
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
	mockClientWithResponsesInterface := apiclient.NewMockClientWithResponsesInterface(t)
	mockClientWithResponsesInterface.EXPECT().
		DeleteClusterWithResponse(mock.Anything, mock.Anything, mock.Anything).
		Return(
			&apiclient.DeleteClusterResponse{
				Body: make([]byte, 0),
				HTTPResponse: &http.Response{
					Status:     "204 No Content",
					StatusCode: 204,
				},
			}, nil)
	apiClient := mockClientWithResponsesInterface
	ac.APIClient = apiClient

	cmd := newRootCmd(ac)

	cmd.SetArgs([]string{"--force", "delete", "cluster", "my-cluster"})
	err := cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)
	t.Log(got.String())
	assert.Contains(t, got.String(), "Logging in")
	assert.Contains(t, got.String(), "Deleting cluster")
	assert.Contains(t, got.String(), "Cluster deleted")
}
