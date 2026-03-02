package cmd

import (
	"fmt"
	"os"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch active project",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		projects, err := config.ReadRegistry()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		var matches []string
		for name := range projects {
			if len(toComplete) == 0 || (len(name) >= len(toComplete) && name[:len(toComplete)] == toComplete) {
				matches = append(matches, name)
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		projects, err := config.ReadRegistry()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		path, ok := projects[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: project '%s' not found in registry\n", name)
			os.Exit(1)
		}

		if err := config.SetActiveProject(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Switched active project to '%s' (%s)\n", name, path)
	},
}

func init() {
	projectCmd.AddCommand(switchCmd)
}
