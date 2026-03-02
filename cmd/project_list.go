package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered projects",
	Run: func(cmd *cobra.Command, args []string) {
		projects, err := config.ReadRegistry()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		home, _ := os.UserHomeDir()
		activePath := filepath.Join(home, ".config", "dutils", "active")
		activeRoot := ""
		if data, err := os.ReadFile(activePath); err == nil {
			activeRoot = strings.TrimSpace(string(data))
		}

		if len(projects) == 0 {
			fmt.Println("No projects registered. Run 'project init' or 'project add'.")
			return
		}

		for name, path := range projects {
			if path == activeRoot {
				fmt.Printf("* %s (%s)\n", name, path)
			} else {
				fmt.Printf("  %s (%s)\n", name, path)
			}
		}
	},
}

func init() {
	projectCmd.AddCommand(projectListCmd)
}
