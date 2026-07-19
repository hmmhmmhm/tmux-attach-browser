package ui

import "charm.land/bubbles/v2/key"

type keyMap struct {
	newSession key.Binding
	refresh    key.Binding
	quit       key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		newSession: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new session"),
		),
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}
