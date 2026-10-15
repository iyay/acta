---
parent: scratch/2026-10-15-optional-tui-install
id: SPC-0131
created: "2026-10-15 09:16:52"
hash: a45zqwx
started: "2026-10-15 09:18:23"
---
# Strip the release binary

Status: Bounded, approved by the user in chat on 2026-10-15.

## Problem

A user asked for a smaller acta binary, first as an option to skip the TUI. Measured on 2026-10-15: the release build is 24M. Building with `-s -w` makes it 17M. The TUI libraries add only about 1.5M, so dropping the TUI saves 2-3M at the cost of a second artifact per platform. `.goreleaser.yaml` has no `-s -w` today.

## Design

- Add `-s` and `-w` to `ldflags` in `.goreleaser.yaml`, before the existing `-X ...release=true`. A short comment says why: they drop the symbol table and debug info, so the download is smaller.
- `runtime/debug` build info stays, so `acta doctor` still reads the version. Panic traces keep function names but lose file and line.
- The no-TUI build stays on hold in SCR-0068. Reopen it only if 17M is still too big.

## Testing

A new test in `internal/plugincheck/release_test.go` parses `.goreleaser.yaml` and fails when the build `ldflags` miss `-s` or `-w`. Red first, then green.
