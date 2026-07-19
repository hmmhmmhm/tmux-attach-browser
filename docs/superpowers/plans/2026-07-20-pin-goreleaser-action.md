# Pin GoReleaser Action Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close CodeQL alert 1 by replacing the mutable GoReleaser Action `v7` reference with the exact full commit SHA currently selected by that tag.

**Architecture:** Keep the release workflow behavior unchanged and modify only the GoReleaser Action reference. Preserve an inline `# v7` comment so the pinned version stays readable and Dependabot can maintain the version documentation.

**Tech Stack:** GitHub Actions YAML, CodeQL default setup, Dependabot, Ruby YAML parser, Go toolchain, GitHub CLI

---

### Task 1: Prove the Existing Reference Violates the Invariant

**Files:**
- Inspect: `.github/workflows/release.yml:22`
- Reference: `docs/superpowers/specs/2026-07-20-pin-goreleaser-action-design.md`

- [ ] **Step 1: Confirm the open alert and upstream tag target**

Run:

```bash
gh api repos/hmmhmmhm/tmux-attach-browser/code-scanning/alerts \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  -f state=open \
  --method GET \
  --jq '.[] | select(.number == 1) | [.rule.id, .rule.security_severity_level, .most_recent_instance.location.path, .most_recent_instance.location.start_line] | join("|")'

gh api repos/goreleaser/goreleaser-action/git/ref/tags/v7 \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --jq '[.object.type, .object.sha] | join("|")'
```

Expected:

```text
actions/unpinned-tag|medium|.github/workflows/release.yml|22
commit|f06c13b6b1a9625abc9e6e439d9c05a8f2190e94
```

- [ ] **Step 2: Run the full-SHA assertion and verify that it fails**

Run:

```bash
/usr/bin/ruby -e '
line = File.readlines(".github/workflows/release.yml").find { |candidate| candidate.include?("goreleaser/goreleaser-action@") }
abort "GoReleaser Action reference missing" unless line
ref = line[/goreleaser\/goreleaser-action@(\S+)/, 1]
abort "expected full SHA, got #{ref}" unless ref&.match?(/\A[0-9a-f]{40}\z/)
'
```

Expected: exit 1 with `expected full SHA, got v7`.

### Task 2: Pin the GoReleaser Action

**Files:**
- Modify: `.github/workflows/release.yml:22`

- [ ] **Step 1: Replace the mutable tag with the verified full SHA**

Apply this exact change:

```diff
-      - uses: goreleaser/goreleaser-action@v7
+      - uses: goreleaser/goreleaser-action@f06c13b6b1a9625abc9e6e439d9c05a8f2190e94 # v7
```

- [ ] **Step 2: Run the security invariant and exact-value assertions**

Run:

```bash
/usr/bin/ruby -e '
line = File.readlines(".github/workflows/release.yml").find { |candidate| candidate.include?("goreleaser/goreleaser-action@") }
abort "GoReleaser Action reference missing" unless line
ref = line[/goreleaser\/goreleaser-action@(\S+)/, 1]
abort "expected full SHA, got #{ref}" unless ref&.match?(/\A[0-9a-f]{40}\z/)
abort "unexpected SHA #{ref}" unless ref == "f06c13b6b1a9625abc9e6e439d9c05a8f2190e94"
abort "missing version comment" unless line.match?(/# v7\s*\z/)
puts "action_pin=ok"
'
```

Expected: `action_pin=ok`.

- [ ] **Step 3: Verify the upstream ref has not moved**

Run:

```bash
test "$(gh api repos/goreleaser/goreleaser-action/git/ref/tags/v7 \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  --jq '.object.type + " " + .object.sha')" \
  = 'commit f06c13b6b1a9625abc9e6e439d9c05a8f2190e94'
```

Expected: exit 0. If it fails, stop and resolve the current official `v7` ref before continuing.

- [ ] **Step 4: Verify workflow syntax and repository health**

Run:

```bash
/usr/bin/ruby -e 'require "yaml"; YAML.load_file(".github/workflows/release.yml"); puts "yaml=ok"'
git diff --check
go test ./...
go vet ./...
```

Expected: YAML prints `yaml=ok`; diff check, all Go tests, and vet exit 0.

- [ ] **Step 5: Verify the implementation diff is exactly one workflow line**

Run:

```bash
git diff -- .github/workflows/release.yml
git diff --name-only
```

Expected: the workflow diff contains one removed line and one added line; the only uncommitted file is `.github/workflows/release.yml`.

- [ ] **Step 6: Commit the implementation**

Run:

```bash
git add .github/workflows/release.yml
git commit -m "ci: pin GoReleaser action"
```

Expected: one commit containing only `.github/workflows/release.yml`.

### Task 3: Deliver and Review the Pull Request

**Files:**
- Include: `.github/workflows/release.yml`
- Include: `docs/superpowers/specs/2026-07-20-pin-goreleaser-action-design.md`
- Include: `docs/superpowers/plans/2026-07-20-pin-goreleaser-action.md`

- [ ] **Step 1: Verify the complete branch before pushing**

Run:

```bash
git diff --check origin/main...HEAD
go test ./...
go vet ./...
git status --short --branch
git diff --name-only origin/main...HEAD
```

Expected: all verification commands exit 0, the worktree is clean, and the branch contains exactly the workflow, spec, and plan files listed above.

- [ ] **Step 2: Push and create a draft pull request**

Run:

```bash
git push -u origin fix/pin-goreleaser-action

gh pr create \
  --repo hmmhmmhm/tmux-attach-browser \
  --base main \
  --head fix/pin-goreleaser-action \
  --draft \
  --title "ci: pin GoReleaser action" \
  --body "## Summary

- pin the third-party GoReleaser Action to the full SHA currently selected by v7
- retain the v7 comment for readability and Dependabot updates
- document and verify closure criteria for CodeQL alert 1

## Verification

- Ruby workflow YAML parse
- full-SHA and upstream-ref assertions
- git diff --check
- go test ./...
- go vet ./..."
```

Expected: a new draft pull request URL.

- [ ] **Step 3: Wait for all pull request checks**

Run:

```bash
gh pr checks --repo hmmhmmhm/tmux-attach-browser --watch --interval 10
```

Expected: CI, Dependency Review, and CodeQL checks finish successfully.

- [ ] **Step 4: Request independent review**

Use `superpowers:requesting-code-review` against the range `origin/main...HEAD`.

Expected: no Critical or Important findings remain. Resolve any such finding before continuing.

### Task 4: Merge, Verify Alert Closure, and Clean Up

**Files:**
- Update outside repository: `/Users/hm/Documents/personal-agent/projects/tmux-attach-browser/PROJECT.md`

- [ ] **Step 1: Mark the pull request ready and squash merge it**

Run:

```bash
pr_number=$(gh pr view \
  --repo hmmhmmhm/tmux-attach-browser \
  fix/pin-goreleaser-action \
  --json number \
  --jq '.number')

gh pr ready "$pr_number" --repo hmmhmmhm/tmux-attach-browser
gh pr merge "$pr_number" \
  --repo hmmhmmhm/tmux-attach-browser \
  --squash \
  --subject "ci: pin GoReleaser action" \
  --body ""
```

Expected: pull request state is `MERGED`.

- [ ] **Step 2: Fast-forward the primary worktree and rerun local verification**

Run from `/Users/hm/Documents/personal-agent/workspaces/tmux-attach-browser`:

```bash
git fetch origin main
git merge --ff-only origin/main
/usr/bin/ruby -e 'require "yaml"; YAML.load_file(".github/workflows/release.yml"); puts "yaml=ok"'
go test ./...
go vet ./...
```

Expected: `main` fast-forwards and every verification command exits 0.

- [ ] **Step 3: Wait for successful CodeQL analysis on `main`**

Run:

```bash
pr_number=$(gh pr view \
  --repo hmmhmmhm/tmux-attach-browser \
  fix/pin-goreleaser-action \
  --json number \
  --jq '.number')
merge_sha=$(gh pr view "$pr_number" \
  --repo hmmhmmhm/tmux-attach-browser \
  --json mergeCommit \
  --jq '.mergeCommit.oid')

codeql_run_id=''
for attempt in {1..24}; do
  codeql_run_id=$(gh run list \
    --repo hmmhmmhm/tmux-attach-browser \
    --workflow CodeQL \
    --branch main \
    --limit 10 \
    --json databaseId,headSha \
    --jq ".[] | select(.headSha == \"$merge_sha\") | .databaseId" | head -n 1)
  test -n "$codeql_run_id" && break
  sleep 5
done
test -n "$codeql_run_id"

gh run watch "$codeql_run_id" \
  --repo hmmhmmhm/tmux-attach-browser \
  --interval 10 \
  --exit-status
```

Expected: the latest post-merge CodeQL run completes successfully.

- [ ] **Step 4: Prove CodeQL alert 1 is no longer open**

Run:

```bash
test "$(gh api repos/hmmhmmhm/tmux-attach-browser/code-scanning/alerts \
  -H 'X-GitHub-Api-Version: 2026-03-10' \
  -f state=open \
  --method GET \
  --jq 'map(select(.number == 1)) | length')" -eq 0
```

Expected: exit 0.

- [ ] **Step 5: Update the project action log and security fact**

Use `apply_patch` to add the merge commit, alert closure, and verification result to the top of the Facts and recent action log in `/Users/hm/Documents/personal-agent/projects/tmux-attach-browser/PROJECT.md`.

Expected: the new entry follows the existing `YYYY-MM-DD HH:MM KST · summary · link` format and contains no U+2014 character.

- [ ] **Step 6: Remove the worktree and feature branches**

Run from `/Users/hm/Documents/personal-agent/workspaces/tmux-attach-browser` after confirming both worktrees are clean:

```bash
git worktree remove /Users/hm/Documents/personal-agent/workspaces/tmux-attach-browser/.worktrees/fix/pin-goreleaser-action
git push origin --delete fix/pin-goreleaser-action
git branch -D fix/pin-goreleaser-action
```

Expected: the implementation worktree and both feature branch refs are removed.

- [ ] **Step 7: Perform final completion audit**

Run:

```bash
test "$(git for-each-ref --format='%(refname:short)' refs/heads | sort)" = 'main'
test "$(git ls-remote --heads origin | awk '{sub("refs/heads/", "", $2); print $2}' | sort)" = 'main'
test "$(git worktree list --porcelain | awk '/^worktree /{count++} END{print count+0}')" -eq 1
test "$(git tag --list | sort)" = 'v0.1.0'
test -z "$(git status --porcelain)"
```

Expected: only `main` remains locally and remotely, one worktree remains, `v0.1.0` is preserved, and the primary worktree is clean.
