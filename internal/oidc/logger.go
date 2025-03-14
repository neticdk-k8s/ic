package oidc

import "github.com/neticdk-k8s/ic/internal/logger"

// OIDCSlogAdapter is an slog adapter used by retryablehttp
type OIDCSlogAdapter struct {
	Logger logger.Logger
}

// Error logs messages at error level
func (a OIDCSlogAdapter) Error(msg string, keysAndValues ...any) {
	a.Logger.Error(msg, keysAndValues...)
}

// Info logs messages at info level
func (a OIDCSlogAdapter) Info(msg string, keysAndValues ...any) {
	a.Logger.Info(msg, keysAndValues...)
}

// Debug logs messages at debug level
func (a OIDCSlogAdapter) Debug(msg string, keysAndValues ...any) {
	a.Logger.Debug(msg, keysAndValues...)
}

// Warn logs messages at warn level
func (a OIDCSlogAdapter) Warn(msg string, keysAndValues ...any) {
	a.Logger.Warn(msg, keysAndValues...)
}
