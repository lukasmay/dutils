package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lukasmay/dutils/pkg/config"
)

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

// --- 2.6 runInit ---

func TestRunInit(t *testing.T) {
	t.Run("no config creates file registers and sets active", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		projDir := localTempDir(t)

		var buf bytes.Buffer
		if err := runInit(projDir, &buf); err != nil {
			t.Fatalf("runInit: %v", err)
		}

		// config file created
		if _, err := os.Stat(filepath.Join(projDir, ".dutils.yml")); err != nil {
			t.Errorf(".dutils.yml not created: %v", err)
		}

		// registered in registry
		reg, err := config.ReadRegistry()
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, path := range reg {
			if path == projDir {
				found = true
			}
		}
		if !found {
			t.Errorf("project not in registry: %v", reg)
		}

		// active project set
		active, err := config.ReadActiveProject()
		if err != nil {
			t.Fatalf("ReadActiveProject: %v", err)
		}
		if active != projDir {
			t.Errorf("active = %q, want %q", active, projDir)
		}
	})

	t.Run("existing config not overwritten still registers and sets active", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		projDir := localTempDir(t)
		existingContent := "project_name: existing\n"
		writeFile(t, filepath.Join(projDir, ".dutils.yml"), existingContent)

		var buf bytes.Buffer
		if err := runInit(projDir, &buf); err != nil {
			t.Fatalf("runInit: %v", err)
		}

		// file not overwritten
		data, _ := os.ReadFile(filepath.Join(projDir, ".dutils.yml"))
		if string(data) != existingContent {
			t.Errorf("config file was overwritten; got %q", string(data))
		}

		// registered
		reg, _ := config.ReadRegistry()
		if reg["existing"] != projDir {
			t.Errorf("registry entry missing or wrong: %v", reg)
		}
	})

	t.Run("project_name in config used as registry key", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		projDir := localTempDir(t)
		writeFile(t, filepath.Join(projDir, ".dutils.yml"), "project_name: customname\n")

		var buf bytes.Buffer
		if err := runInit(projDir, &buf); err != nil {
			t.Fatalf("runInit: %v", err)
		}

		reg, _ := config.ReadRegistry()
		if reg["customname"] != projDir {
			t.Errorf("want registry key 'customname', got %v", reg)
		}
	})

	t.Run("no project_name uses dir basename as key", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		// Create a named subdir so basename is predictable
		base := localTempDir(t)
		projDir := filepath.Join(base, "myservice")
		if err := os.Mkdir(projDir, 0755); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		if err := runInit(projDir, &buf); err != nil {
			t.Fatalf("runInit: %v", err)
		}

		reg, _ := config.ReadRegistry()
		if reg["myservice"] != projDir {
			t.Errorf("want registry key 'myservice', got %v", reg)
		}
	})
}

// --- 2.7 runAdd ---

func TestRunAdd(t *testing.T) {
	t.Run("path with config registers under project_name", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		projDir := localTempDir(t)
		writeFile(t, filepath.Join(projDir, ".dutils.yml"), "project_name: myapp\n")

		var buf bytes.Buffer
		if err := runAdd(projDir, &buf); err != nil {
			t.Fatalf("runAdd: %v", err)
		}

		reg, _ := config.ReadRegistry()
		if reg["myapp"] != projDir {
			t.Errorf("want myapp -> %s, got %v", projDir, reg)
		}
	})

	t.Run("path without config uses dir basename", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		base := localTempDir(t)
		projDir := filepath.Join(base, "svcname")
		if err := os.Mkdir(projDir, 0755); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		if err := runAdd(projDir, &buf); err != nil {
			t.Fatalf("runAdd: %v", err)
		}

		reg, _ := config.ReadRegistry()
		if reg["svcname"] != projDir {
			t.Errorf("want svcname -> %s, got %v", projDir, reg)
		}
	})

	t.Run("dot resolves to absolute CWD", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		projDir := localTempDir(t)
		t.Chdir(projDir)

		var buf bytes.Buffer
		if err := runAdd(".", &buf); err != nil {
			t.Fatalf("runAdd: %v", err)
		}

		reg, _ := config.ReadRegistry()
		found := false
		for _, path := range reg {
			if path == projDir {
				found = true
			}
		}
		if !found {
			t.Errorf("expected projDir %q in registry, got %v", projDir, reg)
		}
	})
}

// --- 2.8 runSwitch ---

func TestRunSwitch(t *testing.T) {
	t.Run("known project sets active", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		projDir := localTempDir(t)
		config.AddToRegistry("myproj", projDir)

		var buf bytes.Buffer
		if err := runSwitch("myproj", &buf); err != nil {
			t.Fatalf("runSwitch: %v", err)
		}

		active, err := config.ReadActiveProject()
		if err != nil {
			t.Fatal(err)
		}
		if active != projDir {
			t.Errorf("active = %q, want %q", active, projDir)
		}
	})

	t.Run("unknown project returns error", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		var buf bytes.Buffer
		if err := runSwitch("nosuchproject", &buf); err == nil {
			t.Error("expected error for unknown project, got nil")
		}
	})
}

// --- 2.9 runStatus ---

func TestRunStatus(t *testing.T) {
	t.Run("active project in registry shows name and path", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		projDir := localTempDir(t)
		config.AddToRegistry("webapp", projDir)
		config.SetActiveProject(projDir)

		var buf bytes.Buffer
		runStatus(&buf)
		out := buf.String()

		if !strings.Contains(out, "webapp") {
			t.Errorf("expected project name 'webapp' in output:\n%s", out)
		}
		if !strings.Contains(out, projDir) {
			t.Errorf("expected path %q in output:\n%s", projDir, out)
		}
	})

	t.Run("active project path not in registry shows unknown", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		projDir := localTempDir(t)
		// set active but do NOT register
		config.SetActiveProject(projDir)

		var buf bytes.Buffer
		runStatus(&buf)
		out := buf.String()

		if !strings.Contains(out, "unknown") {
			t.Errorf("expected 'unknown' in output:\n%s", out)
		}
	})

	t.Run("no active project shows message", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		var buf bytes.Buffer
		runStatus(&buf)
		out := buf.String()

		if !strings.Contains(out, "No active project") {
			t.Errorf("expected 'No active project' in output:\n%s", out)
		}
	})
}

// --- 2.10 runProjectList ---

func TestRunProjectList(t *testing.T) {
	t.Run("two projects one active starred", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		dir1 := localTempDir(t)
		dir2 := localTempDir(t)
		config.AddToRegistry("alpha", dir1)
		config.AddToRegistry("beta", dir2)
		config.SetActiveProject(dir1)

		var buf bytes.Buffer
		if err := runProjectList(&buf); err != nil {
			t.Fatalf("runProjectList: %v", err)
		}
		out := buf.String()

		// active (dir1/alpha) should have * prefix
		if !strings.Contains(out, "* alpha") {
			t.Errorf("expected '* alpha' in output:\n%s", out)
		}
		// inactive (dir2/beta) should not have * prefix
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "beta") && strings.HasPrefix(strings.TrimLeft(line, " "), "*") {
				t.Errorf("beta should not be starred:\n%s", out)
			}
		}
	})

	t.Run("no projects shows message", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		var buf bytes.Buffer
		if err := runProjectList(&buf); err != nil {
			t.Fatalf("runProjectList: %v", err)
		}
		out := buf.String()

		if !strings.Contains(out, "No projects registered") {
			t.Errorf("expected 'No projects registered' in output:\n%s", out)
		}
	})
}
