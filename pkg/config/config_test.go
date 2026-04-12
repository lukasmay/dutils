package config_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lukasmay/dutils/pkg/config"
)

// localTempDir creates a temp directory inside the package directory and
// registers cleanup. Avoids macOS /var/folders symlink issues with path
// comparisons.
func localTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp(".", "testwork_")
	if err != nil {
		t.Fatalf("localTempDir: %v", err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(abs) })
	return abs
}

// normPath resolves symlinks for reliable cross-platform path comparison.
func normPath(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		return p
	}
	return r
}

// writeFile writes content to path, creating parent dirs as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// --- 1.3 GetProjectName ---

func TestGetProjectName(t *testing.T) {
	tests := []struct {
		name       string
		configName string
		dirName    string
		want       string
	}{
		{"config name used", "myapp", "anything", "myapp"},
		{"dir lowercased and stripped", "", "My-App_2", "myapp2"},
		{"plain dir name passthrough", "", "hello", "hello"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := localTempDir(t)
			subDir := filepath.Join(base, tc.dirName)
			if err := os.Mkdir(subDir, 0755); err != nil {
				t.Fatal(err)
			}
			proj := &config.ProjectInfo{Root: subDir}
			if tc.configName != "" {
				proj.Config = &config.Config{ProjectName: tc.configName}
			}
			if got := proj.GetProjectName(); got != tc.want {
				t.Errorf("GetProjectName() = %q, want %q", got, tc.want)
			}
		})
	}
}

// --- 2.1 LoadConfig ---

func TestLoadConfig(t *testing.T) {
	t.Run("valid YAML all fields", func(t *testing.T) {
		dir := localTempDir(t)
		path := filepath.Join(dir, ".dutils.yml")
		writeFile(t, path, `project_name: myapp
groups:
  backend:
    - db
    - worker
compose:
  files:
    - docker-compose.yml
`)
		cfg, err := config.LoadConfig(path)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.ProjectName != "myapp" {
			t.Errorf("ProjectName = %q, want %q", cfg.ProjectName, "myapp")
		}
		if len(cfg.Groups["backend"]) != 2 {
			t.Errorf("Groups[backend] len = %d, want 2", len(cfg.Groups["backend"]))
		}
		if len(cfg.Compose.Files) == 0 || cfg.Compose.Files[0] != "docker-compose.yml" {
			t.Errorf("Compose.Files = %v", cfg.Compose.Files)
		}
	})

	t.Run("minimal YAML project_name only", func(t *testing.T) {
		dir := localTempDir(t)
		path := filepath.Join(dir, ".dutils.yml")
		writeFile(t, path, "project_name: minimal\n")

		cfg, err := config.LoadConfig(path)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.ProjectName != "minimal" {
			t.Errorf("ProjectName = %q", cfg.ProjectName)
		}
		if len(cfg.Groups) != 0 {
			t.Errorf("Groups should be empty, got %v", cfg.Groups)
		}
	})

	t.Run("invalid YAML returns error", func(t *testing.T) {
		dir := localTempDir(t)
		path := filepath.Join(dir, ".dutils.yml")
		writeFile(t, path, ":\t:bad yaml{{\n")

		if _, err := config.LoadConfig(path); err == nil {
			t.Error("expected error for invalid YAML, got nil")
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		if _, err := config.LoadConfig("/nonexistent/path/.dutils.yml"); err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})

	t.Run("empty file returns zero config no error", func(t *testing.T) {
		dir := localTempDir(t)
		path := filepath.Join(dir, ".dutils.yml")
		writeFile(t, path, "")

		cfg, err := config.LoadConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "" || len(cfg.Groups) != 0 {
			t.Errorf("expected zero config, got %+v", cfg)
		}
	})
}

// --- 2.2 ResolveProject ---

func TestResolveProject(t *testing.T) {
	// Most sub-tests change the working directory; each uses t.Chdir() which
	// automatically restores on cleanup.

	t.Run("dutils.yml at git root returns git-config", func(t *testing.T) {
		dir := localTempDir(t)
		if err := exec.Command("git", "init", dir).Run(); err != nil {
			t.Skip("git not available:", err)
		}
		writeFile(t, filepath.Join(dir, ".dutils.yml"), "project_name: gitrepo\n")
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))
		t.Setenv("DUTILS_PROJECT_ROOT", "") // ensure env doesn't interfere
		t.Chdir(dir)

		proj, err := config.ResolveProject()
		if err != nil {
			t.Fatal(err)
		}
		if proj.Source != "git-config" {
			t.Errorf("Source = %q, want %q", proj.Source, "git-config")
		}
		if normPath(t, proj.Root) != normPath(t, dir) {
			t.Errorf("Root = %q, want %q", proj.Root, dir)
		}
	})

	t.Run("dutils.yml in CWD only returns pwd-config", func(t *testing.T) {
		// No git init — git check will fail, falls to CWD check
		dir := localTempDir(t)
		writeFile(t, filepath.Join(dir, ".dutils.yml"), "project_name: cwdrepo\n")
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))
		t.Setenv("DUTILS_PROJECT_ROOT", "")
		t.Chdir(dir)

		proj, err := config.ResolveProject()
		if err != nil {
			t.Fatal(err)
		}
		if proj.Source != "pwd-config" {
			t.Errorf("Source = %q, want %q", proj.Source, "pwd-config")
		}
	})

	t.Run("DUTILS_PROJECT_ROOT set no config returns env", func(t *testing.T) {
		envRoot := localTempDir(t)
		cwd := localTempDir(t) // no .dutils.yml, no git
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))
		t.Setenv("DUTILS_PROJECT_ROOT", envRoot)
		t.Chdir(cwd)

		proj, err := config.ResolveProject()
		if err != nil {
			t.Fatal(err)
		}
		if proj.Source != "env" {
			t.Errorf("Source = %q, want %q", proj.Source, "env")
		}
	})

	t.Run("DUTILS_PROJECT_ROOT with dutils.yml returns env with config", func(t *testing.T) {
		envRoot := localTempDir(t)
		writeFile(t, filepath.Join(envRoot, ".dutils.yml"), "project_name: envrepo\n")
		cwd := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))
		t.Setenv("DUTILS_PROJECT_ROOT", envRoot)
		t.Chdir(cwd)

		proj, err := config.ResolveProject()
		if err != nil {
			t.Fatal(err)
		}
		if proj.Source != "env" {
			t.Errorf("Source = %q, want %q", proj.Source, "env")
		}
		if proj.Config == nil || proj.Config.ProjectName != "envrepo" {
			t.Errorf("expected config with project_name=envrepo, got %+v", proj.Config)
		}
	})

	t.Run("active project set valid path returns active", func(t *testing.T) {
		activeRoot := localTempDir(t)
		cfgDir := localTempDir(t)
		cwd := localTempDir(t) // no .dutils.yml, no git
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)
		t.Setenv("DUTILS_PROJECT_ROOT", "")
		// write active file
		writeFile(t, filepath.Join(cfgDir, "active"), activeRoot)
		t.Chdir(cwd)

		proj, err := config.ResolveProject()
		if err != nil {
			t.Fatal(err)
		}
		if proj.Source != "active" {
			t.Errorf("Source = %q, want %q", proj.Source, "active")
		}
	})

	t.Run("active project path does not exist falls through", func(t *testing.T) {
		cfgDir := localTempDir(t)
		cwd := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)
		t.Setenv("DUTILS_PROJECT_ROOT", "")
		// active points at non-existent path
		writeFile(t, filepath.Join(cfgDir, "active"), "/nonexistent/path/that/does/not/exist")
		t.Chdir(cwd)

		proj, err := config.ResolveProject()
		if err != nil {
			t.Fatal(err)
		}
		// should fall through to "pwd" since no git and no other config
		if proj.Source == "active" {
			t.Error("should not have resolved to active for nonexistent path")
		}
	})

	t.Run("git root no dutils.yml returns git", func(t *testing.T) {
		dir := localTempDir(t)
		if err := exec.Command("git", "init", dir).Run(); err != nil {
			t.Skip("git not available:", err)
		}
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))
		t.Setenv("DUTILS_PROJECT_ROOT", "")
		t.Chdir(dir)

		proj, err := config.ResolveProject()
		if err != nil {
			t.Fatal(err)
		}
		if proj.Source != "git" {
			t.Errorf("Source = %q, want %q", proj.Source, "git")
		}
	})

	t.Run("no git no config no env returns pwd", func(t *testing.T) {
		// Must be outside any git repo — use system temp dir, not localTempDir
		// which creates inside the dutil repo tree.
		cwd, err := os.MkdirTemp("", "dutils_testwork_")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.RemoveAll(cwd) })

		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))
		t.Setenv("DUTILS_PROJECT_ROOT", "")
		t.Chdir(cwd)

		proj, err := config.ResolveProject()
		if err != nil {
			t.Fatal(err)
		}
		if proj.Source != "pwd" {
			t.Errorf("Source = %q, want %q", proj.Source, "pwd")
		}
	})

	t.Run("malformed dutils.yml returns error", func(t *testing.T) {
		cwd := localTempDir(t)
		writeFile(t, filepath.Join(cwd, ".dutils.yml"), ":\t:bad yaml{{\n")
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))
		t.Setenv("DUTILS_PROJECT_ROOT", "")
		t.Chdir(cwd)

		if _, err := config.ResolveProject(); err == nil {
			t.Error("expected error for malformed YAML, got nil")
		}
	})
}

// --- 2.3 GetComposeFiles ---

func TestGetComposeFiles(t *testing.T) {
	t.Run("config specifies file that exists", func(t *testing.T) {
		dir := localTempDir(t)
		writeFile(t, filepath.Join(dir, "docker-compose.yml"), "services: {}")
		proj := &config.ProjectInfo{
			Root:   dir,
			Config: &config.Config{Compose: config.ComposeConfig{Files: []string{"docker-compose.yml"}}},
		}
		files := proj.GetComposeFiles()
		if len(files) != 1 || files[0] != filepath.Join(dir, "docker-compose.yml") {
			t.Errorf("got %v", files)
		}
	})

	t.Run("config specifies relative path resolved to root", func(t *testing.T) {
		dir := localTempDir(t)
		sub := filepath.Join(dir, "sub")
		os.Mkdir(sub, 0755)
		writeFile(t, filepath.Join(sub, "compose.yml"), "services: {}")
		proj := &config.ProjectInfo{
			Root:   dir,
			Config: &config.Config{Compose: config.ComposeConfig{Files: []string{"sub/compose.yml"}}},
		}
		files := proj.GetComposeFiles()
		if len(files) != 1 || files[0] != filepath.Join(dir, "sub", "compose.yml") {
			t.Errorf("got %v", files)
		}
	})

	t.Run("no config compose.yml auto-discovered", func(t *testing.T) {
		dir := localTempDir(t)
		writeFile(t, filepath.Join(dir, "compose.yml"), "services: {}")
		proj := &config.ProjectInfo{Root: dir}
		files := proj.GetComposeFiles()
		if len(files) != 1 || files[0] != filepath.Join(dir, "compose.yml") {
			t.Errorf("got %v", files)
		}
	})

	t.Run("no config multiple candidates both returned", func(t *testing.T) {
		dir := localTempDir(t)
		writeFile(t, filepath.Join(dir, "compose.yaml"), "services: {}")
		writeFile(t, filepath.Join(dir, "docker-compose.yml"), "services: {}")
		proj := &config.ProjectInfo{Root: dir}
		files := proj.GetComposeFiles()
		if len(files) != 2 {
			t.Errorf("want 2 files, got %v", files)
		}
	})

	t.Run("no config no candidates returns empty", func(t *testing.T) {
		dir := localTempDir(t)
		proj := &config.ProjectInfo{Root: dir}
		files := proj.GetComposeFiles()
		if len(files) != 0 {
			t.Errorf("want empty, got %v", files)
		}
	})

	t.Run("config specifies nonexistent file still returned", func(t *testing.T) {
		dir := localTempDir(t)
		proj := &config.ProjectInfo{
			Root:   dir,
			Config: &config.Config{Compose: config.ComposeConfig{Files: []string{"missing.yml"}}},
		}
		files := proj.GetComposeFiles()
		if len(files) != 1 || files[0] != filepath.Join(dir, "missing.yml") {
			t.Errorf("got %v", files)
		}
	})
}
