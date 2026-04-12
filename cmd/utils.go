package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

func runDockerCommand(args ...string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func completeServices(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	proj, err := config.ResolveProject()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var matches []string

	if proj.Config != nil {
		for group := range proj.Config.Groups {
			name := "@" + group
			if strings.HasPrefix(name, toComplete) {
				matches = append(matches, name)
			}
		}
	}

	seen := make(map[string]bool)
	for _, file := range proj.GetComposeFiles() {
		out, err := exec.Command("docker", "compose", "-f", file, "config", "--services").Output()
		if err != nil {
			continue
		}
		for _, s := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if s != "" && strings.HasPrefix(s, toComplete) && !seen[s] {
				matches = append(matches, s)
				seen[s] = true
			}
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}

func expandTargets(proj *config.ProjectInfo, inputs []string) []string {
	var expanded []string
	for _, input := range inputs {
		if strings.HasPrefix(input, "@") {
			if proj.Config == nil {
				fmt.Printf("Warning: no config loaded, cannot expand group %s\n", input)
				continue
			}
			groupName := input[1:]
			if members, ok := proj.Config.Groups[groupName]; ok {
				expanded = append(expanded, members...)
			} else {
				fmt.Printf("Warning: unknown group %s\n", input)
			}
		} else {
			expanded = append(expanded, input)
		}
	}
	return expanded
}

// servicesInFile returns which of the given targets exist as services in the compose file.
func servicesInFile(file string, targets []string) []string {
	if len(targets) == 0 {
		return nil
	}
	out, err := exec.Command("docker", "compose", "-f", file, "config", "--services").Output()
	if err != nil {
		return nil
	}
	available := make(map[string]bool)
	for _, s := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		available[s] = true
	}
	var matched []string
	for _, t := range targets {
		if available[t] {
			matched = append(matched, t)
		}
	}
	return matched
}
