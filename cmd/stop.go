package cmd

import (
	"fmt"
	"os"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:               "stop [services|@groups]",
	Short:             "Stop services or containers",
	ValidArgsFunction: completeServices,
	Run: func(cmd *cobra.Command, args []string) {
		down, _ := cmd.Flags().GetBool("down")

		proj, err := config.ResolveProject()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := runStop(proj, args, down); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runStop(proj *config.ProjectInfo, args []string, down bool) error {
	composeFiles := proj.GetComposeFiles()
	if len(composeFiles) == 0 {
		if len(args) == 0 {
			return fmt.Errorf("no services specified and no compose files found")
		}
		for _, container := range args {
			if down {
				if err := runDockerCommand("rm", "-s", "-f", container); err != nil {
					return err
				}
			} else {
				if err := runDockerCommand("stop", container); err != nil {
					return err
				}
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

		var stopArgs []string
		if down {
			if len(services) == 0 {
				stopArgs = []string{"compose", "-f", file, "down"}
			} else {
				stopArgs = append([]string{"compose", "-f", file, "rm", "-s", "-f"}, services...)
			}
		} else {
			stopArgs = append([]string{"compose", "-f", file, "stop"}, services...)
		}

		if err := runDockerCommand(stopArgs...); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolP("down", "d", false, "Remove services (rm -s -f) or down whole compose file")
}
