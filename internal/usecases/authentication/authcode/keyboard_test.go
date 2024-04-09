package authcode

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/reader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestKeyboard_Login(t *testing.T) {
	timeout := 5 * time.Second

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()

		o := &KeyboardLoginInput{
			RedirectURI: "urn:ietf:wg:oauth:2.0:oob",
		}

		mockClient := oidc.NewMockClient(t)
		mockClient.EXPECT().
			GetAuthCodeURL(mock.Anything, mock.Anything).
			Run(func(_ context.Context, in oidc.GetAuthCodeURLInput) {
				assert.Equal(t, o.RedirectURI, in.RedirectURI)
			}).
			Return("https://issuer.example.com/auth", nil)
		mockClient.EXPECT().
			ExchangeAuthCode(mock.Anything, mock.Anything).
			Run(func(_ context.Context, in oidc.ExchangeAuthCodeInput) {
				assert.Equal(t, in.Code, "YOUR_AUTH_CODE")
			}).
			Return(&oidc.TokenSet{
				IDToken:      "YOUR_ID_TOKEN",
				RefreshToken: "YOUR_REFRESH_TOKEN",
			}, nil)
		mockReader := reader.NewMockReader(t)
		mockReader.EXPECT().
			ReadString(keyboardPrompt).
			Return("YOUR_AUTH_CODE", nil)
		u := Keyboard{
			Reader: mockReader,
			Logger: logger.NewTestLogger(t),
		}
		got, err := u.Login(ctx, o, mockClient)
		assert.NoError(t, err, "Login returned error")

		want := &oidc.TokenSet{
			IDToken:      "YOUR_ID_TOKEN",
			RefreshToken: "YOUR_REFRESH_TOKEN",
		}
		assert.Equal(t, want, got)
	})

	t.Run("AuthError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()

		o := &KeyboardLoginInput{
			RedirectURI: "urn:ietf:wg:oauth:2.0:oob",
		}

		mockClient := oidc.NewMockClient(t)
		mockClient.EXPECT().
			GetAuthCodeURL(mock.Anything, mock.Anything).
			Return("https://issuer.example.com/auth", nil)
		mockClient.EXPECT().
			ExchangeAuthCode(mock.Anything, mock.Anything).
			Return(nil, errors.New("invalid auth code"))
		mockReader := reader.NewMockReader(t)
		mockReader.EXPECT().
			ReadString(keyboardPrompt).
			Return("YOUR_INVALID_AUTH_CODE", nil)
		u := Keyboard{
			Reader: mockReader,
			Logger: logger.NewTestLogger(t),
		}
		got, err := u.Login(ctx, o, mockClient)
		assert.Error(t, err, "Login returned error")
		assert.Nil(t, got)
	})
}
