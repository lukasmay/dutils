package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lukasmay/dutils/pkg/config"
	"github.com/spf13/cobra"
)

// localTempDir creates a temporary directory inside the calling package's
// directory and registers cleanup. Using a local (non-/tmp) path avoids
// macOS symlink issues when comparing absolute paths.
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

// --- 1.1 parseContainerRows ---

func TestParseContainerRows(t *testing.T) {
	t.Run("single row all fields", func(t *testing.T) {
		rows := parseContainerRows("web\tUp 2 hours\t0.0.0.0:80->80/tcp")
		if len(rows) != 1 {
			t.Fatalf("want 1 row, got %d", len(rows))
		}
		r := rows[0]
		if r.Name != "web" || r.Status != "Up 2 hours" || r.Ports != "0.0.0.0:80->80/tcp" {
			t.Errorf("unexpected row: %+v", r)
		}
	})

	t.Run("three rows", func(t *testing.T) {
		input := "web\tUp 2 hours\t0.0.0.0:80->80/tcp\ndb\tUp 1 hour\t\nworker\tUp 30 minutes\t"
		rows := parseContainerRows(input)
		if len(rows) != 3 {
			t.Fatalf("want 3 rows, got %d", len(rows))
		}
	})

	t.Run("empty string returns nil", func(t *testing.T) {
		rows := parseContainerRows("")
		if rows != nil {
			t.Errorf("want nil, got %v", rows)
		}
	})

	t.Run("malformed line skipped", func(t *testing.T) {
		// only two fields — no third tab
		rows := parseContainerRows("web\tUp 2 hours")
		if len(rows) != 0 {
			t.Errorf("want 0 rows for malformed line, got %d", len(rows))
		}
	})

	t.Run("empty ports field included", func(t *testing.T) {
		rows := parseContainerRows("web\tUp\t")
		if len(rows) != 1 {
			t.Fatalf("want 1 row, got %d", len(rows))
		}
		if rows[0].Ports != "" {
			t.Errorf("want empty Ports, got %q", rows[0].Ports)
		}
	})
}

// --- 1.2 expandTargets ---

func TestExpandTargets(t *testing.T) {
	backendCfg := &config.Config{
		Groups: map[string][]string{
			"backend": {"db", "worker"},
		},
	}

	t.Run("single group expanded", func(t *testing.T) {
		proj := &config.ProjectInfo{Config: backendCfg}
		got := expandTargets(proj, []string{"@backend"})
		if len(got) != 2 || got[0] != "db" || got[1] != "worker" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("service and group mixed", func(t *testing.T) {
		proj := &config.ProjectInfo{Config: backendCfg}
		got := expandTargets(proj, []string{"web", "@backend"})
		if len(got) != 3 || got[0] != "web" || got[1] != "db" || got[2] != "worker" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("unknown group warns and skips", func(t *testing.T) {
		proj := &config.ProjectInfo{Config: backendCfg}
		got := expandTargets(proj, []string{"@unknown"})
		if len(got) != 0 {
			t.Errorf("want empty, got %v", got)
		}
	})

	t.Run("group with nil config warns and skips", func(t *testing.T) {
		proj := &config.ProjectInfo{Config: nil}
		got := expandTargets(proj, []string{"@backend"})
		if len(got) != 0 {
			t.Errorf("want empty, got %v", got)
		}
	})

	t.Run("empty inputs returns empty", func(t *testing.T) {
		proj := &config.ProjectInfo{Config: backendCfg}
		got := expandTargets(proj, []string{})
		if len(got) != 0 {
			t.Errorf("want empty, got %v", got)
		}
	})

	t.Run("plain service no groups", func(t *testing.T) {
		proj := &config.ProjectInfo{Config: &config.Config{}}
		got := expandTargets(proj, []string{"web"})
		if len(got) != 1 || got[0] != "web" {
			t.Errorf("got %v", got)
		}
	})
}

// --- 3.3 project switch completion ---

func TestSwitchCompletion(t *testing.T) {
	t.Run("all projects returned for empty prefix", func(t *testing.T) {
		dir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", dir)

		config.AddToRegistry("foo", "/path/a")
		config.AddToRegistry("bar", "/path/b")

		got, directive := switchCmd.ValidArgsFunction(switchCmd, []string{}, "")
		if directive != 0 { // ShellCompDirectiveNoFileComp == 4, but let's check it's not Error
			// ShellCompDirectiveError = 1
			if directive == 1 {
				t.Fatal("got ShellCompDirectiveError")
			}
		}
		names := map[string]bool{}
		for _, n := range got {
			names[n] = true
		}
		if !names["foo"] || !names["bar"] {
			t.Errorf("want foo and bar, got %v", got)
		}
	})

	t.Run("prefix filters completions", func(t *testing.T) {
		dir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", dir)

		config.AddToRegistry("foo", "/path/a")
		config.AddToRegistry("bar", "/path/b")

		got, _ := switchCmd.ValidArgsFunction(switchCmd, []string{}, "f")
		if len(got) != 1 || got[0] != "foo" {
			t.Errorf("want [foo], got %v", got)
		}
	})

	t.Run("empty registry returns empty slice", func(t *testing.T) {
		dir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", dir)

		got, _ := switchCmd.ValidArgsFunction(switchCmd, []string{}, "")
		if len(got) != 0 {
			t.Errorf("want empty, got %v", got)
		}
	})

	t.Run("registry read error returns directive error", func(t *testing.T) {
		dir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", dir)

		// Write a corrupt JSON registry
		registryFile := filepath.Join(dir, "registry.json")
		os.WriteFile(registryFile, []byte("not json"), 0644)

		_, directive := switchCmd.ValidArgsFunction(switchCmd, []string{}, "")
		if directive != cobra.ShellCompDirectiveError {
			t.Errorf("expected ShellCompDirectiveError, got %v", directive)
		}
	})
}
