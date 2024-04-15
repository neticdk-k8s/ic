package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra/doc"
)

// GenDocs generates the CLI documentation
func GenDocs() error {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("could not get filename of caller")
		os.Exit(1)
	}
	docPath := path.Dir(filename)
	fmt.Printf("Generating documentation in: %s\n", docPath)
	in := ExecutionContextInput{
		Version: "doc",
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
	ec := NewExecutionContext(in)
	rootCmd := NewRootCmd(ec)
	return doc.GenMarkdownTree(rootCmd, docPath)
}
