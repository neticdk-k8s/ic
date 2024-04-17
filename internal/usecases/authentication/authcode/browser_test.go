package authcode

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/neticdk-k8s/ic/internal/logger"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBrowser_Login(t *testing.T) {
	timeout := 5 * time.Second

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()

		o := &BrowserLoginInput{
			BindAddress:         "127.0.0.1:18000",
			RedirectURLHostname: "localhost",
		}

		mockClient := oidc.NewMockClient(t)
		mockClient.EXPECT().
			GetTokenByAuthCode(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, in oidc.GetTokenByAuthCodeInput, readyChan chan<- string) {
				if diff := cmp.Diff(o.BindAddress, in.BindAddress); diff != "" {
					t.Errorf("BindAddress mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(o.RedirectURLHostname, in.RedirectURLHostname); diff != "" {
					t.Errorf("RedirectURLHostname mismatch (-want +got):\n%s", diff)
				}
				readyChan <- "LOCAL_SERVER_URL"
			}).
			Return(&oidc.TokenSet{
				AccessToken:  "YOUR_ACCESS_TOKEN",
				IDToken:      "YOUR_ID_TOKEN",
				RefreshToken: "YOUR_REFRESH_TOKEN",
			}, nil)
		u := Browser{
			Logger: logger.NewTestLogger(t),
		}
		got, err := u.Login(ctx, o, mockClient)
		assert.NoError(t, err, "Login returned error")

		want := &oidc.TokenSet{
			AccessToken:  "YOUR_ACCESS_TOKEN",
			IDToken:      "YOUR_ID_TOKEN",
			RefreshToken: "YOUR_REFRESH_TOKEN",
		}
		assert.Equal(t, want, got)
	})

	t.Run("AuthError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()

		o := &BrowserLoginInput{
			BindAddress:         "127.0.0.1:18000",
			RedirectURLHostname: "localhost",
		}

		mockClient := oidc.NewMockClient(t)
		mockClient.EXPECT().
			GetTokenByAuthCode(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, in oidc.GetTokenByAuthCodeInput, readyChan chan<- string) {
				if diff := cmp.Diff(o.BindAddress, in.BindAddress); diff != "" {
					t.Errorf("BindAddress mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(o.RedirectURLHostname, in.RedirectURLHostname); diff != "" {
					t.Errorf("RedirectURLHostname mismatch (-want +got):\n%s", diff)
				}
				readyChan <- "LOCAL_SERVER_URL"
			}).
			Return(nil, errors.New("oauth2 error: bad credentials"))
		u := Browser{
			Logger: logger.NewTestLogger(t),
		}
		got, err := u.Login(ctx, o, mockClient)
		assert.Error(t, err, "Login returned error")
		assert.Nil(t, got)
	})
}
