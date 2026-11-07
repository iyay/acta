---
parent: specs/2026-11-07-stray-worktrees-design
depth: minimal
id: PLN-0216
created: "2026-11-07 17:36:26"
hash: wna91fm
after: [PLN-0215]
started: "2026-11-07 18:23:43"
finished: "2026-11-07 18:32:57"
---
# Stray worktrees: rules, land and tidy

**Goal:** Tell agents to remove the worktrees they make, make land and tidy clean up after themselves with `acta worktrees --clean`, and add the glossary term.

**Spec:** `.acta/specs/2026-11-07-stray-worktrees-design.md`

**Tests:** fast `scripts/test ./internal/hook ./internal/plugincheck`; full `scripts/test --full`. Evals: `scripts/eval --case <name>` for the cases this plan adds.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Skill text stays plain, short and free of repo paths and user names. The plugincheck size caps still pass.
- No text this plan adds tells an agent to use `--force`, `-f` on `worktree remove`, or `git branch -D` for a stray.
- The eval hooks call the `acta` on PATH, so put this branch's binary, with `acta worktrees` in it, on PATH before `scripts/eval`.

## Waves

- Wave 1: Task 1, Task 2

### Task 1: Core rule 10

**Files:** `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`
**Seam:** the session-start text from `internal/hook`, and `plugin/hooks/default-rules.md` that the fallback hook prints.
**verify:** Both the session-start text and `default-rules.md` hold rule 10 with the spec's wording, the two stay the same text, and no rule anywhere in them tells the agent to force a worktree remove.
- [x] Red: a hook test that looks for rule 10 fails because `coreRules` stops at rule 9.
- [x] Change: add rule 10 to `coreRules`, then regenerate `default-rules.md` with `go test ./internal/hook -run TestDefaultRulesFile -update`.
- [x] Commit: feat(hook): agents remove the worktrees they make

### Task 2: Land and tidy clean up worktrees

**Files:** `plugin/skills/land/SKILL.md`, `plugin/skills/land/tidy.md`, `plugin/evals/land-stray-worktrees/`, `plugin/evals/land-tidy-temp-worktree/`, `.acta/wiki/glossary.md`
**Seam:** `scripts/eval --case land-stray-worktrees` and `scripts/eval --case land-tidy-temp-worktree`.
**verify:** On every land path (full and tidy, green or stopped after the merge) the agent runs `acta worktrees --clean` from the main checkout once its own worktree and branch are gone, and the report lists what it removed and each dirty stray it kept. In tidy, the temp gate worktree is removed with a plain `git worktree remove` after the gate run, red or green. List each path checked. The glossary holds the approved `stray worktree` line.
- [x] Red: the two new eval cases fail because land and tidy do not run the cleanup.
- [x] Change: edit land step 7, tidy steps 3 and 7, add the two eval cases in the layout of the existing land cases, and add the glossary line.
- [x] Commit: feat(land): clean stray worktrees after landing

## Polish

### Task 3: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] [fix] move the stray worktree line in .acta/wiki/glossary.md after stamp, so the glossary stays in order.
- [x] Commit: `polish: review notes for PLN-0216`

## Review notes

- Review: deep (delete path, deep lens), round 1 CLEAN on both axes; polish (glossary order) light tier CLEAN. Cleanup moved to land step 10 on purpose. Notes left: no full-mode eval, temp-removed needs a literal --detach, report grader is loose, dirty-kept passes when no cleanup runs, land caps now tight.

## State

### Next

Build done: tasks 1-2 committed, evals red on old text and green on new. Next: acta:review over main..HEAD, then land.

### Findings

Eval grader traps went to the wiki page eval-sandbox-no-git-hooks.
The cleanup went into land SKILL.md step 10, not land step 7 and tidy step 7: tidy runs step 10 too, and one place fits the byte caps. Task 2 box text still names steps 7.
