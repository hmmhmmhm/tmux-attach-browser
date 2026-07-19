// Package ui provides the Bubble Tea tmux session browser.
package ui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hmmhmmhm/tmux-attach-browser/internal/tmux"
)

type viewMode int

const (
	modeList viewMode = iota
	modeCreate
)

type sessionsLoadedMsg struct {
	sessions []tmux.Session
	err      error
}

type sessionCreatedMsg struct {
	name string
	err  error
}

// Model is the complete terminal UI state.
type Model struct {
	client   tmux.Client
	cwd      string
	list     list.Model
	input    textinput.Model
	keys     keyMap
	mode     viewMode
	err      error
	loading  bool
	creating bool
	chosen   string
	selected bool
}

// New creates a session browser model.
func New(client tmux.Client, cwd string) Model {
	keys := newKeyMap()
	delegate := list.NewDefaultDelegate()
	sessionList := list.New(nil, delegate, 80, 24)
	sessionList.Title = "tmux sessions"
	sessionList.SetStatusBarItemName("session", "sessions")
	sessionList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.newSession, keys.refresh}
	}
	sessionList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keys.newSession, keys.refresh}
	}

	input := textinput.New()
	input.Prompt = "session name: "
	input.Placeholder = "project"
	input.CharLimit = 100
	input.Validate = tmux.ValidateSessionName
	input.SetWidth(48)

	return Model{
		client:  client,
		cwd:     cwd,
		list:    sessionList,
		input:   input,
		keys:    keys,
		mode:    modeList,
		loading: true,
	}
}

// Init loads tmux sessions asynchronously.
func (m Model) Init() tea.Cmd {
	return loadSessions(m.client)
}

// Result reports the session chosen by the user.
func (m Model) Result() (string, bool) {
	return m.chosen, m.selected
}

// Update applies a Bubble Tea message.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sessionsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = fmt.Errorf("load sessions: %w", msg.err)
			return m, nil
		}
		m.err = nil
		items := make([]list.Item, len(msg.sessions))
		for index, session := range msg.sessions {
			items[index] = sessionItem{session: session}
		}
		return m, m.list.SetItems(items)

	case sessionCreatedMsg:
		m.creating = false
		if msg.err != nil {
			m.err = fmt.Errorf("create session: %w", msg.err)
			return m, nil
		}
		m.chosen = msg.name
		m.selected = true
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.input.SetWidth(max(20, min(60, msg.Width-8)))

	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.quit) {
			m.selected = false
			return m, tea.Quit
		}
	}

	if m.mode == modeCreate {
		return m.updateCreate(msg)
	}
	return m.updateList(msg)
}

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && !m.list.SettingFilter() {
		switch {
		case key.Matches(keyMsg, m.keys.newSession):
			m.mode = modeCreate
			m.err = nil
			m.input.Reset()
			return m, m.input.Focus()

		case key.Matches(keyMsg, m.keys.refresh):
			m.loading = true
			m.err = nil
			return m, loadSessions(m.client)

		case keyMsg.Code == tea.KeyEnter:
			item, ok := m.list.SelectedItem().(sessionItem)
			if ok {
				m.chosen = item.session.Name
				m.selected = true
				return m, tea.Quit
			}
		}
	}

	updated, cmd := m.list.Update(msg)
	m.list = updated
	return m, cmd
}

func (m Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.Code {
		case tea.KeyEscape:
			m.mode = modeList
			m.err = nil
			m.creating = false
			m.input.Reset()
			m.input.Blur()
			return m, nil

		case tea.KeyEnter:
			if m.creating {
				return m, nil
			}
			name := strings.TrimSpace(m.input.Value())
			if err := tmux.ValidateSessionName(name); err != nil {
				m.err = err
				return m, nil
			}
			m.err = nil
			m.creating = true
			return m, createSession(m.client, name, m.cwd)
		}
	}

	updated, cmd := m.input.Update(msg)
	m.input = updated
	if m.input.Err != nil {
		m.err = m.input.Err
	} else {
		m.err = nil
	}
	return m, cmd
}

// View renders the active list or create prompt in the alternate screen.
func (m Model) View() tea.View {
	content := m.list.View()
	if m.mode == modeCreate {
		body := "Create a new tmux session\n\n" + m.input.View()
		if m.creating {
			body += "\n\nCreating session..."
		} else {
			body += "\n\nEnter create | Esc cancel"
		}
		if m.err != nil {
			body += "\n\n" + errorStyle.Render(m.err.Error())
		}
		content = promptStyle.Render(body)
	} else {
		if len(m.list.Items()) == 0 && !m.loading && m.err == nil {
			content += "\nNo tmux sessions. Press n to create one."
		}
		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}
	}

	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

var (
	promptStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

func loadSessions(client tmux.Client) tea.Cmd {
	return func() tea.Msg {
		sessions, err := client.List(context.Background())
		return sessionsLoadedMsg{sessions: sessions, err: err}
	}
}

func createSession(client tmux.Client, name, cwd string) tea.Cmd {
	return func() tea.Msg {
		err := client.Create(context.Background(), name, cwd)
		return sessionCreatedMsg{name: name, err: err}
	}
}
