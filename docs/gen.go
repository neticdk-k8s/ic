package main

import (
	"fmt"
	"os"

	"github.com/neticdk-k8s/ic/cmd"
)

func main() {
	if err := cmd.GenDocs(); err != nil {
		fmt.Println("Failed docs generation")
		os.Exit(1)
	}
}
