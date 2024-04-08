package oidc

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

func (ts TokenSet) DecodeWithoutVerify() (*Claims, error) {
	return DecodeWithoutVerify(ts.IDToken)
}
