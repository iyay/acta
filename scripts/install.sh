#!/bin/sh
# Installs acta: downloads the release archive, checks its SHA-256 against
# checksums.txt, and copies the binary into the install dir. It never edits
# rc files or PATH.
set -eu

base="${ACTA_DOWNLOAD_URL:-https://github.com/iyay/acta/releases}"
dir="${ACTA_INSTALL_DIR:-$HOME/.local/bin}"

# Say what is wrong and stop. Nothing is installed yet when this runs.
fail() {
	echo "install.sh: $*" >&2
	exit 1
}

# Check the platform first, so an unknown one fails before any download.
case "$(uname -s)" in
Darwin) os=darwin ;;
Linux) os=linux ;;
*) fail "unsupported OS: $(uname -s)" ;;
esac
case "$(uname -m)" in
x86_64 | amd64) arch=amd64 ;;
arm64 | aarch64) arch=arm64 ;;
*) fail "unsupported CPU: $(uname -m)" ;;
esac

# Pick the build before any download. ACTA_TUI wins. With no ACTA_TUI, ask on
# the terminal. With no terminal, nobody can answer, so take the full build,
# which is what every install got before this choice existed.
name=acta
if [ -n "${ACTA_TUI+set}" ]; then
	case "$ACTA_TUI" in
	1) ;;
	0) name=acta-cli ;;
	*) fail "ACTA_TUI must be 0 or 1" ;;
	esac
elif ( : </dev/tty ) 2>/dev/null; then
	cat >/dev/tty <<'EOF'

This will install acta, a development toolkit with a command-line tool
and AI agent skills for tasks such as planning, building, testing, and reviewing code.

You can also install a terminal user interface (TUI) to browse specs, plans, and bugs.
EOF
	while :; do
		printf '\033[1mInstall the TUI? Adds about 7 MB. [Y/n]\033[0m ' >/dev/tty
		read -r answer </dev/tty || break
		case "$answer" in
		"" | y | Y) break ;;
		n | N)
			name=acta-cli
			break
			;;
		esac
	done
fi

if [ -n "${ACTA_VERSION:-}" ]; then
	# Release tags start with v, so 0.1.44 and v0.1.44 mean the same.
	case "$ACTA_VERSION" in
	v*) tag="$ACTA_VERSION" ;;
	*) tag="v$ACTA_VERSION" ;;
	esac
	url="$base/download/$tag"
else
	url="$base/latest/download"
fi

# Linux has sha256sum, macOS has shasum.
if command -v sha256sum >/dev/null 2>&1; then
	hash_of() { sha256sum "$1" | awk '{print $1}'; }
elif command -v shasum >/dev/null 2>&1; then
	hash_of() { shasum -a 256 "$1" | awk '{print $1}'; }
else
	fail "need sha256sum or shasum to check the download"
fi

asset="${name}_${os}_${arch}.tar.gz"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# https only, on the first request and on every redirect. A step down to plain
# http would let someone swap both the archive and checksums.txt.
echo "Downloading $asset"
curl -fsSL --proto '=https' --tlsv1.2 "$url/$asset" -o "$tmp/$asset"
curl -fsSL --proto '=https' --tlsv1.2 "$url/checksums.txt" -o "$tmp/checksums.txt"

# A "*" before the name marks binary mode in some checksum files.
want="$(awk -v n="$asset" '$2 == n || $2 == "*" n { print $1; exit }' "$tmp/checksums.txt")"
[ -n "$want" ] || fail "no checksum line for $asset"
got="$(hash_of "$tmp/$asset")"
[ "$got" = "$want" ] || fail "checksum mismatch for $asset"

tar -xzf "$tmp/$asset" -C "$tmp" acta
mkdir -p "$dir"
# Copy under a temp name, then rename, so a half-copied acta never sits in the dir.
cp "$tmp/acta" "$dir/.acta.$$"
chmod 755 "$dir/.acta.$$"
mv "$dir/.acta.$$" "$dir/acta"
echo "Installed $dir/acta"

case ":$PATH:" in
*":$dir:"*) ;;
*)
	echo "$dir is not on your PATH. Add this line to your shell rc file:"
	echo "  export PATH=\"$dir:\$PATH\""
	;;
esac

# Under "curl | sh" stdin is the script, so setup reads from the terminal.
if ( : </dev/tty ) 2>/dev/null; then
	"$dir/acta" setup </dev/tty || echo "run: acta setup"
else
	echo "run: acta setup"
fi
