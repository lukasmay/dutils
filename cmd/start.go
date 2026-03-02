package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:               "start [services|@groups]",
	Short:             "Start services or containers",
	ValidArgsFunction: completeServices,
	Run: func(cmd *cobra.Command, args []string) {
		build, _ := cmd.Flags().GetBool("build")
		proj, err := config.ResolveProject()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		composeFiles := proj.GetComposeFiles()
		if len(composeFiles) == 0 {
			// Fallback to docker start if no compose files found
			if len(args) == 0 {
				fmt.Println("No services/containers specified and no compose files found.")
				os.Exit(1)
			}
			for _, container := range args {
				dockerCmd := exec.Command("docker", "start", container)
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

			if build {
				buildArgs := []string{"compose", "-f", file, "build", "--no-cache"}
				buildArgs = append(buildArgs, fileServices...)
				bCmd := exec.Command("docker", buildArgs...)
				bCmd.Stdout = os.Stdout
				bCmd.Stderr = os.Stderr
				bCmd.Run()
			}

			upArgs := []string{"compose", "-f", file, "up", "-d"}
			upArgs = append(upArgs, fileServices...)
			uCmd := exec.Command("docker", upArgs...)
			uCmd.Stdout = os.Stdout
			uCmd.Stderr = os.Stderr
			if err := uCmd.Run(); err != nil {
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("build", "b", false, "Build with --no-cache before starting")
}
