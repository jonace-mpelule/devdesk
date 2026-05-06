package devdesk

import (
	"fmt"
	"io"
	"strings"
)

type ActionKind string

const (
	ActionCloseAll ActionKind = "close_all"
	ActionCloseApp ActionKind = "close_app"
	ActionOpenApp  ActionKind = "open_app"
	ActionCommand  ActionKind = "command"
)

type Action struct {
	Kind   ActionKind
	Target string
}

func PlanActions(preset Preset) []Action {
	actions := make([]Action, 0, 1+len(preset.Close)+len(preset.Open)+len(preset.Commands))

	for _, app := range preset.Close {
		actions = append(actions, Action{Kind: ActionCloseApp, Target: app})
	}

	if preset.CloseAll {
		actions = append(actions, Action{Kind: ActionCloseAll})
	}

	for _, app := range preset.Open {
		actions = append(actions, Action{Kind: ActionOpenApp, Target: app})
	}

	for _, command := range preset.Commands {
		actions = append(actions, Action{Kind: ActionCommand, Target: command})
	}

	return actions
}

type Executor interface {
	CloseAllApps() error
	CloseApp(name string) error
	OpenApp(name string) error
	RunCommand(command string) error
}

func ApplyPreset(preset Preset, executor Executor, stdout io.Writer) error {
	for _, action := range PlanActions(preset) {
		switch action.Kind {
		case ActionCloseAll:
			fmt.Fprintln(stdout, "Closing all regular apps...")
			if err := executor.CloseAllApps(); err != nil {
				fmt.Fprintf(stdout, "  failed to close all apps: %v\n", err)
			}
		case ActionCloseApp:
			fmt.Fprintf(stdout, "Closing %s...\n", action.Target)
			if err := executor.CloseApp(action.Target); err != nil {
				fmt.Fprintf(stdout, "  failed to close %s: %v\n", action.Target, err)
			}
		case ActionOpenApp:
			fmt.Fprintf(stdout, "Opening %s...\n", action.Target)
			if err := executor.OpenApp(action.Target); err != nil {
				fmt.Fprintf(stdout, "  failed to open %s: %v\n", action.Target, err)
			}
		case ActionCommand:
			fmt.Fprintf(stdout, "Starting command: %s\n", action.Target)
			if err := executor.RunCommand(action.Target); err != nil {
				return fmt.Errorf("start command %q: %w", action.Target, err)
			}
		}
	}

	return nil
}

func PrintDryRun(preset Preset, stdout io.Writer) {
	printBanner(stdout)
	fmt.Fprintln(stdout, "Dry run:")
	for _, action := range PlanActions(preset) {
		switch action.Kind {
		case ActionCloseAll:
			fmt.Fprintln(stdout, "- close all regular apps")
		case ActionCloseApp:
			fmt.Fprintf(stdout, "- close app: %s\n", action.Target)
		case ActionOpenApp:
			fmt.Fprintf(stdout, "- open app: %s\n", action.Target)
		case ActionCommand:
			fmt.Fprintf(stdout, "- run in background: %s\n", action.Target)
		}
	}
}

func ActionSummary(actions []Action) []string {
	summary := make([]string, 0, len(actions))
	for _, action := range actions {
		parts := []string{string(action.Kind)}
		if action.Target != "" {
			parts = append(parts, action.Target)
		}
		summary = append(summary, strings.Join(parts, ":"))
	}
	return summary
}
