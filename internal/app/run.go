// Package app coordinates command-line behavior around tmux and the TUI.
package app

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/hmmhmmhm/tmux-attach-browser/internal/buildinfo"
	"github.com/hmmhmmhm/tmux-attach-browser/internal/tmux"
)

const usage = `Usage: tab [session]

Browse tmux sessions or attach directly to a named session.

Options:
  -h, --help       Show this help
  -v, --version    Show version information
`

// Dependencies contains the process resources used by Run.
type Dependencies struct {
	Stdout io.Writer
	Stderr io.Writer
	Getwd  func() (string, error)
	Getenv func(string) string
	Client tmux.Client
	Browse func(tmux.Client, string) (string, bool, error)
}

// Run executes the non-interactive command orchestration and returns an exit code.
func Run(args []string, deps Dependencies) int {
	switch {
	case len(args) == 1 && (args[0] == "-h" || args[0] == "--help"):
		fmt.Fprint(deps.Stdout, usage)
		return 0
	case len(args) == 1 && (args[0] == "-v" || args[0] == "--version"):
		fmt.Fprintf(deps.Stdout, "tab %s (%s, %s)\n", buildinfo.Version, buildinfo.Commit, buildinfo.Date)
		return 0
	case len(args) > 1:
		fmt.Fprintln(deps.Stderr, "tab accepts at most one session name")
		fmt.Fprint(deps.Stderr, usage)
		return 2
	case len(args) == 1 && strings.HasPrefix(args[0], "-"):
		fmt.Fprintf(deps.Stderr, "unknown option: %s\n", args[0])
		fmt.Fprint(deps.Stderr, usage)
		return 2
	}

	name := ""
	selected := false
	if len(args) == 1 {
		name, selected = args[0], true
	} else {
		cwd, err := deps.Getwd()
		if err != nil {
			fmt.Fprintf(deps.Stderr, "tab: determine current directory: %v\n", err)
			return 1
		}
		name, selected, err = deps.Browse(deps.Client, cwd)
		if err != nil {
			fmt.Fprintf(deps.Stderr, "tab: %v\n", err)
			return 1
		}
	}

	if !selected {
		return 0
	}
	if err := deps.Client.Connect(context.Background(), name, deps.Getenv("TMUX") != ""); err != nil {
		fmt.Fprintf(deps.Stderr, "tab: %v\n", err)
		return 1
	}
	return 0
}
