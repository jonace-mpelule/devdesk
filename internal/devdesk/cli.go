package devdesk

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

var workerStarter WorkerStarter = ProcessWorkerStarter{}

func RunCLI(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("devdesk", flag.ContinueOnError)
	flags.SetOutput(stderr)

	configPath := flags.String("config", "", "Path to devdesk YAML config")
	dryRun := flags.Bool("dry-run", false, "Print planned actions without executing them")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() == 0 {
		printUsage(stdout)
		return nil
	}

	command := flags.Arg(0)
	if command == "create" {
		return createWorkspace(*configPath, flags.Args()[1:], stdin, stdout, stderr)
	}
	if command == "edit" {
		return editWorkspace(*configPath, flags.Args()[1:], stdin, stdout, stderr)
	}
	if command == "worker" {
		return runWorkerCommand(flags.Args()[1:], stdout, stderr)
	}

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		return err
	}

	if command == "list" {
		printBanner(stdout)
		names := cfg.PresetNames()
		if len(names) == 0 {
			fmt.Fprintln(stdout, "No presets configured.")
			return nil
		}

		fmt.Fprintln(stdout, "Presets:")
		for _, name := range names {
			fmt.Fprintf(stdout, "- %s\n", name)
		}
		return nil
	}

	preset, err := cfg.Preset(command)
	if err != nil {
		return err
	}

	if *dryRun {
		PrintDryRun(preset, stdout)
		return nil
	}

	return StartPresetInBackground(*configPath, command, workerStarter, stdout)
}

func runWorkerCommand(args []string, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("devdesk worker", flag.ContinueOnError)
	flags.SetOutput(stderr)

	configPath := flags.String("config", "", "Path to devdesk YAML config")
	presetName := flags.String("preset", "", "Preset name to run")
	runID := flags.String("run-id", "", "Run id for logs")

	if err := flags.Parse(args); err != nil {
		return err
	}
	return RunWorker(*configPath, *presetName, *runID, stdout)
}

func createWorkspace(configPath string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("devdesk create", flag.ContinueOnError)
	flags.SetOutput(stderr)

	var closeApps stringList
	var openApps stringList
	var commands stringList
	flags.Var(&closeApps, "close", "App to close. Can be used multiple times.")
	flags.Var(&openApps, "open", "App to open. Can be used multiple times.")
	flags.Var(&commands, "command", "Setup command to run. Can be used multiple times.")
	closeAll := flags.Bool("close-all", false, "Close regular visible apps before opening the workspace")
	force := flags.Bool("force", false, "Replace an existing workspace with the same name")

	if len(args) == 0 {
		return createWorkspaceWizard(configPath, "", stdin, stdout)
	}

	name := args[0]
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: devdesk create <workspace> [--open app] [--close app] [--command command] [--close-all] [--force]")
	}
	if !hasCreateFlags(args[1:]) {
		return createWorkspaceWizard(configPath, name, stdin, stdout)
	}

	preset := Preset{
		CloseAll: *closeAll,
		Close:    closeApps,
		Open:     openApps,
		Commands: commands,
	}

	return saveWorkspace(configPath, name, preset, *force, stdout)
}

func editWorkspace(configPath string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("devdesk edit", flag.ContinueOnError)
	flags.SetOutput(stderr)

	var closeApps stringList
	var openApps stringList
	var commands stringList
	flags.Var(&closeApps, "close", "App to close. Can be used multiple times.")
	flags.Var(&openApps, "open", "App to open. Can be used multiple times.")
	flags.Var(&commands, "command", "Setup command to run. Can be used multiple times.")
	closeAll := flags.Bool("close-all", false, "Close regular visible apps before opening the workspace")

	if len(args) == 0 {
		return errors.New("usage: devdesk edit <workspace> [--open app] [--close app] [--command command] [--close-all]")
	}

	name := args[0]
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: devdesk edit <workspace> [--open app] [--close app] [--command command] [--close-all]")
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	existing, err := cfg.Preset(name)
	if err != nil {
		return err
	}

	if !hasCreateFlags(args[1:]) {
		return editWorkspaceWizard(configPath, name, existing, stdin, stdout)
	}

	preset := Preset{
		CloseAll: *closeAll,
		Close:    closeApps,
		Open:     openApps,
		Commands: commands,
	}

	return saveWorkspace(configPath, name, preset, true, stdout)
}

func createWorkspaceWizard(configPath string, initialName string, stdin io.Reader, stdout io.Writer) error {
	reader := bufio.NewReader(stdin)

	printBanner(stdout)
	fmt.Fprintln(stdout, "Create a DevDesk workspace")
	fmt.Fprintln(stdout, "Press Enter to accept defaults shown in brackets.")
	fmt.Fprintln(stdout)

	name, err := promptRequired(reader, stdout, "Workspace name", initialName)
	if err != nil {
		return err
	}

	preset, err := promptPreset(reader, stdout, Preset{})
	if err != nil {
		return err
	}

	cfg, err := LoadConfigForUpdate(configPath)
	if err != nil {
		return err
	}

	overwrite := false
	if _, exists := cfg.Presets[name]; exists {
		overwrite, err = promptBool(reader, stdout, fmt.Sprintf("Workspace %q already exists. Replace it?", name), false)
		if err != nil {
			return err
		}
		if !overwrite {
			return errors.New("workspace creation cancelled")
		}
	}

	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Workspace preview:")
	PrintDryRun(preset, stdout)

	confirm, err := promptBool(reader, stdout, "Save this workspace?", true)
	if err != nil {
		return err
	}
	if !confirm {
		return errors.New("workspace creation cancelled")
	}

	return saveWorkspace(configPath, name, preset, overwrite, stdout)
}

func editWorkspaceWizard(configPath string, name string, initialPreset Preset, stdin io.Reader, stdout io.Writer) error {
	reader := bufio.NewReader(stdin)

	printBanner(stdout)
	fmt.Fprintf(stdout, "Edit workspace %q\n", name)
	fmt.Fprintln(stdout, "Press Enter to keep the current values shown in brackets.")
	fmt.Fprintln(stdout)

	preset, err := promptPreset(reader, stdout, initialPreset)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Workspace preview:")
	PrintDryRun(preset, stdout)

	confirm, err := promptBool(reader, stdout, "Save this workspace?", true)
	if err != nil {
		return err
	}
	if !confirm {
		return errors.New("workspace edit cancelled")
	}

	return saveWorkspace(configPath, name, preset, true, stdout)
}

func promptPreset(reader *bufio.Reader, stdout io.Writer, initialPreset Preset) (Preset, error) {
	for {
		closeAll, err := promptBool(reader, stdout, "Close all regular apps before opening this workspace?", initialPreset.CloseAll)
		if err != nil {
			return Preset{}, err
		}

		var closeApps []string
		if !closeAll {
			closeFallback := initialPreset.Close
			if initialPreset.CloseAll {
				closeFallback = nil
			}
			closeApps, err = promptList(reader, stdout, "Apps to close, separated by commas", closeFallback)
			if err != nil {
				return Preset{}, err
			}
		}

		openApps, err := promptList(reader, stdout, "Apps to open, separated by commas", initialPreset.Open)
		if err != nil {
			return Preset{}, err
		}

		commands, err := promptList(reader, stdout, "Setup commands to run, separated by commas", initialPreset.Commands)
		if err != nil {
			return Preset{}, err
		}

		preset := Preset{
			CloseAll: closeAll,
			Close:    closeApps,
			Open:     openApps,
			Commands: commands,
		}
		if err := preset.Validate(); err == nil {
			return preset, nil
		}

		fmt.Fprintln(stdout, "Add at least one action: close all apps, close an app, open an app, or run a setup command.")
		fmt.Fprintln(stdout)
	}
}

func saveWorkspace(configPath string, name string, preset Preset, overwrite bool, stdout io.Writer) error {
	cfg, err := LoadConfigForUpdate(configPath)
	if err != nil {
		return err
	}
	if err := cfg.SetPreset(name, preset, overwrite); err != nil {
		return err
	}
	if err := SaveConfig(configPath, cfg); err != nil {
		return err
	}

	resolved, err := ExpandPath(configPath)
	if err != nil {
		return err
	}
	verb := "Created"
	if overwrite {
		verb = "Updated"
	}
	fmt.Fprintf(stdout, "%s workspace %q in %s\n", verb, name, resolved)
	fmt.Fprintf(stdout, "Preview it with: devdesk --dry-run %s\n", name)
	return nil
}

func hasCreateFlags(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return true
		}
	}
	return false
}

func promptRequired(reader *bufio.Reader, stdout io.Writer, label string, fallback string) (string, error) {
	for {
		value, err := promptString(reader, stdout, label, fallback)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(value) != "" {
			return value, nil
		}
		fmt.Fprintln(stdout, "Please enter a value.")
	}
}

func promptString(reader *bufio.Reader, stdout io.Writer, label string, fallback string) (string, error) {
	if fallback == "" {
		fmt.Fprintf(stdout, "%s: ", label)
	} else {
		fmt.Fprintf(stdout, "%s [%s]: ", label, fallback)
	}

	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if errors.Is(err, io.EOF) && value == "" && fallback == "" {
		return "", io.EOF
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	return value, nil
}

func promptBool(reader *bufio.Reader, stdout io.Writer, label string, fallback bool) (bool, error) {
	suffix := "y/N"
	if fallback {
		suffix = "Y/n"
	}

	for {
		fmt.Fprintf(stdout, "%s (%s): ", label, suffix)
		value, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return false, err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return fallback, nil
		}

		switch strings.ToLower(value) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(stdout, "Please answer yes or no.")
		}
	}
}

func promptList(reader *bufio.Reader, stdout io.Writer, label string, fallback []string) ([]string, error) {
	value, err := promptString(reader, stdout, label, strings.Join(fallback, ", "))
	if err != nil {
		return nil, err
	}
	return splitList(value), nil
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func printUsage(stdout io.Writer) {
	printBanner(stdout)
	fmt.Fprintln(stdout, `Usage:
  devdesk [--config path] [--dry-run] <preset>
  devdesk [--config path] list
  devdesk [--config path] create [workspace] [options]
  devdesk [--config path] edit <workspace> [options]

Default config:
  ~/.devdesk.yaml

Create options:
  --open app         App to open. Can be used multiple times.
  --close app        App to close. Can be used multiple times.
  --command command  Setup command to run. Can be used multiple times.
  --close-all        Close regular visible apps before opening the workspace.
  --force            Replace an existing workspace.

Run devdesk create or devdesk edit without options for the guided setup.`)
}
