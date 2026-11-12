---
parent: specs/2026-11-12-debug-evidence
depth: minimal
after: [PLN-0235]
id: PLN-0233
created: "2026-11-12 23:03:50"
hash: e12p7s9
started: "2026-11-13 00:59:44"
finished: "2026-11-13 01:11:03"
---
# Debug proof is required, and each text change is measured

**Goal:** The debug skill's rules no longer clash, its probes cannot write to the repo, its proof steps are required, and every text change is kept only when the eval cases do not get worse.

**Spec:** `.acta/specs/2026-11-12-debug-evidence.md` (SPC-0204)

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`; evals `scripts/eval --case <name> --runs 10 -j 8`.

**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched; whole when the diff touches `plugin/hooks/` or `plugin/references/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Before Task 1 changes any text, run each of the seven `debug-*` cases 10 times on the debug text as it is, and record the pass counts with `acta state set <plan id> findings`. That is the old arm for every task.
- The new text is measured once, after Task 5: every `debug-*` case, 10 runs each, against the old arm. Only when a case scores below the old arm is that case run again on each task's commit, to find the change that lowered it. A change found to lower a case is taken out with one commit that carries its task's trailer, and `acta state set <plan id> findings` says which.
- The `acta` on PATH must be the branch binary before any eval run. Build it to a folder of its own and put that folder first on PATH for the run; never overwrite the user's installed binary.
- No line may let an agent do less when a step fails. The wiki page `skill-escape-lines-skip-proof.md` records what such a line cost.
- Skill text stays general for every acta user, in plain English. When a file passes its byte cap, raise the cap in `internal/plugincheck/budget_test.go`; never trim words an agent acts on.
- The four phases and the race recipe keep their structure.

## Build brief

- Land order: this plan starts after PLN-0235 is on main, because both change `internal/plugincheck/budget_test.go` and `internal/plugincheck/evals_test.go`.
- Land notes: BUG-0074 gets its `fixed_in` when Task 1's change lands, and BUG-0068 when Task 2's change lands. A change that was taken out leaves its bug open.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3
- Wave 4: Task 4
- Wave 5: Task 5

### Task 1: Bisect runs in a temp clone

**Files:** `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the debug skill, and the cases `debug-misleading-path` and `debug-exact-symptom`
**verify:** No line of the debug skill asks for a step that changes the repo's working tree in Phases 1 to 3: every place that names bisect or an older commit says it runs in a clone, and shows `git clone . "$TMPDIR/<name>"` as a command to copy. List every line that names a git command that moves HEAD or the working tree, and where each one runs.
- [x] Red: `skill_debug_test.go` fails because the skill names `git bisect` with no temp clone.
- [x] Change: say in Phase 2 that bisect runs in a temp clone and show the clone command.
- [x] Commit: fix(debug): bisect in a temp clone, never in the repo

### Task 2: A probe folder that is checked before any write

**Files:** `plugin/skills/debug/SKILL.md`, `plugin/skills/debug/root-cause-tracing.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the debug skill, and the cases `debug-independent-check`, `debug-race` and `debug-order-dependent`
**verify:** No probe can write while the shell is still in the repo: the ground rules show one copyable pattern that makes the folder with `mkdir -p "$TMPDIR/<name>"`, changes into it, and checks `pwd` prints that folder before any command that writes. The text holds no fallback that allows less work when a step fails. Every other place that speaks of a temp copy uses the same pattern or points at it. List every line that tells an agent to write a file in Phases 1 to 3.
- [x] Red: `skill_debug_test.go` fails because the skill lacks the `mkdir -p "$TMPDIR/` pattern and the `pwd` check.
- [x] Change: add the checked-folder pattern to the ground rules and point the other temp-copy lines at it.
- [x] Commit: fix(debug): check the probe folder before a probe writes

### Task 3: Required control, a falsifying probe, and a four-line report

**Files:** `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the debug skill, and the cases `debug-independent-check`, `debug-order-dependent`, `debug-no-repro` and `debug-performance`
**verify:** No Phase 3 report can state a cause without a control or a written reason under Unverified for why the control could not run. Before the cause is stated, the result that would show it wrong is written down and that probe is run. Every Phase 3 report, proven or not, has the four lines Cause, Repro, Evidence and Unverified, and Repro is a command the user can paste. The words "when you can" no longer weaken the control. List every exit of Phase 3 (proven, not proven, cannot reproduce) and the lines each one prints.
- [x] Red: `skill_debug_test.go` fails because the control is still optional and the skill lacks the falsifying probe and the four report lines.
- [x] Change: make the control required, add the falsifying probe step and the four-line report to Phase 3.
- [x] Commit: feat(debug): required control, falsifying probe and a four-line report

### Task 4: What changed lately, and the same defect in other callers

**Files:** `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the debug skill, and the cases `debug-misleading-path` and `debug-exact-symptom`
**verify:** Phase 2 starts with what changed lately on every kind of input that can change: `git log` on the files involved, dependency versions, and the environment. Once a cause is proven, Phase 3 searches the other callers for the same pattern and the report lists what it found, including "none found" with the search that was run. Neither step writes to the repo. List where each step sits and what its output looks like in a report.
- [x] Red: `skill_debug_test.go` fails because the skill lacks the changed-lately step and the other-callers step.
- [x] Change: add the changed-lately step to Phase 2 and the other-callers step to Phase 3.
- [x] Commit: feat(debug): check recent changes first and look for the same defect elsewhere

### Task 5: A hang recipe, and "Three failed fixes" moves to Phase 4

**Files:** `plugin/skills/debug/SKILL.md`, `plugin/evals/debug-hang/`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** `TestEvalCases` and the plugincheck required strings for the debug skill, and the cases `debug-hang` and `debug-race`
**verify:** Every symptom the skill's description names has a recipe that says which evidence counts; for a hang that is a stack dump of the stuck process and a statement of who waits on whom, never a longer timeout. The `debug-hang` case fails a run that names a cause with no stack dump, has a `# guards:` line for the hang recipe, a scaffold that refuses a folder that is not empty, and caps on turns and time. "Three failed fixes" stands in Phase 4 and nowhere in Phase 3. List each symptom in the description and its recipe.
- [x] Red: `TestEvalCases` fails because the `debug-hang` folder is missing, and `skill_debug_test.go` fails because the skill has no hang recipe.
- [x] Change: add the hang recipe and the `debug-hang` case, and move "Three failed fixes" to Phase 4.
- [x] Commit: feat(debug): hang recipe with its eval case

## Fix round 1

### Task 6: A probe write shares one shell line with its folder check

**verify:** No probe can write while the shell stands in the repo, in a harness where every shell call starts in the repo again: the chain that makes the folder, steps into it and prints pwd is on the same line as the write, with a default folder when TMPDIR is unset or empty, and the same holds for the clone that bisect or an older commit runs in. Every exit of Phase 1 and Phase 3 reports in the four lines. List every line that tells an agent to write a file in Phases 1 to 3 and the chain it rides on.

- [x] BLOCKER debug SKILL.md ground rules: the checked-folder pattern proves the folder in one shell call, and the write runs in the next call, which starts in the repo again; proved by running the pattern and then pwd alone in a second call, which printed the repo. The write must sit in the same line as the mkdir, cd and pwd chain, and the folder needs a default when TMPDIR is unset or empty.
- [x] [fix] the bisect line chains clone, cd and bisect in one line, and an older commit is checked out only in such a clone
- [x] [fix] the Phase 1 cannot-reproduce exit reports in the four lines too
- [x] [fix] the other-callers list has a fixed place: after the four lines
- [x] [fix] the Phase 3 control line points at the checked folder and says to copy what the probe needs into it
- [x] [fix] race.md points its temp copy at the checked folder of the ground rules
- [x] [fix] the hang recipe says faulthandler must be set up before the hang and that sample is for macOS
- [x] [fix] debug-hang scaffold header comment describes the deadlock, not the race it was copied from
- [x] Commit with the trailer `Task: PLN-e12p7s9#6`

## Polish

### Task 7: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] [fix] ground rule: say the write is joined to the chain with &&, since a write after ; or on a second line still runs when the chain fails
- [x] [fix] bisect line: show the form for the next step in the clone, cd, pwd, then the git command
- [x] [fix] race.md: keep the word copy, the probe runs on a copy of the code inside the checked temp folder
- [x] [fix] Phase 1 step 2 and the control line say plainly which line they mean and how the copy gets there
- [x] Commit: `polish: review notes for PLN-0233`

## Fix round 2

### Task 8: The feedback loop runs in a temp copy again

**verify:** Phase 1 step 2 says in the old words that a command that writes files runs only in a temp copy, so the user own command is never run in the repo; debug-race over 20 runs on this text scores no lower than the old text. List every sentence in SKILL.md that holds the words temp copy.

- [x] BLOCKER debug SKILL.md Phase 1 step 2: debug-race scored 37 of 40 on the new text against 20 of 20 on the old, and 7 of 10 with traces kept. In the failing runs the agent ran the user own repro command in the repo, which wrote hits.json and __pycache__ there. The polish had replaced "runs only in a temp copy" with "runs only on a line that starts with the folder check", so the step no longer says the command runs on a copy. Take that sentence back to the old words.
- [x] Commit with the trailer `Task: PLN-e12p7s9#8`

## Review notes

- Deep, full tier (skill text). Round 1 BLOCKED with 1 BLOCKER, kept by the fact-check: the folder check and the write sat in two shell calls. Fix round 1, round 2 CLEAN, polish of 4 [fix] NOTEs reviewed CLEAN. The eval measure then showed debug-race lower on the polished text; fix round 2 put one sentence back and round 3 was CLEAN. BLOCKER 2, NOTE 52, pre-existing 9. 3 NOTEs went to debt (the copy line uses OLDPWD, writes that ride on no folder check, /tmp as a shared default); the rest are wording nits. A reviewer in round 2 left stray files in the main checkout; they were moved out, nothing tracked changed, and the reviewer wiki page holds the lesson.

## State

### Findings

Old arm 2026-11-13, debug text before this plan, 10 runs per case at -j 8: exact-symptom 9, independent-check 7, misleading-path 10, no-repro 10, order-dependent 10, performance 10, race 10. 66 of 70. Controls later added race 10/10 and no-repro 9/10 on the old text.
New text after Task 5: 80 of 80 over 8 cases (debug-hang is new). After fix round 1: 77 of 80. After the polish: 78 of 80, with debug-race at 9, 9, 9 and 7 of 10 in four runs of 10 against 20 of 20 on the old text.
Cause of the race drop, read from kept traces: the polish replaced "A command that writes files runs only in a temp copy." in Phase 1 step 2, and agents then ran the user own repro command in the repo, which wrote hits.json and __pycache__ there. Fix round 2 put the sentence back: debug-race 20 of 20, then 10 of 10.
Final text: 79 of 80 over 8 cases. The one red run is debug-no-repro, which scores 19 of 20 on the old text too and is a recorded flake. independent-check went from 7 to 10, exact-symptom from 9 to 10.
Read a failing run trace with scripts/eval --case <name> --runs 10 --keep-temp --json <file> before guessing at the cause; the json holds tracePath per run.
