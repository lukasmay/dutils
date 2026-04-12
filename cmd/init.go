package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var initProjCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .dutils.yml in current project and register it",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		if err := runInit(cwd, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runInit(cwd string, w io.Writer) error {
	configPath := filepath.Join(cwd, ".dutils.yml")
	projName := filepath.Base(cwd)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		content := fmt.Sprintf(`project_name: %s

# --- dutils configuration ---
#
# Define service groups, referenced in commands with the @ prefix.
# groups:
#   frontend:
#     - web
#     - nginx
#
# Override which compose files dutils uses:
# compose:
#   files:
#     - docker-compose.yml
`, projName)
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("creating config: %w", err)
		}
		fmt.Fprintf(w, "Initialized dutils project at %s\n", cwd)
	} else {
		fmt.Fprintf(w, "%s already exists, registering existing config.\n", configPath)
	}

	if cfg, err := config.LoadConfig(configPath); err == nil && cfg.ProjectName != "" {
		projName = cfg.ProjectName
	}

	if err := config.AddToRegistry(projName, cwd); err != nil {
		return fmt.Errorf("registering project: %w", err)
	}

	if err := config.SetActiveProject(cwd); err != nil {
		return fmt.Errorf("setting active project: %w", err)
	}

	fmt.Fprintf(w, "Active project set to '%s'\n", projName)
	return nil
}

func init() {
	projectCmd.AddCommand(initProjCmd)
}
