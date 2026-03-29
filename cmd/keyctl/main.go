// Package main is the entry point for keyctl CLI.
package main

import (
	"fmt"
	"os"

	"github.com/skys-mission/key-agent/internal/client/commands"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := commands.Execute(version, commit, date); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
