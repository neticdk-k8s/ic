package apiclient

import (
	"context"
	"fmt"
	"net/http"
)

type bearerToken struct {
	token string
}

func NewAuthProvider(token string) *bearerToken {
	return &bearerToken{token: token}
}

func (s *bearerToken) Intercept(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
	return nil
}
