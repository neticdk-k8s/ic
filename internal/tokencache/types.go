package tokencache

// Key is used to generate a unique ID for a cached token
type Key struct {
	IssuerURL   string
	ClientID    string
	ExtraScopes []string
}
