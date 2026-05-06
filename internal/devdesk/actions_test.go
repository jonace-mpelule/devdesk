package devdesk

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestPlanActionsOrdersCloseOpenCommands(t *testing.T) {
	preset := Preset{
		CloseAll: true,
		Close:    []string{"Example Chat", "Example Mail"},
		Open:     []string{"Example Editor", "Example Database"},
		Commands: []string{"echo main-workspace"},
	}

	got := ActionSummary(PlanActions(preset))
	want := []string{
		"close_app:Example Chat",
		"close_app:Example Mail",
		"close_all",
		"open_app:Example Editor",
		"open_app:Example Database",
		"command:echo main-workspace",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PlanActions() = %v, want %v", got, want)
	}
}

func TestPrintDryRun(t *testing.T) {
	preset := Preset{
		Close:    []string{"Example Chat"},
		Open:     []string{"Example Editor"},
		Commands: []string{"pwd"},
	}

	var out bytes.Buffer
	PrintDryRun(preset, &out)

	for _, want := range []string{
		"Dry run:",
		"- close app: Example Chat",
		"- open app: Example Editor",
		"- run in background: pwd",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("PrintDryRun() = %q, want %q", out.String(), want)
		}
	}
}

func TestApplyPresetContinuesOnAppFailuresButStopsOnCommandFailure(t *testing.T) {
	executor := &fakeExecutor{commandErr: errBoom{}}
	preset := Preset{
		Close:    []string{"Example Chat"},
		Open:     []string{"Missing"},
		Commands: []string{"pwd"},
	}

	var out bytes.Buffer
	err := ApplyPreset(preset, executor, &out)
	if err == nil {
		t.Fatal("ApplyPreset() error = nil, want command error")
	}

	if got, want := executor.calls, []string{"close:Example Chat", "open:Missing", "command:pwd"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("executor calls = %v, want %v", got, want)
	}
}

type fakeExecutor struct {
	calls      []string
	commandErr error
}

func (f *fakeExecutor) CloseAllApps() error {
	f.calls = append(f.calls, "close_all")
	return nil
}

func (f *fakeExecutor) CloseApp(name string) error {
	f.calls = append(f.calls, "close:"+name)
	return nil
}

func (f *fakeExecutor) OpenApp(name string) error {
	f.calls = append(f.calls, "open:"+name)
	return nil
}

func (f *fakeExecutor) RunCommand(command string) error {
	f.calls = append(f.calls, "command:"+command)
	return f.commandErr
}

type errBoom struct{}

func (errBoom) Error() string {
	return "boom"
}
