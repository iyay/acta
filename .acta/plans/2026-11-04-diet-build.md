---
parent: specs/2026-11-04-skill-token-diet-design
depth: minimal
id: PLN-0209
created: "2026-11-04 19:45:47"
hash: boe2a2o
started: "2026-11-04 19:49:24"
finished: "2026-11-04 21:05:39"
---
# Diet: build skill implementation plan

**Goal:** Build loads only the text of the executor it runs.
**Spec:** `.acta/specs/2026-11-04-skill-token-diet-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only `./internal/plugincheck/`.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Skill frontmatter descriptions do not change.
Executor order stays: the `/build <executor>` argument, then `build_executor` from `acta config show`, then `inline`.
`skills/build/SKILL.md` byte cap is its real size after the move and tightening; the user ruled on 2026-11-04 that 7000 cannot be met without dropping rules.

## Waves

Wave 1: Task 1.

### Task 1: one file per executor

**Files:** Create `plugin/skills/build/subagent.md`, `plugin/skills/build/inline.md`; modify `plugin/skills/build/SKILL.md`, `plugin/skills/build/dispatch.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/build_dispatch_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the build skill folder as plugincheck reads it, plus `scripts/eval --case 'build-*'`.
**verify:** Each executor gets every rule it has today and no text of another executor. Cases: every executor-only sentence of today's body (Models, Inline task loop, Worker hand-off, the per-executor fix round) sits in exactly one mode file; text all executors use (worktree, plan, task loop core, close) stays in the body; the body's executor table points each row at its file and says to read only that one; existing skill_build and build_dispatch assertions still find their text, now in the mode file; byte caps hold; `build-*` evals give the same results on branch and main, and the inline-run transcript has no Read of `subagent.md` or `dispatch.md`.
- [x] Red: a plugincheck test expects `subagent.md` and `inline.md` and a body under 7000 bytes, and fails because neither file exists.
- [x] Change: move the executor text, point the table, retarget the moved assertions, set caps, run `scripts/test ./internal/plugincheck/` then the evals on branch and main.
- [x] Commit: `feat(build): load only the running executor's text`.

## Fix round 1

### Task 2: Restore cut build rules

**verify:** Every rule the base build body held is reachable on its path: wiki read-and-fix at every close, no ask on executor pick, plan changes in the worktree, no install when checks already run; each pinned

- [x] build Close dropped 'Read affected pages and diffs; fix wrong facts' as an always-on step, so a stale-only wiki check skips the read; put it back in the body and pin it in TestBuildCloseWritesWikiFirst
- [x] [fix] restore 'Never ask' on the executor pick order and the omp and HERDR fallbacks
- [x] [fix] restore 'A plan change goes in the worktree.'
- [x] [fix] restore 'and nothing when they already run' in Setup
- [x] Commit with the trailer `Task: PLN-boe2a2o#2`

## Polish

### Task 3: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] fix the stale doc comment of TestBuildCloseWritesWikiFirst: SKILL.md Close now also carries 'Read affected pages and diffs; fix wrong facts.'
- [x] build SKILL.md omp fallback: 'runs as `subagent` and do not ask.' becomes 'runs as `subagent`. Do not ask.'; update its pin
- [x] Commit: `polish: review notes for PLN-0209`

## Fix round 2

### Task 4: Stop rule pointer in inline.md

**verify:** The inline executor file names the stop rule for a failing tick, a failing state set and a read-back not done; build-tick-failure passes on the branch

- [x] eval build-tick-failure passes 3 of 5 on the branch and 3 of 3 on main; inline.md never points at the stop rule, so add one line under its Inline task loop: a tick or acta state set that exits non-zero, or a read-back not done, stops the build (SKILL.md Task loop); pin it
- [x] Commit with the trailer `Task: PLN-boe2a2o#4`

## Review notes

- Deep depth, full tier; Spec axis opus, Standards axis sonnet. Round 1 BLOCKED (wiki read-and-fix step lost from Close, confirmed by fact-check), fix round restored it plus three cut rules, round 2 CLEAN. Polish fixed a stale test comment and the omp fallback grammar; polish review CLEAN. Size: body 10409 bytes, user ruled 7000 cannot be met without dropping rules. One [debt]: folder-wide pins. Notes [note]: table lost its when-to-read column, covered by fallback lines.

## State

### Next

Paused 2026-11-04 by user. Review CLEAN, polish done, debt filed, full suite green at e5e1485 (4668 pass).
Next: run scripts/eval (plugin/skills changed), rerun red cases alone up to 2 times, then land per tidy.md.
