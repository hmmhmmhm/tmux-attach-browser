# tmux-attach-browser Design

## Summary

tmux-attach-browser is a small Go command named `tab`. Running `tab` opens a terminal session browser that lists existing tmux sessions, attaches to the selected session, or creates a new session without requiring the user to remember tmux commands.

The first release targets macOS and Linux directly. Windows 10 and 11 are supported through WSL2, where tmux has a native Linux environment. A Windows binary and PowerShell installer are also published for users who already provide tmux through MSYS2 or Cygwin, but WSL2 is the documented and tested Windows path.

## Goals

- Make `tab` the only command a user needs to remember.
- List tmux sessions with enough context to choose the right one.
- Attach to an existing session with Enter.
- Create and attach to a named session from the current directory.
- Work correctly both outside and inside tmux.
- Install from a one-line curl command on macOS, Ubuntu, other common Linux distributions, and Windows through WSL2.
- Publish reproducible release archives and checksums for macOS, Linux, and Windows on amd64 and arm64.
- Keep the tmux integration and Bubble Tea state transitions independently testable.

## Non-goals for v0.1

- Managing windows or panes.
- Renaming or deleting sessions.
- Editing tmux configuration.
- Remote host discovery.
- Persisted themes or user configuration.
- Installing tmux itself.

## Approaches Considered

### Bubbles list plus text input

Use the Bubbles `list` component for session navigation, filtering, pagination, status messages, and help. Use `textinput` for the create-session prompt. This keeps application-specific code focused on tmux behavior. This is the selected approach.

### Bubbles table plus a custom prompt

A table can align session metadata into columns, but it degrades more abruptly in narrow terminals and requires additional work for filtering, pagination, empty states, and help. The extra information density is not important enough for the first release.

### Custom viewport and renderer

A custom viewport provides complete control over appearance and interaction, but recreates navigation, filtering, and accessibility behaviors already provided by Bubbles. It increases code and test surface without improving the core workflow.

## User Experience

### Launch

The installed executable is `tab`. With no arguments, it validates that `tmux` is available, loads the current sessions, and opens the Bubble Tea alternate screen.

The list is ordered by most recent activity. Each session item shows:

- session name;
- window count;
- attached or detached state;
- relative or concise last activity time.

The title is `tmux sessions`. A zero-session state explains that `n` creates the first session.

### Key bindings

- Up, Down, `j`, `k`: move through sessions.
- Enter: attach to the selected session.
- `n`: open the new-session prompt.
- `/`: use the list component's fuzzy filter.
- `r`: refresh sessions.
- `?`: expand help.
- Esc: cancel filtering or the new-session prompt.
- `q`, Ctrl+C: quit when no input field is active.

The new-session prompt uses the Bubbles `textinput` component. Enter validates the name, creates the session detached with the invoking process's current directory, exits the TUI cleanly, and attaches to the new session. Duplicate names and tmux errors remain in the TUI as actionable messages.

### Attaching

Bubble Tea exits and restores the terminal before tmux starts. Outside tmux, the command runs `tmux attach-session -t <name>`. If the `TMUX` environment variable is set, it runs `tmux switch-client -t <name>` to avoid nested tmux clients.

The command also accepts `tab <session>` as a direct attach shortcut. `tab --version` prints build version information and `tab --help` prints concise usage.

## Architecture

### `cmd/tab`

Owns argument parsing, process exit codes, Bubble Tea startup, and execution of the final attach or switch action after the terminal UI has shut down.

### `internal/tmux`

Defines a `Client` interface and an `ExecClient` implementation. It lists sessions with a stable tab-separated tmux format, parses command output, creates detached sessions, and attaches or switches clients. Commands are passed as argument arrays, never through a shell.

The list format includes session name, window count, attached client count, creation timestamp, and activity timestamp. A tmux exit status that specifically means there is no running server is treated as an empty list. Missing executables, permission failures, malformed output, and other errors are surfaced.

### `internal/ui`

Owns the Bubble Tea model and uses:

- `charm.land/bubbles/v2/list` for navigation, fuzzy filtering, pagination, help, and status messages;
- `charm.land/bubbles/v2/textinput` for session creation;
- `charm.land/lipgloss/v2` for adaptive layout and restrained styling.

The model receives tmux operations through the `Client` interface. Commands return typed messages so tmux I/O does not block or leak into state transition logic. The model produces a final action describing which session to attach to. It does not start an interactive tmux client itself.

### Build metadata

An internal version package exposes version, commit, and build date values injected at release time. Development builds report `dev`.

## Data Flow

1. `tab` validates arguments and tmux availability.
2. The UI initializes and asynchronously requests sessions.
3. The tmux client runs `list-sessions` with the stable format.
4. Parsed sessions become Bubbles list items.
5. The user selects a session or submits a new name.
6. For new sessions, the tmux client creates a detached session in the current directory.
7. The model stores the chosen session as its result and asks Bubble Tea to quit.
8. After terminal restoration, the command attaches or switches to that session.

## Error Handling

- If tmux is not installed, exit with code 1 and print platform-specific installation guidance.
- If there is no tmux server, render an empty list rather than an error.
- Reject blank session names and names containing `:` or `.` before invoking tmux because those characters make tmux target syntax ambiguous.
- Show duplicate-name and create failures inline while preserving the entered value.
- Show refresh errors as list status messages while retaining the last successful session list.
- If attach fails after leaving the UI, print the exact tmux error and return a nonzero exit code.
- Treat Ctrl+C as a clean user cancellation with exit code 0.

## Installation and Releases

The primary installation command is:

```sh
curl -fsSL https://raw.githubusercontent.com/hmmhmmhm/tmux-attach-browser/main/install.sh | sh
```

The POSIX installer detects Darwin or Linux and amd64 or arm64, downloads the latest matching GitHub release archive, verifies it against `checksums.txt`, and installs `tab` into `$TAB_INSTALL_DIR` or `$HOME/.local/bin`.

The repository also includes `install.ps1`. Its documented one-liner downloads the Windows archive and verifies its SHA-256 checksum. Windows users are directed to run the POSIX installer inside WSL2 for the supported tmux experience.

GoReleaser builds static archives for Darwin, Linux, and Windows on amd64 and arm64. GitHub Actions runs formatting, vetting, tests, and builds on pull requests and main. Version tags create GitHub releases.

## Testing

### Unit tests

- Parse valid, empty, Unicode, and malformed tmux session output.
- Distinguish a missing server from other tmux failures.
- Validate session names.
- Verify attach versus switch selection from the environment.
- Verify list loading, navigation, refresh, create prompt, validation, cancellation, create success, and create failure model transitions.
- Verify argument parsing and version output.

### Integration tests

- Use a temporary fake `tmux` executable to verify command arguments and exit-code behavior without requiring tmux on every platform.
- On Unix CI, run installer tests against local fixture archives and checksums.
- Run `go test ./...`, `go vet ./...`, and build checks on Linux, macOS, and Windows.
- Run a real tmux smoke test on Ubuntu CI to list, create, and target a temporary isolated server socket.

### Manual release verification

- Install the published archive with the public one-line installer on macOS and Ubuntu or WSL2.
- Create two sessions, confirm their metadata, attach, detach, re-open `tab`, filter, and reattach.
- Run `tab` from inside tmux and confirm it switches clients rather than nesting.

## Open Source Project Files

The repository includes an MIT license, English README, contributing guide, code of conduct, security policy, issue templates, and CI/release workflows. The README states the tmux prerequisite and the exact Windows support boundary.

## Success Criteria

- A new user with tmux installed can install `tab` from the documented one-liner and run it without additional configuration.
- Existing sessions appear and can be entered with keyboard navigation and Enter.
- A named session can be created and entered from the same UI.
- Running inside tmux switches clients without a nesting error.
- Unit, integration, and cross-platform CI checks pass.
- A public `hmmhmmhm/tmux-attach-browser` repository and an installable tagged release exist.
