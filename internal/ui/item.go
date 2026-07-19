package ui

import (
	"fmt"
	"time"

	"github.com/hmmhmmhm/tmux-attach-browser/internal/tmux"
)

type sessionItem struct {
	session tmux.Session
}

func (i sessionItem) Title() string {
	return i.session.Name
}

func (i sessionItem) FilterValue() string {
	return i.session.Name
}

func (i sessionItem) Description() string {
	windows := "window"
	if i.session.Windows != 1 {
		windows = "windows"
	}
	state := "detached"
	if i.session.Attached > 0 {
		state = fmt.Sprintf("%d attached", i.session.Attached)
	}
	activity := i.session.ActivityAt.Local().Format(time.DateTime)
	return fmt.Sprintf("%d %s | %s | active %s", i.session.Windows, windows, state, activity)
}
