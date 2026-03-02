package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ProjectName string              `yaml:"project_name"`
	Groups      map[string][]string `yaml:"groups"`
	Compose     struct {
		Files []string `yaml:"files"`
	} `yaml:"compose"`
	Defaults struct {
		Dlist struct {
			Scope string `yaml:"scope"`
		} `yaml:"dlist"`
	} `yaml:"defaults"`
}

type ProjectInfo struct {
	Root   string
	Source string
	Config *Config
}

func ResolveProject() (*ProjectInfo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// 1. Check for git root and .dutils.yml there
	gitRoot, _ := getGitRoot(cwd)
	if gitRoot != "" {
		cfgPath := filepath.Join(gitRoot, ".dutils.yml")
		if _, err := os.Stat(cfgPath); err == nil {
			cfg, _ := LoadConfig(cfgPath)
			return &ProjectInfo{Root: gitRoot, Source: "git-config", Config: cfg}, nil
		}
	}

	// 2. Check for .dutils.yml in current dir
	cfgPath := filepath.Join(cwd, ".dutils.yml")
	if _, err := os.Stat(cfgPath); err == nil {
		cfg, _ := LoadConfig(cfgPath)
		return &ProjectInfo{Root: cwd, Source: "pwd-config", Config: cfg}, nil
	}

	// 3. Environment variable DUTILS_PROJECT_ROOT
	if envRoot := os.Getenv("DUTILS_PROJECT_ROOT"); envRoot != "" {
		if _, err := os.Stat(envRoot); err == nil {
			cfgPath := filepath.Join(envRoot, ".dutils.yml")
			cfg, _ := LoadConfig(cfgPath)
			return &ProjectInfo{Root: envRoot, Source: "env", Config: cfg}, nil
		}
	}

	// 4. Active project file ~/.config/dutils/active
	home, _ := os.UserHomeDir()
	activePath := filepath.Join(home, ".config", "dutils", "active")
	if data, err := os.ReadFile(activePath); err == nil {
		activeRoot := strings.TrimSpace(string(data))
		if _, err := os.Stat(activeRoot); err == nil {
			cfgPath := filepath.Join(activeRoot, ".dutils.yml")
			cfg, _ := LoadConfig(cfgPath)
			return &ProjectInfo{Root: activeRoot, Source: "active", Config: cfg}, nil
		}
	}

	// 5. Git root (fallback)
	if gitRoot != "" {
		return &ProjectInfo{Root: gitRoot, Source: "git"}, nil
	}

	// 6. Current directory (fallback)
	return &ProjectInfo{Root: cwd, Source: "pwd"}, nil
}

func getGitRoot(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
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

	// Default discovery
	defaults := []string{"compose.yml", "compose.yaml", "docker-compose.yml", "docker-compose.yaml"}
	var found []string
	for _, f := range defaults {
		path := filepath.Join(p.Root, f)
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
	
	// Fallback to directory name
	name := filepath.Base(p.Root)
	// Sanitize: lowercase and remove non-alphanumeric
	name = strings.ToLower(name)
	var sb strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
