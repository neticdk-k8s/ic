package main

import (
	"context"
	"os"

	"github.com/neticdk-k8s/k8s-inventory-cli/cmd"
)

var version = "HEAD"

func main() {
	os.Exit(cmd.NewCmd().Run(context.Background(), os.Args, version))
}
