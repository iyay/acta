---
parent: specs/2026-11-18-fix-two-flaky-evals
depth: minimal
id: PLN-0248
created: "2026-11-18 08:01:07"
hash: z708ok7
started: "2026-11-18 08:09:08"
finished: "2026-11-18 08:10:32"
---
# slice-split grader counts any read of many.md

**Goal:** `slice-split-parallel-set` passes when the agent opens `build/many.md` with Read or with `cat`, and fails when it never opens it.

**Spec:** `.acta/specs/2026-11-18-fix-two-flaky-evals.md` (SPC-0215), Fix 1
**Tests:** fast `scripts/test ./internal/plugincheck ./internal/evalomp`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched and `scripts/eval --case <name>` for one case; whole when the diff touches `plugin/hooks/` or `plugin/references/`. This plan touches only `plugin/evals/`, so the gate runs as `scripts/eval --case slice-split-parallel-set`, 3 times.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Change no other grader of the case, and no skill text.

## Waves

- Wave 1: Task 1

### Task 1: Trace regex in place of the Read-only grader

**Files:** `plugin/evals/slice-split-parallel-set/graders/reads-many-md.md`, `plugin/evals/slice-split-parallel-set/prompt.md`
**Seam:** `scripts/eval --case slice-split-parallel-set`, and `scripts/test ./internal/plugincheck ./internal/evalomp` for the case layout and the omp suite check
**verify:** The grader passes for every way a run can open `skills/build/many.md` (a Read call, `cat` or another Bash read) and fails for a run that never opens it. List each case checked: the kept failing trace `/private/tmp/e-emWJ1Y/out/trace.jsonl` (opened with `cat`), a kept passing trace that used Read, and a trace with no read of the file. The omp suite test accepts the case.
- [x] Red: the old pattern run over the kept `cat` trace finds no Read match, so the grader fails a run that did open the file.
- [x] Change: make the grader a `regex` on `target: trace` with pattern `skills/build/many\.md`, and add `claude-only` to the case tags.
- [x] Commit: test(evals): slice-split grader counts any read of many.md

## Review notes

- Review depth auto, picked standard (two reviewers): prompt.md frontmatter change, in doubt. Round 1 CLEAN. BLOCKER 0, NOTE 3, pre-existing 0. Notes: pattern needs the full path, so a cat of a relative many.md would miss; Findings list only the cat trace path; skill text must never hold the literal path.

## State

### Next

Task done; review then land.

### Findings

New grader, 3 runs: 3 of 3 green (2 opened many.md with Read, 1 inside a Bash command). Old Read-only grader on kept cat trace /private/tmp/e-emWJ1Y: 0 matches; new pattern matches it. Debug-case trace with no read: 0 matches.
