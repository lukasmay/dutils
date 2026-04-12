package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func registryPath() string {
	return filepath.Join(dutilsConfigDir(), "registry.json")
}

func AddToRegistry(name, path string) error {
	if err := os.MkdirAll(dutilsConfigDir(), 0755); err != nil {
		return err
	}
	projects, err := ReadRegistry()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	projects[name] = path
	return writeRegistry(projects)
}

func ReadRegistry() (map[string]string, error) {
	data, err := os.ReadFile(registryPath())
	if err != nil {
		if os.IsNotExist(err) {
			return migrateLegacyRegistry()
		}
		return make(map[string]string), err
	}
	var projects map[string]string
	if err := json.Unmarshal(data, &projects); err != nil {
		return make(map[string]string), err
	}
	return projects, nil
}

// migrateLegacyRegistry reads the old name:path text format and writes it to the
// new JSON format. The legacy file is removed after a successful migration.
func migrateLegacyRegistry() (map[string]string, error) {
	legacyPath := filepath.Join(dutilsConfigDir(), "registry")
	data, err := os.ReadFile(legacyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return make(map[string]string), err
	}
	projects := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && parts[0] != "" {
			projects[parts[0]] = parts[1]
		}
	}
	if len(projects) > 0 {
		if err := writeRegistry(projects); err != nil {
			return projects, err
		}
		os.Remove(legacyPath)
	}
	return projects, nil
}

func writeRegistry(projects map[string]string) error {
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(registryPath(), data, 0644)
}

func SetActiveProject(path string) error {
	if err := os.MkdirAll(dutilsConfigDir(), 0755); err != nil {
		return err
	}
	return os.WriteFile(ActiveProjectPath(), []byte(path), 0644)
}

func ClearActiveProject() error {
	return os.Remove(ActiveProjectPath())
}
