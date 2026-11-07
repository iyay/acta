---
parent: specs/2026-10-31-block-destructive-commands-design
depth: minimal
closes: [SCR-0076]
id: PLN-0207
created: "2026-10-31 20:24:08"
hash: n78af12
started: "2026-10-31 20:25:45"
finished: "2026-10-31 20:30:35"
---
# Test output and cleanup skill text implementation plan

**Goal:** Teach build, land and tidy what to do with untracked test output and a refused `git worktree remove`, ban amend at land, and prove it with the state-resume eval.
**Spec:** `.acta/specs/2026-10-31-block-destructive-commands-design.md` (The skills say what to do with test output, Testing)
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only their owning packages; the eval case runs once with `scripts/eval --case state-resume`.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Skill prose stays plain, short, and free of repo paths and user names, and fits the plugincheck size caps; the user ruled that caps in `internal/plugincheck/budget_test.go` rise to the exact size the text needs.
Untracked files a test or tool made are never deleted by the agent. The orchestrator, never an implementer subagent, asks the user. The three land choices are: add a `.gitignore` line, the user removes them, or stop the land.

## Waves

Wave 1: Task 1. Wave 2: Task 2.

### Task 1: build, land and tidy text

**Files:** Modify `plugin/skills/build/SKILL.md`, `plugin/skills/land/SKILL.md`, `plugin/skills/land/tidy.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/budget_test.go`, `.acta/wiki/omp-report-file.md`.
**Seam:** The skill text read by plugincheck.
**verify:** Build step 3 says untracked test or tool output is not uncommitted work, is never deleted, and the orchestrator offers one `.gitignore` line after the user's yes. Land's clean-worktree precondition gives the three choices and never picks; a `.gitignore` commit at land reruns the gates for that file. Land step 7 and tidy step 7 stop on a refused `git worktree remove`, show `git status --porcelain --ignored`, and ask with the same choices; neither says `--force` or deletes first. Land never amends. The wiki page says moving acta's own omp report is not a delete, the empty folder goes with `rmdir`, and other untracked files follow the rule. Each new sentence has a plugincheck guard that goes red when it is removed.
- [x] Red: the new plugincheck guards fail because the sentences are missing.
- [x] Change: Write the sentences and the wiki page, raise the caps to the exact sizes, then run `scripts/test ./internal/plugincheck/` to green.
- [x] Commit: `feat(land): keep test output and stop on a refused worktree remove`.

### Task 2: state-resume keeps the test folder

**Files:** Create `plugin/evals/state-resume/graders/keeps-test-output.md`; modify `internal/plugincheck/evals_test.go` only if the new grader must be registered there.
**Seam:** The state-resume eval case run by `scripts/eval`.
**verify:** The grader fails when `src/__pycache__` is gone at the end of a run, whatever command removed it, and passes when it is still there; it is a file-presence or regex grader, not an llm grader.
- [x] Red: plugincheck or a hand check shows the case has no grader for the test folder.
- [x] Change: Add the grader, run `scripts/test ./internal/plugincheck/` to green, then `scripts/eval --case state-resume` once with the branch binary first on PATH.
- [x] Commit: `test(land): state-resume keeps the test folder`.

## Build brief

Land order: PLN-0206 and PLN-0207 land on main in either order; they share no file.

## Fix round 1

### Task 3: After the merge a refused worktree remove offers no gitignore commit

**verify:** No path after the parent moved can add a commit to the branch: land step 7 and tidy step 7 offer only two choices on a refused git worktree remove (the user removes the files, or stop the land), the gitignore choice stays only at land step 1 before the merge, the tidy branch -D exemption still holds, and the spec says the same.

- [x] plugin/skills/land/SKILL.md step 7 and plugin/skills/land/tidy.md step 7: a refused git worktree remove points to the three choices of step 1, so picking the gitignore line adds a commit to the branch after the merge; in tidy mode the following git branch -D drops it silently. Offer only two choices there: the user removes the files, or stop the land. Keep the gitignore choice at step 1 only. Update the plugincheck guards to match.
- [x] .acta/specs/2026-10-31-block-destructive-commands-design.md: the land step 7 / tidy step 7 bullet says "ask with the same three choices"; change it to the two choices, per the user's ruling.
- [x] Commit with the trailer `Task: PLN-n78af12#3`

## Review notes

- Review: deep level, round 1 BLOCKED (step 7 offered a .gitignore commit after the merge that tidy branch -D dropped; the user ruled two choices there), fix round 1, round 2 CLEAN, no polish. Notes kept, not tracked: the wiki page omp-report-file has no guard and its paths do not match land; the .gitignore commit at land comes after the CLEAN review range; scaffold.sh comment should give the pytest tag as the reason; scaffold.sh was not named in the plan; step 7 does not say to retry the remove after the user cleans up; stopping at step 7 skips steps 8-9; the spec commit bbcfae3 has no trailer.

## State

### Next

Land resumed 2026-11-07: main merged into the branch (22fbf40), conflicts resolved by intent and reviewed CLEAN.
The old tidy tip 31f2e14 and its temp worktree are gone; run acta tidy main destructive-skill again on parent a79c884, then the gates, then tidy steps 4-7.

### Open rulings

User ruled A (2026-10-31 20:39, WhatsApp): at land step 7 and tidy step 7 a refused git worktree remove offers only two choices, the user removes the files or stop the land; the .gitignore choice is offered only before the merge (land step 1). The spec changes the same way.
