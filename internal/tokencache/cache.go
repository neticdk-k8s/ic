package tokencache

import (
	"github.com/neticdk-k8s/ic/internal/oidc"
)

// Cache represents a token caching interface
type Cache interface {
    // Lookup retrieves a token set from the cache
	Lookup(key Key) (*oidc.TokenSet, error)
    // Save stores a token set in the cache
	Save(key Key, tokenSet oidc.TokenSet) error
    // Invalidate removes a token set from the cache
	Invalidate(key Key) error
}
