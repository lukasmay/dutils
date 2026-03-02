package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Register a project",
	Run: func(cmd *cobra.Command, args []string) {
		targetDir := "."
		if len(args) > 0 {
			targetDir = args[0]
		}

		absPath, err := filepath.Abs(targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cfgPath := filepath.Join(absPath, ".dutils.yml")
		projName := filepath.Base(absPath)

		if cfg, err := config.LoadConfig(cfgPath); err == nil && cfg.ProjectName != "" {
			projName = cfg.ProjectName
		}

		if err := config.AddToRegistry(projName, absPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Added project '%s' -> %s\n", projName, absPath)
	},
}

func init() {
	projectCmd.AddCommand(addCmd)
}
