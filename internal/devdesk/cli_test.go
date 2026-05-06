package devdesk

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCLICreateWorkspaceCreatesConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devdesk.yaml")

	var out bytes.Buffer
	err := RunCLI([]string{
		"--config", path,
		"create", "api-workspace",
		"--open", "Example Editor",
		"--open", "Example Database",
		"--close", "Example Chat",
		"--command", "echo api-workspace",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("RunCLI() error = %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	preset, err := cfg.Preset("api-workspace")
	if err != nil {
		t.Fatalf("Preset() error = %v", err)
	}

	if got, want := preset.Open, []string{"Example Editor", "Example Database"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Open = %v, want %v", got, want)
	}
	if got, want := preset.Close, []string{"Example Chat"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Close = %v, want %v", got, want)
	}
	if got, want := preset.Commands, []string{"echo api-workspace"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Commands = %v, want %v", got, want)
	}
	if !strings.Contains(out.String(), `Created workspace "api-workspace"`) {
		t.Fatalf("output = %q, want creation message", out.String())
	}
}

func TestRunCLIVersionDoesNotRequireConfig(t *testing.T) {
	var out bytes.Buffer
	err := RunCLI([]string{"version"}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("RunCLI() error = %v", err)
	}

	if !strings.Contains(out.String(), "devdesk ") {
		t.Fatalf("output = %q, want version output", out.String())
	}
}

func TestRunCLICreateWorkspaceRejectsCloseAllWithCloseApps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devdesk.yaml")

	err := RunCLI([]string{
		"--config", path,
		"create", "api-workspace",
		"--close-all",
		"--close", "Example Chat",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("RunCLI() error = nil, want close_all conflict")
	}

	if !strings.Contains(err.Error(), "close_all cannot be combined") {
		t.Fatalf("RunCLI() error = %q, want close_all conflict", err.Error())
	}
}

func TestRunCLICreateWorkspaceRefusesDuplicate(t *testing.T) {
	path := writeConfig(t, `presets:
  api-workspace:
    open:
      - Example Editor
`)

	err := RunCLI([]string{"--config", path, "create", "api-workspace", "--open", "Example Replacement"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("RunCLI() error = nil, want duplicate error")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("RunCLI() error = %q, want duplicate error", err.Error())
	}
}

func TestRunCLIEditWorkspaceErrorsWhenMissing(t *testing.T) {
	path := writeConfig(t, "presets: {}\n")

	err := RunCLI([]string{"--config", path, "edit", "missing"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("RunCLI() error = nil, want missing workspace")
	}

	if !strings.Contains(err.Error(), `unknown preset "missing"`) {
		t.Fatalf("RunCLI() error = %q, want missing preset", err.Error())
	}
}

func TestRunCLICreateWorkspaceForceReplacesDuplicate(t *testing.T) {
	path := writeConfig(t, `presets:
  api-workspace:
    open:
      - Example Editor
`)

	err := RunCLI([]string{"--config", path, "create", "api-workspace", "--force", "--open", "Example Replacement"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("RunCLI() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(data), "Example Replacement") {
		t.Fatalf("config = %q, want replacement open app", string(data))
	}
}

func TestRunCLIEditWorkspaceWithFlagsReplacesPreset(t *testing.T) {
	path := writeConfig(t, `presets:
  api-workspace:
    close:
      - Example Chat
    open:
      - Example Editor
`)

	var out bytes.Buffer
	err := RunCLI([]string{
		"--config", path,
		"edit", "api-workspace",
		"--close-all",
		"--open", "Example Database",
		"--command", "echo edited",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("RunCLI() error = %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	preset, err := cfg.Preset("api-workspace")
	if err != nil {
		t.Fatalf("Preset() error = %v", err)
	}

	if !preset.CloseAll {
		t.Fatal("preset.CloseAll = false, want true")
	}
	if len(preset.Close) != 0 {
		t.Fatalf("preset.Close = %v, want empty", preset.Close)
	}
	if got, want := preset.Open, []string{"Example Database"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Open = %v, want %v", got, want)
	}
	if got, want := preset.Commands, []string{"echo edited"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Commands = %v, want %v", got, want)
	}
	if !strings.Contains(out.String(), `Updated workspace "api-workspace"`) {
		t.Fatalf("output = %q, want updated message", out.String())
	}
}

func TestRunCLICreateWorkspaceWizardRequiresAtLeastOneAction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devdesk.yaml")

	input := strings.NewReader(strings.Join([]string{
		"api-workspace",
		"",
		"",
		"",
		"",
		"",
		"",
		"Example Editor",
		"",
		"yes",
	}, "\n") + "\n")
	var out bytes.Buffer
	err := RunCLI([]string{"--config", path, "create"}, input, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("RunCLI() error = %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	preset, err := cfg.Preset("api-workspace")
	if err != nil {
		t.Fatalf("Preset() error = %v", err)
	}

	if got, want := preset.Open, []string{"Example Editor"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Open = %v, want %v", got, want)
	}
	if len(preset.Commands) != 0 {
		t.Fatalf("preset.Commands = %v, want empty", preset.Commands)
	}
	if !strings.Contains(out.String(), "Create a DevDesk workspace") {
		t.Fatalf("output = %q, want wizard heading", out.String())
	}
	if !strings.Contains(out.String(), "Add at least one action") {
		t.Fatalf("output = %q, want validation prompt", out.String())
	}
}

func TestRunCLICreateWorkspaceWizardAcceptsCustomValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devdesk.yaml")

	input := strings.NewReader(strings.Join([]string{
		"api-workspace",
		"yes",
		"Example Editor, Example Database",
		"echo api-workspace, date",
		"yes",
	}, "\n") + "\n")

	err := RunCLI([]string{"--config", path, "create"}, input, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("RunCLI() error = %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	preset, err := cfg.Preset("api-workspace")
	if err != nil {
		t.Fatalf("Preset() error = %v", err)
	}

	if !preset.CloseAll {
		t.Fatal("preset.CloseAll = false, want true")
	}
	if len(preset.Close) != 0 {
		t.Fatalf("preset.Close = %v, want empty when close_all is true", preset.Close)
	}
	if got, want := preset.Open, []string{"Example Editor", "Example Database"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Open = %v, want %v", got, want)
	}
	if got, want := preset.Commands, []string{"echo api-workspace", "date"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Commands = %v, want %v", got, want)
	}
}

func TestRunCLIEditWorkspaceWizardUsesExistingValuesAsDefaults(t *testing.T) {
	path := writeConfig(t, `presets:
  api-workspace:
    close:
      - Example Chat
    open:
      - Example Editor
    commands:
      - echo existing
`)

	input := strings.NewReader(strings.Join([]string{
		"",
		"",
		"Example Editor, Example Database",
		"",
		"yes",
	}, "\n") + "\n")

	var out bytes.Buffer
	err := RunCLI([]string{"--config", path, "edit", "api-workspace"}, input, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("RunCLI() error = %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	preset, err := cfg.Preset("api-workspace")
	if err != nil {
		t.Fatalf("Preset() error = %v", err)
	}

	if got, want := preset.Close, []string{"Example Chat"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Close = %v, want %v", got, want)
	}
	if got, want := preset.Open, []string{"Example Editor", "Example Database"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Open = %v, want %v", got, want)
	}
	if got, want := preset.Commands, []string{"echo existing"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("preset.Commands = %v, want %v", got, want)
	}
	if !strings.Contains(out.String(), `Edit workspace "api-workspace"`) {
		t.Fatalf("output = %q, want edit heading", out.String())
	}
	if !strings.Contains(out.String(), `Updated workspace "api-workspace"`) {
		t.Fatalf("output = %q, want updated message", out.String())
	}
}

func TestRunCLIApplyStartsBackgroundWorker(t *testing.T) {
	path := writeConfig(t, `presets:
  api-workspace:
    open:
      - Example Editor
`)
	starter := &fakeWorkerStarter{}
	previous := workerStarter
	workerStarter = starter
	t.Cleanup(func() {
		workerStarter = previous
	})

	var out bytes.Buffer
	err := RunCLI([]string{"--config", path, "api-workspace"}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("RunCLI() error = %v", err)
	}

	if got, want := starter.presetName, "api-workspace"; got != want {
		t.Fatalf("starter preset = %q, want %q", got, want)
	}
	if starter.configPath != path {
		t.Fatalf("starter config path = %q, want %q", starter.configPath, path)
	}
	if starter.runID == "" {
		t.Fatal("starter run id is empty")
	}
	if !strings.Contains(out.String(), `Started workspace "api-workspace" in the background`) {
		t.Fatalf("output = %q, want background start message", out.String())
	}
	if !strings.Contains(out.String(), "Logs:") {
		t.Fatalf("output = %q, want log directory", out.String())
	}
}

type fakeWorkerStarter struct {
	configPath string
	presetName string
	runID      string
	logDir     string
}

func (f *fakeWorkerStarter) StartWorker(configPath string, presetName string, runID string, logDir string) error {
	f.configPath = configPath
	f.presetName = presetName
	f.runID = runID
	f.logDir = logDir
	return nil
}
