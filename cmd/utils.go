package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

func completeContainers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	out, err := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}").Output()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	names := strings.Split(strings.TrimSpace(string(out)), "\n")
	var matches []string
	for _, name := range names {
		if name != "" && strings.HasPrefix(name, toComplete) {
			matches = append(matches, name)
		}
	}
	return matches, cobra.ShellCompDirectiveNoFileComp
}

func completeServices(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	proj, err := config.ResolveProject()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var matches []string

	// Add groups with @ prefix
	if proj.Config != nil {
		for group := range proj.Config.Groups {
			name := "@" + group
			if strings.HasPrefix(name, toComplete) {
				matches = append(matches, name)
			}
		}
	}

	// Add services from compose files
	composeFiles := proj.GetComposeFiles()
	seen := make(map[string]bool)
	for _, file := range composeFiles {
		out, err := exec.Command("docker", "compose", "-f", file, "config", "--services").Output()
		if err != nil {
			continue
		}
		services := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, s := range services {
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
		if strings.HasPrefix(input, "@") && proj.Config != nil {
			groupName := input[1:]
			if group, ok := proj.Config.Groups[groupName]; ok {
				expanded = append(expanded, group...)
			} else {
				fmt.Printf("Warning: Unknown group %s\n", input)
			}
		} else if !strings.HasPrefix(input, "@") {
			expanded = append(expanded, input)
		}
	}
	return expanded
}

func getServicesForFile(file string, targets []string) []string {
	if len(targets) == 0 {
		return nil
	}
	out, err := exec.Command("docker", "compose", "-f", file, "config", "--services").Output()
	if err != nil {
		return nil
	}
	allServices := strings.Split(strings.TrimSpace(string(out)), "\n")
	serviceMap := make(map[string]bool)
	for _, s := range allServices {
		serviceMap[s] = true
	}

	var found []string
	for _, t := range targets {
		if serviceMap[t] {
			found = append(found, t)
		}
	}
	return found
}
