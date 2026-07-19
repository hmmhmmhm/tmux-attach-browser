# Repository Security Enablement Design

**Date:** 2026-07-19
**Status:** Approved for implementation

## Goal

Enable the full set of practical GitHub-native dependency and security checks available to the public `hmmhmmhm/tmux-attach-browser` repository, then remove every local and remote branch except `main`.

## Scope

The change covers four areas:

1. GitHub repository security settings
2. Dependabot version-update configuration
3. Pull-request dependency review
4. Local and remote branch cleanup

External scanners such as OpenSSF Scorecard and custom `govulncheck` workflows are out of scope. GitHub CodeQL, Dependabot, dependency review, and secret scanning provide the desired coverage without adding overlapping third-party maintenance. New branch-protection or ruleset requirements are also out of scope because they would change merge permissions rather than enable analysis.

## Repository Security Settings

Use GitHub's repository APIs to enable every applicable native feature:

- Dependency graph and Dependabot vulnerability alerts
- Dependabot security updates
- CodeQL default setup for `go` and `actions`
  - Query suite: `extended`
  - Threat model: `remote_and_local`
  - Runner: GitHub-hosted standard runner
- Secret scanning and push protection, preserving their existing enabled state
- Secret scanning validity checks and non-provider patterns when GitHub permits them for this user-owned public repository
- Private vulnerability reporting

Some advanced secret-scanning options may be restricted to organization-owned repositories with GitHub Secret Protection. An unsupported response is recorded as unavailable, not treated as permission to disable or replace the existing secret-scanning protections.

## Repository Files

### `.github/dependabot.yml`

Add Dependabot version updates for both ecosystems used by the repository:

- `gomod` in `/`
- `github-actions` in `/`

Both checks run weekly on Monday at 09:00 in `Asia/Seoul`. The configuration stays minimal so Dependabot can apply its normal security-update behavior and create separate, reviewable pull requests.

### `.github/workflows/dependency-review.yml`

Add a pull-request workflow using `actions/dependency-review-action@v4`. It receives only `contents: read` permission and uses the action's default policy, which fails when a pull request introduces a dependency with a known vulnerability.

### `.github/workflows/ci.yml`

Remove the stale `feature/initial-release` push trigger. After branch cleanup, CI should run on pushes to `main` and on pull requests.

## Delivery Flow

1. Commit the approved design on `chore/enable-repository-security`.
2. Add and locally validate the Dependabot and workflow configuration.
3. Push the branch and open a draft pull request.
4. Enable the repository security settings through GitHub APIs.
5. Verify the returned settings and wait for CI, dependency review, and CodeQL checks.
6. Mark the approved pull request ready and squash merge it into `main`.
7. Fast-forward the local `main` worktree and verify the merged configuration.
8. Remove the clean `docs/add-terminal-demo` and implementation worktrees.
9. Delete every non-`main` local and remote branch, including the temporary security branch.
10. Confirm that both local and remote branch listings contain only `main`.

The user's approval of this design includes the pull-request merge and deletion of all non-`main` branches. Tags and releases are untouched.

## Safety and Failure Handling

- Read each setting before and after mutation, and never turn off an already enabled protection.
- Treat CodeQL setup as asynchronous and wait for the initial analysis result before cleanup.
- Do not merge if repository checks fail.
- Do not delete branches until the security configuration is confirmed on `main`.
- Verify every worktree is clean before removal.
- Resolve branch names explicitly. Do not use wildcard deletion or force push.
- Report any GitHub plan limitation instead of attempting a paid feature or changing account settings.

## Verification

Completion requires all of the following evidence:

- `go test ./...` and `go vet ./...` pass.
- The pull request CI and dependency review checks pass.
- CodeQL default setup reports `configured`, with `go`, `actions`, `extended`, and `remote_and_local`.
- Dependabot vulnerability alerts and security updates report enabled.
- Secret scanning and push protection remain enabled; supported additional checks report enabled.
- Private vulnerability reporting reports enabled.
- `.github/dependabot.yml` is present on `main` with both ecosystems.
- Local and remote branch listings contain only `main`.
