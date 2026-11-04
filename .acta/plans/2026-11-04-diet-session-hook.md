---
parent: specs/2026-11-04-skill-token-diet-design
depth: minimal
id: PLN-0211
created: "2026-11-04 19:45:47"
hash: a503tk7
started: "2026-11-04 19:48:42"
finished: "2026-11-04 19:50:01"
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
- [x] Red: a board test with two worktrees on one plan gets two entries and fails.
- [x] Change: drop later entries of the same plan in `Running`, run `scripts/test ./internal/board/ -run Running` to green.
- [x] Commit: `fix(board): list each running plan once`.

### Task 2: brainstorm rule lives in shape only

**Files:** Modify `internal/hook/hook.go`, `internal/hook/hook_test.go`, `internal/plugincheck/brainstorm_continuity_test.go`.
**Seam:** `hook.SessionStart(hook.Input{...})`.
**verify:** The session note has no brainstorm rule, and shape still offers every move way. Cases: core rules hold no "Several Architectural brainstorms" text and are numbered 1 to 8 with no gap; shape still names `claude --bg`, a new session, and the herdr tab gated on the `herdr:` line; a herdr session still prints the `herdr:` line and a non-herdr one names no herdr in the rules.
- [x] Red: a hook test expects no brainstorm rule and fails on today's rule 8.
- [x] Change: drop rule 8, renumber, move the continuity checks to shape's text, run `scripts/test ./internal/hook/ ./internal/plugincheck/ -run 'Session|Brainstorm'` to green.
- [x] Commit: `feat(hook): leave the brainstorm rule to shape`.

## Fix round 1

### Task 3: Short brainstorm move rule in the session note

**verify:** An agent that cannot load shape still moves a brainstorm right: log the talk to the item, give its real id and a way this harness has (new session typing brainstorm <id>; Claude Code only claude --bg), launch nothing; the case types stay out of the note; no herdr word outside herdr

- [x] eval brainstorm-user-handoff fails 3 of 3: with no tools the agent cannot load shape, and the session note no longer says how to move a brainstorm; add back a short move-only core rule
- [x] Commit with the trailer `Task: PLN-a503tk7#3`

## Polish

### Task 4: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] hook_test.go doc comment of TestSessionStartLeavesBrainstormRuleToShape says rules count 1 to 8 and the note carries no brainstorm rule; make it say 1 to 9 and that only the short move rule stays
- [x] Commit: `polish: review notes for PLN-0211`

## Review notes

- Round 1 CLEAN at deep depth (full tier; Spec axis opus, Standards axis sonnet). No BLOCKER, so no fact-check. 9 notes, all [note]: dedup also covers PlanStates, which helps every caller; NewestCommitTime is tested only through board; rule 8 saves about 150 tokens, not 400; brainstorm evals now rely on shape being loaded and should run at land.
- Round 1 deep, full tier: CLEAN. Land eval gate then failed brainstorm-user-handoff 3 of 3 (agent with no tools lost the move rule); fix round 1 put back a short move-only rule 8; round 2 CLEAN (opus Spec, sonnet Standards). Polish fixed a stale test comment, light tier, reviewed inline CLEAN. Notes [note]: rule 8 names no log command; sessionStartCap now loose; dispatch executor text names herdr outside herdr (old).

## State

### Next

Paused 2026-11-04 by user. Review CLEAN, polish done, full suite green at 008f1b8 (4670 pass).
Land blocked by evals: brainstorm-user-handoff PASS FAIL FAIL and brainstorm-mixed-context FAIL x3 with the short rule 8.
Next: rerun both with scripts/eval --case <name> --keep-temp using a branch-built acta on PATH, read why the judge fails, run mixed-context on main as control, then offer the user: full rule 8 back, or reworded short rule.
