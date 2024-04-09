package jwt_test

import (
	"testing"
	"time"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/jwt"
	"github.com/stretchr/testify/assert"
)

func TestClaims_IsExpired(t *testing.T) {
	claims := jwt.Claims{
		Expiry: time.Now().Local().Add(time.Duration(-1) * time.Hour),
	}

	t.Run("Expired", func(t *testing.T) {
		got := claims.IsExpired()
		assert.True(t, got, "IsExpired() wants true but is false")
	})

	claims = jwt.Claims{
		Expiry: time.Now().Local().Add(time.Hour * 1),
	}

	t.Run("NotExpired", func(t *testing.T) {
		got := claims.IsExpired()
		assert.False(t, got, "IsExpired() wants false but is true")
	})
}
