package oidc

import "github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"

// OIDCSlogAdapter is an slog adapter used by retryablehttp
type OIDCSlogAdapter struct {
	Logger logger.Logger
}

// Error logs messages at error level
func (a OIDCSlogAdapter) Error(msg string, keysAndValues ...interface{}) {
	a.Logger.Error(msg, keysAndValues...)
}

// Info logs messages at info level
func (a OIDCSlogAdapter) Info(msg string, keysAndValues ...interface{}) {
	a.Logger.Info(msg, keysAndValues...)
}

// Debug logs messages at debug level
func (a OIDCSlogAdapter) Debug(msg string, keysAndValues ...interface{}) {
	a.Logger.Debug(msg, keysAndValues...)
}

// Warn logs messages at warn level
func (a OIDCSlogAdapter) Warn(msg string, keysAndValues ...interface{}) {
	a.Logger.Warn(msg, keysAndValues...)
}
