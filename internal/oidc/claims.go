package oidc

import "time"

// Claims represents claims of an ID token.
type Claims struct {
	Subject string
	Expiry  time.Time
	Pretty  string // string representation for debug and logging
}

// IsExpired returns true if the token is expired.
func (c *Claims) IsExpired() bool {
	return c.Expiry.Before(time.Now())
}
