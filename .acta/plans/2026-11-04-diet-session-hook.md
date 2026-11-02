---
parent: specs/2026-11-04-skill-token-diet-design
depth: minimal
id: PLN-0211
created: "2026-11-04 19:45:47"
hash: a503tk7
---
# Diet: session start hook implementation plan

**Goal:** The session note lists each running plan once and leaves the brainstorm rule to shape.
**Spec:** `.acta/specs/2026-11-04-skill-token-diet-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only the packages they touch.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
The cap of 3 plans and the `+N more` line already exist in `runningBlock`; keep them.
The `herdr:` line (`herdrExtra`) stays: shape uses it to know a herdr tab is a way to move.

## Waves

Wave 1: Task 1, Task 2. They share no file.

### Task 1: each running plan once

**Files:** Modify `internal/board/state.go` and its test.
**Seam:** `board.Running(cfg)`.
**verify:** No plan id comes back twice, whatever worktrees carry it. Cases: two worktrees whose branches both touched the same plan file give one entry; distinct plans all come back; the kept entry is the worktree whose branch made the plan's newest commit. `acta state` and the session note both read `Running`, so both get the fix.
- [ ] Red: a board test with two worktrees on one plan gets two entries and fails.
- [ ] Change: drop later entries of the same plan in `Running`, run `scripts/test ./internal/board/ -run Running` to green.
- [ ] Commit: `fix(board): list each running plan once`.

### Task 2: brainstorm rule lives in shape only

**Files:** Modify `internal/hook/hook.go`, `internal/hook/hook_test.go`, `internal/plugincheck/brainstorm_continuity_test.go`.
**Seam:** `hook.SessionStart(hook.Input{...})`.
**verify:** The session note has no brainstorm rule, and shape still offers every move way. Cases: core rules hold no "Several Architectural brainstorms" text and are numbered 1 to 8 with no gap; shape still names `claude --bg`, a new session, and the herdr tab gated on the `herdr:` line; a herdr session still prints the `herdr:` line and a non-herdr one names no herdr in the rules.
- [ ] Red: a hook test expects no brainstorm rule and fails on today's rule 8.
- [ ] Change: drop rule 8, renumber, move the continuity checks to shape's text, run `scripts/test ./internal/hook/ ./internal/plugincheck/ -run 'Session|Brainstorm'` to green.
- [ ] Commit: `feat(hook): leave the brainstorm rule to shape`.
