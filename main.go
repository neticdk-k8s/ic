package main

import (
	"context"
	"os"

	"github.com/neticdk-k8s/k8s-inventory-cli/cmd"
)

var version = "HEAD"

func main() {
	os.Exit(cmd.NewCLI().Run(context.Background(), os.Args, version))
}
