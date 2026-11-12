---
parent: specs/2026-11-12-land-gates-scoped
depth: minimal
after: [PLN-0231]
id: PLN-0236
created: "2026-11-12 23:14:21"
hash: wo216ep
started: "2026-11-12 23:30:45"
---
# Land runs the gates a diff needs, in any repo

**Goal:** A plan names its slow gates, land runs each one only as far as the diff calls for, a moved parent no longer repeats a slow gate for nothing, and no skill text names one repo's own gate.

**Spec:** `.acta/specs/2026-11-12-land-gates-scoped.md` (SPC-0207)

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`; evals `scripts/eval --case <name> --runs 10 -j 8`.

**Gates:** evals: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag <skill folder>` for each skill folder touched; whole when the diff touches `plugin/hooks/` or `plugin/references/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Skill text holds for every repo that uses acta. No line of the slice or land skill names this repo's eval script, its folders or its tags; those live in `CLAUDE.md`.
- No line may let land skip a gate because a line is missing, a path is unclear or a step failed: each of those means the whole gate.
- The full suite at land, the `run-one` lock and the flake recording rules stay as they are.
- Evals are measured once, after Task 3, on `land-gate-narrow` and every case tagged `land`, 10 runs each, against 10 runs on the text before this plan. A case that scores lower sends its change back out with one commit, and `acta state set <plan id> findings` says which.
- The `acta` on PATH must be the branch binary before any eval run. Build it to a folder of its own and put that folder first on PATH for the run; never overwrite the user's installed binary.
- When a file passes its byte cap, raise the cap in `internal/plugincheck/budget_test.go`; never trim words an agent acts on.

## Build brief

- Land order: this plan starts after PLN-0231 is on main, because both change `plugin/skills/land/SKILL.md` and the plugincheck files.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3, Task 4

### Task 1: A plan names its slow gates

**Files:** `plugin/skills/slice/SKILL.md`, `plugin/skills/slice/minimal.md`, `plugin/skills/slice/full.md`, `internal/plugincheck/skill_slice_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the slice skill
**verify:** Every plan for a repo whose docs name a slow gate carries a `**Gates:**` line, at either depth, with one sentence per gate holding the command, the paths that make it run, the narrow form when the gate's tool has one, and the paths that force the whole gate. A repo with no slow gate gets no line and no mention. The text says where the facts come from (the repo's own docs) and that the yes to the plan is the yes to its gates. No word in it names this repo's eval script or folders. List both depths and what each header holds.
- [x] Red: `skill_slice_test.go` fails because the slice text has no `**Gates:**` line and no rule for where its facts come from.
- [x] Change: add the `**Gates:**` rule to the slice skill and name the line in both depth files.
- [x] Commit: feat(slice): a plan names its slow gates

### Task 2: Land runs each gate as far as the diff calls for

**Files:** `plugin/skills/land/SKILL.md`, `plugin/evals/land-eval-flake/`, `plugin/evals/land-gate-narrow/`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the land skill, `TestEvalCases`, and the cases `land-gate-narrow` and `land-eval-flake`
**verify:** No path through land skips a gate the diff can affect: a gate runs when the diff touches a path its line names, narrow only when the line gives a narrow form and no touched path forces the whole gate, and whole in every other case, including doubt about the touched paths and a plan with no `**Gates:**` line in a repo whose docs name a gate. The land report says for each gate which form ran or why it did not run. The land skill no longer names `scripts/eval`, `plugin/skills/` or `plugin/hooks/`, and a test fails when one comes back. List every combination of line present or missing, path touched or not, and narrow form given or not, with what runs in each.
- [x] Red: `skill_land_test.go` fails because the land text still names this repo's eval gate and has no rule for a `**Gates:**` line, and `TestEvalCases` fails because `land-gate-narrow` does not exist.
- [x] Change: replace the eval gate text in the land skill with the general gate rule, give `land-eval-flake` a plan with a `**Gates:**` line, and add the `land-gate-narrow` case.
- [x] Commit: feat(land): run each slow gate as far as the diff calls for

### Task 3: A moved parent repeats only what it can break, and a known flake blocks once

**Files:** `plugin/skills/land/tidy.md`, `plugin/skills/land/SKILL.md`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the land skill
**verify:** On every path where the gates run a second time (the merge result in full history, the tip in tidy history), the full suite runs again as before, and a slow gate runs again only when the commits the parent gained touch a path its line names, in the form those paths call for; otherwise the report says the slow gate is reused and names the parent commits it read. Doubt about what the parent gained means the gate runs. A red case that an open bug lists under `flaky_tests` runs again alone once; every other red case keeps 2 reruns; a case red on every rerun still stops the land. List each second-run path and each kind of red case, with what happens.
- [ ] Red: `skill_land_test.go` fails because the second-run text repeats every gate and the rerun rule does not tell a known flake from a new red case.
- [ ] Change: narrow the second run of a slow gate in `tidy.md` and step 6, and add the one-rerun rule for a known flake.
- [ ] Commit: feat(land): a moved parent repeats only the gates it can break

### Task 4: This repo's eval gate as data

**Files:** `CLAUDE.md`, the `prompt.md` of each case under `plugin/evals/`, `.acta/wiki/eval-gate-by-tag.md`, `internal/plugincheck/evals_scope_test.go`
**Seam:** a new test in plugincheck over the `evalCases` table and each case's `tags:` line, and `scripts/eval --tag <skill folder>`
**verify:** Every eval case carries, as a tag, the name of each skill folder it guards, so `scripts/eval --tag <skill folder>` runs every case of that skill and no skill's case is left out of its own tag; a case that guards a file outside `plugin/skills/` needs no such tag, since a change there runs the whole gate. `CLAUDE.md` names the gate with all four parts: the command, the paths that make it run, the narrow form, and the paths that force the whole gate. The wiki page says why a narrow run is allowed and when the whole gate still runs. List every case, its guarded file and its tags.
- [ ] Red: the new plugincheck test fails because cases such as the brainstorm ones lack the tag of the skill folder they guard.
- [ ] Change: add the missing tags, the gate line in `CLAUDE.md` and the wiki page.
- [ ] Commit: chore(evals): tag each case with its skill and name the eval gate in CLAUDE.md

## State

### Next

Task 3: narrow the second run of a slow gate in tidy.md step 3 and SKILL.md step 6; known flake reruns once.
Task 4: tags, CLAUDE.md gate line, wiki page, evals_scope_test.go.

### Findings

baseline.md still says "An `eval:<case>` record follows the same rule" while SKILL.md now says <gate>:<case>; baseline.md is outside the plan files, so it is left for review to rule on.
