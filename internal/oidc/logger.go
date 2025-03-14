package oidc

import "github.com/neticdk-k8s/ic/internal/logger"

// OIDCSlogAdapter is an slog adapter used by retryablehttp
type SlogAdapter struct {
	// Looger is the logger to use
	Logger logger.Logger
}

// Error logs messages at error level
func (a SlogAdapter) Error(msg string, keysAndValues ...any) {
	a.Logger.Error(msg, keysAndValues...)
}

// Info logs messages at info level
func (a SlogAdapter) Info(msg string, keysAndValues ...any) {
	a.Logger.Info(msg, keysAndValues...)
}

// Debug logs messages at debug level
func (a SlogAdapter) Debug(msg string, keysAndValues ...any) {
	a.Logger.Debug(msg, keysAndValues...)
}

// Warn logs messages at warn level
func (a SlogAdapter) Warn(msg string, keysAndValues ...any) {
	a.Logger.Warn(msg, keysAndValues...)
}
