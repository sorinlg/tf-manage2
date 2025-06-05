package main

import (
	"fmt"
	"os"

	"github.com/sorinlg/tf-manage2/internal/cli"
)

// Version information - set by GoReleaser during build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Set version info for CLI
	cli.SetVersionInfo(version, commit, date, builtBy)

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
