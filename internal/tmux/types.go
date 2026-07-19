package tmux

import (
	"context"
	"time"
)

// Session contains the tmux metadata displayed by the browser.
type Session struct {
	Name       string
	Windows    int
	Attached   int
	CreatedAt  time.Time
	ActivityAt time.Time
}

// Client is the tmux behavior needed by the command and terminal UI.
type Client interface {
	Check(context.Context) error
	List(context.Context) ([]Session, error)
	Create(context.Context, string, string) error
	Connect(context.Context, string, bool) error
}
