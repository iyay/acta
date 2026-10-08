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

asset="acta_${os}_${arch}.tar.gz"
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
