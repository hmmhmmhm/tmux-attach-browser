package tmux

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type commandRunner interface {
	Output(context.Context, string, ...string) ([]byte, []byte, error)
	Run(context.Context, string, ...string) error
}

type osRunner struct{}

func (osRunner) Output(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	return stdout, stderr.Bytes(), err
}

func (osRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ExecClient invokes a tmux executable directly without shell evaluation.
type ExecClient struct {
	executable string
	runner     commandRunner
}

// NewExecClient creates a client that invokes executable.
func NewExecClient(executable string) *ExecClient {
	return newExecClient(executable, osRunner{})
}

func newExecClient(executable string, runner commandRunner) *ExecClient {
	return &ExecClient{executable: executable, runner: runner}
}

// List returns sessions ordered by recent activity.
func (c *ExecClient) List(ctx context.Context) ([]Session, error) {
	stdout, stderr, err := c.runner.Output(ctx, c.executable, "list-sessions", "-F", sessionFormat)
	if err != nil {
		message := strings.TrimSpace(string(stderr))
		lower := strings.ToLower(message)
		if strings.Contains(lower, "no server running") || strings.Contains(lower, "failed to connect to server") {
			return []Session{}, nil
		}
		return nil, commandError("list sessions", message, err)
	}

	sessions, err := ParseSessions(stdout)
	if err != nil {
		return nil, fmt.Errorf("parse tmux sessions: %w", err)
	}
	return sessions, nil
}

// Create starts a detached session in cwd.
func (c *ExecClient) Create(ctx context.Context, name, cwd string) error {
	if err := ValidateSessionName(name); err != nil {
		return err
	}
	if err := c.runner.Run(ctx, c.executable, "new-session", "-d", "-s", name, "-c", cwd); err != nil {
		return commandError("create session", "", err)
	}
	return nil
}

// Connect attaches outside tmux and switches clients inside tmux.
func (c *ExecClient) Connect(ctx context.Context, name string, insideTmux bool) error {
	if err := ValidateSessionName(name); err != nil {
		return err
	}
	command := "attach-session"
	if insideTmux {
		command = "switch-client"
	}
	if err := c.runner.Run(ctx, c.executable, command, "-t", name); err != nil {
		return commandError("connect to session", "", err)
	}
	return nil
}

func commandError(action, stderr string, err error) error {
	if errors.Is(err, exec.ErrNotFound) || strings.Contains(strings.ToLower(err.Error()), "executable file not found") {
		return fmt.Errorf("tmux is not installed or is not on PATH")
	}
	if stderr != "" {
		return fmt.Errorf("%s: %s", action, stderr)
	}
	return fmt.Errorf("%s: %w", action, err)
}
