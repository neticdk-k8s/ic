package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/neticdk-k8s/ic/internal/apiclient"
	"github.com/neticdk-k8s/ic/internal/logger"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk-k8s/ic/internal/reader"
	"github.com/neticdk-k8s/ic/internal/tokencache"
	"github.com/neticdk-k8s/ic/internal/ui"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// OIDCConfig holds flag values for OIDC settings
type OIDCConfig struct {
	IssuerURL                   string
	ClientID                    string
	GrantType                   string
	RedirectURLHostname         string
	RedirectURIAuthCodeKeyboard string
	AuthBindAddr                string
	TokenCacheDir               string
}

// ExecutionContext holds configuration that can be used (and modified) across
// the application
type ExecutionContext struct {
	Stdout, Stderr io.Writer

	// Command is the current command
	Command *cobra.Command

	// Version is the CLI version
	Version string

	// Logger is the global logger object to print logs.
	Logger logger.Logger

	// Spinner is the global spinner object used to show progress across the cli
	Spinner *ui.Spinner

	// LogLevel is the log level used for the logger
	LogLevel string

	// APIServer is the inventory api server endpoint
	APIServer string

	// Interactive can be used to force interactive/non-interactive use
	Interactive string

	// IsTerminal indicates whether the current session is a terminal or not
	IsTerminal bool

	// OutputFormat is the format used for outputting data
	OutputFormat string

	// NoHeaders is used to control whether headers are printed
	NoHeaders bool

	// OIDC is the OIDC settings
	OIDC OIDCConfig

	// OIDC is the OIDC provider settings
	OIDCProvider *oidc.Provider

	// Authenticator is the Authenticator
	Authenticator authentication.Authenticator

	// TokenCache is the token cache
	TokenCache tokencache.Cache

	// APIClient is an inventory server api client
	APIClient apiclient.ClientWithResponsesInterface
}

type ExecutionContextInput struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Version string
}

// NewExecutionContext creates a new ExecutionContext
func NewExecutionContext(in ExecutionContextInput) *ExecutionContext {
	ec := &ExecutionContext{
		Stdout:       in.Stdout,
		Stderr:       in.Stderr,
		Version:      in.Version,
		OutputFormat: "text",
		OIDC:         OIDCConfig{},
		LogLevel:     "info",
	}
	ec.Logger = logger.New(in.Stderr, ec.LogLevel)

	stdout, ok := ec.Stdout.(*os.File)
	if !ok {
		ec.Logger.Debug("using default stdout")
		stdout = os.Stdout
	}
	ec.IsTerminal = term.IsTerminal(int(stdout.Fd()))

	ec.Logger.SetInteractive("auto", ec.IsTerminal)

	ec.setupSpinner()
	return ec
}

// SetupDefaultAPIClient sets up ec.APIClient from flags if it's not already set
func (ec *ExecutionContext) SetupDefaultAPIClient(token string) (err error) {
	if ec.APIClient != nil {
		return
	}

	provider := apiclient.NewBearerTokenProvider(token)
	ec.APIClient, err = apiclient.NewClientWithResponses(
		ec.APIServer,
		apiclient.WithRequestEditorFn(provider.WithAuthHeader))
	return
}

// SetupDefaultAuthenticator sets up ec.Authenticator from flags if it's not already set
// It should be called from rootCmd.PersistentPreRunE
func (ec *ExecutionContext) SetupDefaultAuthenticator() {
	if ec.Authenticator != nil {
		return
	}

	authn := authentication.NewAuthentication(
		ec.Logger,
		nil,
		&authcode.Browser{Logger: ec.Logger},
		&authcode.Keyboard{Reader: reader.NewReader(), Logger: ec.Logger})
	ec.Authenticator = authentication.NewAuthenticator(ec.Logger, authn)
}

// SetupDefaultOIDCProvider sets up ec.OIDCProvider from flags if it's not already set
// It should be called from rootCmd.PersistentPreRunE
func (ec *ExecutionContext) SetupDefaultOIDCProvider() {
	if ec.OIDCProvider != nil {
		return
	}
	ec.OIDCProvider = &oidc.Provider{
		IssuerURL:   ec.OIDC.IssuerURL,
		ClientID:    ec.OIDC.ClientID,
		ExtraScopes: []string{"profile", "email", "roles", "offline_access"},
	}
}

// SetupDefaultTokenCache sets up ec.TokenCache from flags if it's not already set
// It should be called from rootCmd.PersistentPreRunE
func (ec *ExecutionContext) SetupDefaultTokenCache() (err error) {
	if ec.TokenCache != nil {
		return
	}

	if ec.TokenCache, err = tokencache.NewFSCache(ec.OIDC.TokenCacheDir); err != nil {
		return fmt.Errorf("creating token cache: %w", err)
	}

	return
}

// SetLogLevel sets the ec.Logger log level
func (ec *ExecutionContext) SetLogLevel() {
	if err := ec.Logger.SetLevel(ec.LogLevel); err != nil {
		ec.Logger.Error("Failed to set loglevel", "level", ec.LogLevel)
	}
	ec.Logger.SetInteractive(ec.Interactive, ec.IsTerminal)
}

func (ec *ExecutionContext) setupSpinner() {
	if ec.Spinner == nil {
		ec.Spinner = ui.NewSpinner(ec.Stderr, ec.Logger)
	}
}

// Spin starts the spinnner and sets its text
func (ec *ExecutionContext) Spin(t string) {
	if !ec.Spinner.Running() {
		ec.Spinner.Run(t)
	}
	ec.Spinner.Text(t)
}
