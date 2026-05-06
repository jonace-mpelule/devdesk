package devdesk

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type MacExecutor struct {
	LogDir       string
	commandIndex int
}

var protectedApps = map[string]bool{
	"Control Center":      true,
	"DevDesk":             true,
	"Dock":                true,
	"Finder":              true,
	"loginwindow":         true,
	"Notification Center": true,
	"System Events":       true,
	"System Settings":     true,
	"WindowServer":        true,
}

var closeAllExcludedApps = []string{
	"Control Center",
	"DevDesk",
	"Dock",
	"Finder",
	"loginwindow",
	"Notification Center",
	"System Events",
	"System Settings",
	"WindowServer",
}

func (MacExecutor) CloseAllApps() error {
	appNames, err := runningAppNames()
	if err != nil {
		return err
	}

	var errs []error
	for _, appName := range appNames {
		if protectedApps[appName] {
			continue
		}
		if err := quitApplication(appName); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", appName, err))
		}
	}

	return errors.Join(errs...)
}

func runningAppNames() ([]string, error) {
	output, err := runOSAOutput(runningAppNamesScript())
	if err != nil {
		return nil, err
	}

	return parseApplicationNames(output), nil
}

func quitApplication(appName string) error {
	return runOSA(quitApplicationScript(appName))
}

func quitApplicationScript(appName string) string {
	return fmt.Sprintf(`try
  ignoring application responses
    tell application %q to quit
  end ignoring
on error errMsg
  error errMsg
end try`, appName)
}

func runningAppNamesScript() string {
	return `
tell application "System Events"
  set appNames to name of every application process whose background only is false
end tell
set AppleScript's text item delimiters to linefeed
return appNames as text`
}

func parseApplicationNames(output string) []string {
	lines := strings.Split(output, "\n")
	appNames := make([]string, 0, len(lines))
	for _, line := range lines {
		appName := strings.TrimSpace(line)
		if appName != "" {
			appNames = append(appNames, appName)
		}
	}
	return appNames
}

func (MacExecutor) CloseApp(name string) error {
	if protectedApps[name] {
		return fmt.Errorf("%s is protected and will not be closed", name)
	}

	script := fmt.Sprintf(`try
  tell application %q to quit
on error errMsg
  error errMsg
end try`, name)

	return runOSA(script)
}

func (MacExecutor) OpenApp(name string) error {
	cmd := exec.Command("open", "-a", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (executor *MacExecutor) RunCommand(command string) error {
	executor.commandIndex++
	logPath := filepath.Join(executor.LogDir, fmt.Sprintf("command-%d.log", executor.commandIndex))

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open command log %s: %w", logPath, err)
	}

	cmd := exec.Command("sh", "-lc", command)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return err
	}

	_, _ = fmt.Fprintf(logFile, "\n[devdesk] started pid %d\n", cmd.Process.Pid)
	_ = cmd.Process.Release()
	return logFile.Close()
}

func runOSA(script string) error {
	cmd := exec.Command("osascript", "-e", script)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func runOSAOutput(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func appleScriptList(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, fmt.Sprintf("%q", item))
	}
	return strings.Join(quoted, ", ")
}
