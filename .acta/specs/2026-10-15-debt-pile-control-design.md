---
parent: scratch/2026-10-15-debt-pile-control
id: SPC-0135
created: "2026-10-15 10:22:39"
hash: z4ct5tl
started: "2026-10-15 10:27:58"
---
# Debt pile control: only real debt goes in, review notes stop piling up

Status: Architectural, approved by the user in chat on 2026-10-15. Full answers are in the Log of SCR-0071.

## Why

On 2026-10-15 the repo had 616 open debt items and 158 closed. 542 of the open items were `(low)`, 69 `(medium)`, 0 `(high)`, 5 had no tag. Plans also held 438 `## Review notes` lines in 69 files, and no commit ever acted on one. Both piles grow every review and nobody reads them.

## Part 1: review skill

File: `plugin/skills/review/SKILL.md`, `## Where findings go` and `## After a CLEAN round`.

- A `[debt]` NOTE must carry `(medium)` or `(high)`. A NOTE that would be `(low)` or has no priority falls to `[note]`.
- `[note]` NOTEs are no longer written line by line. The plan gets one summary line per review under `## Review notes`, for example `Round 1 CLEAN; 5 notes dropped.`
- Old `## Review notes` sections in existing plans stay as they are.

## Part 2: `acta debt new` guard

Files: `internal/write` (debt writer) and its tests in `internal/write/debt_test.go`; `acta debt` help text in `internal/cli`.

- Every non-empty stdin line must start with `(medium) ` or `(high) `.
- If any line fails, the whole batch is refused: nothing is written, nothing is committed, the exit code is non-zero, and the error names the first bad line number.
- The help text says the rule.

## Part 3: one-time cleanup

- Close every open debt item that is `(low)` or has no priority tag (547 on 2026-10-15) in one commit. Each one is marked dropped with the reason `low, under new debt bar`.
- Use an existing acta write path. If no bulk path exists, the plan adds the smallest flag that does it, with a test.
- After the cleanup, only `(medium)` and `(high)` items stay open.

## Out of scope

TUI changes, a cap on open debt, a doctor warning.

## Done when

- `acta debt new` refuses a batch with a `(low)` or untagged line, and the test proves nothing was written.
- The review skill text says debt takes only medium or high, and notes get one summary line.
- `acta list` shows no open `(low)` or untagged debt item.
- `scripts/test --full` passes.
