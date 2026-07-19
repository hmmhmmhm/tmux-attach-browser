# Repository Security Enablement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable GitHub-native dependency and security checks for `hmmhmmhm/tmux-attach-browser`, merge the configuration into `main`, and leave only the `main` branch locally and remotely.

**Architecture:** Repository-owned YAML enables recurring Dependabot updates and pull-request dependency review, while GitHub repository APIs enable vulnerability alerts, security updates, CodeQL default setup, secret protection, and private vulnerability reporting. All settings are read back after mutation, the approved pull request is merged only after checks pass, and branch deletion happens last with explicit branch names.

**Tech Stack:** GitHub Actions, Dependabot, CodeQL default setup, GitHub REST API, `gh`, Go, YAML

---

### Task 1: Add repository-owned dependency checks

**Files:**
- Create: `.github/dependabot.yml`
- Create: `.github/workflows/dependency-review.yml`
- Modify: `.github/workflows/ci.yml:4-7`

- [ ] **Step 1: Verify the expected configuration is absent and the stale branch trigger is present**

Run:

```sh
set +e
test -f .github/dependabot.yml
dependabot_status=$?
test -f .github/workflows/dependency-review.yml
dependency_review_status=$?
set -e
printf 'dependabot_status=%s\ndependency_review_status=%s\n' "$dependabot_status" "$dependency_review_status"
test "$dependabot_status" -ne 0
test "$dependency_review_status" -ne 0
rg -F 'feature/initial-release' .github/workflows/ci.yml
```

Expected: both file checks report status `1`, and `feature/initial-release` is found in `ci.yml`.

- [ ] **Step 2: Create `.github/dependabot.yml`**

Add exactly:

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
      timezone: "Asia/Seoul"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
      timezone: "Asia/Seoul"
```

- [ ] **Step 3: Create `.github/workflows/dependency-review.yml`**

Add exactly:

```yaml
name: Dependency Review

on:
  pull_request:

permissions:
  contents: read

jobs:
  dependency-review:
    name: Dependency Review
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/dependency-review-action@v4
```

- [ ] **Step 4: Remove the deleted release branch from the CI push trigger**

Change the beginning of `.github/workflows/ci.yml` to:

```yaml
name: CI

on:
  push:
    branches:
      - main
  pull_request:
```

- [ ] **Step 5: Verify the configuration is syntactically valid and GREEN**

Run:

```sh
set -e
ruby -e "require 'yaml'; ARGV.each { |file| YAML.parse_file(file) }" \
  .github/dependabot.yml \
  .github/workflows/dependency-review.yml \
  .github/workflows/ci.yml
test -z "$(rg -F 'feature/initial-release' .github/workflows/ci.yml || true)"
test "$(rg -c 'package-ecosystem:' .github/dependabot.yml)" = "2"
rg -F 'package-ecosystem: "gomod"' .github/dependabot.yml
rg -F 'package-ecosystem: "github-actions"' .github/dependabot.yml
rg -F 'actions/dependency-review-action@v4' .github/workflows/dependency-review.yml
git diff --check
go test ./...
go vet ./...
```

Expected: YAML parsing succeeds, both ecosystems and the dependency-review action are found, no stale branch reference remains, and all Go checks pass.

- [ ] **Step 6: Commit the repository configuration**

Run:

```sh
git add .github/dependabot.yml .github/workflows/dependency-review.yml .github/workflows/ci.yml
git diff --cached --check
git commit -m "ci: enable dependency security checks"
```

Expected: one commit contains only the three repository configuration files.

### Task 2: Publish the approved draft pull request

**Files:**
- Verify: `docs/superpowers/specs/2026-07-19-repository-security-design.md`
- Verify: `docs/superpowers/plans/2026-07-20-repository-security.md`
- Verify: `.github/dependabot.yml`
- Verify: `.github/workflows/dependency-review.yml`
- Verify: `.github/workflows/ci.yml`

- [ ] **Step 1: Verify the branch scope before publishing**

Run:

```sh
set -e
test "$(git branch --show-current)" = "chore/enable-repository-security"
git diff --check main...HEAD
git diff --name-only main...HEAD
go test ./...
go vet ./...
```

Expected changed paths:

```text
.github/dependabot.yml
.github/workflows/ci.yml
.github/workflows/dependency-review.yml
docs/superpowers/plans/2026-07-20-repository-security.md
docs/superpowers/specs/2026-07-19-repository-security-design.md
```

- [ ] **Step 2: Push the branch and open a draft pull request**

Run:

```sh
git push -u origin chore/enable-repository-security
gh pr create --draft \
  --base main \
  --head chore/enable-repository-security \
  --title "chore: enable repository security" \
  --body $'## Summary\n\n- add weekly Dependabot updates for Go modules and GitHub Actions\n- add pull-request dependency review and remove a stale CI branch trigger\n- enable and verify GitHub-native repository security settings\n\n## Verification\n\n- parsed all changed YAML files\n- ran `go test ./...`\n- ran `go vet ./...`'
```

Expected: GitHub creates a draft pull request targeting `main`.

- [ ] **Step 3: Record and verify the pull request number**

Run:

```sh
pr_number=$(gh pr view chore/enable-repository-security --json number --jq .number)
test -n "$pr_number"
gh pr view "$pr_number" --json number,url,isDraft,state,baseRefName,headRefName
```

Expected: the pull request is open, draft, based on `main`, and headed by `chore/enable-repository-security`.

### Task 3: Enable GitHub repository security settings

**Files:**
- Modify: none

- [ ] **Step 1: Capture the pre-change settings**

Run:

```sh
gh api repos/hmmhmmhm/tmux-attach-browser \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --jq '{visibility, default_branch, security_and_analysis}'
gh api repos/hmmhmmhm/tmux-attach-browser/code-scanning/default-setup \
  -H 'X-GitHub-Api-Version: 2026-03-10'
gh api repos/hmmhmmhm/tmux-attach-browser/private-vulnerability-reporting \
  -H 'X-GitHub-Api-Version: 2026-03-10'
```

Expected: the repository is public, the default branch is `main`, CodeQL is not configured, secret scanning and push protection are enabled, and other disabled or unavailable features are visible.

- [ ] **Step 2: Enable dependency alerts and automatic security fixes**

Run:

```sh
gh api --method PUT repos/hmmhmmhm/tmux-attach-browser/vulnerability-alerts \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --silent
gh api --method PUT repos/hmmhmmhm/tmux-attach-browser/automated-security-fixes \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --silent
```

Expected: both commands exit `0`.

- [ ] **Step 3: Preserve core secret protection and attempt every applicable extended check**

Run core settings first:

```sh
gh api --method PATCH repos/hmmhmmhm/tmux-attach-browser \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  -F 'security_and_analysis[secret_scanning][status]=enabled' \
  -F 'security_and_analysis[secret_scanning_push_protection][status]=enabled' \
  --silent
```

Then attempt each plan-dependent extension independently:

```sh
set +e
gh api --method PATCH repos/hmmhmmhm/tmux-attach-browser \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  -F 'security_and_analysis[secret_scanning_validity_checks][status]=enabled' \
  --silent
validity_status=$?
gh api --method PATCH repos/hmmhmmhm/tmux-attach-browser \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  -F 'security_and_analysis[secret_scanning_non_provider_patterns][status]=enabled' \
  --silent
non_provider_status=$?
set -e
printf 'validity_status=%s\nnon_provider_status=%s\n' "$validity_status" "$non_provider_status"
```

Expected: supported settings exit `0`. A nonzero result is acceptable only when GitHub reports that the feature is unavailable to this user-owned public repository; retain secret scanning and push protection and record the limitation in the final report.

- [ ] **Step 4: Enable private vulnerability reporting**

Run:

```sh
gh api --method PUT repos/hmmhmmhm/tmux-attach-browser/private-vulnerability-reporting \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --silent
```

Expected: the command exits `0`.

- [ ] **Step 5: Enable CodeQL default setup with the approved maximum query scope**

Run:

```sh
gh api --method PATCH repos/hmmhmmhm/tmux-attach-browser/code-scanning/default-setup \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  -F state=configured \
  -F 'languages[]=go' \
  -F 'languages[]=actions' \
  -F query_suite=extended \
  -F threat_model=remote_and_local \
  -F runner_type=standard
```

Expected: GitHub returns `200` with the configured settings or `202` with a validation run URL.

- [ ] **Step 6: Read back and verify every setting**

Run:

```sh
set -e
gh api repos/hmmhmmhm/tmux-attach-browser/vulnerability-alerts \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --silent
gh api repos/hmmhmmhm/tmux-attach-browser/automated-security-fixes \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --silent
gh api repos/hmmhmmhm/tmux-attach-browser/private-vulnerability-reporting \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --jq '.enabled'
gh api repos/hmmhmmhm/tmux-attach-browser \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --jq '.security_and_analysis'
gh api repos/hmmhmmhm/tmux-attach-browser/code-scanning/default-setup \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --jq '{state, languages, query_suite, threat_model, runner_type}'
```

Expected: alerts and automatic fixes respond successfully, private vulnerability reporting is `true`, Dependabot security updates report enabled, core secret protection remains enabled, and CodeQL reports `configured`, `go` plus `actions`, `extended`, `remote_and_local`, and `standard`.

### Task 4: Verify and merge the approved security pull request

**Files:**
- Verify: `.github/dependabot.yml`
- Verify: `.github/workflows/dependency-review.yml`
- Verify: `.github/workflows/ci.yml`

- [ ] **Step 1: Wait for pull-request checks**

Run:

```sh
pr_number=$(gh pr view chore/enable-repository-security --json number --jq .number)
gh pr checks "$pr_number" --watch --interval 10
```

Expected: Ubuntu, macOS, Windows, tmux integration, Dependency Review, and CodeQL checks all pass. If GitHub schedules the initial CodeQL validation outside the pull request, verify that run separately before continuing.

- [ ] **Step 2: Verify the exact files GitHub received**

Run:

```sh
gh api "repos/hmmhmmhm/tmux-attach-browser/contents/.github/dependabot.yml?ref=chore/enable-repository-security" --jq .path
gh api "repos/hmmhmmhm/tmux-attach-browser/contents/.github/workflows/dependency-review.yml?ref=chore/enable-repository-security" --jq .path
gh pr diff "$pr_number" --name-only
```

Expected: both new configuration files exist on the branch and the pull request contains only the five approved paths.

- [ ] **Step 3: Mark the approved pull request ready and squash merge it**

Run:

```sh
gh pr ready "$pr_number"
gh pr merge "$pr_number" --squash --subject "chore: enable repository security" --body ""
gh pr view "$pr_number" --json state,mergedAt,mergeCommit,url
```

Expected: the pull request state is `MERGED` with a nonempty merge commit.

- [ ] **Step 4: Synchronize local `main` and verify the merged repository**

Run:

```sh
git switch main
git fetch origin main
git merge --ff-only origin/main
test "$(git rev-parse HEAD)" = "$(git rev-parse origin/main)"
ruby -e "require 'yaml'; ARGV.each { |file| YAML.parse_file(file) }" \
  .github/dependabot.yml \
  .github/workflows/dependency-review.yml \
  .github/workflows/ci.yml
go test ./...
go vet ./...
```

Expected: local `main` matches `origin/main`, YAML parsing succeeds, and all Go checks pass.

- [ ] **Step 5: Verify CodeQL's initial analysis and the final setting state**

Run:

```sh
gh run list --limit 20 --json databaseId,name,workflowName,status,conclusion,url,headBranch,event
gh api repos/hmmhmmhm/tmux-attach-browser/code-scanning/default-setup \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --jq '{state, languages, query_suite, threat_model, runner_type, updated_at}'
gh api repos/hmmhmmhm/tmux-attach-browser/code-scanning/alerts \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --jq 'length'
```

Expected: the CodeQL run completes successfully, default setup remains configured, and the alerts endpoint is readable even when it returns `0` alerts.

### Task 5: Remove every non-main branch

**Files:**
- Remove worktree: `.worktrees/docs/add-terminal-demo`
- Delete local branches: `docs/add-terminal-demo`, `docs/modernize-readme`, `feature/initial-release`, `chore/enable-repository-security`
- Delete remote branches: `docs/add-terminal-demo`, `docs/modernize-readme`, `feature/initial-release`, `chore/enable-repository-security`

- [ ] **Step 1: Resolve and verify every deletion target**

Run:

```sh
set -e
test "$(git branch --show-current)" = "main"
git status --short --branch
test -z "$(git status --porcelain)"
git worktree list --porcelain
git branch --list docs/add-terminal-demo docs/modernize-readme feature/initial-release chore/enable-repository-security
git ls-remote --heads origin docs/add-terminal-demo docs/modernize-readme feature/initial-release chore/enable-repository-security
git tag --list
```

Expected: `main` is clean, the four explicit branches are the only deletion targets, the demo worktree is visible, and release tags are listed separately and remain out of scope.

- [ ] **Step 2: Verify and remove the old demo worktree**

Run:

```sh
set -e
test -d /Users/hm/Documents/personal-agent/workspaces/tmux-attach-browser/.worktrees/docs/add-terminal-demo
test -z "$(git -C /Users/hm/Documents/personal-agent/workspaces/tmux-attach-browser/.worktrees/docs/add-terminal-demo status --porcelain)"
git worktree remove /Users/hm/Documents/personal-agent/workspaces/tmux-attach-browser/.worktrees/docs/add-terminal-demo
test ! -d /Users/hm/Documents/personal-agent/workspaces/tmux-attach-browser/.worktrees/docs/add-terminal-demo
```

Expected: only the clean, explicitly named worktree is removed.

- [ ] **Step 3: Delete the four remote branches explicitly**

Run:

```sh
git push origin --delete \
  docs/add-terminal-demo \
  docs/modernize-readme \
  feature/initial-release \
  chore/enable-repository-security
```

Expected: GitHub confirms deletion of all four named remote branches. No wildcard or force push is used.

- [ ] **Step 4: Delete the four local branches explicitly**

Run:

```sh
git branch -D \
  docs/add-terminal-demo \
  docs/modernize-readme \
  feature/initial-release \
  chore/enable-repository-security
```

Expected: Git deletes the four named local branches. Forced local deletion is necessary because squash-merged branch tips are not ancestors of `main`; the user explicitly approved this cleanup.

- [ ] **Step 5: Verify only `main` remains and tags were preserved**

Run:

```sh
set -e
git fetch --prune origin
test "$(git for-each-ref --format='%(refname:short)' refs/heads)" = "main"
remote_branches=$(git ls-remote --heads origin | awk '{sub("refs/heads/", "", $2); print $2}')
test "$remote_branches" = "main"
test "$(git worktree list --porcelain | rg '^worktree ' | wc -l | tr -d ' ')" = "1"
test -n "$(git tag --list v0.1.0)"
git status --short --branch
git branch -a
git tag --list
```

Expected: local and remote branch listings contain only `main`, exactly one worktree remains, `v0.1.0` still exists, and the repository is clean.

### Task 6: Final verification

**Files:**
- Verify: `.github/dependabot.yml`
- Verify: `.github/workflows/dependency-review.yml`
- Verify: `.github/workflows/ci.yml`

- [ ] **Step 1: Run the complete local verification suite on `main`**

Run:

```sh
set -e
test "$(git branch --show-current)" = "main"
test "$(git rev-parse HEAD)" = "$(git rev-parse origin/main)"
git diff --check
ruby -e "require 'yaml'; ARGV.each { |file| YAML.parse_file(file) }" \
  .github/dependabot.yml \
  .github/workflows/dependency-review.yml \
  .github/workflows/ci.yml
go test ./...
go vet ./...
git status --short --branch
```

Expected: all commands pass and the worktree is clean and synchronized.

- [ ] **Step 2: Run the complete remote verification suite**

Run:

```sh
set -e
gh api repos/hmmhmmhm/tmux-attach-browser/vulnerability-alerts \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --silent
gh api repos/hmmhmmhm/tmux-attach-browser/automated-security-fixes \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --silent
test "$(gh api repos/hmmhmmhm/tmux-attach-browser/private-vulnerability-reporting -H 'X-GitHub-Api-Version: 2026-03-10' --jq .enabled)" = "true"
test "$(gh api repos/hmmhmmhm/tmux-attach-browser/code-scanning/default-setup -H 'X-GitHub-Api-Version: 2026-03-10' --jq .state)" = "configured"
gh api repos/hmmhmmhm/tmux-attach-browser --jq '.security_and_analysis'
gh api repos/hmmhmmhm/tmux-attach-browser/contents/.github/dependabot.yml --jq .path
gh api repos/hmmhmmhm/tmux-attach-browser/contents/.github/workflows/dependency-review.yml --jq .path
gh run list --limit 10 --json workflowName,status,conclusion,url,headBranch,event
```

Expected: all required security endpoints are enabled and readable, both configuration files exist on `main`, and the latest security and CI runs succeeded.
