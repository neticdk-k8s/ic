package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultConfigFilename = "ic"
	envPrefix             = "IC"
	oobRedirectURI        = "urn:ietf:wg:oauth:2.0:oob"
)

type rootOptions struct {
	LogLevel    string
	Server      string
	Interactive string
}

func (o *rootOptions) addFlags(f *pflag.FlagSet) {
	f.StringVarP(&o.LogLevel, "log-level", "l", "info", "Set log level")
	_ = viper.BindPFlag("log-level", f.Lookup("log-level"))

	f.StringVarP(&o.Server, "server", "s", "http://localhost:8086", "URL for the inventory server.")
	_ = viper.BindPFlag("server", f.Lookup("server"))

	f.StringVarP(&o.Interactive, "interactive", "i", "auto", "Run in interactive mode. One of (yes|no|auto)")
	_ = viper.BindPFlag("interactive", f.Lookup("interactive"))
}

// Root represents the root command
type Root struct {
	Logger logger.Logger
}

// New creates a new Root command
func (c *Root) New() *cobra.Command {
	var o rootOptions
	command := &cobra.Command{
		Use:   "ic",
		Short: "Inventory Client",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := initConfig(cmd); err != nil {
				return err
			}
			c.initLog(cmd)
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				os.Exit(0)
			}
			return nil
		},
	}
	command.Flags().SortFlags = false
	o.addFlags(command.PersistentFlags())
	return command
}

func (c *Root) initLog(cmd *cobra.Command) {
	logLevel, err := cmd.Flags().GetString("log-level")
	if err == nil {
		if err := c.Logger.SetLevel(logLevel); err != nil {
			c.Logger.Error("Failed to set loglevel", "level", logLevel)
		}
	}
	interactive, err := cmd.Flags().GetString("interactive")
	if err == nil {
		c.Logger.SetInteractive(interactive)
	}
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
