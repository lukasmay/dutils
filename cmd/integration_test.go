//go:build integration

package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lukasmay/dutils/pkg/config"
)

// absTestdata returns the absolute path to cmd/testdata.
// Must be called before any t.Chdir() because the test runner starts in cmd/.
var absTestdata string

func init() {
	abs, err := filepath.Abs("testdata")
	if err != nil {
		panic("integration_test: cannot resolve testdata path: " + err.Error())
	}
	absTestdata = abs
}

// composeDown tears down the test fixture after each integration test.
func composeDown(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		exec.Command(
			"docker", "compose",
			"-f", filepath.Join(absTestdata, "compose.yml"),
			"down", "--remove-orphans", "--timeout", "5",
		).Run()
	})
}

// runningContainers returns the names of currently running containers whose
// names contain the given substring. Pass "" to get all running containers.
func runningContainers(t *testing.T, filter string) []string {
	t.Helper()
	out, err := exec.Command("docker", "ps", "--format", "{{.Names}}").Output()
	if err != nil {
		t.Fatalf("docker ps: %v", err)
	}
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && (filter == "" || strings.Contains(line, filter)) {
			names = append(names, line)
		}
	}
	return names
}

// allContainers is like runningContainers but includes stopped containers.
func allContainers(t *testing.T, filter string) []string {
	t.Helper()
	out, err := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}").Output()
	if err != nil {
		t.Fatalf("docker ps -a: %v", err)
	}
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && (filter == "" || strings.Contains(line, filter)) {
			names = append(names, line)
		}
	}
	return names
}

// testdataProject returns a ProjectInfo rooted at testdata with the fixture config.
func testdataProject(t *testing.T) *config.ProjectInfo {
	t.Helper()
	cfg, err := config.LoadConfig(filepath.Join(absTestdata, ".dutils.yml"))
	if err != nil {
		t.Fatalf("load fixture config: %v", err)
	}
	return &config.ProjectInfo{
		Root:   absTestdata,
		Source: "test",
		Config: cfg,
	}
}

// --- 3.1 servicesInFile ---

func TestServicesInFile(t *testing.T) {
	composeFile := filepath.Join(absTestdata, "compose.yml")

	t.Run("subset of services returned", func(t *testing.T) {
		got := servicesInFile(composeFile, []string{"web", "db"})
		names := map[string]bool{}
		for _, n := range got {
			names[n] = true
		}
		if !names["web"] || !names["db"] {
			t.Errorf("want web and db, got %v", got)
		}
		if names["worker"] {
			t.Errorf("worker should not be in result, got %v", got)
		}
	})

	t.Run("nonexistent service filtered out", func(t *testing.T) {
		got := servicesInFile(composeFile, []string{"web", "nonexistent"})
		if len(got) != 1 || got[0] != "web" {
			t.Errorf("want [web], got %v", got)
		}
	})

	t.Run("empty targets returns nil", func(t *testing.T) {
		got := servicesInFile(composeFile, []string{})
		if got != nil {
			t.Errorf("want nil, got %v", got)
		}
	})

	t.Run("all three services returned", func(t *testing.T) {
		got := servicesInFile(composeFile, []string{"web", "db", "worker"})
		if len(got) != 3 {
			t.Errorf("want 3, got %v", got)
		}
	})

	t.Run("nonexistent compose file returns nil no panic", func(t *testing.T) {
		got := servicesInFile("/nonexistent/compose.yml", []string{"web"})
		if got != nil {
			t.Errorf("want nil, got %v", got)
		}
	})
}

// --- 3.2 completeServices ---

func TestCompleteServices(t *testing.T) {
	// completeServices calls config.ResolveProject() internally, so we point
	// DUTILS_PROJECT_ROOT at testdata to control which project is resolved.
	// We also need to be in a dir that has no .dutils.yml above it in priority,
	// so we use DUTILS_PROJECT_ROOT which is checked after git-config and pwd-config.
	// The safest approach: cd into testdata so pwd-config picks it up.
	t.Chdir(absTestdata)
	t.Setenv("DUTILS_CONFIG_DIR", localTempDir(t))
	t.Setenv("DUTILS_PROJECT_ROOT", "")

	t.Run("empty prefix returns groups and all services", func(t *testing.T) {
		got, directive := completeServices(nil, nil, "")
		if directive != 0 && directive != 4 { // 0=default,4=NoFileComp
			t.Errorf("unexpected directive %v", directive)
		}
		names := map[string]bool{}
		for _, n := range got {
			names[n] = true
		}
		for _, want := range []string{"@backend", "@all", "web", "db", "worker"} {
			if !names[want] {
				t.Errorf("missing %q in completions %v", want, got)
			}
		}
	})

	t.Run("prefix w returns web and worker", func(t *testing.T) {
		got, _ := completeServices(nil, nil, "w")
		for _, n := range got {
			if !strings.HasPrefix(n, "w") {
				t.Errorf("completion %q does not match prefix 'w'", n)
			}
		}
		names := map[string]bool{}
		for _, n := range got {
			names[n] = true
		}
		if !names["web"] || !names["worker"] {
			t.Errorf("want web and worker, got %v", got)
		}
	})

	t.Run("@ prefix returns only groups", func(t *testing.T) {
		got, _ := completeServices(nil, nil, "@")
		for _, n := range got {
			if !strings.HasPrefix(n, "@") {
				t.Errorf("non-group %q returned for @ prefix", n)
			}
		}
		if len(got) == 0 {
			t.Error("expected group completions for @ prefix")
		}
	})

	t.Run("@b prefix returns only @backend", func(t *testing.T) {
		got, _ := completeServices(nil, nil, "@b")
		if len(got) != 1 || got[0] != "@backend" {
			t.Errorf("want [@backend], got %v", got)
		}
	})
}

// --- 3.4 runStart ---

func TestRunStart(t *testing.T) {
	t.Run("no args starts all services", func(t *testing.T) {
		composeDown(t)
		proj := testdataProject(t)

		if err := runStart(proj, nil, false); err != nil {
			t.Fatalf("runStart: %v", err)
		}

		// Docker Compose prefixes containers with the directory name "testdata"
		running := runningContainers(t, "testdata")
		if len(running) < 3 {
			t.Errorf("expected 3 running containers, got %v", running)
		}
	})

	t.Run("single service arg starts only that service", func(t *testing.T) {
		composeDown(t)
		proj := testdataProject(t)

		if err := runStart(proj, []string{"web"}, false); err != nil {
			t.Fatalf("runStart: %v", err)
		}

		running := runningContainers(t, "testdata")
		hasWeb := false
		for _, n := range running {
			if strings.Contains(n, "-web-") {
				hasWeb = true
			}
			if strings.Contains(n, "-db-") || strings.Contains(n, "-worker-") {
				t.Errorf("unexpected container running: %s", n)
			}
		}
		if !hasWeb {
			t.Errorf("web not running, got %v", running)
		}
	})

	t.Run("group arg starts group members", func(t *testing.T) {
		composeDown(t)
		proj := testdataProject(t)

		if err := runStart(proj, []string{"@backend"}, false); err != nil {
			t.Fatalf("runStart: %v", err)
		}

		running := runningContainers(t, "testdata")
		hasDB := false
		hasWorker := false
		for _, n := range running {
			if strings.Contains(n, "-db-") {
				hasDB = true
			}
			if strings.Contains(n, "-worker-") {
				hasWorker = true
			}
		}
		if !hasDB || !hasWorker {
			t.Errorf("expected db and worker running, got %v", running)
		}
	})

	t.Run("no compose files and no args returns error", func(t *testing.T) {
		proj := &config.ProjectInfo{Root: localTempDir(t)}
		if err := runStart(proj, nil, false); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("--build flag builds then starts service", func(t *testing.T) {
		// Uses compose-build.yml which has a build: context pointing at testdata/Dockerfile
		buildComposeDown := func() {
			exec.Command("docker", "compose",
				"-f", filepath.Join(absTestdata, "compose-build.yml"),
				"down", "--remove-orphans", "--timeout", "5").Run()
		}
		buildComposeDown()
		t.Cleanup(buildComposeDown)

		buildProj := &config.ProjectInfo{
			Root: absTestdata,
			Config: &config.Config{
				Compose: config.ComposeConfig{
					Files: []string{"compose-build.yml"},
				},
			},
		}

		if err := runStart(buildProj, []string{"buildable"}, true); err != nil {
			t.Fatalf("runStart --build: %v", err)
		}

		running := runningContainers(t, "testdata")
		hasBuildable := false
		for _, n := range running {
			if strings.Contains(n, "buildable") {
				hasBuildable = true
			}
		}
		if !hasBuildable {
			t.Errorf("buildable container not running after --build start, got %v", running)
		}
	})

	t.Run("no compose files starts plain container by name", func(t *testing.T) {
		containerName := "dutils-test-plain-start"
		// Create a stopped container to start
		exec.Command("docker", "rm", "-f", containerName).Run()
		exec.Command("docker", "create", "--name", containerName, "busybox", "sleep", "infinity").Run()
		t.Cleanup(func() {
			exec.Command("docker", "rm", "-f", containerName).Run()
		})

		proj := &config.ProjectInfo{Root: localTempDir(t)} // no compose files

		if err := runStart(proj, []string{containerName}, false); err != nil {
			t.Fatalf("runStart plain container: %v", err)
		}

		running := runningContainers(t, containerName)
		if len(running) == 0 {
			t.Errorf("expected %s to be running", containerName)
		}
	})
}

// --- 3.5 runStop ---

func TestRunStop(t *testing.T) {
	t.Run("no args stops all services", func(t *testing.T) {
		composeDown(t)
		proj := testdataProject(t)

		exec.Command("docker", "compose", "-f",
			filepath.Join(absTestdata, "compose.yml"), "up", "-d").Run()

		if err := runStop(proj, nil, false); err != nil {
			t.Fatalf("runStop: %v", err)
		}

		running := runningContainers(t, "testdata")
		if len(running) > 0 {
			t.Errorf("expected no running containers, got %v", running)
		}
	})

	t.Run("single service arg stops only that service", func(t *testing.T) {
		composeDown(t)
		proj := testdataProject(t)

		exec.Command("docker", "compose", "-f",
			filepath.Join(absTestdata, "compose.yml"), "up", "-d").Run()

		if err := runStop(proj, []string{"web"}, false); err != nil {
			t.Fatalf("runStop: %v", err)
		}

		running := runningContainers(t, "testdata")
		for _, n := range running {
			if strings.Contains(n, "-web-") {
				t.Errorf("web should be stopped, but found %s in running", n)
			}
		}
	})

	t.Run("--down flag removes all services", func(t *testing.T) {
		composeDown(t)
		proj := testdataProject(t)

		exec.Command("docker", "compose", "-f",
			filepath.Join(absTestdata, "compose.yml"), "up", "-d").Run()

		if err := runStop(proj, nil, true); err != nil {
			t.Fatalf("runStop --down: %v", err)
		}

		all := allContainers(t, "testdata")
		if len(all) > 0 {
			t.Errorf("expected no containers after down, got %v", all)
		}
	})

	t.Run("no compose files and no args returns error", func(t *testing.T) {
		proj := &config.ProjectInfo{Root: localTempDir(t)}
		if err := runStop(proj, nil, false); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("single service --down removes only that service", func(t *testing.T) {
		composeDown(t)
		proj := testdataProject(t)

		exec.Command("docker", "compose", "-f",
			filepath.Join(absTestdata, "compose.yml"), "up", "-d").Run()

		if err := runStop(proj, []string{"db"}, true); err != nil {
			t.Fatalf("runStop db --down: %v", err)
		}

		// db should be fully removed
		all := allContainers(t, "testdata")
		for _, n := range all {
			if strings.Contains(n, "-db-") {
				t.Errorf("db container should be removed, but found %s", n)
			}
		}
		// web and worker should still exist
		stillHasWeb := false
		for _, n := range all {
			if strings.Contains(n, "-web-") {
				stillHasWeb = true
			}
		}
		if !stillHasWeb {
			t.Errorf("web should still exist after removing only db, got %v", all)
		}
	})

	t.Run("no compose files stops plain container by name", func(t *testing.T) {
		containerName := "dutils-test-plain-stop"
		exec.Command("docker", "rm", "-f", containerName).Run()
		exec.Command("docker", "run", "-d", "--name", containerName, "busybox", "sleep", "infinity").Run()
		t.Cleanup(func() {
			exec.Command("docker", "rm", "-f", containerName).Run()
		})

		proj := &config.ProjectInfo{Root: localTempDir(t)} // no compose files

		if err := runStop(proj, []string{containerName}, false); err != nil {
			t.Fatalf("runStop plain container: %v", err)
		}

		running := runningContainers(t, containerName)
		if len(running) > 0 {
			t.Errorf("expected %s to be stopped, still running", containerName)
		}
	})
}

// --- 3.6 runRestart ---

func TestRunRestart(t *testing.T) {
	t.Run("no args restarts all services", func(t *testing.T) {
		composeDown(t)
		proj := testdataProject(t)

		// Start first so we have something to restart
		exec.Command("docker", "compose", "-f",
			filepath.Join(absTestdata, "compose.yml"), "up", "-d").Run()

		// Capture StartedAt before restart
		beforeOut, _ := exec.Command("docker", "inspect",
			"--format", "{{.State.StartedAt}}",
			"testproject-web-1").Output()
		startedBefore := strings.TrimSpace(string(beforeOut))

		if err := runRestart(proj, nil); err != nil {
			t.Fatalf("runRestart: %v", err)
		}

		afterOut, _ := exec.Command("docker", "inspect",
			"--format", "{{.State.StartedAt}}",
			"testproject-web-1").Output()
		startedAfter := strings.TrimSpace(string(afterOut))

		if startedBefore == startedAfter && startedBefore != "" {
			t.Errorf("StartedAt unchanged after restart: %s", startedBefore)
		}
	})

	t.Run("single service arg restarts only that service", func(t *testing.T) {
		composeDown(t)
		proj := testdataProject(t)

		exec.Command("docker", "compose", "-f",
			filepath.Join(absTestdata, "compose.yml"), "up", "-d").Run()

		// Capture StartedAt for web and worker before restart
		webBefore, _ := exec.Command("docker", "inspect",
			"--format", "{{.State.StartedAt}}", "testproject-web-1").Output()
		workerBefore, _ := exec.Command("docker", "inspect",
			"--format", "{{.State.StartedAt}}", "testproject-worker-1").Output()

		if err := runRestart(proj, []string{"web"}); err != nil {
			t.Fatalf("runRestart web: %v", err)
		}

		webAfter, _ := exec.Command("docker", "inspect",
			"--format", "{{.State.StartedAt}}", "testproject-web-1").Output()
		workerAfter, _ := exec.Command("docker", "inspect",
			"--format", "{{.State.StartedAt}}", "testproject-worker-1").Output()

		wBefore := strings.TrimSpace(string(webBefore))
		wAfter := strings.TrimSpace(string(webAfter))
		if wBefore == wAfter && wBefore != "" {
			t.Errorf("web StartedAt unchanged after restart: %s", wBefore)
		}

		// worker should not have been restarted
		wkBefore := strings.TrimSpace(string(workerBefore))
		wkAfter := strings.TrimSpace(string(workerAfter))
		if wkBefore != wkAfter {
			t.Errorf("worker should not have been restarted: before=%s after=%s", wkBefore, wkAfter)
		}
	})

	t.Run("no compose files returns error", func(t *testing.T) {
		proj := &config.ProjectInfo{Root: localTempDir(t)}
		if err := runRestart(proj, nil); err == nil {
			t.Error("expected error, got nil")
		}
	})
}

// --- 3.7 runDlist ---

func TestRunDlist(t *testing.T) {
	composeDown(t)

	// Start web (has ports) and worker (no ports) for the test
	exec.Command("docker", "compose", "-f",
		filepath.Join(absTestdata, "compose.yml"),
		"up", "-d", "web", "worker").Run()

	t.Run("running containers shows table with header", func(t *testing.T) {
		var buf bytes.Buffer
		if err := runDlist(false, &buf); err != nil {
			t.Fatalf("runDlist: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "NAME") || !strings.Contains(out, "STATUS") || !strings.Contains(out, "PORTS") {
			t.Errorf("missing header in output:\n%s", out)
		}
	})

	t.Run("--all shows stopped containers too", func(t *testing.T) {
		// Stop web, leave worker running
		exec.Command("docker", "compose", "-f",
			filepath.Join(absTestdata, "compose.yml"),
			"stop", "web").Run()

		var bufRunning, bufAll bytes.Buffer
		runDlist(false, &bufRunning)
		runDlist(true, &bufAll)

		// --all output should have more or equal lines than running-only
		runningLines := strings.Count(bufRunning.String(), "\n")
		allLines := strings.Count(bufAll.String(), "\n")
		if allLines < runningLines {
			t.Errorf("--all (%d lines) should have >= lines than default (%d lines)", allLines, runningLines)
		}
	})

	t.Run("default excludes stopped containers exact count", func(t *testing.T) {
		// Stop web, leave worker running — both are up from the outer setup
		exec.Command("docker", "compose", "-f",
			filepath.Join(absTestdata, "compose.yml"),
			"stop", "web").Run()

		var bufRunning bytes.Buffer
		runDlist(false, &bufRunning)
		out := bufRunning.String()

		// Count data rows (lines that are not the header and not empty)
		lines := strings.Split(strings.TrimSpace(out), "\n")
		dataRows := 0
		for _, l := range lines {
			if l != "" && !strings.HasPrefix(l, "NAME") {
				dataRows++
			}
		}
		// Only worker is running; web (stopped) must not appear
		if dataRows != 1 {
			t.Errorf("expected 1 running row, got %d rows:\n%s", dataRows, out)
		}
		if strings.Contains(out, "web") {
			t.Errorf("stopped web container should not appear in default output:\n%s", out)
		}

		// Restart web for subsequent tests
		exec.Command("docker", "compose", "-f",
			filepath.Join(absTestdata, "compose.yml"),
			"start", "web").Run()
	})

	t.Run("no containers running shows message", func(t *testing.T) {
		// Bring everything down
		exec.Command("docker", "compose", "-f",
			filepath.Join(absTestdata, "compose.yml"),
			"down", "--remove-orphans").Run()

		var buf bytes.Buffer
		if err := runDlist(false, &buf); err != nil {
			t.Fatalf("runDlist: %v", err)
		}
		if !strings.Contains(buf.String(), "No containers running") {
			t.Errorf("expected 'No containers running' message, got:\n%s", buf.String())
		}
	})
}

// Ensure the integration test package resolves the testdata path even when
// t.Chdir changes the working directory mid-test.
func TestIntegrationTestdataPathStable(t *testing.T) {
	original := absTestdata
	t.Chdir(localTempDir(t))
	if absTestdata != original {
		t.Errorf("absTestdata changed after Chdir: was %s, now %s", original, absTestdata)
	}
	if _, err := os.Stat(absTestdata); err != nil {
		t.Errorf("absTestdata not accessible: %v", err)
	}
}
