package cmd

import (
	"fmt"
	"os"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show resolved project root and source",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := config.ResolveProject()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Project root: %s\n", proj.Root)
		fmt.Printf("Source: %s\n", proj.Source)
	},
}

func init() {
	projectCmd.AddCommand(showCmd)
}
