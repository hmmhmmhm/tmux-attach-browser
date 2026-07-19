package main

import (
	"os"

	"github.com/hmmhmmhm/tmux-attach-browser/internal/app"
	"github.com/hmmhmmhm/tmux-attach-browser/internal/tmux"
	"github.com/hmmhmmhm/tmux-attach-browser/internal/ui"
)

func main() {
	client := tmux.NewExecClient("tmux")
	code := app.Run(os.Args[1:], app.Dependencies{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Getwd:  os.Getwd,
		Getenv: os.Getenv,
		Client: client,
		Browse: ui.Run,
	})
	os.Exit(code)
}
