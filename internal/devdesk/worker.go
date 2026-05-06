package devdesk

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type WorkerStarter interface {
	StartWorker(configPath string, presetName string, runID string, logDir string) error
}

type ProcessWorkerStarter struct{}

func NewRunID() string {
	return fmt.Sprintf("%s-%d", time.Now().Format("20060102-150405"), os.Getpid())
}

func LogRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".devdesk", "logs"), nil
}

func RunLogDir(runID string) (string, error) {
	root, err := LogRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, runID), nil
}

func (ProcessWorkerStarter) StartWorker(configPath string, presetName string, runID string, logDir string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return fmt.Errorf("create log directory %s: %w", logDir, err)
	}

	workerLogPath := filepath.Join(logDir, "worker.log")
	workerLog, err := os.OpenFile(workerLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open worker log %s: %w", workerLogPath, err)
	}
	defer workerLog.Close()

	args := []string{"worker", "--config", configPath, "--preset", presetName, "--run-id", runID}
	cmd := exec.Command(executable, args...)
	cmd.Stdin = nil
	cmd.Stdout = workerLog
	cmd.Stderr = workerLog
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start worker: %w", err)
	}

	_, _ = fmt.Fprintf(workerLog, "[devdesk] worker pid %d\n", cmd.Process.Pid)
	_ = cmd.Process.Release()
	return nil
}

func StartPresetInBackground(configPath string, presetName string, starter WorkerStarter, stdout io.Writer) error {
	resolvedConfig, err := ExpandPath(configPath)
	if err != nil {
		return err
	}

	runID := NewRunID()
	logDir, err := RunLogDir(runID)
	if err != nil {
		return err
	}

	if err := starter.StartWorker(resolvedConfig, presetName, runID, logDir); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Started workspace %q in the background.\n", presetName)
	fmt.Fprintf(stdout, "Logs: %s\n", logDir)
	return nil
}

func RunWorker(configPath string, presetName string, runID string, stdout io.Writer) error {
	if presetName == "" {
		return fmt.Errorf("worker preset is required")
	}
	if runID == "" {
		return fmt.Errorf("worker run id is required")
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	preset, err := cfg.Preset(presetName)
	if err != nil {
		return err
	}

	logDir, err := RunLogDir(runID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return fmt.Errorf("create log directory %s: %w", logDir, err)
	}

	fmt.Fprintf(stdout, "Running workspace %q with run id %s\n", presetName, runID)
	executor := &MacExecutor{LogDir: logDir}
	return ApplyPreset(preset, executor, stdout)
}
