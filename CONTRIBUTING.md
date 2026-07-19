# Contributing

Thank you for improving tmux-attach-browser.

## Before opening an issue

- Confirm tmux itself works with `tmux -V` and `tmux list-sessions`.
- Search existing issues for the same behavior.
- Include the operating system, terminal, tmux version, and `tab --version` output in bug reports.

## Development

The project requires Go 1.25 or newer and tmux for the real smoke test.

```sh
git clone https://github.com/hmmhmmhm/tmux-attach-browser.git
cd tmux-attach-browser
go mod download
go test ./...
```

Write a failing test before changing behavior. Keep tmux command execution in `internal/tmux`, Bubble Tea state in `internal/ui`, and CLI orchestration in `internal/app`.

Run all checks before opening a pull request:

```sh
test -z "$(gofmt -l .)"
go vet ./...
go test -race ./...
go build ./cmd/tab
sh scripts/test-install.sh
sh scripts/tmux-smoke.sh
```

Pull requests should be small, explain user-visible behavior, and include tests. Do not include generated `dist` artifacts.

## Commit messages

Use a short imperative subject, for example:

```text
feat: add session sorting option
fix: retain sessions after refresh failure
docs: clarify WSL installation
```

## Reporting security issues

Do not open public issues for vulnerabilities. Follow [SECURITY.md](SECURITY.md).
