package devdesk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCloseAllDoesNotExcludeUserTerminalApps(t *testing.T) {
	for _, app := range []string{"Warp", "Terminal", "iTerm2"} {
		if protectedApps[app] {
			t.Fatalf("protectedApps contains %q, want user terminal apps closable", app)
		}
		for _, excludedApp := range closeAllExcludedApps {
			if excludedApp == app {
				t.Fatalf("closeAllExcludedApps contains %q, want user terminal apps closable", app)
			}
		}
	}
}

func TestQuitApplicationScriptUsesConcreteAppName(t *testing.T) {
	script := quitApplicationScript("Warp")
	if !strings.Contains(script, `tell application "Warp" to quit`) {
		t.Fatalf("quitApplicationScript() = %q, want concrete app name quit", script)
	}
}

func TestParseApplicationNames(t *testing.T) {
	got := parseApplicationNames("Finder\nWarp\n\nExample Editor\n")
	want := []string{"Finder", "Warp", "Example Editor"}

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("parseApplicationNames() = %v, want %v", got, want)
	}
}

func TestMacExecutorRunCommandLaunchesShellWithCommandLog(t *testing.T) {
	logDir := t.TempDir()
	executor := &MacExecutor{LogDir: logDir}

	if err := executor.RunCommand(`printf '%s' devdesk-shell-test`); err != nil {
		t.Fatalf("RunCommand() error = %v", err)
	}

	logPath := filepath.Join(logDir, "command-1.log")
	var data []byte
	var err error
	for i := 0; i < 20; i++ {
		data, err = os.ReadFile(logPath)
		if err == nil && strings.Contains(string(data), "devdesk-shell-test") {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	t.Fatalf("command log = %q, want shell command output", string(data))
}
