package cmd

import (
	"fmt"
	"os"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear active project override",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.ClearActiveProject(); err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		fmt.Println("Active project cleared")
	},
}

func init() {
	projectCmd.AddCommand(clearCmd)
}
