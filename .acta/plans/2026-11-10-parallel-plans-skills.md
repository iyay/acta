---
parent: specs/2026-11-10-parallel-plans-subagent-ask-design
depth: minimal
id: PLN-0224
created: "2026-11-10 18:16:19"
hash: o8oqi8d
after: [PLN-0223]
started: "2026-11-10 18:34:16"
finished: "2026-11-10 18:46:40"
---
# Parallel plans: slice ask and build many.md

**Goal:** Slice asks subagent or inline when two or more approved plans are ready at once, and build has a `many.md` that runs them with one subagent per plan.

**Spec:** `.acta/specs/2026-11-10-parallel-plans-subagent-ask-design.md` (SPC-0197), Part 1 steps 1 to 7 and Part 2.

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`. Evals: `scripts/eval --case slice-split-parallel-ask` (and the other three new cases), with the branch `acta` on PATH.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Skill prose stays plain and short, with no repo paths or user names (wiki `skill-text-fits-everyone`).
- Building one plan does not change: `inline.md`, `subagent.md` and `dispatch.md` stay as they are.
- Needs `acta config pick subagent` and `acta config set --parallel-plans` from the config plan on PATH before the evals run.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2, Task 3

### Task 1: Slice asks after the plans are approved

**Files:** `plugin/skills/slice/split.md`, `plugin/skills/slice/SKILL.md`, `internal/plugincheck/skill_split_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** `scripts/test ./internal/plugincheck -run 'TestSkillSplit|TestBudget'`
**verify:** The ask fires only after the yes to every plan and the Build briefs, only with 2 or more ready plans, and the only build slice may start is the parallel one after that ask or a `subagent` setting. List each rule checked in the text: no ask with one ready plan; no ask and straight to parallel with `parallel_plans: subagent` or `build_executor: subagent`; `dispatch` not offered; inline prints the build lines and is not counted; subagent runs `acta config pick subagent`; the offer at 3 picks unless `subagent_offer: declined`, with the yes and no commands; both "never starts a build" lines name the exception.
- [x] Red: the split test fails because `split.md` does not name `acta config pick subagent`.
- [x] Change: add the ask, the count and the offer after step 5 of `split.md`, name the exception in step 6 and in slice `SKILL.md`, and raise the two size caps only as far as the new text needs.
- [x] Commit: feat(slice): ask subagent or inline for ready parallel plans

### Task 2: Build runs several plans with many.md

**Files:** `plugin/skills/build/many.md`, `plugin/skills/build/SKILL.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** `scripts/test ./internal/plugincheck -run 'TestSkillBuild|TestBudget'`
**verify:** `many.md` holds every rule of spec Part 2 and build `SKILL.md` says when to read it and not an executor file. List each rule checked: Build brief wins for repo, parent, land method and land order; one subagent per plan, all in one message, `agent="task"` on omp; the brief is `implementer-prompt.md` with every task of the plan, its worktree check, house rules, state step and model line kept; no nested subagent; the close check per plan before `acta:review`; new subagent per fix round, three rounds at most; lands one at a time; `after:` plans start once their list has landed; a stuck plan stops with its dependents; state text holds the spec id, `parallel run` and each sibling's status, and `/build` on any sibling resumes through `many.md`.
- [x] Red: the build test fails because `plugin/skills/build/many.md` does not exist.
- [x] Change: write `many.md`, add the one line beside the executor table in build `SKILL.md`, and add caps for `many.md` and the new `SKILL.md` size.
- [x] Commit: feat(build): add many.md for parallel plans with one subagent each

### Task 3: Eval cases for the ask

**Files:** `plugin/evals/slice-split-parallel-ask/`, `plugin/evals/slice-split-parallel-offer/`, `plugin/evals/slice-split-parallel-declined/`, `plugin/evals/slice-split-parallel-set/`, `internal/plugincheck/evals_test.go`
**Seam:** `scripts/eval --case <name>` for each of the four cases, and `scripts/test ./internal/plugincheck -run TestEval`
**verify:** Each case passes only on its own behaviour, from a scaffold with two approved ready plans. List each case and what fails it: `ask` with `build_executor: inline` fails when no subagent or inline ask appears or `dispatch` is offered; `offer` with `subagent_picks: 3` fails without the default offer; `declined` with `subagent_offer: declined` fails when the offer appears or the plain ask is missing; `set` with `parallel_plans: subagent` fails when any ask appears or no parallel build starts.
- [x] Red: the four cases fail on the slice text from before Task 1.
- [x] Change: add the four case folders in the shape of `slice-minimal-one-file`, and tie each to its guarded phrase in `evalCases`.
- [x] Commit: test(evals): pin the parallel plans ask, offer, decline and setting

## Polish

### Task 4: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] [fix] split.md step 7: the offer's Yes branch says to build the ready plans here with many.md of acta:build, the same words as the other branches, so no extra pick is counted.
- [x] [fix] many.md step 7: acta state set runs from each plan's own worktree, since the main checkout refuses a plan that is building.
- [x] Commit: `polish: review notes for PLN-0224`

## Review notes

- Review: deep, round 1 CLEAN on both axes; polish (offer Yes branch wording, state set from the plan worktree) full tier, CLEAN on both axes. BLOCKER 0, NOTE 16, pre-existing 0 (implementer-prompt line 3 and the review-round rule went to debt as diff-made conflicts). Eval offer case caught agents skipping the offer on the first step 7 text (3 of 5); the reordered text passed 5 of 6. Task 3 red was not run on the old text for ask, declined and set. Notes left: no-ask grader matches only 'or inline'; no-dispatch-offer can false-red on a mention of dispatch; set case proves the build only by a Read of many.md; the resume trigger sits above ## Plan.

## State

### Next

All tasks done. Close: wiki check, then acta:review over the merge base..HEAD.

### Findings

Eval offer case: on the first step 7 text agents skipped the offer 3 of 5 runs even with subagent_picks 3 in config show; step 7 now lists the cases in order and passed 5 of 6. Regex graders take no inline (?i) and no s flag; use flags: i and [\s\S].
