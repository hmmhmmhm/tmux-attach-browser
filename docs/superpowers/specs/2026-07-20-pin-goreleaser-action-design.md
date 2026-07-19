# Pin GoReleaser Action Design

## Context

CodeQL default setup reports one open medium-severity alert on
`.github/workflows/release.yml`. The release job references
`goreleaser/goreleaser-action@v7`, and the `v7` tag is mutable.

The alert is:

- Rule: `actions/unpinned-tag`
- Alert: <https://github.com/hmmhmmhm/tmux-attach-browser/security/code-scanning/1>
- Location: `.github/workflows/release.yml:22`

The official `goreleaser/goreleaser-action` repository currently resolves the
`v7` tag to the verified commit
`f06c13b6b1a9625abc9e6e439d9c05a8f2190e94`.

## Decision

Replace the moving `v7` reference with its full 40-character commit SHA and
retain `# v7` on the same line:

```yaml
- uses: goreleaser/goreleaser-action@f06c13b6b1a9625abc9e6e439d9c05a8f2190e94 # v7
```

The full SHA makes the executed third-party code immutable. The inline version
comment keeps the workflow readable and lets Dependabot update the documented
version when it updates the pinned SHA.

## Scope

The implementation changes only the GoReleaser `uses` line in
`.github/workflows/release.yml`.

It does not change:

- `actions/checkout` or `actions/setup-go`
- workflow triggers or permissions
- GoReleaser distribution, version constraint, or arguments
- release artifacts, archive names, or checksums
- Go source code, installers, documentation, tags, or existing releases

## Verification

Verification will prove both the security property and the absence of unrelated
behavior changes:

1. A pre-change assertion must fail because the GoReleaser reference is not a
   full SHA.
2. A post-change assertion must require the exact 40-character SHA and the
   `# v7` comment.
3. The current upstream `v7` ref must still resolve to the pinned SHA before the
   branch is pushed.
4. The workflow YAML must parse successfully.
5. `git diff --check`, `go test ./...`, and `go vet ./...` must pass.
6. Pull request CI, Dependency Review, and CodeQL checks must pass.
7. After squash merge, CodeQL must analyze `main` successfully and alert 1 must
   no longer be open.

## Delivery and Cleanup

Work will be delivered through a draft pull request. After independent review
and all checks pass, the pull request will be marked ready and squash merged.
The local `main` branch will then be fast-forwarded, the implementation
worktree will be removed, and the local and remote feature branches will be
deleted. The `v0.1.0` tag and existing release remain untouched.

## Failure and Rollback

If upstream moves `v7` before the branch is pushed, the implementation will
stop and re-resolve the official ref instead of pinning stale evidence. If CI or
CodeQL fails, the pull request will not be merged until the root cause is
understood. The functional change is one line and can be rolled back by
restoring `goreleaser/goreleaser-action@v7`, although that rollback would reopen
the security alert.
