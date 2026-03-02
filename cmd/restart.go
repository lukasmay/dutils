package cmd

import (
	"fmt"
	"os"
	"os/exec"

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

		composeFiles := proj.GetComposeFiles()
		if len(composeFiles) == 0 {
			fmt.Println("No compose files found.")
			os.Exit(1)
		}

		expandedTargets := expandTargets(proj, args)

		for _, file := range composeFiles {
			fileServices := getServicesForFile(file, expandedTargets)
			if len(expandedTargets) > 0 && len(fileServices) == 0 {
				continue
			}

			// Build
			buildArgs := []string{"compose", "-f", file, "build", "--no-cache"}
			buildArgs = append(buildArgs, fileServices...)
			fmt.Printf("Building services in %s...\n", file)
			bCmd := exec.Command("docker", buildArgs...)
			bCmd.Stdout = os.Stdout
			bCmd.Stderr = os.Stderr
			bCmd.Run()

			// Up
			upArgs := []string{"compose", "-f", file, "up", "-d", "--force-recreate"}
			upArgs = append(upArgs, fileServices...)
			fmt.Printf("Recreating services in %s...\n", file)
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
	rootCmd.AddCommand(restartCmd)
}
