package main

import (
	"os"

	"github.com/neticdk-k8s/ic/cmd"
)

var version = "HEAD"

func main() {
	os.Exit(cmd.Execute(os.Args, version))
}
