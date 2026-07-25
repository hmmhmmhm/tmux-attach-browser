# Project Health Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the code, test, CI, documentation, and repository-governance gaps found in the 2026-07-25 health review.

**Architecture:** Keep the existing package boundaries. Add request generations and a UI-lifetime context inside `internal/ui`, add focused POSIX test scripts for source length and the compiled CLI, and collect dedicated static checks in one Ubuntu CI job. Apply the GitHub ruleset only after every new pull-request check succeeds.

**Tech Stack:** Go 1.25, Bubble Tea v2, POSIX shell, GitHub Actions, `govulncheck` v1.6.0, `actionlint` v1.7.12, ShellCheck, GitHub repository rulesets.

---

## File map

- Modify `internal/ui/model.go`: load generations and context-aware commands.
- Modify `internal/ui/model_test.go`: stale-result and cancellation regression tests.
- Modify `internal/ui/run.go`: bind the command context to the Bubble Tea program lifetime.
- Create `scripts/check-source-lines.sh`: enforce the source-length policy.
- Create `scripts/test-source-lines.sh`: test the 450/451 boundaries.
- Create `scripts/tab-binary-smoke.sh`: exercise the compiled CLI with a fake tmux executable.
- Modify `CONTRIBUTING.md`: document policy and local checks.
- Modify `.github/workflows/ci.yml`: run policy, vulnerability, workflow, shell, and binary checks.
- Modify `.github/workflows/dependency-review.yml`: pin Dependency Review v5.0.0.
- Create GitHub repository ruleset after merge: protect `main`.

### Task 1: Reject stale refresh results

**Files:**
- Modify: `internal/ui/model_test.go`
- Modify: `internal/ui/model.go`

- [ ] **Step 1: Write the failing stale-result test**

Add a test that executes two refresh commands in request order but delivers
their messages in reverse order:

```go
func TestOlderRefreshResultCannotReplaceNewerResult(t *testing.T) {
	client := &fakeClient{lists: [][]tmux.Session{
		{session("initial", 1, 0)},
		{session("older", 1, 0)},
		{session("newer", 1, 0)},
	}}
	model := loadModel(t, client)

	model, olderCmd := updateModel(t, model, keyMsg('r', "r"))
	model, newerCmd := updateModel(t, model, keyMsg('r', "r"))
	olderMsg := olderCmd()
	newerMsg := newerCmd()

	model, _ = updateModel(t, model, newerMsg)
	model, _ = updateModel(t, model, olderMsg)

	item := model.list.Items()[0].(sessionItem)
	if item.Title() != "newer" || model.loading {
		t.Fatalf("title = %q, loading = %v", item.Title(), model.loading)
	}
}
```

- [ ] **Step 2: Run the test and verify RED**

Run:

```sh
go test ./internal/ui -run TestOlderRefreshResultCannotReplaceNewerResult -count=1
```

Expected: FAIL because the older result replaces `newer`.

- [ ] **Step 3: Add generation metadata**

Change the result and model state:

```go
type sessionsLoadedMsg struct {
	generation uint64
	sessions   []tmux.Session
	err        error
}
```

Add `loadGeneration uint64` after `loading bool` in `Model`. Initialize
generation 1, pass it from `Init`, and increment it for every `r` request:

```go
loadGeneration: 1

func (m Model) Init() tea.Cmd {
	return loadSessions(m.client, m.loadGeneration)
}

m.loadGeneration++
return m, loadSessions(m.client, m.loadGeneration)
```

Ignore mismatched messages before changing loading, errors, or items:

```go
case sessionsLoadedMsg:
	if msg.generation != m.loadGeneration {
		return m, nil
	}
	m.loading = false
```

Return the generation from the command:

```go
func loadSessions(client tmux.Client, generation uint64) tea.Cmd {
	return func() tea.Msg {
		sessions, err := client.List(context.Background())
		return sessionsLoadedMsg{
			generation: generation,
			sessions:   sessions,
			err:        err,
		}
	}
}
```

- [ ] **Step 4: Run focused and package tests**

Run:

```sh
go test ./internal/ui -run 'TestOlderRefreshResultCannotReplaceNewerResult|TestRefresh' -count=1
go test ./internal/ui -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```sh
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "fix: ignore stale session refreshes"
```

### Task 2: Cancel tmux work when the UI exits

**Files:**
- Modify: `internal/ui/model_test.go`
- Modify: `internal/ui/model.go`
- Modify: `internal/ui/run.go`

- [ ] **Step 1: Extend the fake client and write failing tests**

Add these fields to `fakeClient`:

```go
listContext   context.Context
createContext context.Context
```

Assign the context at the start of each existing fake method:

```go
func (f *fakeClient) List(ctx context.Context) ([]tmux.Session, error) {
	f.listContext = ctx
	index := f.listCalls
	f.listCalls++
	if index < len(f.listErrors) && f.listErrors[index] != nil {
		return nil, f.listErrors[index]
	}
	if len(f.lists) == 0 {
		return []tmux.Session{}, nil
	}
	if index >= len(f.lists) {
		index = len(f.lists) - 1
	}
	return f.lists[index], nil
}

func (f *fakeClient) Create(ctx context.Context, name, dir string) error {
	f.createContext = ctx
	f.createName = name
	f.createDir = dir
	return f.createError
}
```

Add focused tests:

```go
func TestLoadUsesModelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &fakeClient{}
	model := newModel(ctx, client, "/tmp/project")
	cancel()

	model.Init()()

	if !errors.Is(client.listContext.Err(), context.Canceled) {
		t.Fatalf("context error = %v", client.listContext.Err())
	}
}

func TestCreateUsesModelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &fakeClient{}
	model := newModel(ctx, client, "/tmp/project")
	model = loadModelFromModel(t, model)
	model, _ = updateModel(t, model, keyMsg('n', "n"))
	model.input.SetValue("work")
	model, cmd := updateModel(t, model, keyMsg(tea.KeyEnter, ""))
	cancel()

	cmd()

	if !errors.Is(client.createContext.Err(), context.Canceled) {
		t.Fatalf("context error = %v", client.createContext.Err())
	}
}
```

Use a helper that runs `Init` on a supplied model so the context-aware model is
not replaced by `New`:

```go
func loadModelFromModel(t *testing.T, model Model) Model {
	t.Helper()
	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init returned nil command")
	}
	model, _ = updateModel(t, model, cmd())
	return model
}
```

- [ ] **Step 2: Run the tests and verify RED**

Run:

```sh
go test ./internal/ui -run 'TestLoadUsesModelContext|TestCreateUsesModelContext' -count=1
```

Expected: compilation failure because `newModel` and the model context do not
exist.

- [ ] **Step 3: Add the model context**

Add `ctx context.Context` to `Model`. Preserve the public constructor:

```go
func New(client tmux.Client, cwd string) Model {
	return newModel(context.Background(), client, cwd)
}

func newModel(ctx context.Context, client tmux.Client, cwd string) Model {
	keys := newKeyMap()
	delegate := list.NewDefaultDelegate()
	sessionList := list.New(nil, delegate, 80, 24)
	sessionList.Title = "tmux sessions"
	sessionList.SetStatusBarItemName("session", "sessions")
	sessionList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.newSession, keys.refresh}
	}
	sessionList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keys.newSession, keys.refresh}
	}

	input := textinput.New()
	input.Prompt = "session name: "
	input.Placeholder = "project"
	input.CharLimit = 100
	input.Validate = tmux.ValidateSessionName
	input.SetWidth(48)

	return Model{
		ctx:            ctx,
		client:         client,
		cwd:            cwd,
		list:           sessionList,
		input:          input,
		keys:           keys,
		mode:           modeList,
		loading:        true,
		loadGeneration: 1,
	}
}
```

Change `Init` and the refresh path to call
`loadSessions(m.ctx, m.client, m.loadGeneration)`. Change both factories to:

```go
func loadSessions(ctx context.Context, client tmux.Client, generation uint64) tea.Cmd {
	return func() tea.Msg {
		sessions, err := client.List(ctx)
		return sessionsLoadedMsg{
			generation: generation,
			sessions:   sessions,
			err:        err,
		}
	}
}

func createSession(ctx context.Context, client tmux.Client, name, cwd string) tea.Cmd {
	return func() tea.Msg {
		err := client.Create(ctx, name, cwd)
		return sessionCreatedMsg{name: name, err: err}
	}
}
```

- [ ] **Step 4: Tie cancellation to `Run`**

Update `internal/ui/run.go`:

```go
func Run(client tmux.Client, cwd string) (string, bool, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return runProgram(tea.NewProgram(newModel(ctx, client, cwd)))
}
```

- [ ] **Step 5: Run UI tests and race tests**

Run:

```sh
go test ./internal/ui -count=1
go test -race ./internal/ui -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```sh
git add internal/ui/model.go internal/ui/model_test.go internal/ui/run.go
git commit -m "fix: cancel tmux work with the UI"
```

### Task 3: Enforce the 450-line source policy

**Files:**
- Create: `scripts/check-source-lines.sh`
- Create: `scripts/test-source-lines.sh`
- Modify: `CONTRIBUTING.md`

- [ ] **Step 1: Write the checker boundary test first**

Create `scripts/test-source-lines.sh` with a temporary git repository. Generate
450 lines, expect success, then append line 451 and expect failure:

```sh
#!/bin/sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
fixture=$(mktemp -d)
cleanup() {
	rm -rf "$fixture"
}
trap cleanup EXIT INT TERM

git -C "$fixture" init -q
awk 'BEGIN { for (i = 1; i <= 450; i++) print "# line" }' >"$fixture/boundary.sh"
git -C "$fixture" add boundary.sh

sh "$project_root/scripts/check-source-lines.sh" "$fixture"
printf '%s\n' "# line 451" >>"$fixture/boundary.sh"

if sh "$project_root/scripts/check-source-lines.sh" "$fixture"; then
	printf '%s\n' "expected a 451-line file to fail" >&2
	exit 1
fi

printf '%s\n' "source line limit tests passed"
```

- [ ] **Step 2: Run the test and verify RED**

Run:

```sh
sh scripts/test-source-lines.sh
```

Expected: FAIL because `scripts/check-source-lines.sh` does not exist.

- [ ] **Step 3: Implement the checker**

Create `scripts/check-source-lines.sh`:

```sh
#!/bin/sh
set -eu

max_lines=450
repository=${1:-$(git rev-parse --show-toplevel)}
scratch=$(mktemp -d)
cleanup() {
	rm -rf "$scratch"
}
trap cleanup EXIT INT TERM

git -C "$repository" ls-files -- \
	'*.go' '*.sh' '*.ps1' '*.yml' '*.yaml' >"$scratch/files"

status=0
while IFS= read -r file; do
	lines=$(awk 'END { print NR }' "$repository/$file")
	if [ "$lines" -gt "$max_lines" ]; then
		printf '%s: %s lines (maximum %s)\n' "$file" "$lines" "$max_lines" >&2
		status=1
	fi
done <"$scratch/files"

exit "$status"
```

- [ ] **Step 4: Document scope and commands**

Add to `CONTRIBUTING.md`:

```markdown
Keep every tracked, hand-written `.go`, `.sh`, `.ps1`, `.yml`, and `.yaml`
file at or below 450 lines. This includes tests and workflow configuration.
Documentation, generated output, vendored code, and binary assets are excluded.
Split a file by responsibility before it exceeds the limit.
```

Add both source-line scripts to the local check block.

- [ ] **Step 5: Run tests and policy check**

Run:

```sh
sh scripts/test-source-lines.sh
sh scripts/check-source-lines.sh
```

Expected: the fixture reports its intentional 451-line failure, the test ends
with `source line limit tests passed`, and the real repository returns zero.

- [ ] **Step 6: Commit**

```sh
git add CONTRIBUTING.md scripts/check-source-lines.sh scripts/test-source-lines.sh
git commit -m "ci: enforce source file line limit"
```

### Task 4: Exercise the real CLI binary

**Files:**
- Create: `scripts/tab-binary-smoke.sh`
- Modify: `.github/workflows/ci.yml`
- Modify: `CONTRIBUTING.md`

- [ ] **Step 1: Add the binary smoke script**

Create a temporary fake `tmux`, build the real binary, and check both connection
modes:

```sh
#!/bin/sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
scratch=$(mktemp -d)
cleanup() {
	rm -rf "$scratch"
}
trap cleanup EXIT INT TERM

go build -o "$scratch/tab" "$project_root/cmd/tab"

cat >"$scratch/tmux" <<'SCRIPT'
#!/bin/sh
set -eu
case ${1:-} in
	-V)
		printf '%s\n' "tmux 3.6"
		;;
	attach-session|switch-client)
		printf '%s\n' "$*" >>"$TAB_TMUX_LOG"
		;;
	*)
		printf 'unexpected tmux arguments: %s\n' "$*" >&2
		exit 1
		;;
esac
SCRIPT
chmod +x "$scratch/tmux"

: >"$scratch/tmux.log"
PATH="$scratch:$PATH" TAB_TMUX_LOG="$scratch/tmux.log" TMUX= \
	"$scratch/tab" alpha
grep -Fx 'attach-session -t alpha' "$scratch/tmux.log"

: >"$scratch/tmux.log"
PATH="$scratch:$PATH" TAB_TMUX_LOG="$scratch/tmux.log" TMUX=inside \
	"$scratch/tab" alpha
grep -Fx 'switch-client -t alpha' "$scratch/tmux.log"

printf '%s\n' "tab binary smoke test passed"
```

- [ ] **Step 2: Demonstrate and run the new coverage**

Run:

```sh
sh scripts/tab-binary-smoke.sh
```

Expected: PASS with both exact tmux invocations printed.

- [ ] **Step 3: Wire it into local and CI checks**

Add `sh scripts/tab-binary-smoke.sh` to `CONTRIBUTING.md` and to the Ubuntu
integration job after installer tests.

- [ ] **Step 4: Run all shell smoke tests**

Run:

```sh
sh scripts/test-install.sh
sh scripts/tmux-smoke.sh
sh scripts/tab-binary-smoke.sh
```

Expected: all three scripts print their success line.

- [ ] **Step 5: Commit**

```sh
git add scripts/tab-binary-smoke.sh CONTRIBUTING.md .github/workflows/ci.yml
git commit -m "test: exercise the compiled tab binary"
```

### Task 5: Add dedicated CI quality checks and upgrade Dependency Review

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/dependency-review.yml`
- Modify: `CONTRIBUTING.md`

- [ ] **Step 1: Add the Ubuntu quality job**

Add a `quality` job with these exact steps:

```yaml
  quality:
    name: Quality (Ubuntu)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
          cache-dependency-path: go.sum
      - name: Verify source file line limit
        run: sh scripts/check-source-lines.sh
      - name: Test source file line checker
        run: sh scripts/test-source-lines.sh
      - name: Check Go vulnerabilities
        run: go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
      - name: Check shell scripts
        run: shellcheck install.sh scripts/*.sh
      - name: Check GitHub Actions workflows
        run: go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
```

- [ ] **Step 2: Upgrade and pin Dependency Review**

Replace v4 with:

```yaml
- uses: actions/dependency-review-action@a1d282b36b6f3519aa1f3fc636f609c47dddb294 # v5.0.0
```

- [ ] **Step 3: Document the extra local checks**

Add these commands to the contributor checklist:

```sh
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
shellcheck install.sh scripts/*.sh
```

- [ ] **Step 4: Run the quality suite**

Run:

```sh
sh scripts/check-source-lines.sh
sh scripts/test-source-lines.sh
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
shellcheck install.sh scripts/*.sh
```

Expected: all commands return zero and `govulncheck` reports no reachable
vulnerabilities.

- [ ] **Step 5: Commit**

```sh
git add .github/workflows/ci.yml .github/workflows/dependency-review.yml CONTRIBUTING.md
git commit -m "ci: expand project quality checks"
```

### Task 6: Verify, review, merge, and protect `main`

**Files:**
- Verify all modified files.
- Update: `projects/tmux-attach-browser/PROJECT.md` in the parent workspace after
  repository work completes.

- [ ] **Step 1: Run the complete local verification suite**

Run:

```sh
test -z "$(gofmt -l .)"
git diff --check origin/main...HEAD
go mod verify
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...
go build ./...
sh scripts/check-source-lines.sh
sh scripts/test-source-lines.sh
sh scripts/test-install.sh
sh scripts/tmux-smoke.sh
sh scripts/tab-binary-smoke.sh
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
shellcheck install.sh scripts/*.sh
```

Expected: zero failures, formatting output empty, and no reachable
vulnerabilities.

- [ ] **Step 2: Push and open a pull request**

```sh
git push -u origin feat/project-health-improvements
gh pr create --base main --head feat/project-health-improvements \
  --title "Improve project health checks and runtime safety" \
  --body "$(printf '%s\n' \
    '## Summary' \
    '- prevent stale refresh results and cancel tmux work on UI exit' \
    '- enforce the 450-line source policy and exercise the compiled binary' \
    '- add vulnerability, workflow, and shell checks; upgrade Dependency Review v5' \
    '' \
    '## Verification' \
    '- complete local unit, race, build, smoke, and quality suites pass')"
```

The pull request body summarizes behavior, quality tooling, documentation, and
the planned post-merge ruleset.

- [ ] **Step 3: Verify every pull-request check**

Run:

```sh
pr_number=$(gh pr view feat/project-health-improvements --json number --jq .number)
gh pr checks "$pr_number" --watch
```

Expected: CI on Ubuntu, macOS, and Windows, integration, quality, Dependency
Review, and CodeQL all complete successfully.

- [ ] **Step 4: Merge the pull request**

Use squash merge and delete the remote feature branch:

```sh
pr_number=$(gh pr view feat/project-health-improvements --json number --jq .number)
gh pr merge "$pr_number" --squash --delete-branch
```

Then synchronize local `main` and rerun the core verification suite.

- [ ] **Step 5: Create the active default-branch ruleset**

Confirm the successful pull-request check names match the contexts below, then
create one active branch ruleset named `Protect main`:

```sh
ruleset_file=$(mktemp)
trap 'rm -f "$ruleset_file"' EXIT INT TERM
printf '%s\n' '{
  "name": "Protect main",
  "target": "branch",
  "enforcement": "active",
  "bypass_actors": [
    {
      "actor_id": 5,
      "actor_type": "RepositoryRole",
      "bypass_mode": "always"
    }
  ],
  "conditions": {
    "ref_name": {
      "include": ["~DEFAULT_BRANCH"],
      "exclude": []
    }
  },
  "rules": [
    {"type": "deletion"},
    {"type": "non_fast_forward"},
    {
      "type": "pull_request",
      "parameters": {
        "dismiss_stale_reviews_on_push": false,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_approving_review_count": 0,
        "required_review_thread_resolution": false
      }
    },
    {
      "type": "required_status_checks",
      "parameters": {
        "do_not_enforce_on_create": true,
        "required_status_checks": [
          {"context": "Test (ubuntu-latest)"},
          {"context": "Test (macos-latest)"},
          {"context": "Test (windows-latest)"},
          {"context": "Integration (Ubuntu)"},
          {"context": "Quality (Ubuntu)"},
          {"context": "Dependency Review"},
          {"context": "Analyze (go)"}
        ],
        "strict_required_status_checks_policy": true
      }
    }
  ]
}' >"$ruleset_file"

gh api --method POST repos/hmmhmmhm/tmux-attach-browser/rulesets \
  --input "$ruleset_file"
gh api repos/hmmhmmhm/tmux-attach-browser/rulesets \
  --jq '.[] | select(.name == "Protect main") |
    {name, target, enforcement, bypass_actors, conditions, rules}'
```

Expected: one active branch ruleset targeting `~DEFAULT_BRANCH`, with the
administrator bypass and all four rule types.

- [ ] **Step 6: Clean branches and update project memory**

Verify only `main` remains locally and remotely. Add a dated fact and action log
entry to `projects/tmux-attach-browser/PROJECT.md` describing the merged PR,
successful checks, and ruleset.
