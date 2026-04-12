package cmd

import (
	"fmt"
	"io"
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
		if err := runAdd(targetDir, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runAdd(targetDir string, w io.Writer) error {
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}

	cfgPath := filepath.Join(absPath, ".dutils.yml")
	projName := filepath.Base(absPath)

	if cfg, err := config.LoadConfig(cfgPath); err == nil && cfg.ProjectName != "" {
		projName = cfg.ProjectName
	}

	if err := config.AddToRegistry(projName, absPath); err != nil {
		return err
	}

	fmt.Fprintf(w, "Added project '%s' -> %s\n", projName, absPath)
	return nil
}

func init() {
	projectCmd.AddCommand(addCmd)
}
