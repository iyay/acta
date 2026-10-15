---
parent: specs/2026-10-15-strip-release-binary-design
depth: minimal
id: PLN-0141
created: "2026-10-15 09:18:01"
hash: ekuf4fr
started: "2026-10-15 09:18:23"
---
# Strip the release binary Implementation Plan

**Goal:** The release build passes `-s -w`, so the downloaded acta binary drops from about 24M to about 17M.

**Spec:** `.acta/specs/2026-10-15-strip-release-binary-design.md`

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- The existing `-X github.com/iyay/acta/internal/update.release=true` ldflag stays, unchanged.
- No version bump.

## Waves

- Wave 1: Task 01

### Task 01: Release build strips symbols and debug info

**Files:**
- Modify: `.goreleaser.yaml`
- Test: `internal/plugincheck/release_test.go`

**verify:** Every build in `.goreleaser.yaml` passes both `-s` and `-w` in `ldflags` and still passes the `update.release=true` flag; the test fails if any one of the three is missing, in any build. Run `go test ./internal/plugincheck -run TestReleaseGoreleaser`.

- [x] **Failing test:** `TestReleaseGoreleaserStrips` reads `ldflags` of each build and wants `-s`, `-w` and the release flag. `scripts/test ./internal/plugincheck -run TestReleaseGoreleaserStrips` fails.
- [x] **Change:** add `- -s -w` before the `-X` line in `.goreleaser.yaml`, with a comment on why.

## Polish

### Task 02: Review polish

**verify:** Each NOTE below applied, and nothing else changes.

- [x] `.goreleaser.yaml` header names the strip next to the release mark.
- [x] Size comment in `.goreleaser.yaml` and `release_test.go` says about 6M, as measured.
- [x] Commit: `polish: review notes PLN-0141`

## Review notes

- Round 1 (deep, two reviewers): CLEAN.
- The test checks the release value as its own word, not that `-X` comes right before it; a missing `-X` makes the linker fail the release loudly.
- `.acta/wiki/release.md` and `.acta/wiki/.checked` changed outside the task Files list, under the wiki rule.
- The verify line says `go test`; this repo runs `scripts/test`.
- goreleaser is not installed locally, so `goreleaser check` was not run; YAML parse and the new test cover the shape.
- Polish (light tier, comments only): read the diff, plugincheck tests and gofmt green.
