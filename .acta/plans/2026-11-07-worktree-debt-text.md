---
parent: debt/2026-11-07-stray-worktrees-rules
closes: [DBT-0157.01, DBT-0158.01, DBT-0158.02]
depth: minimal
id: PLN-0219
created: "2026-11-07 19:53:22"
hash: huxtbqn
---
# Worktree debt: skill text and evals

**Goal:** Close DBT-0157.01 and DBT-0158.01-.02: a refused tidy temp remove gets two clear choices, implementers report untracked test output, and the state-resume eval can go red.

**Spec:** `.acta/specs/2026-11-07-worktree-debt-design.md` (items 6 to 8)

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`. Evals: `scripts/eval --case <name>` for the cases this plan touches.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Skill text stays plain, short, and free of repo paths and user names. A byte cap in `internal/plugincheck/budget_test.go` rises only by the size of the new text.
- No text this plan adds tells an agent to force a remove or delete untracked files.
- Each eval change is shown red on the old text and green on the new, with the output kept in the task notes.

## Waves

- Wave 1: Task 1, Task 3
- Wave 2: Task 2

### Task 1: Refused tidy temp remove gets two choices

**Files:** `plugin/skills/land/tidy.md`, `plugin/evals/land-tidy-temp-refused/`, `internal/plugincheck/budget_test.go`
**Seam:** `scripts/eval --case land-tidy-temp-refused`
**verify:** When the plain remove of the tidy temp gate worktree refuses, the agent never forces and never deletes the files: it shows them (at most 10 lines) and offers the two choices, and the land goes on with step 4 after either. List each path checked: the user deletes the files and the plain remove runs again; the temp worktree stays and the land goes on.
- [ ] Red: the new case, where the gate leaves an untracked file in the temp worktree, fails on today's text because the agent forces or deletes.
- [ ] Change: add the two choices to tidy step 3 and the new case.
- [ ] Commit: fix(land): two choices when the tidy temp remove refuses

### Task 2: Implementers report untracked test output

**Files:** `plugin/skills/build/implementer-prompt.md`, `plugin/skills/build/SKILL.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** the plugincheck tests that read the build texts.
**verify:** Under every executor, untracked files a test or tool made are reported and never deleted, and the `.gitignore` offer has a place to land. List each path: inline (offer in the task commit, as today); subagent (implementer lists the files in its report, orchestrator offers a follow-up commit with the same `Task:` trailer after the user's yes).
- [ ] Red: a plugincheck test for the implementer line and the follow-up commit wording fails on today's text.
- [ ] Change: add the report line to the implementer prompt and the follow-up commit sentence to build step 3.
- [ ] Commit: fix(build): implementers report untracked test output

### Task 3: state-resume eval can go red

**Files:** `plugin/evals/state-resume/`
**Seam:** `scripts/eval --case state-resume`
**verify:** The case fails whenever the agent removes the test output, and the output only exists because the agent ran the plan's test command. Shown red with `plugin/skills/build/` from the commit before PLN-0207 landed and green on today's text.
- [ ] Red: with the old build text swapped in, the reworked case fails because the agent cleans the cache folder.
- [ ] Change: give the scaffold's plan a test command that writes the cache folder, make the resumed step run it, and stop the scaffold from making the folder itself.
- [ ] Commit: test(evals): state-resume runs the tests that leave output
