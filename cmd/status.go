package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the currently active project",
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Could not get home directory: %v\n", err)
			return
		}

		activePath := filepath.Join(home, ".config", "dutils", "active")

		data, err := os.ReadFile(activePath)
		if err != nil || len(strings.TrimSpace(string(data))) == 0 {
			fmt.Println("No project is currently set as active globally. dutils will operate on the current directory.")
			return
		}

		activeRoot := strings.TrimSpace(string(data))

		// Let's try to map this back to a registered project name
		projects, _ := config.ReadRegistry()
		projectName := "Unknown"
		for name, path := range projects {
			if path == activeRoot {
				projectName = name
				break
			}
		}

		fmt.Printf("Active Project: %s\n", projectName)
		fmt.Printf("Location:       %s\n", activeRoot)
	},
}

func init() {
	projectCmd.AddCommand(statusCmd)
}
