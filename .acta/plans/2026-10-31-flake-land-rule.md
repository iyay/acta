---
parent: specs/2026-10-31-flake-tracker-design
depth: minimal
closes: [SCR-0062, SCR-0079, DBT-0147.02, DBT-0145.03]
id: PLN-0203
created: "2026-10-31 17:59:57"
hash: pvlb4at
after: [PLN-0204]
---
# Flake land rule implementation plan

**Goal:** Teach land, build and debug to record FLAKY lines and rerun known flakes alone, rerun red eval cases alone, and seed bugs for the known flakes.
**Spec:** `.acta/specs/2026-10-31-flake-tracker-design.md` (Section 3, Seed and cleanup)
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only their owning packages; each new eval case runs once red and once green with `scripts/eval --case <name>`.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Skill prose stays plain, short, and free of repo paths and user names, and fits the plugincheck size caps.
Every command the skills name exists: `acta flake record <name> --plan <plan id>` from the flake record plan, `scripts/eval --case <name>`.
A red full suite (exit 1) keeps today's rule; a repo whose test command prints no FLAKY line sees no change.

## Waves

Wave 1: Task 1, Task 3. Wave 2: Task 2. Wave 3: Task 4.

### Task 1: land records FLAKY lines

**Files:** Modify `plugin/skills/land/SKILL.md`, `plugin/skills/land/baseline.md`; create `plugin/evals/land-records-flake/case.yaml`, `plugin/evals/land-records-flake/prompt.md`, `plugin/evals/land-records-flake/scaffold.sh`, `plugin/evals/land-records-flake/graders/records-flake.md`; modify `internal/plugincheck/evals_test.go`.
**Seam:** The land skill text, and a greenfield eval case run by `scripts/eval`.
**verify:** After a full suite that exits 0, each FLAKY line in its log gets one `acta flake record <name> --plan <plan id>` call in the checkout where the suite ran, before the merge, and the land report names each flake and its bug; no rule lets a FLAKY line pass without that call. The case's scaffold suite prints one FLAKY line, and the grader checks the record call comes before `git merge`.
- [ ] Red: the evals layout test fails because the new case is registered but its files do not exist.
- [ ] Change: Add the rule to land step 1 and `baseline.md`, add the case files, run `scripts/test ./internal/plugincheck/` to green, then the case once with the branch binary on PATH.
- [ ] Commit: `feat(flake): land records each FLAKY line`.

### Task 2: red eval case reruns alone at the land gate

**Files:** Modify `plugin/skills/land/SKILL.md`; create `plugin/evals/land-eval-flake/case.yaml`, `plugin/evals/land-eval-flake/prompt.md`, `plugin/evals/land-eval-flake/scaffold.sh`, `plugin/evals/land-eval-flake/graders/eval-flake.md`; modify `internal/plugincheck/evals_test.go`.
**Seam:** The land skill's eval gate text, and a greenfield eval case whose fake `scripts/eval` fails one case once.
**verify:** A red eval case at the gate reruns alone with `--case <name>` up to 2 times; a pass runs `acta flake record eval:<case> --plan <plan id>` and the land goes on; red 3 of 3 stops the land as today. No path reruns the whole eval set for one red case.
- [ ] Red: the evals layout test fails because the new case is registered but its files do not exist.
- [ ] Change: Rewrite the eval gate sentence in land step 1, add the case files, run `scripts/test ./internal/plugincheck/` to green, then the case once with the branch binary on PATH.
- [ ] Commit: `feat(flake): land reruns a red eval case alone`.

### Task 3: build and debug rerun a known flake once

**Files:** Modify `plugin/skills/debug/SKILL.md`, `plugin/skills/build/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/skill_build_test.go`.
**Seam:** The debug and build skill text, read by plugincheck.
**verify:** In build and debug, a failing test named in an open bug's `flaky_tests` runs again alone once before any debugging, and a pass runs `acta flake record`; a test in no flake bug follows today's rules. Both skills stay within their size caps.
- [ ] Red: a plugincheck must-hold test for the flake rerun sentence fails because neither skill has it.
- [ ] Change: Add one sentence to each skill and the must-hold check, then run `scripts/test ./internal/plugincheck/` to green.
- [ ] Commit: `feat(flake): build and debug rerun a known flake once`.

### Task 4: seed the known flakes

**Files:** Create bug files under `.acta/bugs/` through `acta flake record` and `acta bug new`; modify `.acta/scratch/2026-10-31-flake-tracker.md`.
**Seam:** `acta list` and `acta show` on the new bugs.
**verify:** Each known flake has one open bug with its `flaky_tests` name: `internal/tui.TestClickOnTheBarTakesTheHighlightOff`, `internal/tui.TestWheelNotchesBelongToTheScreenTheyWereGatheredOn`, `eval:task-trailer`, `eval:brainstorm-mixed-context`, `eval:brainstorm-resume-topic`, `eval:debug-race`. The real red `answers-appended` (red on main) has its own non-flake bug that also names DBT-0139.01 for `build-close-no-suite`. No other bug changes.
- [ ] Red: `acta list --json` shows no bug whose `flaky_tests` names these tests.
- [ ] Change: Run `acta flake record` for each name and `acta bug new` for the real red, from the worktree, then check `acta list --json`.
- [ ] Commit: the `acta` commands commit each bug file; no extra commit.

## Build brief

Land order: PLN-0204 and PLN-0205 land on main in either order; they share no file. PLN-0203 lands after PLN-0204.
