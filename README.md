<div align="center">

# tmux-attach-browser

**Choose or create a tmux session with one command.**

[![CI](https://github.com/hmmhmmhm/tmux-attach-browser/actions/workflows/ci.yml/badge.svg)](https://github.com/hmmhmmhm/tmux-attach-browser/actions/workflows/ci.yml)
[![Release](https://img.shields.io/badge/release-v0.1.0-blue)](https://github.com/hmmhmmhm/tmux-attach-browser/releases/latest)
[![Go](https://img.shields.io/badge/go-1.25-blue)](go.mod)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

</div>

## Install

Already have [tmux](https://github.com/tmux/tmux) on your `PATH`? Install the latest release on macOS, Linux, or Windows WSL2:

```sh
curl -fsSL https://raw.githubusercontent.com/hmmhmmhm/tmux-attach-browser/main/install.sh | sh
```

The installer verifies the release checksum and places `tab` in `~/.local/bin`. It never runs `sudo`.

## Run `tab`

![Demo showing tab browsing an existing tmux session and creating a new one](docs/assets/tab-demo.gif)

That is the whole workflow:

1. Run `tab` to see your tmux sessions.
2. Select a session and press Enter to enter it.
3. Press `n` to create a session in the current directory.

To skip the browser and open a named session directly:

```sh
tab work
```

When `tab` runs inside tmux, it switches the current client to your chosen session instead of nesting another tmux client.

## Keys

| Key | Action |
| --- | --- |
| Up, Down, `j`, `k` | Select a session |
| Enter | Enter the selected session |
| `n` | Create a session in the current directory |
| `/` | Filter sessions by name |
| `r` | Refresh the session list |
| `?` | Show or hide expanded help |
| Esc | Cancel filtering or session creation; quit from the session list |
| `q`, Ctrl+C | Quit |

<details>
<summary><strong>Install tmux</strong></summary>

```sh
# macOS
brew install tmux

# Ubuntu or Debian, including WSL2
sudo apt update && sudo apt install tmux

# Fedora
sudo dnf install tmux

# Arch Linux
sudo pacman -S tmux
```

</details>

<details>
<summary><strong>Other installation options</strong></summary>

Choose another writable installation directory:

```sh
curl -fsSL https://raw.githubusercontent.com/hmmhmmhm/tmux-attach-browser/main/install.sh | TAB_INSTALL_DIR="$HOME/bin" sh
```

Add the directory to your `PATH` if your shell does not already include it.

Install with Go 1.25 or newer:

```sh
go install github.com/hmmhmmhm/tmux-attach-browser/cmd/tab@latest
```

Native Windows PowerShell is intended for MSYS2 or Cygwin environments where `tmux.exe` is already available:

```powershell
curl.exe -fsSL https://raw.githubusercontent.com/hmmhmmhm/tmux-attach-browser/main/install.ps1 | powershell -NoProfile -ExecutionPolicy Bypass -Command -
```

WSL2 remains the recommended Windows setup.

</details>

<details>
<summary><strong>Troubleshooting</strong></summary>

### `tab: command not found`

The default installer destination is `~/.local/bin`. Add it to your shell profile, then open a new terminal:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

### `tmux is not installed or is not on PATH`

Install tmux using the commands above, then confirm that `tmux -V` works in the same terminal.

### `tab: connect to session: exit status 1`

The session may have ended after the list was loaded. Run `tab` again or press `r` before selecting.

</details>

<details>
<summary><strong>Build and verify</strong></summary>

```sh
git clone https://github.com/hmmhmmhm/tmux-attach-browser.git
cd tmux-attach-browser
go test -race ./...
go vet ./...
go build -o tab ./cmd/tab
```

</details>

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Contributions are described in [CONTRIBUTING.md](CONTRIBUTING.md). Licensed under the [MIT License](LICENSE).
