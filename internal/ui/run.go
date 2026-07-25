package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/hmmhmmhm/tmux-attach-browser/internal/tmux"
)

type programRunner interface {
	Run() (tea.Model, error)
}

// Run starts the interactive browser and returns its selected session.
func Run(client tmux.Client, cwd string) (string, bool, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return runProgram(tea.NewProgram(newModel(ctx, client, cwd)))
}

func runProgram(program programRunner) (string, bool, error) {
	result, err := program.Run()
	if err != nil {
		return "", false, fmt.Errorf("run terminal UI: %w", err)
	}
	model, ok := result.(Model)
	if !ok {
		return "", false, fmt.Errorf("unexpected UI result %T", result)
	}
	name, selected := model.Result()
	return name, selected, nil
}
