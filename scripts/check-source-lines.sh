#!/bin/sh
set -eu

max_lines=450
repository=${1:-$(git rev-parse --show-toplevel)}
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT INT TERM

git -C "$repository" ls-files -- \
	'*.go' '*.sh' '*.ps1' '*.yml' '*.yaml' \
	':(exclude)vendor/**' ':(exclude)dist/**' >"$scratch/files"

status=0
while IFS= read -r file; do
	if [ "${file##*.}" = "go" ] &&
		sed -n '1,20p' "$repository/$file" |
			grep -q '^// Code generated .* DO NOT EDIT\.$'; then
		continue
	fi
	lines=$(awk 'END { print NR }' "$repository/$file")
	if [ "$lines" -gt "$max_lines" ]; then
		printf '%s: %s lines (maximum %s)\n' "$file" "$lines" "$max_lines" >&2
		status=1
	fi
done <"$scratch/files"

exit "$status"
