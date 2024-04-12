package oidc

import "github.com/neticdk-k8s/ic/internal/jwt"

// Provider represents an OICD provider
type Provider struct {
	IssuerURL   string
	ClientID    string
	ExtraScopes []string
}

// TokenSet represents a set of ID token and refresh token
type TokenSet struct {
	IDToken      string
	RefreshToken string
}

// DecodeWithoutVerify decodes the JWT string and returns the claims.
// Note that this method does not verify the signature and always trust it.
func (ts TokenSet) DecodeWithoutVerify() (*jwt.Claims, error) {
	return jwt.DecodeWithoutVerify(ts.IDToken)
}
