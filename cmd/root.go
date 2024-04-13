package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultConfigFilename = "ic"
	envPrefix             = "IC"
	oobRedirectURI        = "urn:ietf:wg:oauth:2.0:oob"

	groupBase    = "group-base"
	groupAuth    = "group-auth"
	groupOther   = "group-other"
	groupCluster = "group-cluster"
)

// ec is the Execution Context for the current run
var ec *ExecutionContext

var rootCmd = &cobra.Command{
	Use:                   "ic",
	DisableFlagsInUseLine: true,
	Short:                 "Inventory CLI",
	SilenceUsage:          true,
	SilenceErrors:         true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initConfig(cmd); err != nil {
			return err
		}
		if err := ec.Setup(); err != nil {
			return fmt.Errorf("settings up execution context: %w", err)
		}
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func initConfig(cmd *cobra.Command) error {
	v := viper.New()
	v.SetConfigName(defaultConfigFilename)
	v.AddConfigPath(".")
	if dir, err := os.UserConfigDir(); err == nil {
		v.AddConfigPath(filepath.Join(dir, "ic"))
	}
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func getDefaultTokenCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "ic", "oidc-login")
	}
	return filepath.Join(cacheDir, "ic", "oidc-login")
}

func init() {
	ec = NewExecutionContext()

	f := rootCmd.PersistentFlags()
	f.StringVarP(&ec.LogLevel, "log-level", "l", "info", "Log level")
	f.StringVarP(&ec.APIServer, "api-server", "s", "https://api.k8s.netic.dk", "URL for the inventory server.")
	f.StringVarP(&ec.Interactive, "interactive", "i", "auto", "Run in interactive mode. One of (yes|no|auto)")
	f.StringVarP(&ec.OutputFormat, "output-format", "o", "text", "Output format. One of (text|json)")
	f.StringVar(&ec.OIDC.IssuerURL, "oidc-issuer-url", "https://keycloak.netic.dk/auth/realms/services", "Issuer URL for the OIDC Provider")
	f.StringVar(&ec.OIDC.ClientID, "oidc-client-id", "inventory-cli", "OIDC client ID")
	f.StringVar(&ec.OIDC.GrantType, "oidc-grant-type", "authcode-browser", "OIDC authorization grant type. One of (authcode-browser|authcode-keyboard)")
	f.StringVar(&ec.OIDC.RedirectURLHostname, "oidc-redirect-url-hostname", "localhost", "[authcode-browser] Hostname of the redirect URL")
	f.StringVar(&ec.OIDC.AuthBindAddr, "oidc-auth-bind-addr", "localhost:18000", "[authcode-browser] Bind address and port for local server used for OIDC redirect")
	f.StringVar(&ec.OIDC.RedirectURIAuthCodeKeyboard, "oidc-redirect-uri-authcode-keyboard", oobRedirectURI, "[authcode-keyboard] Redirect URI when using authcode keyboard")
	f.StringVar(&ec.OIDC.TokenCacheDir, "oidc-token-cache-dir", getDefaultTokenCacheDir(), "Directory used to store cached tokens")

	viper.BindPFlags(rootCmd.PersistentFlags()) // nolint:errcheck

	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.AddGroup(
		&cobra.Group{
			ID:    groupBase,
			Title: "Basic Commands:",
		},
		&cobra.Group{
			ID:    groupAuth,
			Title: "Authentication Commands:",
		},
		&cobra.Group{
			ID:    groupOther,
			Title: "Other Commands:",
		},
	)

	rootCmd.SetHelpCommandGroupID(groupOther)
	rootCmd.SetCompletionCommandGroupID(groupOther)

	rootCmd.AddCommand(
		NewLoginCmd(ec),
		NewLogoutCmd(ec),
		NewGetCmd(ec),
	)
}

func Execute(args []string, version string) int {
	rootCmd.Version = version
	rootCmd.SilenceUsage = true
	rootCmd.SetArgs(args[1:])
	if err := ec.Prepare(); err != nil {
		ec.Logger.Error("Preparing execution context", "err", err)
		return 1
	}
	err := rootCmd.ExecuteContext(context.Background())
	if ec.Spinner.Running() {
		ec.Spinner.Stop()
	}
	if err != nil {
		fmt.Fprintln(ec.Stderr, err)
		return 1
	}
	return 0
}
