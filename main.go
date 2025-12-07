// ABOUTME: Entry point for gh-context CLI extension
// ABOUTME: Provides kubectx-style GitHub account switching via gh CLI

package main

import (
	"os"

	"github.com/peterjmorgan/gh-context/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
