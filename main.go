package main

import (
	"os"

	"github.com/neticdk-k8s/k8s-inventory-cli/cmd"
)

var version = "HEAD"

func main() {
	os.Exit(cmd.Execute(os.Args, version))
}
