package cmd

import (
	"fmt"
	"os"
	"os/exec"

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

		composeFiles := proj.GetComposeFiles()
		if len(composeFiles) == 0 {
			// Fallback to docker stop/rm if no compose files found
			if len(args) == 0 {
				fmt.Println("No services/containers specified and no compose files found.")
				os.Exit(1)
			}
			for _, container := range args {
				action := "stop"
				argsArr := []string{action}
				if down {
					action = "rm"
					argsArr = []string{action, "-s", "-f"}
				}
				argsArr = append(argsArr, container)
				dockerCmd := exec.Command("docker", argsArr...)
				dockerCmd.Stdout = os.Stdout
				dockerCmd.Stderr = os.Stderr
				dockerCmd.Run()
			}
			return
		}

		expandedTargets := expandTargets(proj, args)

		for _, file := range composeFiles {
			fileServices := getServicesForFile(file, expandedTargets)
			if len(expandedTargets) > 0 && len(fileServices) == 0 {
				continue
			}

			var stopArgs []string
			if down {
				if len(fileServices) == 0 {
					stopArgs = []string{"compose", "-f", file, "down"}
				} else {
					stopArgs = []string{"compose", "-f", file, "rm", "-s", "-f"}
					stopArgs = append(stopArgs, fileServices...)
				}
			} else {
				stopArgs = []string{"compose", "-f", file, "stop"}
				stopArgs = append(stopArgs, fileServices...)
			}

			sCmd := exec.Command("docker", stopArgs...)
			sCmd.Stdout = os.Stdout
			sCmd.Stderr = os.Stderr
			if err := sCmd.Run(); err != nil {
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolP("down", "d", false, "Remove services (rm -s -f) or down whole compose file")
}
