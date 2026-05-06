package devdesk

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadConfigAndPresetNames(t *testing.T) {
	path := writeConfig(t, `presets:
  main-workspace:
    close:
      - Example Chat
    open:
      - Example Editor
  zodiak:
    close_all: true
    open:
      - Example Database
`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if got, want := cfg.PresetNames(), []string{"main-workspace", "zodiak"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("PresetNames() = %v, want %v", got, want)
	}
}

func TestUnknownPresetShowsAvailablePresets(t *testing.T) {
	cfg := Config{Presets: map[string]Preset{"main-workspace": {Open: []string{"Example Editor"}}}}

	_, err := cfg.Preset("missing")
	if err == nil {
		t.Fatal("Preset() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "available presets: main-workspace") {
		t.Fatalf("Preset() error = %q, want available preset", err.Error())
	}
}

func TestPresetValidationAllowsCommandsOnly(t *testing.T) {
	err := Preset{Commands: []string{"pwd"}}.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestPresetValidationRequiresAtLeastOneAction(t *testing.T) {
	err := Preset{}.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "must define") {
		t.Fatalf("Validate() error = %q, want action error", err.Error())
	}
}

func TestPresetValidationRejectsCloseAllWithCloseApps(t *testing.T) {
	err := Preset{CloseAll: true, Close: []string{"Example Chat"}}.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want close_all conflict")
	}

	if !strings.Contains(err.Error(), "close_all cannot be combined") {
		t.Fatalf("Validate() error = %q, want close_all conflict", err.Error())
	}
}

func TestSaveConfigOmitsEmptyCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devdesk.yaml")
	cfg := Config{Presets: map[string]Preset{
		"main-workspace": {
			Open: []string{"Example Editor"},
		},
	}}

	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if strings.Contains(string(data), "commands:") {
		t.Fatalf("config = %q, want commands omitted", string(data))
	}
}

func TestExpandPathExpandsHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	got, err := ExpandPath("~/devdesk.yaml")
	if err != nil {
		t.Fatalf("ExpandPath() error = %v", err)
	}

	want := filepath.Join(home, "devdesk.yaml")
	if got != want {
		t.Fatalf("ExpandPath() = %q, want %q", got, want)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "devdesk.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	return path
}
