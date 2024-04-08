package ui

import (
	"os"

	"github.com/mattn/go-isatty"
)

var isInteractive = IsTerminal(os.Stdin) && IsTerminal(os.Stdout)

func IsTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
