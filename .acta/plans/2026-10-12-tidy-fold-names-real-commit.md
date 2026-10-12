---
parent: specs/2026-10-12-tidy-fold-names-real-commit
depth: minimal
id: PLN-0137
created: "2026-10-12 08:54:21"
hash: e4bdwzw
started: "2026-10-12 08:56:34"
finished: "2026-10-12 08:57:33"
---
# acta tidy names a full fold after its newest non-chore commit

**Goal:** When tidy keeps no commit of a branch, the one folded commit takes its message from the newest non-chore commit of the range, so a real change is not hidden behind a planning chore (BUG-0043).

**Spec:** `.acta/specs/2026-10-12-tidy-fold-names-real-commit.md`

**Tests:** `scripts/test ./internal/tidy`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Only `chore` is passed over when picking the source. `docs`, `wiki` and `polish` count as real. `foldTypes` does not change.
- Comments are plain English a 10-year-old reads back without stopping, and say why.

## Waves

- Wave 1: Task 1

### Task 1: Full fold takes its message from the newest non-chore commit

**Files:**
- Modify: `internal/tidy/tidy.go` (the `len(nodes) == 0` fallback at the end of `plan`)
- Test: `internal/tidy/tidy_test.go` (new test next to `TestNoKeptCommitMakesOneCommit`)

**verify:** When no commit is kept, the folded commit's full message (subject, body, trailers) is the message of the newest commit in the range whose subject type is not `chore`, and when every commit is a `chore` it is the newest commit's message, as before. On every path the folded tree still equals the branch tree, and runs with at least one kept commit are unchanged. List every case checked: the BUG-0043 repro (`docs(wiki): narrow paths` with a body and `Task: PLN-x#1`, then `chore(plan): tick and review notes`, then `chore(spec): mark started`), two non-chore commits followed by a chore (the newer non-chore wins), and the existing all-chore tests `TestNoKeptCommitMakesOneCommit` and `TestWikiCheckedKeptWhenNoCommitIsKept` passing unchanged.

- [x] **Step 1:** Write the failing test from the BUG-0043 repro: assert the tidy tip's `%B` equals the docs commit's `%B` and its tree equals `feat^{tree}`. Run `scripts/test ./internal/tidy -run TestNoKeptCommit`; it fails with the chore subject.
- [x] **Step 2:** In the fallback, walk `commits` from newest to oldest and take the first whose `subjectType` is not `chore`; keep `commits[len(commits)-1]` when none is found. Run the same command; it passes.
- [x] **Step 3:** Run `scripts/test ./internal/tidy`, `go vet ./internal/tidy`, `gofmt -l internal/tidy`, then commit with trailer `Task: <plan id>#1`.

## Review notes

- Round 1, full tier, deep level (setting), two reviewers on opus: both CLEAN, no BLOCKER, so no fact-check.
- The spec's `started`/`finished` stamp is left uncommitted by tick and is committed at land.
- No test pairs a mid-range `src` with a newer `chore(wiki): re-check` carrying `Wiki-Checked`; `withFoldedWikiChecks` adds those through `olds`, and the plan names no such case.
- Step 1 runs `-run TestNoKeptCommit`, which misses the new `TestFullFoldTakesNewestNonChoreMessage`; Step 3 runs the whole package.
- The folded commit's author and committer now come from the chosen non-chore commit; the date stays the newest in the range.
- `subjectType("chore!: x")` is `chore!`, so a breaking-change chore counts as real, the same as `foldTypes` treats it.
