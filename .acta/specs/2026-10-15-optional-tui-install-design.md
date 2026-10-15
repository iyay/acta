---
parent: scratch/2026-10-15-optional-tui-install
id: SPC-0133
created: "2026-10-15 09:34:43"
hash: m1t1xxh
started: "2026-10-15 09:37:25"
finished: "2026-10-15 09:45:27"
---
# Optional TUI at install time

Status: Architectural, approved by the user in chat on 2026-10-15, section by section.

## Problem

Every user gets one acta binary that holds both the CLI and the TUI board. Skills and hooks only call the CLI. A user who never opens the board still downloads it.

Measured on 2026-10-15 (darwin/arm64, stripped, in a scratch copy with `runTUI` stubbed out): the full build is 16.6M and a build without the TUI is 9.4M. That is 7.2M less, or 43%. Most of the gap is the markdown stack only the TUI uses: glamour, chroma, goldmark and bluemonday. `setup` (huh) and `doctor` (lipgloss) keep bubbletea and lipgloss in both builds.

## Goal

1. A user can pick, at install time, a smaller acta with no TUI.
2. Nothing changes for anyone who does not pick it.
3. `acta update` keeps each user on the build they picked.

## Design

### 1. Build and release

- `runTUI` and its TUI imports move out of `internal/cli/cli.go` into `internal/cli/tui_full.go`, with `//go:build !notui`. The code stays the same. `internal/cli/tuitrace.go` gets the same tag.
- New `internal/cli/tui_none.go`, with `//go:build notui`. Its `runTUI` prints the usual help and one more line: the TUI is not in this build; to get it, run the installer again with `ACTA_TUI=1`. It exits 0.
- A constant `cli.Variant` is `"full"` or `"cli"`, set by the tag. `acta --version` and `acta doctor` show it.
- `.goreleaser.yaml` adds a second build, `acta-cli`, with `tags: [notui]` and the same ldflags, plus a second archive named `acta-cli_{{ .Os }}_{{ .Arch }}`. The existing `acta_<os>_<arch>.tar.gz` asset keeps its name, so install links already shared keep working. Both archives hold a file named `acta`.

### 2. Installer (`scripts/install.sh`)

- New env `ACTA_TUI`. `1` gets the full build and `0` the small one. Any other value fails before any download, with `ACTA_TUI must be 0 or 1`.
- `ACTA_TUI` unset and `/dev/tty` readable (the same check as the setup step): ask once, before the download, `Install the TUI board too? It adds about 7M. [Y/n]`. Enter or `y` gets full, `n` gets small, anything else asks again.
- `ACTA_TUI` unset and no tty: full build, no prompt. This is the same as today.
- The asset is `acta_${os}_${arch}.tar.gz` or `acta-cli_${os}_${arch}.tar.gz`. The checksum check, the extract step and the installed file name `acta` stay the same, so PATH, hooks and skills see no change.
- `scripts/install.ps1` is not touched. Windows is not released yet (SCR-0058).

### 3. `acta update` and `acta doctor`

- `update.Fetch` takes the asset prefix as a new argument, `"acta"` or `"acta-cli"`. `internal/cli/update_cmd.go` passes the one that matches `cli.Variant`. Checksum, size caps and the binary swap stay as they are.
- A release that has no `acta-cli_*` asset (one cut before this change) makes `update` on the small build fail with a clear error. Nothing is replaced.
- `acta doctor` adds one info line, `build: full` or `build: cli (no TUI)`. It never warns or fails.
- To switch builds later, run the installer again with `ACTA_TUI`. `acta update` has no flag for this.
- `.acta/wiki/release.md` gets one sentence on the `acta-cli_*` assets and `ACTA_TUI`.

## Testing

- `internal/plugincheck/release_test.go`: two builds, both stripped and both with the release flag. Only `acta-cli` has `notui`. Archive names are `acta_{{ .Os }}_{{ .Arch }}` and `acta-cli_{{ .Os }}_{{ .Arch }}`.
- A test runs `go list -deps -tags notui ./cmd/acta` and fails if the list holds `internal/tui` or `github.com/charmbracelet/glamour`. This stops a later import from pulling the TUI back into the small build without anyone noticing.
- `.github/workflows/ci.yml` also builds with `-tags notui` on the four os/arch pairs.
- `scripts/install_test.go` covers `ACTA_TUI=0`, `ACTA_TUI=1`, a bad value (fails, no download), no tty (full build), and the prompt through a pty: Enter, `n`, and a bad answer that asks again. Each case checks which asset was fetched.
- `internal/update`: `Fetch` with prefix `acta-cli` fetches and checks `acta-cli_*`. A missing `acta-cli_*` line in the release fails clearly and returns no bytes.
- The `doctor` line and the bare `acta` output are tested in both builds. To make that real, `scripts/test --full` also runs `go test -tags notui ./internal/cli`, and CI runs the same.

## Out of scope

- A setting or flag inside `acta` to switch builds.
- Windows.
