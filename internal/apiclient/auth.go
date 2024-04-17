package apiclient

import (
	"context"
	"fmt"
	"net/http"
)

type bearerToken struct {
	token string
}

// NewBearerTokenProvider creates a new bearer token authentication provider
func NewBearerTokenProvider(token string) *bearerToken {
	return &bearerToken{token: token}
}

// WithAuthHeader adds Authorization: Bearer header to the request
func (s *bearerToken) WithAuthHeader(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
	return nil
}
