#!/bin/sh
set -eu

project_root=$(CDPATH='' cd -- "$(dirname "$0")/.." && pwd)
fixture=$(mktemp -d)
trap 'rm -rf "$fixture"' EXIT INT TERM

git -C "$fixture" init -q
awk 'BEGIN { for (i = 1; i <= 450; i++) print "# line" }' >"$fixture/boundary.sh"
git -C "$fixture" add boundary.sh

sh "$project_root/scripts/check-source-lines.sh" "$fixture"
printf '%s\n' "# line 451" >>"$fixture/boundary.sh"

if sh "$project_root/scripts/check-source-lines.sh" "$fixture"; then
	printf '%s\n' "expected a 451-line file to fail" >&2
	exit 1
fi

printf '%s\n' "source line limit tests passed"
