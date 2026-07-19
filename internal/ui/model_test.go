package ui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/hmmhmmhm/tmux-attach-browser/internal/tmux"
)

type fakeClient struct {
	lists       [][]tmux.Session
	listErrors  []error
	listCalls   int
	createName  string
	createDir   string
	createError error
}

type fakeProgram struct {
	model tea.Model
	err   error
}

func (f fakeProgram) Run() (tea.Model, error) {
	return f.model, f.err
}

func (f *fakeClient) List(context.Context) ([]tmux.Session, error) {
	index := f.listCalls
	f.listCalls++
	if index < len(f.listErrors) && f.listErrors[index] != nil {
		return nil, f.listErrors[index]
	}
	if len(f.lists) == 0 {
		return []tmux.Session{}, nil
	}
	if index >= len(f.lists) {
		index = len(f.lists) - 1
	}
	return f.lists[index], nil
}

func (f *fakeClient) Check(context.Context) error { return nil }

func (f *fakeClient) Create(_ context.Context, name, dir string) error {
	f.createName = name
	f.createDir = dir
	return f.createError
}

func (f *fakeClient) Connect(context.Context, string, bool) error { return nil }

func session(name string, windows, attached int) tmux.Session {
	return tmux.Session{
		Name:       name,
		Windows:    windows,
		Attached:   attached,
		CreatedAt:  time.Unix(100, 0),
		ActivityAt: time.Unix(200, 0),
	}
}

func keyMsg(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text}
}

func updateModel(t *testing.T, model Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	updated, cmd := model.Update(msg)
	result, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model type = %T", updated)
	}
	return result, cmd
}

func loadModel(t *testing.T, client *fakeClient) Model {
	t.Helper()
	model := New(client, "/tmp/project")
	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init returned nil command")
	}
	model, _ = updateModel(t, model, cmd())
	return model
}

func TestInitLoadsSessions(t *testing.T) {
	client := &fakeClient{lists: [][]tmux.Session{{session("work", 2, 0)}}}

	model := loadModel(t, client)

	if client.listCalls != 1 || len(model.list.Items()) != 1 {
		t.Fatalf("list calls = %d, items = %d", client.listCalls, len(model.list.Items()))
	}
}

func TestSessionsAreRenderedAsListItems(t *testing.T) {
	client := &fakeClient{lists: [][]tmux.Session{{session("work", 3, 1)}}}

	model := loadModel(t, client)
	item, ok := model.list.Items()[0].(sessionItem)
	if !ok {
		t.Fatalf("item type = %T", model.list.Items()[0])
	}
	if item.Title() != "work" || item.FilterValue() != "work" {
		t.Fatalf("title = %q, filter = %q", item.Title(), item.FilterValue())
	}
	if description := item.Description(); !strings.Contains(description, "3 windows") || !strings.Contains(description, "1 attached") {
		t.Fatalf("description = %q", description)
	}
}

func TestEnterChoosesSelectedSession(t *testing.T) {
	client := &fakeClient{lists: [][]tmux.Session{{session("work", 1, 0)}}}
	model := loadModel(t, client)

	model, cmd := updateModel(t, model, keyMsg(tea.KeyEnter, ""))

	name, selected := model.Result()
	if name != "work" || !selected || cmd == nil {
		t.Fatalf("result = %q, %v; cmd nil = %v", name, selected, cmd == nil)
	}
}

func TestRefreshReplacesItems(t *testing.T) {
	client := &fakeClient{lists: [][]tmux.Session{
		{session("old", 1, 0)},
		{session("new", 2, 0)},
	}}
	model := loadModel(t, client)

	model, cmd := updateModel(t, model, keyMsg('r', "r"))
	if cmd == nil {
		t.Fatal("refresh returned nil command")
	}
	model, _ = updateModel(t, model, cmd())

	item := model.list.Items()[0].(sessionItem)
	if item.Title() != "new" || client.listCalls != 2 {
		t.Fatalf("title = %q, calls = %d", item.Title(), client.listCalls)
	}
}

func TestRefreshErrorKeepsExistingItems(t *testing.T) {
	client := &fakeClient{
		lists:      [][]tmux.Session{{session("work", 1, 0)}},
		listErrors: []error{nil, errors.New("socket denied")},
	}
	model := loadModel(t, client)

	model, cmd := updateModel(t, model, keyMsg('r', "r"))
	model, _ = updateModel(t, model, cmd())

	if len(model.list.Items()) != 1 || model.err == nil || !strings.Contains(model.err.Error(), "socket denied") {
		t.Fatalf("items = %d, error = %v", len(model.list.Items()), model.err)
	}
}

func TestWindowSizeUpdatesListDimensions(t *testing.T) {
	model := New(&fakeClient{}, "/tmp/project")

	model, _ = updateModel(t, model, tea.WindowSizeMsg{Width: 100, Height: 40})

	if model.list.Width() != 100 || model.list.Height() != 40 {
		t.Fatalf("size = %dx%d", model.list.Width(), model.list.Height())
	}
}

func TestCtrlCCancels(t *testing.T) {
	model := New(&fakeClient{}, "/tmp/project")
	message := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}

	model, cmd := updateModel(t, model, message)

	_, selected := model.Result()
	if selected || cmd == nil {
		t.Fatalf("selected = %v, cmd nil = %v", selected, cmd == nil)
	}
}

func TestSlashStartsFiltering(t *testing.T) {
	model := loadModel(t, &fakeClient{lists: [][]tmux.Session{{session("work", 1, 0)}}})

	model, _ = updateModel(t, model, keyMsg('/', "/"))

	if model.list.FilterState() != list.Filtering {
		t.Fatalf("filter state = %v, want filtering", model.list.FilterState())
	}
}

func TestEmptyListExplainsCreation(t *testing.T) {
	model := loadModel(t, &fakeClient{})

	view := model.View().Content

	if !strings.Contains(view, "No tmux sessions") || !strings.Contains(view, "n") {
		t.Fatalf("view = %q", view)
	}
}

func TestNewKeyOpensCreatePrompt(t *testing.T) {
	model := loadModel(t, &fakeClient{})

	model, _ = updateModel(t, model, keyMsg('n', "n"))

	if model.mode != modeCreate || !model.input.Focused() {
		t.Fatalf("mode = %v, focused = %v", model.mode, model.input.Focused())
	}
}

func TestEscapeCancelsCreatePrompt(t *testing.T) {
	model := loadModel(t, &fakeClient{})
	model, _ = updateModel(t, model, keyMsg('n', "n"))
	model.input.SetValue("draft")

	model, _ = updateModel(t, model, keyMsg(tea.KeyEscape, ""))

	if model.mode != modeList || model.input.Value() != "" {
		t.Fatalf("mode = %v, input = %q", model.mode, model.input.Value())
	}
}

func TestBlankNameStaysInPrompt(t *testing.T) {
	model := loadModel(t, &fakeClient{})
	model, _ = updateModel(t, model, keyMsg('n', "n"))

	model, cmd := updateModel(t, model, keyMsg(tea.KeyEnter, ""))

	if model.mode != modeCreate || model.err == nil || cmd != nil {
		t.Fatalf("mode = %v, error = %v, cmd nil = %v", model.mode, model.err, cmd == nil)
	}
}

func TestInvalidNameStaysInPrompt(t *testing.T) {
	model := loadModel(t, &fakeClient{})
	model, _ = updateModel(t, model, keyMsg('n', "n"))
	model.input.SetValue("bad.name")

	model, cmd := updateModel(t, model, keyMsg(tea.KeyEnter, ""))

	if model.mode != modeCreate || model.err == nil || cmd != nil {
		t.Fatalf("mode = %v, error = %v, cmd nil = %v", model.mode, model.err, cmd == nil)
	}
}

func TestCreateSuccessChoosesNewSession(t *testing.T) {
	client := &fakeClient{}
	model := loadModel(t, client)
	model, _ = updateModel(t, model, keyMsg('n', "n"))
	model, _ = updateModel(t, model, keyMsg('w', "w"))
	model, _ = updateModel(t, model, keyMsg('o', "o"))
	model, _ = updateModel(t, model, keyMsg('r', "r"))
	model, _ = updateModel(t, model, keyMsg('k', "k"))

	model, cmd := updateModel(t, model, keyMsg(tea.KeyEnter, ""))
	if cmd == nil {
		t.Fatal("create returned nil command")
	}
	model, quit := updateModel(t, model, cmd())

	name, selected := model.Result()
	if name != "work" || !selected || client.createName != "work" || client.createDir != "/tmp/project" || quit == nil {
		t.Fatalf("result = %q/%v, create = %q/%q, quit nil = %v", name, selected, client.createName, client.createDir, quit == nil)
	}
}

func TestCreateFailurePreservesInput(t *testing.T) {
	client := &fakeClient{createError: errors.New("duplicate session")}
	model := loadModel(t, client)
	model, _ = updateModel(t, model, keyMsg('n', "n"))
	model.input.SetValue("work")

	model, cmd := updateModel(t, model, keyMsg(tea.KeyEnter, ""))
	model, quit := updateModel(t, model, cmd())

	if model.mode != modeCreate || model.input.Value() != "work" || model.err == nil || quit != nil {
		t.Fatalf("mode = %v, input = %q, error = %v, quit nil = %v", model.mode, model.input.Value(), model.err, quit == nil)
	}
}

func TestRunProgramReturnsSelection(t *testing.T) {
	model := New(&fakeClient{}, "/tmp/project")
	model.chosen = "work"
	model.selected = true

	name, selected, err := runProgram(fakeProgram{model: model})

	if err != nil || name != "work" || !selected {
		t.Fatalf("result = %q/%v, error = %v", name, selected, err)
	}
}

func TestRunProgramReturnsProgramError(t *testing.T) {
	_, _, err := runProgram(fakeProgram{err: errors.New("terminal failed")})

	if err == nil || !strings.Contains(err.Error(), "terminal failed") {
		t.Fatalf("error = %v", err)
	}
}

func TestRunProgramRejectsUnexpectedModel(t *testing.T) {
	_, _, err := runProgram(fakeProgram{})

	if err == nil || !strings.Contains(err.Error(), "unexpected UI result") {
		t.Fatalf("error = %v", err)
	}
}
