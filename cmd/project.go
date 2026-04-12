package cmd

import "github.com/spf13/cobra"

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage dutils projects",
	Long:  `Register, list, and switch between dutils projects.`,
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
