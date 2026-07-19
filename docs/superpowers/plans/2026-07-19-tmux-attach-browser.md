# tmux-attach-browser Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build, test, document, publish, and release a Go command named `tab` that browses, creates, and enters tmux sessions through a Bubble Tea interface.

**Architecture:** The command delegates orchestration to `internal/app`, tmux operations to an injected `internal/tmux` client, and terminal state to `internal/ui`. The UI uses Bubbles `list` and `textinput`; it returns a selected session before the app starts an interactive tmux client so Bubble Tea restores the terminal first.

**Tech Stack:** Go 1.24 or newer, Bubble Tea v2, Bubbles v2, Lip Gloss v2, GoReleaser v2, POSIX shell, PowerShell, GitHub Actions

---

## File Map

- `cmd/tab/main.go`: production wiring and process exit.
- `internal/app/run.go`: CLI parsing, UI orchestration, help, version, and exit codes.
- `internal/buildinfo/buildinfo.go`: release values injected by linker flags.
- `internal/tmux/session.go`: session data, parsing, sorting, and name validation.
- `internal/tmux/client.go`: executable-backed tmux operations.
- `internal/ui/model.go`: Bubble Tea state, messages, views, and selection result.
- `internal/ui/item.go`: Bubbles list item representation.
- `internal/ui/keys.go`: custom key bindings.
- `internal/ui/run.go`: production Bubble Tea program setup.
- Matching `_test.go` files: behavior tests for each package.
- `install.sh`, `install.ps1`: public installers.
- `scripts/test-install.sh`, `scripts/tmux-smoke.sh`: installer and real-tmux tests.
- `.goreleaser.yaml`: release archives and checksums.
- `.github/workflows/*.yml`: cross-platform CI and tagged releases.
- `README.md` and community files: public documentation and contribution policy.

### Task 1: Module, build metadata, and command orchestration

**Files:**
- Create: `go.mod`
- Create: `cmd/tab/main.go`
- Create: `internal/buildinfo/buildinfo.go`
- Create: `internal/app/run.go`
- Test: `internal/app/run_test.go`

- [ ] **Step 1: Initialize dependencies**

```bash
go mod init github.com/hmmhmmhm/tmux-attach-browser
go get charm.land/bubbletea/v2@v2.0.8
go get charm.land/bubbles/v2@v2.1.1
go get charm.land/lipgloss/v2@v2.0.5
```

Expected: `go.mod` and `go.sum` contain the three direct dependencies.

- [ ] **Step 2: Write failing orchestration tests**

Drive this API from `run_test.go`:

```go
type Dependencies struct {
	Stdout io.Writer
	Stderr io.Writer
	Getwd  func() (string, error)
	Getenv func(string) string
	Client tmux.Client
	Browse func(tmux.Client, string) (string, bool, error)
}

func Run(args []string, deps Dependencies) int
```

Cover `--version`, `--help`, too many arguments, direct attach, browser attach, browser cancellation, browser failure, and connect failure. Assert output, selected session, nested-tmux flag, and exit code.

- [ ] **Step 3: Verify red**

Run: `go test ./internal/app -run TestRun -v`

Expected: FAIL because the package does not exist.

- [ ] **Step 4: Implement the minimal command**

Use these development build values:

```go
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
```

`Run` supports `-h`, `--help`, `-v`, `--version`, zero positional arguments, and one session name. It detects nesting with `Getenv("TMUX") != ""`. `cmd/tab/main.go` constructs the executable tmux client, wires `ui.Run`, and exits with the returned code.

- [ ] **Step 5: Verify and commit**

```bash
go test ./internal/app -v
go build ./cmd/tab
git add go.mod go.sum cmd/tab internal/app internal/buildinfo
git commit -m "feat: add tab command entry point"
```

Expected: tests pass and the command builds.

### Task 2: Session parsing and validation

**Files:**
- Create: `internal/tmux/session.go`
- Test: `internal/tmux/session_test.go`

- [ ] **Step 1: Write failing tests for the intended API**

```go
type Session struct {
	Name       string
	Windows    int
	Attached   int
	CreatedAt  time.Time
	ActivityAt time.Time
}

func ParseSessions(data []byte) ([]Session, error)
func ValidateSessionName(name string) error
```

Test empty output, valid rows, Unicode, recent-first sorting, invalid field counts, nonnumeric counts, and nonnumeric timestamps. Table-test blank names, whitespace, colon, period, tab, newline, and valid Unicode or dashed names.

- [ ] **Step 2: Verify red**

Run: `go test ./internal/tmux -run 'Test(Parse|Validate)' -v`

Expected: FAIL because the parser and validator do not exist.

- [ ] **Step 3: Implement strict parsing**

Use this stable list format:

```go
const sessionFormat = "#{session_name}\t#{session_windows}\t#{session_attached}\t#{session_created}\t#{session_activity}"
```

Require five fields, parse numbers with `strconv`, convert Unix seconds, and stable-sort by activity descending with name as the tie-breaker. Trim names and reject blank input, control characters, colon, and period with actionable errors.

- [ ] **Step 4: Verify and commit**

```bash
go test ./internal/tmux -run 'Test(Parse|Validate)' -v
git add internal/tmux/session.go internal/tmux/session_test.go
git commit -m "feat: parse tmux sessions"
```

Expected: all parser and validator tests pass.

### Task 3: Executable-backed tmux client

**Files:**
- Create: `internal/tmux/client.go`
- Test: `internal/tmux/client_test.go`

- [ ] **Step 1: Write failing fake-executable tests**

Define:

```go
type Client interface {
	List(context.Context) ([]Session, error)
	Create(context.Context, string, string) error
	Connect(context.Context, string, bool) error
}
```

A temporary executable records arguments and returns fixture output. Verify `list-sessions -F`, no-server handling, unrelated failures, `new-session -d -s work -c /tmp/project`, `attach-session -t work`, `switch-client -t work`, and missing executable errors.

- [ ] **Step 2: Verify red**

Run: `go test ./internal/tmux -run TestExecClient -v`

Expected: FAIL because `ExecClient` does not exist.

- [ ] **Step 3: Implement without a shell**

Use `exec.CommandContext` and pass all user values as separate arguments. `List` captures stdout and stderr; only known no-server stderr becomes an empty list. `Create` and `Connect` inherit standard streams where needed for interactive tmux.

- [ ] **Step 4: Verify and commit**

```bash
go test ./internal/tmux -v
go test -race ./internal/tmux
git add internal/tmux/client.go internal/tmux/client_test.go
git commit -m "feat: execute tmux session commands"
```

Expected: all tests pass without race reports.

### Task 4: Bubbles session list

**Files:**
- Create: `internal/ui/item.go`
- Create: `internal/ui/keys.go`
- Create: `internal/ui/model.go`
- Test: `internal/ui/model_test.go`

- [ ] **Step 1: Write failing state-transition tests**

Create a fake client and test initialization, list rendering, Enter selection, refresh replacement, refresh error retention, window resizing, and Ctrl+C cancellation. Expose the result through:

```go
func (m Model) Result() (session string, selected bool)
```

- [ ] **Step 2: Verify red**

Run: `go test ./internal/ui -run 'Test(Init|Sessions|Enter|Refresh|Window|Ctrl)' -v`

Expected: FAIL because the UI package does not exist.

- [ ] **Step 3: Implement the list model**

`sessionItem` implements `list.DefaultItem` with name as title and filter value. Its description renders window count, attached or detached state, and concise activity time. Configure `list.New` with the default delegate, fuzzy filtering, pagination, status, help, and `n` plus `r` custom bindings. During filtering, delegate keys to the list. Enter stores the selected session and returns `tea.Quit`. Load and refresh operations return typed messages from `tea.Cmd` functions.

- [ ] **Step 4: Verify and commit**

```bash
go test ./internal/ui -run 'Test(Init|Sessions|Enter|Refresh|Window|Ctrl)' -v
git add internal/ui/item.go internal/ui/keys.go internal/ui/model.go internal/ui/model_test.go
git commit -m "feat: browse tmux sessions"
```

Expected: all list model tests pass.

### Task 5: Create prompt and production UI runner

**Files:**
- Modify: `internal/ui/model.go`
- Modify: `internal/ui/model_test.go`
- Create: `internal/ui/run.go`

- [ ] **Step 1: Write failing create-flow tests**

Test `n` opening the prompt, Esc cancellation, blank and invalid names, successful creation, and failed creation that preserves input. Assert `Client.Create` receives the current directory.

- [ ] **Step 2: Verify red**

Run: `go test ./internal/ui -run 'Test(New|Escape|Blank|Invalid|Create)' -v`

Expected: create-flow assertions fail.

- [ ] **Step 3: Implement create mode using Bubbles textinput**

Use explicit list and create modes. Configure `textinput.New()` with prompt `session name: `, placeholder `project`, 100-rune limit, and `tmux.ValidateSessionName`. Esc returns to the list. Enter validates and runs a typed create command. Other messages go through `textinput.Update`. Render the prompt in a Lip Gloss panel with concise Enter and Esc help.

The production runner must restore the terminal before connection:

```go
func Run(client tmux.Client, cwd string) (string, bool, error) {
	p := tea.NewProgram(New(client, cwd), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil { return "", false, err }
	m, ok := result.(Model)
	if !ok { return "", false, fmt.Errorf("unexpected UI result %T", result) }
	name, selected := m.Result()
	return name, selected, nil
}
```

- [ ] **Step 4: Verify and commit**

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build -o dist/tab ./cmd/tab
./dist/tab --version
./dist/tab --help
git add cmd internal
git commit -m "feat: create sessions from the browser"
```

Expected: tests and vet pass; version and help match the CLI contract.

### Task 6: Cross-platform installers

**Files:**
- Create: `install.sh`
- Create: `install.ps1`
- Create: `scripts/test-install.sh`

- [ ] **Step 1: Write failing installer fixture tests**

Create a temporary fake archive and checksums file. Run the installer with `TAB_RELEASE_BASE_URL`, `TAB_VERSION`, `TAB_INSTALL_DIR`, `TAB_OS`, and `TAB_ARCH` overrides. Assert successful installation, checksum rejection, and unsupported architecture rejection.

- [ ] **Step 2: Verify red**

Run: `sh scripts/test-install.sh`

Expected: FAIL because `install.sh` does not exist.

- [ ] **Step 3: Implement `install.sh`**

Use `set -eu`, `mktemp -d`, and a cleanup trap. Detect Darwin or Linux plus amd64 or arm64, obtain the latest release when the version override is absent, download the matching archive and `checksums.txt`, verify with an available SHA-256 command, extract `tab`, and install mode 0755 into `$TAB_INSTALL_DIR` or `$HOME/.local/bin`. Never invoke sudo automatically and print a PATH hint when required.

- [ ] **Step 4: Implement `install.ps1`**

Accept version and install-directory overrides, map AMD64 and ARM64, download a Windows zip plus checksums, validate with `Get-FileHash`, and extract `tab.exe`. Do not mutate PATH. Explain that WSL2 is the supported Windows path when tmux is absent.

- [ ] **Step 5: Verify and commit**

```bash
sh -n install.sh
sh scripts/test-install.sh
command -v shellcheck >/dev/null && shellcheck install.sh scripts/test-install.sh || true
git add install.sh install.ps1 scripts/test-install.sh
git commit -m "feat: add cross-platform installers"
```

Expected: syntax, fixture, and available shellcheck checks pass.

### Task 7: Release automation and CI

**Files:**
- Create: `.goreleaser.yaml`
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`
- Create: `scripts/tmux-smoke.sh`

- [ ] **Step 1: Add a real tmux smoke test**

Use a unique socket name and cleanup trap. Create two detached sessions, request the same five-field session format, and verify both are targetable. Clean up only that socket.

- [ ] **Step 2: Configure GoReleaser v2**

Build `tab` with `CGO_ENABLED=0` for Darwin, Linux, and Windows on amd64 and arm64. Use `tab_{{ .Os }}_{{ .Arch }}` archives, zip for Windows and tar.gz elsewhere, generate `checksums.txt`, and inject version, commit, and date build fields.

- [ ] **Step 3: Add GitHub workflows**

CI runs gofmt verification, vet, race tests, and builds on Ubuntu, macOS, and Windows. An Ubuntu job installs tmux and runs both shell scripts. The release workflow runs the official GoReleaser action for `v*` tags with `GITHUB_TOKEN`.

- [ ] **Step 4: Verify and commit**

```bash
go install github.com/goreleaser/goreleaser/v2@latest
goreleaser check
goreleaser release --snapshot --clean
find dist -maxdepth 2 -type f -print | sort
git add .goreleaser.yaml .github/workflows scripts/tmux-smoke.sh
git commit -m "ci: test and release cross-platform builds"
```

Expected: six platform archives and `checksums.txt` exist in the snapshot.

### Task 8: Public documentation and policies

**Files:**
- Create: `README.md`
- Create: `CONTRIBUTING.md`
- Create: `CODE_OF_CONDUCT.md`
- Create: `SECURITY.md`
- Create: `LICENSE`
- Create: `.github/ISSUE_TEMPLATE/bug_report.yml`
- Create: `.github/ISSUE_TEMPLATE/feature_request.yml`

- [ ] **Step 1: Write public usage documentation**

Document the exact curl and PowerShell commands, `go install` fallback, tmux prerequisites, WSL2 support boundary, key map, direct attach shortcut, source build, troubleshooting, and license. Include a text terminal example that matches the implemented view.

- [ ] **Step 2: Add community health files**

Use the MIT license with copyright `2026 Hamin`, concise contribution checks matching CI, Contributor Covenant 2.1 with a GitHub contact route, private vulnerability reports through GitHub Security Advisories, and issue forms requesting platform, terminal, tmux version, tab version, steps, and expected behavior.

- [ ] **Step 3: Verify and commit**

```bash
rg -n "tab |Enter|WSL|curl|PowerShell|tmux" README.md
rg -n "TBD|TODO|FIXME|—" README.md CONTRIBUTING.md CODE_OF_CONDUCT.md SECURITY.md LICENSE .github/ISSUE_TEMPLATE || true
git diff --check
git add README.md CONTRIBUTING.md CODE_OF_CONDUCT.md SECURITY.md LICENSE .github/ISSUE_TEMPLATE
git commit -m "docs: prepare the project for contributors"
```

Expected: documentation matches behavior and contains no placeholders.

### Task 9: Full verification and v0.1.0 release

**Files:**
- Modify only if verification exposes defects.

- [ ] **Step 1: Run all local gates**

```bash
test -z "$(gofmt -l .)"
go vet ./...
go test -race ./...
go build -trimpath -o dist/tab ./cmd/tab
sh -n install.sh
sh scripts/test-install.sh
scripts/tmux-smoke.sh
goreleaser check
goreleaser release --snapshot --clean
git diff --check
git status --short
```

Expected: every command exits 0 and only intended files appear.

- [ ] **Step 2: Exercise the TUI in a PTY**

Create uniquely named temporary tmux sessions. Verify list display, navigation, filter, create prompt, cancellation, creation, and selection. Run inside tmux and verify selection chooses `switch-client`. Remove only those temporary sessions.

- [ ] **Step 3: Push main and observe CI**

```bash
git push origin main
gh run list --repo hmmhmmhm/tmux-attach-browser --limit 5
```

Expected: main CI succeeds on Ubuntu, macOS, and Windows.

- [ ] **Step 4: Publish and verify v0.1.0**

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
gh run watch --repo hmmhmmhm/tmux-attach-browser --exit-status
gh release view v0.1.0 --repo hmmhmmhm/tmux-attach-browser
```

Expected: the release has six platform archives and `checksums.txt`.

- [ ] **Step 5: Prove the public installer**

Run the exact public curl installer with a fresh `TAB_INSTALL_DIR`, retain that path, and execute `<path>/tab --version`.

Expected: checksum verification succeeds and the installed binary prints `tab v0.1.0`.

- [ ] **Step 6: Record evidence**

Update `projects/tmux-attach-browser/PROJECT.md` with commit, CI, release, and installer evidence. Confirm the repository is clean and tracking `origin/main`.
