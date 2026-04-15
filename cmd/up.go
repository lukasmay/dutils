package cmd

import (
	"fmt"
	"os"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:               "up [services|@groups]",
	Short:             "Start services or containers",
	ValidArgsFunction: completeServices,
	Run: func(cmd *cobra.Command, args []string) {
		build, _ := cmd.Flags().GetBool("build")

		proj, err := config.ResolveProject()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := runUp(proj, args, build); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runUp(proj *config.ProjectInfo, args []string, build bool) error {
	composeFiles := proj.GetComposeFiles()
	if len(composeFiles) == 0 {
		if len(args) == 0 {
			return fmt.Errorf("no services specified and no compose files found")
		}
		for _, container := range args {
			if err := runDockerCommand("start", container); err != nil {
				return err
			}
		}
		return nil
	}

	targets := expandTargets(proj, args)

	for _, file := range composeFiles {
		services := servicesInFile(file, targets)
		if len(targets) > 0 && len(services) == 0 {
			continue
		}

		if build {
			buildArgs := append([]string{"compose", "-f", file, "build", "--no-cache"}, services...)
			if err := runDockerCommand(buildArgs...); err != nil {
				return fmt.Errorf("build failed for %s: %w", file, err)
			}
		}

		upArgs := append([]string{"compose", "-f", file, "up", "-d"}, services...)
		if err := runDockerCommand(upArgs...); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().BoolP("build", "b", false, "Build with --no-cache before starting")
}
