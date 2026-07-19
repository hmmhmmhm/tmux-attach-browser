# tmux-attach-browser

[![CI](https://github.com/hmmhmmhm/tmux-attach-browser/actions/workflows/ci.yml/badge.svg)](https://github.com/hmmhmmhm/tmux-attach-browser/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/hmmhmmhm/tmux-attach-browser)](https://github.com/hmmhmmhm/tmux-attach-browser/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Browse, create, and enter tmux sessions with one command: `tab`.

```text
tmux sessions

> work
  3 windows | 1 attached | active 2026-07-19 14:03:00

  api
  2 windows | detached | active 2026-07-19 11:42:00

  ↑/k up  ↓/j down  / filter  n new session  r refresh  ? more
```

`tab` is a focused terminal UI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and the Bubbles [list](https://github.com/charmbracelet/bubbles/tree/main/list) and [textinput](https://github.com/charmbracelet/bubbles/tree/main/textinput) components.

## Features

- Fuzzy-search current tmux sessions.
- See window count, attached clients, and recent activity.
- Attach with Enter.
- Create a session in the current directory with `n`.
- Refresh without leaving the browser.
- Switch clients when already inside tmux, avoiding nested sessions.
- Install verified release binaries on macOS, Linux, and Windows.

## Requirements

`tab` controls an existing tmux installation. Install tmux first:

```sh
# macOS
brew install tmux

# Ubuntu or Debian
sudo apt update && sudo apt install tmux

# Fedora
sudo dnf install tmux

# Arch Linux
sudo pacman -S tmux
```

Windows 10 and 11 are supported through WSL2. Run `wsl --install`, open the installed Linux distribution, install tmux there, and then use the Linux installation command below.

The native Windows binary is provided for advanced MSYS2 or Cygwin environments where `tmux.exe` is already on PATH. WSL2 is the tested and recommended Windows setup.

## Install

### macOS, Linux, and Windows WSL2

```sh
curl -fsSL https://raw.githubusercontent.com/hmmhmmhm/tmux-attach-browser/main/install.sh | sh
```

The installer downloads the latest release, verifies its SHA-256 checksum, and places `tab` in `$HOME/.local/bin`. Choose another directory with `TAB_INSTALL_DIR`:

```sh
curl -fsSL https://raw.githubusercontent.com/hmmhmmhm/tmux-attach-browser/main/install.sh | TAB_INSTALL_DIR=/usr/local/bin sh
```

The installer never runs `sudo` automatically.

### Native Windows PowerShell

Use this only when tmux is already available in the same environment:

```powershell
curl.exe -fsSL https://raw.githubusercontent.com/hmmhmmhm/tmux-attach-browser/main/install.ps1 | powershell -NoProfile -ExecutionPolicy Bypass -Command -
```

The default destination is `$HOME\.local\bin\tab.exe`. Set `TAB_INSTALL_DIR` before running the command to change it.

### Go

With Go 1.25 or newer:

```sh
go install github.com/hmmhmmhm/tmux-attach-browser/cmd/tab@latest
```

## Usage

```sh
# Open the session browser
tab

# Attach directly without opening the browser
tab work

# Print version or help
tab --version
tab --help
```

### Keys

| Key | Action |
| --- | --- |
| Up, Down, `j`, `k` | Move through sessions |
| Enter | Attach to the selected session |
| `n` | Create a new session in the current directory |
| `/` | Filter sessions by name |
| `r` | Refresh the session list |
| `?` | Expand or collapse help |
| Esc | Cancel filtering or session creation |
| `q`, Ctrl+C | Quit |

Session names cannot be blank or contain `:` or `.`. tmux rewrites or interprets these characters in target names, so `tab` rejects them before creating a session.

## How it works

`tab` asks tmux for a tab-separated session list, then renders it through the Bubbles list component. Selecting an existing session exits Bubble Tea before starting the interactive tmux client, so the terminal is restored correctly.

Outside tmux, selection runs:

```sh
tmux attach-session -t SESSION
```

Inside tmux, it runs:

```sh
tmux switch-client -t SESSION
```

New sessions are created detached with `tmux new-session -d -s SESSION -c CURRENT_DIRECTORY` and attached after the UI exits.

## Build and test

```sh
git clone https://github.com/hmmhmmhm/tmux-attach-browser.git
cd tmux-attach-browser
go test ./...
go test -race ./...
go vet ./...
go build -o tab ./cmd/tab
```

Additional checks:

```sh
sh scripts/test-install.sh
sh scripts/tmux-smoke.sh
goreleaser check
goreleaser release --snapshot --clean
```

## Troubleshooting

### `tmux is not installed or is not on PATH`

Install tmux using the platform instructions above, then confirm `tmux -V` works in the same terminal where you run `tab`.

### `tab: connect to session: exit status 1`

The session may have ended after the list was loaded. Run `tab` again or press `r` before selecting.

### `$HOME/.local/bin` is not on PATH

Add this line to your shell profile:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Bug reports and focused feature proposals are welcome.

## License

[MIT](LICENSE)
