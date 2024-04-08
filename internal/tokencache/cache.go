package tokencache

import (
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
)

type Interface interface {
	Lookup(key Key) (*oidc.TokenSet, error)
	Save(key Key, tokenSet oidc.TokenSet) error
	Invalidate(key Key) error
}
