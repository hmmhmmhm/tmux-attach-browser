#!/bin/sh
set -eu

max_lines=450
repository=${1:-$(git rev-parse --show-toplevel)}
scratch=$(mktemp -d)
cleanup() {
	rm -rf "$scratch"
}
trap cleanup EXIT INT TERM

git -C "$repository" ls-files -- \
	'*.go' '*.sh' '*.ps1' '*.yml' '*.yaml' >"$scratch/files"

status=0
while IFS= read -r file; do
	lines=$(awk 'END { print NR }' "$repository/$file")
	if [ "$lines" -gt "$max_lines" ]; then
		printf '%s: %s lines (maximum %s)\n' "$file" "$lines" "$max_lines" >&2
		status=1
	fi
done <"$scratch/files"

exit "$status"
