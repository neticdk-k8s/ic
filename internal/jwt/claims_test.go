package jwt_test

import (
	"testing"
	"time"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/jwt"
)

func TestClaims_IsExpired(t *testing.T) {
	claims := jwt.Claims{
		Expiry: time.Now().Local().Add(time.Duration(-1) * time.Hour),
	}

	t.Run("Expired", func(t *testing.T) {
		got := claims.IsExpired()
		if got != true {
			t.Errorf("IsExpired() wants true but false")
		}
	})

	claims = jwt.Claims{
		Expiry: time.Now().Local().Add(time.Hour * 1),
	}

	t.Run("NotExpired", func(t *testing.T) {
		got := claims.IsExpired()
		if got != false {
			t.Errorf("IsExpired() wants false but true")
		}
	})
}
