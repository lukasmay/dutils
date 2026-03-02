package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initProjCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .dutils.yml in current project and register it",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		configPath := filepath.Join(cwd, ".dutils.yml")
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("File %s already exists. Using 'project add' to register it.\n", configPath)
			addCmd.Run(cmd, []string{cwd})
			return
		}

		projName := filepath.Base(cwd)
		content := fmt.Sprintf(`project_name: %s

# --- dutils configuration ---
#
# Define your groups here. You can reference these in commands using the @ prefix.
# groups:
#   frontend:
#     - web
#     - nginx
#
# Override which compose files to use:
# compose:
#   files:
#     - docker-compose.yml
`, projName)

		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Initialized dutils project at %s\n", cwd)
		addCmd.Run(cmd, []string{cwd})
		switchCmd.Run(cmd, []string{projName})
	},
}

func init() {
	projectCmd.AddCommand(initProjCmd)
}
