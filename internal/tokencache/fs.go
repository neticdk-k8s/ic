package tokencache

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/oidc"
	"github.com/pkg/errors"
)

type fsCache struct {
	tokenDir string
}

type cachedToken struct {
	IDToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// NewFSCache creates a new filesystem backed cache
func NewFSCache() (*fsCache, error) {
	tokenDir, err := os.UserCacheDir()
	if err != nil {
		return nil, errors.Wrap(err, "opening cache dir")
	}
	tokenDir = filepath.Join(tokenDir, "ic", "oidc-login")
	cache := &fsCache{
		tokenDir: tokenDir,
	}
	return cache, nil
}

// Lookup retrieves a cached token
func (c *fsCache) Lookup(key Key) (*oidc.TokenSet, error) {
	filename, err := computeFilename(key)
	if err != nil {
		return nil, fmt.Errorf("could not compute the key: %w", err)
	}
	p := filepath.Join(c.tokenDir, filename)
	f, err := os.Open(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &CacheMissError{}
		}
		return nil, fmt.Errorf("could not open file %s: %w", p, err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var e cachedToken
	if err := d.Decode(&e); err != nil {
		return nil, fmt.Errorf("invalid json file %s: %w", p, err)
	}
	return &oidc.TokenSet{
		IDToken:      e.IDToken,
		RefreshToken: e.RefreshToken,
	}, nil
}

// Save stores a cached token
func (c *fsCache) Save(key Key, tokenSet oidc.TokenSet) error {
	if err := os.MkdirAll(c.tokenDir, 0o700); err != nil {
		return fmt.Errorf("could not create directory %s: %w", c.tokenDir, err)
	}
	filename, err := computeFilename(key)
	if err != nil {
		return fmt.Errorf("could not compute the key: %w", err)
	}
	p := filepath.Join(c.tokenDir, filename)
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("could not create file %s: %w", p, err)
	}
	defer f.Close()
	e := cachedToken{
		IDToken:      tokenSet.IDToken,
		RefreshToken: tokenSet.RefreshToken,
	}
	if err := json.NewEncoder(f).Encode(&e); err != nil {
		return fmt.Errorf("json encode error: %w", err)
	}
	return nil
}

// Invalidate deletes a cached token
func (c *fsCache) Invalidate(key Key) error {
	filename, err := computeFilename(key)
	if err != nil {
		return fmt.Errorf("could not compute the key: %w", err)
	}
	p := filepath.Join(c.tokenDir, filename)

	if err := os.Remove(p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &CacheMissError{}
		}
		return fmt.Errorf("could not remove file %s: %w", p, err)
	}
	return nil
}

func computeFilename(key Key) (string, error) {
	s := sha256.New()
	e := gob.NewEncoder(s)
	if err := e.Encode(&key); err != nil {
		return "", fmt.Errorf("could not encode the key: %w", err)
	}
	h := hex.EncodeToString(s.Sum(nil))
	return h, nil
}
