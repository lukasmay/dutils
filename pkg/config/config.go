package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ComposeConfig struct {
	Files []string `yaml:"files"`
}

type Config struct {
	ProjectName string              `yaml:"project_name"`
	Groups      map[string][]string `yaml:"groups"`
	Compose     ComposeConfig       `yaml:"compose"`
	Defaults    struct {
		Ps struct {
			Scope string `yaml:"scope"`
		} `yaml:"ps"`
	} `yaml:"defaults"`
}

type ProjectInfo struct {
	Root   string
	Source string
	Config *Config
}

func dutilsConfigDir() string {
	if dir := os.Getenv("DUTILS_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "dutils")
}

func ActiveProjectPath() string {
	return filepath.Join(dutilsConfigDir(), "active")
}

func ReadActiveProject() (string, error) {
	data, err := os.ReadFile(ActiveProjectPath())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func ResolveProject() (*ProjectInfo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	gitRoot, _ := getGitRoot(cwd)

	// 1. .dutils.yml at git root
	if gitRoot != "" {
		if proj, err := loadProjectAt(gitRoot, "git-config"); err != nil {
			return nil, err
		} else if proj != nil {
			return proj, nil
		}
	}

	// 2. .dutils.yml in current directory
	if proj, err := loadProjectAt(cwd, "pwd-config"); err != nil {
		return nil, err
	} else if proj != nil {
		return proj, nil
	}

	// 3. DUTILS_PROJECT_ROOT environment variable
	if envRoot := os.Getenv("DUTILS_PROJECT_ROOT"); envRoot != "" {
		if _, err := os.Stat(envRoot); err == nil {
			if proj, err := loadProjectAt(envRoot, "env"); err != nil {
				return nil, err
			} else if proj != nil {
				return proj, nil
			}
			return &ProjectInfo{Root: envRoot, Source: "env"}, nil
		}
	}

	// 4. Active project file (~/.config/dutils/active)
	if activeRoot, err := ReadActiveProject(); err == nil && activeRoot != "" {
		if _, err := os.Stat(activeRoot); err == nil {
			if proj, err := loadProjectAt(activeRoot, "active"); err != nil {
				return nil, err
			} else if proj != nil {
				return proj, nil
			}
			return &ProjectInfo{Root: activeRoot, Source: "active"}, nil
		}
	}

	// 5. Git root without config (fallback)
	if gitRoot != "" {
		return &ProjectInfo{Root: gitRoot, Source: "git"}, nil
	}

	// 6. Current directory (fallback)
	return &ProjectInfo{Root: cwd, Source: "pwd"}, nil
}

// loadProjectAt looks for a .dutils.yml at root and returns a ProjectInfo if one exists.
// Returns (nil, nil) when no config file is present at root.
// Returns (nil, error) when a config file exists but cannot be parsed.
func loadProjectAt(root, source string) (*ProjectInfo, error) {
	cfgPath := filepath.Join(root, ".dutils.yml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, nil
	}
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", cfgPath, err)
	}
	return &ProjectInfo{Root: root, Source: source, Config: cfg}, nil
}

func getGitRoot(path string) (string, error) {
	out, err := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (p *ProjectInfo) GetComposeFiles() []string {
	if p.Config != nil && len(p.Config.Compose.Files) > 0 {
		var files []string
		for _, f := range p.Config.Compose.Files {
			if filepath.IsAbs(f) {
				files = append(files, f)
			} else {
				files = append(files, filepath.Join(p.Root, f))
			}
		}
		return files
	}

	candidates := []string{"compose.yml", "compose.yaml", "docker-compose.yml", "docker-compose.yaml"}
	var found []string
	for _, name := range candidates {
		path := filepath.Join(p.Root, name)
		if _, err := os.Stat(path); err == nil {
			found = append(found, path)
		}
	}
	return found
}

func (p *ProjectInfo) GetProjectName() string {
	if p.Config != nil && p.Config.ProjectName != "" {
		return p.Config.ProjectName
	}
	name := strings.ToLower(filepath.Base(p.Root))
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, name)
}
