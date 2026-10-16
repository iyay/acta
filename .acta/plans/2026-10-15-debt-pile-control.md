---
parent: specs/2026-10-15-debt-pile-control-design
depth: minimal
closes: [SCR-0071]
id: PLN-0145
created: "2026-10-15 10:26:30"
hash: gk2cr63
started: "2026-10-15 10:27:58"
status: done
finished: "2026-10-16 12:21:07"
---
# Debt pile control Implementation Plan

**Goal:** Only medium or high review NOTEs become debt, review notes shrink to one summary line, and the existing low and untagged debt is closed once.

**Spec:** .acta/specs/2026-10-15-debt-pile-control-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Accepted priority prefixes are exactly `(medium) ` and `(high) `. Anything else counts as low.
- The cleanup reason text is exactly `low, under new debt bar`.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3

### Task 1: `acta debt new` refuses low and untagged lines

**Files:** `internal/write/ops.go` (`NewDebt`), `internal/write/debt_test.go`, `internal/cli` debt help text and its test

**verify:** No batch that holds a line without a `(medium) ` or `(high) ` prefix writes or commits anything, on both the new-file and the append path; the error names the first bad line number. List every path checked, including a `(low) ` line, an untagged line, a wrong-case `(Medium) ` line, and a batch whose bad line comes last.

### Task 2: review skill sends only medium and high NOTEs to debt

**Files:** `plugin/skills/review/SKILL.md`, the plugincheck or skill test that pins its rules

**verify:** No place in the review skill still tells an agent to file a `(low)` or untagged NOTE as debt, or to write `[note]` NOTEs line by line into the plan; the plan gets one summary line per review instead. List every section checked.

### Task 3: close every open low and untagged debt item in one commit

**Files:** `internal/write` (new bulk wontfix over debt lines), `internal/cli` (`acta debt drop-low`), their tests, then the `.acta/debt/` files

**verify:** After `acta debt drop-low` runs, no open debt line without a `(medium) ` or `(high) ` prefix is left, no medium or high line changes, and the whole change lands in one commit whose message carries `low, under new debt bar`. A second run changes nothing and makes no commit. List every path checked.

## Fix round 1

### Task 4: Round 1 findings

**Files:** `plugin/skills/land/SKILL.md`, `internal/plugincheck/skill_land_test.go`, `internal/write/droplow.go`, `internal/write/droplow_test.go`

**verify:** Land's state gate no longer sends untagged Findings to `acta debt new`; drop-low never changes a box inside a code fence.

## Review notes

- Deep, full tier. Round 1 CLEAN on both axes; land state-gate finding raised to a BLOCKER, fixed in round 1 with the fence [fix]. Round 2 CLEAN on both axes; 11 low notes dropped.
