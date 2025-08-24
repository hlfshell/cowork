package main

import (
	"fmt"
	"os"

	"github.com/hlfshell/cowork/internal/cli"
)

// Version information for the cowork CLI
const (
	Version   = "0.1.0"
	BuildDate = "2024-01-01"
	GitCommit = "development"
)

func main() {
	// Create the CLI application with version information
	app := cli.NewApp(Version, BuildDate, GitCommit)

	// Run the CLI application
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
