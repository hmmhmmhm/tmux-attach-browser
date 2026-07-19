#!/bin/sh
set -eu

command -v tmux >/dev/null 2>&1 || {
	printf '%s\n' "tmux is required for the smoke test" >&2
	exit 1
}

socket="tab-smoke-$$"
cleanup() {
	tmux -L "$socket" kill-server >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

tmux -L "$socket" new-session -d -s alpha
tmux -L "$socket" new-session -d -s beta

format='#{session_name}\t#{session_windows}\t#{session_attached}\t#{session_created}\t#{session_activity}'
sessions=$(tmux -L "$socket" list-sessions -F "$format")

printf '%s\n' "$sessions" | awk -F '\t' 'NF != 5 { exit 1 }'
printf '%s\n' "$sessions" | awk -F '\t' '$1 == "alpha" { found = 1 } END { exit !found }'
printf '%s\n' "$sessions" | awk -F '\t' '$1 == "beta" { found = 1 } END { exit !found }'
tmux -L "$socket" has-session -t alpha
tmux -L "$socket" has-session -t beta

printf '%s\n' "tmux smoke test passed"
