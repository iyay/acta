---
parent: specs/2026-10-31-flake-tracker-design
depth: minimal
closes: [SCR-0062, SCR-0079, DBT-0147.02, DBT-0145.03]
id: PLN-0203
created: "2026-10-31 17:59:57"
hash: pvlb4at
after: [PLN-0204]
started: "2026-10-31 18:36:37"
finished: "2026-10-31 18:52:54"
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
- [x] Red: the evals layout test fails because the new case is registered but its files do not exist.
- [x] Change: Add the rule to land step 1 and `baseline.md`, add the case files, run `scripts/test ./internal/plugincheck/` to green, then the case once with the branch binary on PATH.
- [x] Commit: `feat(flake): land records each FLAKY line`.

### Task 2: red eval case reruns alone at the land gate

**Files:** Modify `plugin/skills/land/SKILL.md`; create `plugin/evals/land-eval-flake/case.yaml`, `plugin/evals/land-eval-flake/prompt.md`, `plugin/evals/land-eval-flake/scaffold.sh`, `plugin/evals/land-eval-flake/graders/eval-flake.md`; modify `internal/plugincheck/evals_test.go`.
**Seam:** The land skill's eval gate text, and a greenfield eval case whose fake `scripts/eval` fails one case once.
**verify:** A red eval case at the gate reruns alone with `--case <name>` up to 2 times; a pass runs `acta flake record eval:<case> --plan <plan id>` and the land goes on; red 3 of 3 stops the land as today. No path reruns the whole eval set for one red case.
- [x] Red: the evals layout test fails because the new case is registered but its files do not exist.
- [x] Change: Rewrite the eval gate sentence in land step 1, add the case files, run `scripts/test ./internal/plugincheck/` to green, then the case once with the branch binary on PATH.
- [x] Commit: `feat(flake): land reruns a red eval case alone`.

### Task 3: build and debug rerun a known flake once

**Files:** Modify `plugin/skills/debug/SKILL.md`, `plugin/skills/build/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/skill_build_test.go`.
**Seam:** The debug and build skill text, read by plugincheck.
**verify:** In build and debug, a failing test named in an open bug's `flaky_tests` runs again alone once before any debugging, and a pass runs `acta flake record`; a test in no flake bug follows today's rules. Both skills stay within their size caps.
- [x] Red: a plugincheck must-hold test for the flake rerun sentence fails because neither skill has it.
- [x] Change: Add one sentence to each skill and the must-hold check, then run `scripts/test ./internal/plugincheck/` to green.
- [x] Commit: `feat(flake): build and debug rerun a known flake once`.

### Task 4: seed the known flakes

**Files:** Create bug files under `.acta/bugs/` through `acta flake record` and `acta bug new`; modify `.acta/scratch/2026-10-31-flake-tracker.md`.
**Seam:** `acta list` and `acta show` on the new bugs.
**verify:** Each known flake has one open bug with its `flaky_tests` name: `internal/tui.TestClickOnTheBarTakesTheHighlightOff`, `internal/tui.TestWheelNotchesBelongToTheScreenTheyWereGatheredOn`, `eval:task-trailer`, `eval:brainstorm-mixed-context`, `eval:brainstorm-resume-topic`, `eval:debug-race`. The real red `answers-appended` (red on main) has its own non-flake bug that also names DBT-0139.01 for `build-close-no-suite`. No other bug changes.
- [x] Red: `acta list --json` shows no bug whose `flaky_tests` names these tests.
- [x] Change: Run `acta flake record` for each name and `acta bug new` for the real red, from the worktree, then check `acta list --json`.
- [x] Commit: the `acta` commands commit each bug file; no extra commit.

## Build brief

Land order: PLN-0204 and PLN-0205 land on main in either order; they share no file. PLN-0203 lands after PLN-0204.

## Fix round 1

### Task 5: Flake records land in a committed checkout, and a tracked flake can still be fixed

**verify:** Every FLAKY record a land makes ends up committed on the parent: step 1 records in the plan worktree before the merge, a step-6 rerun records in the main checkout after the merge, a tidy-tip rerun records in the plan worktree and then runs acta tidy again, and a record that exits non-zero stops the land; the build and debug rerun never stops work on a test that the current bug or plan task is fixing.

- [x] plugin/skills/land/baseline.md step 4: a tidy-tip rerun runs in a detached temp worktree, where acta flake record cannot commit (exit 2, untracked bug file, report names a BUG main never gets). Name the checkout per run: plan worktree before the merge for step 1; main checkout after the merge for a step-6 rerun; plan worktree then acta tidy again for a tidy-tip rerun. A record that exits non-zero stops the land.
- [x] plugin/skills/debug/SKILL.md and plugin/skills/build/SKILL.md: the known-flake rerun says a pass "needs no debugging", so asking to fix a tracked flake (for example BUG-0058) records one more flake and stops. Limit the rule to failures met while doing other work; a test the current bug or task is fixing is not covered.
- [x] plugin/skills/debug/SKILL.md: pass --plan only when there is a plan.
- [x] plugin/skills/land/baseline.md: capture the tested commit and tree right after the run, before step 4 commits a bug file, so the record names what was tested.
- [x] plugin/evals/land-records-flake/graders/records-flake.md: do not eat the newline after the record line, so a merge line right after it still matches; accept git merge --no-ff -m "<msg>" <branch> like land-full-progress does.
- [x] plugin/evals/land-eval-flake/graders/eval-flake.md: refuse 3 or more lone reruns of the case before the pass, and a lone rerun of a different case.
- [x] internal/plugincheck: guard the land SKILL.md step-1 FLAKY sentence, the report clause "each flake with its bug", "never rerun the whole set" and "Red all 3 times stops the land", and the new checkout rule.
- [x] Commit with the trailer `Task: PLN-pvlb4at#5`

## Polish

### Task 6: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] internal/plugincheck/skill_land_test.go: guard the land SKILL.md step-1 eval sentence "in the checkout [baseline.md](baseline.md) step 4 names" with a want in TestLandRecordsFlakes, so a revert of SKILL.md alone goes red.
- [x] plugin/evals/land-baseline-pass/graders/baseline-identity.md: update the comment to say the commit is the tested commit step 2 printed, not git rev-parse HEAD.
- [x] Commit: `polish: review notes for PLN-0203`

## Review notes

- Review: deep level, no floor raised; round 1 BLOCKED (tidy-tip record lost in a detached worktree; the flake rule stopped work on a tracked flake; fact-check kept both), fix round 1, round 2 CLEAN, polish review CLEAN. Notes kept, not tracked: build then debug can rerun a known flake alone twice; the seed bugs folded into one commit and carry a no-plan Flakes line dated today; no red-first eval run on record; the eval sentence says before the merge though a step-6 eval rerun records after it; the tidy reuse clause names FLAKY lines, not eval records; auto_commit off lets a record exit 0 uncommitted; the eval-flake grader accepts a record in main before the merge.

## State

### Next

Fix round 1 (Task 5) is committed; review round 2 over the fix commit is next. land-eval-flake went green, then red once after a wording change that was later reverted, so its final text has no eval run yet.

### Findings

Task 3: build/SKILL.md had 2 bytes under its byte cap (12649 of 12651) and debug/SKILL.md 65 (6166 of 6231).
The flake rerun sentence needs about 210 bytes in each, so budget_test.go caps must rise or other skill text must be cut.

### Open rulings

User ruled A (2026-10-31 18:40, WhatsApp): raise the byte caps in internal/plugincheck/budget_test.go as far as each task needs, for land/SKILL.md, land/baseline.md, build/SKILL.md and debug/SKILL.md, inside each task's own commit.
