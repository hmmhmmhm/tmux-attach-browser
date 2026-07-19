package app

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hmmhmmhm/tmux-attach-browser/internal/tmux"
)

type fakeClient struct {
	connectName   string
	connectInside bool
	connectErr    error
}

func (f *fakeClient) List(context.Context) ([]tmux.Session, error) {
	return nil, nil
}

func (f *fakeClient) Create(context.Context, string, string) error {
	return nil
}

func (f *fakeClient) Connect(_ context.Context, name string, inside bool) error {
	f.connectName = name
	f.connectInside = inside
	return f.connectErr
}

type harness struct {
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	client *fakeClient
	deps   Dependencies
}

func newHarness() *harness {
	h := &harness{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		client: &fakeClient{},
	}
	h.deps = Dependencies{
		Stdout: h.stdout,
		Stderr: h.stderr,
		Getwd:  func() (string, error) { return "/tmp/project", nil },
		Getenv: func(string) string { return "" },
		Client: h.client,
		Browse: func(tmux.Client, string) (string, bool, error) {
			return "", false, nil
		},
	}
	return h
}

func TestRunVersion(t *testing.T) {
	h := newHarness()

	code := Run([]string{"--version"}, h.deps)

	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if got := h.stdout.String(); !strings.Contains(got, "tab dev") {
		t.Fatalf("stdout = %q, want version", got)
	}
}

func TestRunHelp(t *testing.T) {
	h := newHarness()

	code := Run([]string{"--help"}, h.deps)

	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if got := h.stdout.String(); !strings.Contains(got, "Usage: tab [session]") {
		t.Fatalf("stdout = %q, want usage", got)
	}
}

func TestRunTooManyArguments(t *testing.T) {
	h := newHarness()

	code := Run([]string{"one", "two"}, h.deps)

	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if got := h.stderr.String(); !strings.Contains(got, "accepts at most one session") {
		t.Fatalf("stderr = %q, want argument error", got)
	}
}

func TestRunDirectAttach(t *testing.T) {
	h := newHarness()
	h.deps.Getenv = func(key string) string {
		if key == "TMUX" {
			return "/tmp/tmux,1,0"
		}
		return ""
	}

	code := Run([]string{"work"}, h.deps)

	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if h.client.connectName != "work" || !h.client.connectInside {
		t.Fatalf("Connect(%q, %v), want Connect(%q, true)", h.client.connectName, h.client.connectInside, "work")
	}
}

func TestRunBrowsedAttach(t *testing.T) {
	h := newHarness()
	h.deps.Browse = func(_ tmux.Client, cwd string) (string, bool, error) {
		if cwd != "/tmp/project" {
			t.Fatalf("cwd = %q, want /tmp/project", cwd)
		}
		return "chosen", true, nil
	}

	code := Run(nil, h.deps)

	if code != 0 || h.client.connectName != "chosen" {
		t.Fatalf("code = %d, session = %q, want 0 and chosen", code, h.client.connectName)
	}
}

func TestRunCancelled(t *testing.T) {
	h := newHarness()

	code := Run(nil, h.deps)

	if code != 0 || h.client.connectName != "" {
		t.Fatalf("code = %d, session = %q, want clean cancellation", code, h.client.connectName)
	}
}

func TestRunBrowseError(t *testing.T) {
	h := newHarness()
	h.deps.Browse = func(tmux.Client, string) (string, bool, error) {
		return "", false, errors.New("terminal unavailable")
	}

	code := Run(nil, h.deps)

	if code != 1 || !strings.Contains(h.stderr.String(), "terminal unavailable") {
		t.Fatalf("code = %d, stderr = %q", code, h.stderr.String())
	}
}

func TestRunConnectError(t *testing.T) {
	h := newHarness()
	h.client.connectErr = errors.New("session vanished")

	code := Run([]string{"work"}, h.deps)

	if code != 1 || !strings.Contains(h.stderr.String(), "session vanished") {
		t.Fatalf("code = %d, stderr = %q", code, h.stderr.String())
	}
}
