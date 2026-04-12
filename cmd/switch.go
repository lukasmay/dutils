package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

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
			if strings.HasPrefix(name, toComplete) {
				matches = append(matches, name)
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSwitch(args[0], os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runSwitch(name string, w io.Writer) error {
	projects, err := config.ReadRegistry()
	if err != nil {
		return err
	}

	path, ok := projects[name]
	if !ok {
		return fmt.Errorf("project '%s' not found in registry", name)
	}

	if err := config.SetActiveProject(path); err != nil {
		return err
	}

	fmt.Fprintf(w, "Switched active project to '%s' (%s)\n", name, path)
	return nil
}

func init() {
	projectCmd.AddCommand(switchCmd)
}
