package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/tokencache"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/usecases/authentication"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// CLI represents a Command Line Interface
type CLI interface {
	Run(ctx context.Context, args []string, version string) int
	GenDocs() error
}

type cli struct {
	Root   *Root
	Login  *Login
	Logout *Logout
	Logger logger.Logger
}

// NewCLI creates a new CLI
func NewCLI() CLI {
	logger := logger.New(os.Stderr, "info")

	tokenCache, err := tokencache.NewFSCache()
	if err != nil {
		logger.Error("Setting up token cache", "err", err)
		os.Exit(1)
	}

	authenticator := authentication.NewAuthenticator(logger)

	root := &Root{
		Logger: logger,
	}
	login := &Login{
		Authenticator: authenticator,
		TokenCache:    tokenCache,
		Logger:        logger,
	}
	logout := &Logout{
		Authenticator: authenticator,
		TokenCache:    tokenCache,
		Logger:        logger,
	}
	return &cli{
		Root:   root,
		Login:  login,
		Logout: logout,
		Logger: logger,
	}
}

// Run runs the command
func (c *cli) Run(ctx context.Context, args []string, version string) int {
	rootCmd := c.buildRootCmd(args, version)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return 1
	}

	return 0
}

func (c *cli) buildRootCmd(args []string, version string) *cobra.Command {
	rootCmd := c.Root.New()
	rootCmd.Version = version
	rootCmd.SilenceUsage = true

	loginCmd := c.Login.New()
	rootCmd.AddCommand(loginCmd)

	logoutCmd := c.Logout.New()
	rootCmd.AddCommand(logoutCmd)

	rootCmd.SetArgs(args[1:])
	return rootCmd
}

// GenDocs generates the CLI documentation
func (c *cli) GenDocs() error {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("could not get filename of caller")
		os.Exit(1)
	}
	docPath := path.Dir(filename)
	fmt.Printf("Generating documentation in: %s\n", docPath)
	return doc.GenMarkdownTree(c.buildRootCmd([]string{""}, "HEAD"), docPath)
}
