package tokencache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/stretchr/testify/assert"
)

func TestNewFSCache(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cacheDir := t.TempDir()

		got, err := NewFSCache(cacheDir)
		assert.NoError(t, err, "could not create new fsCache")

		want := &fsCache{CacheDir: cacheDir}
		assert.Equal(t, want, got)
	})
}

func TestFSCache_Lookup(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cacheDir := t.TempDir()

		fsCache, err := NewFSCache(cacheDir)
		assert.NoError(t, err, "could not create new fsCache")

		key := Key{
			IssuerURL:   "YOUR_ISSUER",
			ClientID:    "YOUR_CLIENT_ID",
			ExtraScopes: []string{"openid", "email"},
		}
		json := `{"id_token":"YOUR_ID_TOKEN","refresh_token":"YOUR_REFRESH_TOKEN"}`
		filename, err := computeFilename(key)
		assert.NoError(t, err, "could not compute the key")

		p := filepath.Join(cacheDir, filename)
		if err := os.WriteFile(p, []byte(json), 0o600); err != nil {
			t.Fatalf("could not write to the temp file: %s", err)
		}
		got, err := fsCache.Lookup(key)
		assert.NoError(t, err, "could not look up cached token")

		want := &oidc.TokenSet{IDToken: "YOUR_ID_TOKEN", RefreshToken: "YOUR_REFRESH_TOKEN"}
		assert.Equal(t, want, got)
	})
}

func TestFSCache_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cacheDir := t.TempDir()

		fsCache, err := NewFSCache(cacheDir)
		assert.NoError(t, err, "could not create new fsCache")

		key := Key{
			IssuerURL:   "YOUR_ISSUER",
			ClientID:    "YOUR_CLIENT_ID",
			ExtraScopes: []string{"openid", "email"},
		}
		tokenSet := oidc.TokenSet{IDToken: "YOUR_ID_TOKEN", RefreshToken: "YOUR_REFRESH_TOKEN"}
		err = fsCache.Save(key, tokenSet)
		assert.NoError(t, err, "could not save cached token")

		filename, err := computeFilename(key)
		assert.NoError(t, err, "could not compute the key")

		p := filepath.Join(cacheDir, filename)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("could not read the token cache file: %s", err)
		}

		want := "{\"id_token\":\"YOUR_ID_TOKEN\",\"refresh_token\":\"YOUR_REFRESH_TOKEN\"}\n"
		got := string(b)
		assert.Equal(t, want, got)
	})
}

func TestFSCache_Invalidate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cacheDir := t.TempDir()

		fsCache, err := NewFSCache(cacheDir)
		assert.NoError(t, err, "could not create new fsCache")

		key := Key{
			IssuerURL:   "YOUR_ISSUER",
			ClientID:    "YOUR_CLIENT_ID",
			ExtraScopes: []string{"openid", "email"},
		}
		json := `{"id_token":"YOUR_ID_TOKEN","refresh_token":"YOUR_REFRESH_TOKEN"}`
		filename, err := computeFilename(key)
		assert.NoError(t, err, "could not compute the key")

		p := filepath.Join(cacheDir, filename)
		if err := os.WriteFile(p, []byte(json), 0o600); err != nil {
			t.Fatalf("could not write to the temp file: %s", err)
		}
		err = fsCache.Invalidate(key)
		assert.NoError(t, err, "could not invalidate token")

		_, err = os.Stat(p)
		assert.Error(t, err, "cached token file not deleted")
	})

	t.Run("CacheMissError", func(t *testing.T) {
		cacheDir := t.TempDir()

		fsCache, err := NewFSCache(cacheDir)
		assert.NoError(t, err, "could not create new fsCache")

		key := Key{
			IssuerURL:   "YOUR_ISSUER",
			ClientID:    "YOUR_CLIENT_ID",
			ExtraScopes: []string{"openid", "email"},
		}
		err = fsCache.Invalidate(key)
		assert.ErrorIs(t, err, &CacheMissError{})
	})
}
