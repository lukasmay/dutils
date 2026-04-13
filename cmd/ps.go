package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type containerRow struct {
	Name   string
	Status string
	Ports  string
}

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List containers in a formatted table",
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		if err := runDlist(all, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runDlist(all bool, w io.Writer) error {
	dockerArgs := []string{"ps", "--format", "{{.Names}}\t{{.Status}}\t{{.Ports}}"}
	if all {
		dockerArgs = append(dockerArgs, "-a")
	}

	out, err := exec.Command("docker", dockerArgs...).Output()
	if err != nil {
		return fmt.Errorf("error running docker ps: %w", err)
	}

	rows := parseContainerRows(strings.TrimSpace(string(out)))
	if len(rows) == 0 {
		fmt.Fprintln(w, "No containers running.")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tPORTS")
	fmt.Fprintln(tw, "----\t------\t-----")
	for _, row := range rows {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", row.Name, row.Status, row.Ports)
	}
	return tw.Flush()
}

func parseContainerRows(output string) []containerRow {
	if output == "" {
		return nil
	}
	lines := strings.Split(output, "\n")
	rows := make([]containerRow, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		rows = append(rows, containerRow{
			Name:   parts[0],
			Status: parts[1],
			Ports:  parts[2],
		})
	}
	return rows
}

func init() {
	rootCmd.AddCommand(psCmd)
	psCmd.Flags().BoolP("all", "a", false, "Show all containers including stopped ones")
}
