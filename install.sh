#!/bin/sh
set -eu

repository="hmmhmmhm/tmux-attach-browser"

fail() {
	printf '%s\n' "tab installer: $*" >&2
	exit 1
}

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"

raw_os=${TAB_OS:-$(uname -s)}
case $(printf '%s' "$raw_os" | tr '[:upper:]' '[:lower:]') in
	darwin) os=darwin ;;
	linux) os=linux ;;
	*) fail "unsupported operating system: $raw_os" ;;
esac

raw_arch=${TAB_ARCH:-$(uname -m)}
case $(printf '%s' "$raw_arch" | tr '[:upper:]' '[:lower:]') in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) fail "unsupported architecture: $raw_arch" ;;
esac

archive="tab_${os}_${arch}.tar.gz"
if [ -n "${TAB_RELEASE_BASE_URL:-}" ]; then
	release_base=$TAB_RELEASE_BASE_URL
elif [ -n "${TAB_VERSION:-}" ]; then
	release_base="https://github.com/$repository/releases/download/$TAB_VERSION"
else
	release_base="https://github.com/$repository/releases/latest/download"
fi

install_dir=${TAB_INSTALL_DIR:-"$HOME/.local/bin"}
temporary_dir=$(mktemp -d "${TMPDIR:-/tmp}/tab-install.XXXXXX")
trap 'rm -rf "$temporary_dir"' EXIT INT TERM

curl -fsSL "$release_base/$archive" -o "$temporary_dir/$archive" || fail "could not download $archive"
curl -fsSL "$release_base/checksums.txt" -o "$temporary_dir/checksums.txt" || fail "could not download checksums.txt"

expected=$(awk -v file="$archive" '$2 == file || $2 == "*" file { print $1; exit }' "$temporary_dir/checksums.txt")
[ -n "$expected" ] || fail "checksum entry for $archive was not found"

if command -v sha256sum >/dev/null 2>&1; then
	actual=$(sha256sum "$temporary_dir/$archive" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
	actual=$(shasum -a 256 "$temporary_dir/$archive" | awk '{print $1}')
elif command -v openssl >/dev/null 2>&1; then
	actual=$(openssl dgst -sha256 "$temporary_dir/$archive" | awk '{print $NF}')
else
	fail "sha256sum, shasum, or openssl is required to verify the download"
fi

[ "$actual" = "$expected" ] || fail "checksum verification failed for $archive"

tar -xzf "$temporary_dir/$archive" -C "$temporary_dir" tab || fail "could not extract tab"
mkdir -p "$install_dir"
if command -v install >/dev/null 2>&1; then
	install -m 0755 "$temporary_dir/tab" "$install_dir/tab"
else
	cp "$temporary_dir/tab" "$install_dir/tab"
	chmod 0755 "$install_dir/tab"
fi

printf '%s\n' "Installed tab to $install_dir/tab"
case ":${PATH:-}:" in
	*":$install_dir:"*) ;;
	*) printf '%s\n' "Add $install_dir to PATH to run tab from any directory." ;;
esac

if ! command -v tmux >/dev/null 2>&1; then
	printf '%s\n' "tmux was not found. Install tmux before running tab." >&2
fi
