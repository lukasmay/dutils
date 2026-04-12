package cmd

import (
	"fmt"
	"os"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:               "restart [services|@groups]",
	Short:             "Rebuild and force-recreate services",
	ValidArgsFunction: completeServices,
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := config.ResolveProject()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := runRestart(proj, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runRestart(proj *config.ProjectInfo, args []string) error {
	composeFiles := proj.GetComposeFiles()
	if len(composeFiles) == 0 {
		return fmt.Errorf("no compose files found")
	}

	targets := expandTargets(proj, args)

	for _, file := range composeFiles {
		services := servicesInFile(file, targets)
		if len(targets) > 0 && len(services) == 0 {
			continue
		}

		buildArgs := append([]string{"compose", "-f", file, "build", "--no-cache"}, services...)
		fmt.Printf("Building services in %s...\n", file)
		if err := runDockerCommand(buildArgs...); err != nil {
			return fmt.Errorf("build failed for %s: %w", file, err)
		}

		upArgs := append([]string{"compose", "-f", file, "up", "-d", "--force-recreate"}, services...)
		fmt.Printf("Recreating services in %s...\n", file)
		if err := runDockerCommand(upArgs...); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
