package main

import (
	"fmt"
	"os"

	"github.com/neticdk-k8s/k8s-inventory-cli/cmd"
)

func main() {
	cmd := cmd.NewCmd()
	if err := cmd.GenDocs(); err != nil {
		fmt.Println("Failed docs generation")
		os.Exit(1)
	}
}
