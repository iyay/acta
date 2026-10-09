---
parent: scratch/2026-10-07-acta-update-command
id: SPC-0117
created: "2026-10-09 16:11:21"
hash: wmv98xt
---
# acta update: check and install the latest version

Status: Architectural, approved by the user in chat on 2026-10-09, section by section. Full answers are in the Log of SCR-0053.

## Why

Today a user gets a new acta only by running `scripts/install.sh` again, and then `acta setup` to unpack the plugin. `acta update` does both in one step: it replaces the binary with the latest release and refreshes the plugin in each harness.

## Rulings

- Scope: binary and plugin together.
- No update hint anywhere else. Only `acta update` and `acta update --check` touch the network. Other commands, hooks and the TUI never do.
- A binary built from source (`go install`) is refused, so a dev build is never replaced by an older release.
- Download and replace in Go. No call to `install.sh`.

## Release marker

Release builds carry no mark today (`.goreleaser.yaml` says nothing is set at link time). Goreleaser now sets one ldflag: `-X <pkg>.release=true`. A binary without it counts as built from source. `plugin.Version()` stays the version source.

## Flow of `acta update`

1. Not a release build: print `acta was built from source; rebuild it with go install` and exit 1.
2. Read the latest tag from the redirect of `<base>/latest`. `<base>` is `ACTA_DOWNLOAD_URL`, default `https://github.com/iyay/acta/releases`, the same variable `install.sh` reads. Same as `plugin.Version()`: print `acta vX is the latest` and exit 0. With `--check`, print the current and latest version and stop here, exit 0.
3. Download `acta_<os>_<arch>.tar.gz` and `checksums.txt` from `<base>/download/<tag>`. https only, also on redirects. Timeout 60 s. Check the SHA-256 against `checksums.txt`, then take `acta` out of the archive.
4. Resolve `os.Executable()` through symlinks. Write `.acta.<pid>` in that dir, chmod 755, rename over the binary. On any failure remove the temp file; the old binary stays as it was.
5. Exec the new binary as `acta update --refresh-plugin` (hidden). It runs `setup.ExtractPlugin` into `~/.acta/plugin`, with no wizard. omp links that folder, so nothing more is needed there. When `claude` is on PATH, it then runs `claude plugin marketplace update acta-local` and `claude plugin update acta@acta-local`, because Claude Code keeps its own copy of the plugin. The first task of the plan checks these two commands against the real `claude` before code depends on them.

## Errors

- Network error, missing checksum line, checksum mismatch: clear message, binary untouched, exit 1.
- No write access to the binary's dir: `no write access to <dir>`, exit 1. No sudo.
- Plugin refresh fails: the binary stays updated. Print `plugin refresh failed: run acta setup`, exit 1.

## Testing

- `httptest.NewTLSServer` behind `ACTA_DOWNLOAD_URL` serves the redirect, archive and checksums. No test reaches GitHub.
- A fake `claude` on PATH records the commands it gets.
- Cases: source build refused, already latest, `--check`, good update, checksum mismatch leaves the binary, read-only dir, refresh failure exit code.
- The release marker gets a test that the goreleaser config sets the ldflag.

## Out of scope

- Windows (SCR-0058).
- A GitHub marketplace for the plugin.
- Rolling back to an older version.
