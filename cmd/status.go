package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the currently active project",
	Run: func(cmd *cobra.Command, args []string) {
		runStatus(os.Stdout)
	},
}

func runStatus(w io.Writer) {
	activeRoot, err := config.ReadActiveProject()
	if err != nil || activeRoot == "" {
		fmt.Fprintln(w, "No active project set. dutils will operate on the current directory.")
		return
	}

	projects, _ := config.ReadRegistry()
	projectName := "unknown"
	for name, path := range projects {
		if path == activeRoot {
			projectName = name
			break
		}
	}

	fmt.Fprintf(w, "Active Project: %s\n", projectName)
	fmt.Fprintf(w, "Location:       %s\n", activeRoot)
}

func init() {
	projectCmd.AddCommand(statusCmd)
}
