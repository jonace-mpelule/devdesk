# DevDesk

DevDesk is a macOS CLI for switching workspaces without rebooting or manually closing and reopening the same apps. A preset can close apps, optionally close all regular apps, open the apps you need, and run setup commands through a detached worker.

## Install

From this repository:

```sh
go install ./cmd/devdesk
```

Or run it directly while developing:

```sh
go run ./cmd/devdesk create
```

## Configuration

DevDesk reads presets from `~/.devdesk.yaml` by default.

```yaml
presets:
  main-workspace:
    close:
      - Example Chat
    open:
      - Example Editor
```

### Preset fields

- `close_all`: when `true`, closes regular visible apps before opening the preset apps. DevDesk excludes core system apps.
- `close`: app names to quit.
- `open`: app names to launch with `open -a`.
- `commands`: optional shell commands DevDesk starts from the workspace worker.

## Usage

Create a workspace preset with the guided setup:

```sh
devdesk create
```

The wizard asks for:

- Workspace name
- Whether to close all regular apps
- Apps to close, only when close-all is disabled
- Apps to open
- Optional setup commands to run
- Final confirmation before saving

Enter app names as comma-separated macOS application names. This creates `~/.devdesk.yaml` if it does not exist. If you leave commands empty, DevDesk will not save a `commands:` section or run any commands.

You can also provide the workspace name first and still use the guided setup:

```sh
devdesk create main-workspace
```

Edit an existing workspace with the guided setup:

```sh
devdesk edit main-workspace
```

The edit flow loads the current workspace values and lets you keep them by pressing Enter or replace them with new values.

Create a workspace with explicit apps and commands without prompts:

```sh
devdesk create main-workspace \
  --close "Example Chat" \
  --open "Example Editor" \
  --command "echo Workspace started"
```

Replace an existing workspace:

```sh
devdesk create main-workspace --force --open "Example Editor"
```

Edit a workspace without prompts:

```sh
devdesk edit main-workspace --close-all --open "Example Editor" --command "echo Workspace updated"
```

Apply a preset:

```sh
devdesk main-workspace
```

DevDesk starts the workspace run in a detached background worker and prints the log directory:

```text
Started workspace "main-workspace" in the background.
Logs: /Users/you/.devdesk/logs/20260429-120000-12345
```

The whole workspace setup runs in a detached worker, so it can continue even if the terminal that launched `devdesk` is closed. Commands are launched by that worker, and each command writes to a file like `command-1.log` in the run directory.

List configured presets:

```sh
devdesk list
```

Preview actions without closing or opening anything:

```sh
devdesk --dry-run main-workspace
```

Use a different config file:

```sh
devdesk --config ./devdesk.yaml main-workspace
```

Create a workspace in a different config file:

```sh
devdesk --config ./devdesk.yaml create main-workspace --open "Example Editor"
```

## macOS Permissions

DevDesk uses AppleScript and System Events to inspect and quit visible apps for `close_all`. macOS may ask for Automation or Accessibility permission when you use app-closing features.

If app closing fails, enable permission in:

```text
System Settings -> Privacy & Security -> Accessibility
```

Grant access to the app that is running `devdesk`.

## Notes

- DevDesk v1 is macOS-only.
- YAML is the supported preset format.
- App names must match the macOS application name.
- Use `--dry-run` first when creating a new preset, especially with `close_all: true`.
