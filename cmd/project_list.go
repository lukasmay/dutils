package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered projects",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runProjectList(os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runProjectList(w io.Writer) error {
	projects, err := config.ReadRegistry()
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		fmt.Fprintln(w, "No projects registered. Run 'project init' or 'project add'.")
		return nil
	}

	activeRoot, _ := config.ReadActiveProject()

	for name, path := range projects {
		if path == activeRoot {
			fmt.Fprintf(w, "* %s (%s)\n", name, path)
		} else {
			fmt.Fprintf(w, "  %s (%s)\n", name, path)
		}
	}
	return nil
}

func init() {
	projectCmd.AddCommand(projectListCmd)
}
