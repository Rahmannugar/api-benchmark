package main

import (
	"fmt"
	"os"

	"github.com/mbrik/CLI-Benchmarking-Tool/cmd"
)

func main() {
	// Execute the CLI command handler
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
