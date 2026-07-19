package tmux

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type runnerCall struct {
	mode string
	name string
	args []string
}

type fakeRunner struct {
	stdout    []byte
	stderr    []byte
	outputErr error
	runErr    error
	calls     []runnerCall
}

func (f *fakeRunner) Output(_ context.Context, name string, args ...string) ([]byte, []byte, error) {
	f.calls = append(f.calls, runnerCall{mode: "output", name: name, args: append([]string(nil), args...)})
	return f.stdout, f.stderr, f.outputErr
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) error {
	f.calls = append(f.calls, runnerCall{mode: "run", name: name, args: append([]string(nil), args...)})
	return f.runErr
}

func TestExecClientListsSessions(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("work\t2\t0\t10\t20\n")}
	client := newExecClient("tmux-custom", runner)

	sessions, err := client.List(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].Name != "work" {
		t.Fatalf("sessions = %#v", sessions)
	}
	want := runnerCall{mode: "output", name: "tmux-custom", args: []string{"list-sessions", "-F", sessionFormat}}
	if !reflect.DeepEqual(runner.calls, []runnerCall{want}) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, []runnerCall{want})
	}
}

func TestExecClientChecksVersion(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("tmux 3.6a\n")}
	client := newExecClient("tmux-custom", runner)

	err := client.Check(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	want := []runnerCall{{mode: "output", name: "tmux-custom", args: []string{"-V"}}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestExecClientCheckIdentifiesMissingExecutable(t *testing.T) {
	runner := &fakeRunner{outputErr: errors.New("executable file not found in $PATH")}
	client := newExecClient("tmux", runner)

	err := client.Check(context.Background())

	if err == nil || !strings.Contains(err.Error(), "tmux is not installed") {
		t.Fatalf("error = %v", err)
	}
}

func TestExecClientTreatsMissingServerAsEmpty(t *testing.T) {
	for _, stderr := range []string{
		"no server running on /tmp/tmux-501/default",
		"failed to connect to server: Connection refused",
	} {
		t.Run(stderr, func(t *testing.T) {
			runner := &fakeRunner{stderr: []byte(stderr), outputErr: errors.New("exit status 1")}
			client := newExecClient("tmux", runner)

			sessions, err := client.List(context.Background())

			if err != nil || len(sessions) != 0 {
				t.Fatalf("sessions = %#v, error = %v", sessions, err)
			}
		})
	}
}

func TestExecClientReturnsListFailure(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("permission denied"), outputErr: errors.New("exit status 1")}
	client := newExecClient("tmux", runner)

	_, err := client.List(context.Background())

	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("error = %v", err)
	}
}

func TestExecClientCreatesSessionInDirectory(t *testing.T) {
	runner := &fakeRunner{}
	client := newExecClient("tmux", runner)

	err := client.Create(context.Background(), "work", "/tmp/project")

	if err != nil {
		t.Fatal(err)
	}
	want := []runnerCall{{mode: "run", name: "tmux", args: []string{"new-session", "-d", "-s", "work", "-c", "/tmp/project"}}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestExecClientAttachesOutsideTmux(t *testing.T) {
	runner := &fakeRunner{}
	client := newExecClient("tmux", runner)

	err := client.Connect(context.Background(), "work", false)

	if err != nil {
		t.Fatal(err)
	}
	want := []runnerCall{{mode: "run", name: "tmux", args: []string{"attach-session", "-t", "work"}}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestExecClientSwitchesInsideTmux(t *testing.T) {
	runner := &fakeRunner{}
	client := newExecClient("tmux", runner)

	err := client.Connect(context.Background(), "work", true)

	if err != nil {
		t.Fatal(err)
	}
	want := []runnerCall{{mode: "run", name: "tmux", args: []string{"switch-client", "-t", "work"}}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestExecClientIdentifiesMissingExecutable(t *testing.T) {
	runner := &fakeRunner{outputErr: errors.New("executable file not found in $PATH")}
	client := newExecClient("tmux", runner)

	_, err := client.List(context.Background())

	if err == nil || !strings.Contains(err.Error(), "tmux is not installed") {
		t.Fatalf("error = %v", err)
	}
}
