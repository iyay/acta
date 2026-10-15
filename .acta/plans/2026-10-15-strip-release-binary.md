---
parent: specs/2026-10-15-strip-release-binary-design
depth: minimal
id: PLN-0141
created: "2026-10-15 09:18:01"
hash: ekuf4fr
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

Write a failing test that reads `ldflags` of each build and wants `-s`, `-w` and the release flag; watch it fail; add `- -s -w` before the `-X` line with a one-line comment on why; watch it pass; commit.
