# Project Health Improvements Design

## Goal

Close every gap found in the 2026-07-25 project health review without changing
the user-facing `tab` workflow.

The finished project will:

- apply only the newest tmux session refresh result;
- cancel in-flight tmux list and create commands when the UI exits;
- exercise the real `tab` binary in an end-to-end smoke test;
- document and enforce a 450-line limit for hand-written source files;
- run dedicated Go vulnerability, workflow, and shell checks in CI;
- use Dependency Review Action v5 without the Node.js 20 warning; and
- protect `main` with a repository ruleset after the changes are merged.

## Scope

### Included source files

The 450-line policy applies to tracked, hand-written files with these
extensions:

- `.go`
- `.sh`
- `.ps1`
- `.yml`
- `.yaml`

The limit includes tests, installers, issue forms, release configuration, and
workflows. Documentation, generated output, vendored code, and binary assets
are outside the rule.

### Compatibility

The implementation keeps the existing Go 1.25 minimum and does not change the
CLI syntax, key bindings, release matrix, or supported platforms.

## Design

### Refresh ordering

`Model` owns a monotonically increasing load generation. The initial request
uses generation 1. Each refresh increments it and includes the generation in
the asynchronous command result.

`Update` ignores a session result whose generation does not match the current
generation. An old success or failure therefore cannot replace a newer list,
clear its loading state, or surface a stale error.

This keeps refresh responsive instead of disabling the `r` key while a request
is running.

### Command cancellation

The model receives a context when it is constructed. `loadSessions` and
`createSession` pass that context to the tmux client instead of using
`context.Background()`.

The public `New` constructor retains its current signature and uses a
background context for direct model construction and tests. `Run` creates a
cancellable context, builds the production model with it, and cancels the
context when the Bubble Tea program returns.

This confines the context to the UI lifetime without changing the public API.

### Binary smoke test

A new POSIX shell smoke test builds `cmd/tab`, places a deterministic fake
`tmux` executable first on `PATH`, and runs the real binary twice:

1. outside tmux, where `tab alpha` must invoke `attach-session -t alpha`;
2. inside tmux, where it must invoke `switch-client -t alpha`.

The fake executable also answers `tmux -V`, so the path covers `main`,
`app.Run`, `ExecClient.Check`, and `ExecClient.Connect` without taking over the
developer's terminal.

The existing real-tmux smoke test remains in place because it checks session
creation and output parsing against the installed tmux version.

### Source line enforcement

`scripts/check-source-lines.sh` inspects tracked files in the policy scope and
fails with each violating path and line count. It accepts an optional repository
path so its own test can create a temporary fixture repository and verify both
the 450-line pass boundary and the 451-line failure boundary.

The script runs locally through the documented contributor checklist and in a
dedicated Ubuntu quality job.

### Dedicated quality checks

The new CI quality job runs:

- `scripts/check-source-lines.sh`;
- `govulncheck` from `golang.org/x/vuln` v1.6.0;
- `actionlint` v1.7.12; and
- the runner-provided `shellcheck` over installers and scripts.

Tool versions are explicit where Go installs are used. Dependabot continues to
track Go modules and GitHub Actions.

### Dependency Review upgrade

Dependency Review Action is updated from v4 to the immutable commit for
v5.0.0:

`a1d282b36b6f3519aa1f3fc636f609c47dddb294`

The comment records `v5.0.0`. This removes the Node.js 20 runtime warning and
preserves the repository's action-pinning posture.

### Main branch ruleset

After the pull request is merged, an active repository branch ruleset targets
the default branch and:

- blocks branch deletion and non-fast-forward updates;
- requires changes to arrive through a pull request;
- requires the stable CI and security checks added by this change; and
- leaves repository administrators an explicit emergency bypass.

The exact required check names are read from the successful pull-request run
before the ruleset is created. This avoids installing a rule that refers to a
nonexistent status context.

## Error handling

- Stale session results are silently discarded because they are expected
  asynchronous completions, not user-facing failures.
- A current refresh error keeps the existing list and remains visible.
- Context cancellation is reported by the existing error path only if the UI
  is still active; cancellation after exit has no visible effect.
- Quality scripts print the offending file or tool output and return nonzero.
- The ruleset is created only after all required status contexts have completed
  successfully.

## Testing

Behavior changes follow red-green-refactor:

1. add a failing test proving an older refresh result cannot overwrite a newer
   one;
2. add failing tests proving list and create commands receive and observe the
   model context;
3. add boundary tests for the 450-line checker;
4. add the real-binary smoke test and first demonstrate that the previous smoke
   suite did not cover it;
5. run formatting, module verification, vet, unit tests, race tests, builds,
   installer tests, both smoke tests, source-line checks, `govulncheck`,
   `actionlint`, and `shellcheck`;
6. verify the pull-request checks on Ubuntu, macOS, and Windows before merge.

## Non-goals

- No visual redesign or key-binding change.
- No new runtime dependency.
- No broad package refactor.
- No requirement for contributor approvals in a single-maintainer repository.
- No automatic release.
