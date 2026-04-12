package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "dutils",
	Version: Version,
	Short:   "A CLI tool to simplify Docker and Docker Compose workflows",
	Long: `dutils simplifies Docker and Docker Compose workflows with
optimized subcommands and tab auto-completion for container management.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
