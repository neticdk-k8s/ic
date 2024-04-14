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
	Stderr, Stdout io.Writer

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

	// OIDC is the OIDC settings
	OIDC OIDCConfig

	// OIDC is the OIDC provider settings
	OIDCProvider oidc.Provider

	// Authenticator is the Authenticator
	Authenticator authentication.Authenticator

	// TokenCache is the token cache
	TokenCache tokencache.Cache

	// APIClient is an inventory server api client
	APIClient *apiclient.ClientWithResponses
}

// NewExecutionContext creates a new ExecutionContext
func NewExecutionContext() *ExecutionContext {
	ec := &ExecutionContext{
		Stderr:       os.Stderr,
		Stdout:       os.Stdout,
		OutputFormat: "text",
		OIDC:         OIDCConfig{},
	}
	return ec
}

// SetupAPIClient configures the API client
func (ec *ExecutionContext) SetupAPIClient(token string) error {
	var err error
	provider := apiclient.NewBearerTokenProvider(token)
	ec.APIClient, err = apiclient.NewClientWithResponses(
		ec.APIServer,
		apiclient.WithRequestEditorFn(provider.WithAuthHeader))
	if err != nil {
		return err
	}
	return nil
}

// Prepare sets up context that does not depend on flags
// It should be called before rootCmd.Execute
func (ec *ExecutionContext) Prepare() error {
	ec.IsTerminal = term.IsTerminal(int(os.Stdout.Fd()))

	ec.SetupLogger("info")

	ec.setupSpinner()

	return nil
}

// Setup sets up context that depends on the flags being sets
// It should be called from rootCmd.PersistentPreRunE
func (ec *ExecutionContext) Setup() error {
	ec.SetupLogger(ec.LogLevel)

	authn := authentication.NewAuthentication(
		ec.Logger,
		nil,
		&authcode.Browser{Logger: ec.Logger},
		&authcode.Keyboard{Reader: reader.NewReader(), Logger: ec.Logger})
	ec.Authenticator = authentication.NewAuthenticator(ec.Logger, authn)

	ec.OIDCProvider = oidc.Provider{
		IssuerURL:   ec.OIDC.IssuerURL,
		ClientID:    ec.OIDC.ClientID,
		ExtraScopes: []string{"profile", "email", "roles", "offline_access"},
	}

	var err error
	if ec.TokenCache, err = tokencache.NewFSCache(ec.OIDC.TokenCacheDir); err != nil {
		return fmt.Errorf("creating token cache: %w", err)
	}

	return nil
}

func (ec *ExecutionContext) setupSpinner() {
	if ec.Spinner == nil {
		ec.Spinner = ui.NewSpinner(ec.Stderr, ec.Logger)
	}
}

// Spin starts the spinnner and sets its text
func (ec *ExecutionContext) Spin(t string) {
	if !ec.Spinner.Running() {
		ec.Spinner.Run()
	}
	ec.Spinner.Text(t)
}

// SetupLogger configures the logger
func (ec *ExecutionContext) SetupLogger(logLevel string) {
	if ec.Logger == nil {
		ec.Logger = logger.New(ec.Stderr, logLevel)
	} else {
		if err := ec.Logger.SetLevel(logLevel); err != nil {
			ec.Logger.Error("Failed to set loglevel", "level", logLevel)
		}
	}
	if logLevel != ec.LogLevel {
		ec.LogLevel = logLevel
	}
	ec.Logger.SetInteractive(ec.Interactive, ec.IsTerminal)
}
