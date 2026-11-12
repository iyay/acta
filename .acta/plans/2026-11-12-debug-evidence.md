---
parent: specs/2026-11-12-debug-evidence
depth: minimal
after: [PLN-0235]
id: PLN-0233
created: "2026-11-12 23:03:50"
hash: e12p7s9
---
# Debug proof is required, and each text change is measured

**Goal:** The debug skill's rules no longer clash, its probes cannot write to the repo, its proof steps are required, and every text change is kept only when the eval cases do not get worse.

**Spec:** `.acta/specs/2026-11-12-debug-evidence.md` (SPC-0204)

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`; evals `scripts/eval --case <name> --runs 10 -j 8`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Before Task 1 changes any text, run each of the seven `debug-*` cases 10 times on the debug text as it is, and record the pass counts with `acta state set <plan id> findings`. That is the old arm for every task.
- Each task runs only the cases its verify line names, 10 runs, after the change. A change that scores below the old arm on any of those cases is taken out before the commit, and `acta state set <plan id> findings` says which. A task whose whole change is taken out ends with its boxes ticked and no code commit; say so in the report.
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
- [ ] Red: `skill_debug_test.go` fails because the skill names `git bisect` with no temp clone.
- [ ] Change: say in Phase 2 that bisect runs in a temp clone and show the clone command.
- [ ] Commit: fix(debug): bisect in a temp clone, never in the repo

### Task 2: A probe folder that is checked before any write

**Files:** `plugin/skills/debug/SKILL.md`, `plugin/skills/debug/root-cause-tracing.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the debug skill, and the cases `debug-independent-check`, `debug-race` and `debug-order-dependent`
**verify:** No probe can write while the shell is still in the repo: the ground rules show one copyable pattern that makes the folder with `mkdir -p "$TMPDIR/<name>"`, changes into it, and checks `pwd` prints that folder before any command that writes. The text holds no fallback that allows less work when a step fails. Every other place that speaks of a temp copy uses the same pattern or points at it. List every line that tells an agent to write a file in Phases 1 to 3.
- [ ] Red: `skill_debug_test.go` fails because the skill lacks the `mkdir -p "$TMPDIR/` pattern and the `pwd` check.
- [ ] Change: add the checked-folder pattern to the ground rules and point the other temp-copy lines at it.
- [ ] Commit: fix(debug): check the probe folder before a probe writes

### Task 3: Required control, a falsifying probe, and a four-line report

**Files:** `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the debug skill, and the cases `debug-independent-check`, `debug-order-dependent`, `debug-no-repro` and `debug-performance`
**verify:** No Phase 3 report can state a cause without a control or a written reason under Unverified for why the control could not run. Before the cause is stated, the result that would show it wrong is written down and that probe is run. Every Phase 3 report, proven or not, has the four lines Cause, Repro, Evidence and Unverified, and Repro is a command the user can paste. The words "when you can" no longer weaken the control. List every exit of Phase 3 (proven, not proven, cannot reproduce) and the lines each one prints.
- [ ] Red: `skill_debug_test.go` fails because the control is still optional and the skill lacks the falsifying probe and the four report lines.
- [ ] Change: make the control required, add the falsifying probe step and the four-line report to Phase 3.
- [ ] Commit: feat(debug): required control, falsifying probe and a four-line report

### Task 4: What changed lately, and the same defect in other callers

**Files:** `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the debug skill, and the cases `debug-misleading-path` and `debug-exact-symptom`
**verify:** Phase 2 starts with what changed lately on every kind of input that can change: `git log` on the files involved, dependency versions, and the environment. Once a cause is proven, Phase 3 searches the other callers for the same pattern and the report lists what it found, including "none found" with the search that was run. Neither step writes to the repo. List where each step sits and what its output looks like in a report.
- [ ] Red: `skill_debug_test.go` fails because the skill lacks the changed-lately step and the other-callers step.
- [ ] Change: add the changed-lately step to Phase 2 and the other-callers step to Phase 3.
- [ ] Commit: feat(debug): check recent changes first and look for the same defect elsewhere

### Task 5: A hang recipe, and "Three failed fixes" moves to Phase 4

**Files:** `plugin/skills/debug/SKILL.md`, `plugin/evals/debug-hang/`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** `TestEvalCases` and the plugincheck required strings for the debug skill, and the cases `debug-hang` and `debug-race`
**verify:** Every symptom the skill's description names has a recipe that says which evidence counts; for a hang that is a stack dump of the stuck process and a statement of who waits on whom, never a longer timeout. The `debug-hang` case fails a run that names a cause with no stack dump, has a `# guards:` line for the hang recipe, a scaffold that refuses a folder that is not empty, and caps on turns and time. "Three failed fixes" stands in Phase 4 and nowhere in Phase 3. List each symptom in the description and its recipe.
- [ ] Red: `TestEvalCases` fails because the `debug-hang` folder is missing, and `skill_debug_test.go` fails because the skill has no hang recipe.
- [ ] Change: add the hang recipe and the `debug-hang` case, and move "Three failed fixes" to Phase 4.
- [ ] Commit: feat(debug): hang recipe with its eval case
