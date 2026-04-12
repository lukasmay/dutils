package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lukasmay/dutils/pkg/config"
)

// --- 2.4 Registry ---

func TestRegistry(t *testing.T) {
	t.Run("add then read returns entry", func(t *testing.T) {
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))

		if err := config.AddToRegistry("myapp", "/projects/myapp"); err != nil {
			t.Fatal(err)
		}
		projects, err := config.ReadRegistry()
		if err != nil {
			t.Fatal(err)
		}
		if projects["myapp"] != "/projects/myapp" {
			t.Errorf("got %v", projects)
		}
	})

	t.Run("add twice overwrites", func(t *testing.T) {
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))

		config.AddToRegistry("myapp", "/projects/myapp-v1")
		config.AddToRegistry("myapp", "/projects/myapp-v2")

		projects, err := config.ReadRegistry()
		if err != nil {
			t.Fatal(err)
		}
		if projects["myapp"] != "/projects/myapp-v2" {
			t.Errorf("expected /projects/myapp-v2, got %q", projects["myapp"])
		}
	})

	t.Run("read when file absent returns empty map", func(t *testing.T) {
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))

		projects, err := config.ReadRegistry()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(projects) != 0 {
			t.Errorf("want empty map, got %v", projects)
		}
	})

	t.Run("read with corrupt JSON returns empty map and error", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		os.WriteFile(filepath.Join(cfgDir, "registry.json"), []byte("not json"), 0644)

		projects, err := config.ReadRegistry()
		if err == nil {
			t.Error("expected error for corrupt JSON, got nil")
		}
		if len(projects) != 0 {
			t.Errorf("want empty map on error, got %v", projects)
		}
	})

	t.Run("legacy text registry migrated to JSON", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		legacyContent := "foo:/projects/foo\nbar:/projects/bar\n"
		os.WriteFile(filepath.Join(cfgDir, "registry"), []byte(legacyContent), 0644)

		projects, err := config.ReadRegistry()
		if err != nil {
			t.Fatal(err)
		}
		if projects["foo"] != "/projects/foo" || projects["bar"] != "/projects/bar" {
			t.Errorf("migration result: %v", projects)
		}

		// JSON registry should now exist
		if _, err := os.Stat(filepath.Join(cfgDir, "registry.json")); err != nil {
			t.Error("registry.json not created after migration")
		}
		// Legacy file should be removed
		if _, err := os.Stat(filepath.Join(cfgDir, "registry")); err == nil {
			t.Error("legacy registry file should have been removed")
		}
	})

	t.Run("legacy migration skips blank lines", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		legacyContent := "foo:/projects/foo\n\n\nbar:/projects/bar\n"
		os.WriteFile(filepath.Join(cfgDir, "registry"), []byte(legacyContent), 0644)

		projects, err := config.ReadRegistry()
		if err != nil {
			t.Fatal(err)
		}
		if len(projects) != 2 {
			t.Errorf("want 2 entries, got %v", projects)
		}
	})

	t.Run("multiple projects all stored", func(t *testing.T) {
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))

		config.AddToRegistry("alpha", "/a")
		config.AddToRegistry("beta", "/b")
		config.AddToRegistry("gamma", "/c")

		projects, err := config.ReadRegistry()
		if err != nil {
			t.Fatal(err)
		}
		if len(projects) != 3 {
			t.Errorf("want 3, got %d: %v", len(projects), projects)
		}
	})
}

// --- 2.5 Active Project ---

func TestActiveProject(t *testing.T) {
	t.Run("set then read returns same path", func(t *testing.T) {
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))

		if err := config.SetActiveProject("/projects/myapp"); err != nil {
			t.Fatal(err)
		}
		got, err := config.ReadActiveProject()
		if err != nil {
			t.Fatal(err)
		}
		if got != "/projects/myapp" {
			t.Errorf("got %q, want %q", got, "/projects/myapp")
		}
	})

	t.Run("set then read trims whitespace", func(t *testing.T) {
		cfgDir := localTempDir(t)
		t.Setenv("DUTILS_CONFIG_DIR", cfgDir)

		// Write with trailing newline directly to simulate file
		os.WriteFile(filepath.Join(cfgDir, "active"), []byte("/projects/myapp\n"), 0644)

		got, err := config.ReadActiveProject()
		if err != nil {
			t.Fatal(err)
		}
		if got != "/projects/myapp" {
			t.Errorf("got %q, want trimmed path", got)
		}
	})

	t.Run("clear then read returns error", func(t *testing.T) {
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))

		config.SetActiveProject("/projects/myapp")
		if err := config.ClearActiveProject(); err != nil {
			t.Fatal(err)
		}
		if _, err := config.ReadActiveProject(); err == nil {
			t.Error("expected error after clear, got nil")
		}
	})

	t.Run("read when file absent returns error", func(t *testing.T) {
		t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))

		if _, err := config.ReadActiveProject(); err == nil {
			t.Error("expected error when no active file, got nil")
		}
	})
}

