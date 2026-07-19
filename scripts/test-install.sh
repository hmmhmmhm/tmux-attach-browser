#!/bin/sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
test_root=$(mktemp -d "${TMPDIR:-/tmp}/tab-install-test.XXXXXX")
trap 'rm -rf "$test_root"' EXIT INT TERM

release_dir="$test_root/releases/v0.1.0"
mkdir -p "$release_dir/package" "$test_root/bin"

cat >"$release_dir/package/tab" <<'EOF'
#!/bin/sh
printf '%s\n' 'fixture tab v0.1.0'
EOF
chmod 0755 "$release_dir/package/tab"

archive="$release_dir/tab_linux_amd64.tar.gz"
tar -C "$release_dir/package" -czf "$archive" tab

if command -v sha256sum >/dev/null 2>&1; then
	checksum=$(sha256sum "$archive" | awk '{print $1}')
else
	checksum=$(shasum -a 256 "$archive" | awk '{print $1}')
fi
printf '%s  %s\n' "$checksum" "$(basename "$archive")" >"$release_dir/checksums.txt"

TAB_RELEASE_BASE_URL="file://$release_dir" \
	TAB_VERSION=v0.1.0 \
	TAB_INSTALL_DIR="$test_root/bin" \
	TAB_OS=linux \
	TAB_ARCH=amd64 \
	sh "$project_root/install.sh" >/dev/null

actual=$($test_root/bin/tab)
if [ "$actual" != "fixture tab v0.1.0" ]; then
	printf '%s\n' "FAIL: installed binary output was $actual" >&2
	exit 1
fi

printf '%064d  %s\n' 0 "$(basename "$archive")" >"$release_dir/checksums.txt"
if TAB_RELEASE_BASE_URL="file://$release_dir" \
	TAB_INSTALL_DIR="$test_root/bad-checksum" \
	TAB_OS=linux \
	TAB_ARCH=amd64 \
	sh "$project_root/install.sh" >/dev/null 2>&1; then
	printf '%s\n' "FAIL: installer accepted a bad checksum" >&2
	exit 1
fi

if TAB_RELEASE_BASE_URL="file://$release_dir" \
	TAB_INSTALL_DIR="$test_root/unsupported" \
	TAB_OS=linux \
	TAB_ARCH=sparc \
	sh "$project_root/install.sh" >/dev/null 2>&1; then
	printf '%s\n' "FAIL: installer accepted an unsupported architecture" >&2
	exit 1
fi

printf '%s\n' "installer tests passed"
