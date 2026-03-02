package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetRegistryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "dutils", "registry")
}

func AddToRegistry(name, path string) error {
	registryPath := GetRegistryPath()
	os.MkdirAll(filepath.Dir(registryPath), 0755)

	projects, err := ReadRegistry()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	projects[name] = path

	f, err := os.Create(registryPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for k, v := range projects {
		fmt.Fprintf(f, "%s:%s\n", k, v)
	}
	return nil
}

func ReadRegistry() (map[string]string, error) {
	registryPath := GetRegistryPath()
	projects := make(map[string]string)

	f, err := os.Open(registryPath)
	if err != nil {
		return projects, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			projects[parts[0]] = parts[1]
		}
	}
	return projects, scanner.Err()
}

func SetActiveProject(path string) error {
	home, _ := os.UserHomeDir()
	activePath := filepath.Join(home, ".config", "dutils", "active")
	os.MkdirAll(filepath.Dir(activePath), 0755)
	return os.WriteFile(activePath, []byte(path), 0644)
}

func ClearActiveProject() error {
	home, _ := os.UserHomeDir()
	activePath := filepath.Join(home, ".config", "dutils", "active")
	return os.Remove(activePath)
}
