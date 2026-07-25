#!/bin/sh
set -eu

project_root=$(CDPATH='' cd -- "$(dirname "$0")/.." && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT INT TERM

cd "$project_root"
go build -o "$scratch/tab" ./cmd/tab

cat >"$scratch/tmux" <<'SCRIPT'
#!/bin/sh
set -eu
case ${1:-} in
	-V)
		printf '%s\n' "tmux 3.6"
		;;
	attach-session|switch-client)
		printf '%s\n' "$*" >>"$TAB_TMUX_LOG"
		;;
	*)
		printf 'unexpected tmux arguments: %s\n' "$*" >&2
		exit 1
		;;
esac
SCRIPT
chmod +x "$scratch/tmux"

: >"$scratch/tmux.log"
PATH="$scratch:$PATH" TAB_TMUX_LOG="$scratch/tmux.log" TMUX='' \
	"$scratch/tab" alpha
grep -Fx 'attach-session -t alpha' "$scratch/tmux.log"

: >"$scratch/tmux.log"
PATH="$scratch:$PATH" TAB_TMUX_LOG="$scratch/tmux.log" TMUX=inside \
	"$scratch/tab" alpha
grep -Fx 'switch-client -t alpha' "$scratch/tmux.log"

printf '%s\n' "tab binary smoke test passed"
