---
parent: specs/2026-10-08-review-tiers
depth: minimal
id: PLN-0120
created: "2026-10-08 08:56:57"
hash: nxy7zjg
started: "2026-10-08 08:58:19"
finished: "2026-10-08 09:01:47"
---
# Review depth follows what the diff changes

**Goal:** The review skill picks a light inline review or the two reviewers from what the diff changes, for round 1 and for a polish commit, while fix rounds keep the two reviewers.

**Spec:** .acta/specs/2026-10-08-review-tiers.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- `## Review tiers` is the one place that says how deep a review goes. Other lines point to it and never restate the tier rules.
- Skill and eval text is plain English that fits every repo: no repo paths, no user names.
- Exact phrases later tasks rely on, written verbatim in `plugin/skills/review/SKILL.md`: the heading `## Review tiers`, the sentence `Light review: the orchestrator reads the full diff`, and the sentence `A fix round always takes the two reviewers.`
- acta write commands auto-commit, so eval scaffolds and tests run them only in temp folders.

## Waves

- Wave 1: Task 01, Task 03
- Wave 2: Task 02

### Task 01: Review tiers in the review skill

**Files:**
- Modify: `plugin/skills/review/SKILL.md`, `plugin/skills/build/dispatch.md` (only if a line there says the polish always gets the two reviewers)
- Test: `internal/plugincheck/skill_review_test.go`

**verify:** No line in any skill says every polish or every change of more than one file takes the two reviewers, and the skill names every spec rule: light tier lines, full tier cases with "in doubt: full", light review steps with the `## Review notes` line, the dispatch exception, tiers for round 1 and polish only, and fix rounds always full. List each rule and the sentence that carries it, and each old sentence checked as gone.

- [x] Failing test: in `skill_review_test.go` swap the Must strings `Small means one file, and only text or config with no code logic.` and `The polish commit counts as no round. Every polish commit gets the two reviewers.` and `then the two reviewers review the polish range.` for the new phrases (the three from Global Constraints plus `The polish commit counts as no round.`), and add both old sentences to MustNot; run `scripts/test ./internal/plugincheck/ -run Review` and see it fail.
- [x] Code: replace `## Small changes` with `## Review tiers` holding the spec's light tier, full tier, light review and where-it-applies rules; in `## After a CLEAN round` step 2, the polish gets the review its tier picks (`## Review tiers`) in place of the two reviewers, keeping the full test run and every revert and `[debt]` rule; adjust `dispatch.md` only if it restates the old rule; stay under the review `MaxLines` cap.
- [x] Run `scripts/test ./internal/plugincheck/` passes, commit.

### Task 02: Eval cases for both tiers

**Files:**
- Create: `plugin/evals/polish-light-review/` and `plugin/evals/polish-full-review/` (each `case.yaml`, `prompt.md`, `scaffold.sh`, `graders/`)
- Modify: `internal/plugincheck/evals_test.go`

**verify:** Each case fails a run that picks the wrong tier and passes a run that picks the right one: the light case rejects any reviewer subagent dispatch and any reply without a verdict, and the full case rejects a run that dispatches no reviewer. List each grader and the wrong run it rejects.

- [x] Failing test: add `{"polish-light-review", "skills/review/SKILL.md", "Light review: the orchestrator reads the full diff"}` and `{"polish-full-review", "skills/review/SKILL.md", "## Review tiers"}` to `evalCases`, and run `scripts/test ./internal/plugincheck/ -run Eval` and see it fail on the missing folders.
- [x] Code: copy the shape of `plugin/evals/repo-override-ask` (scaffold guards an empty folder, writes the config only when missing, copies acta into `bin/`, git init with a local user). Each scaffold makes a repo with a plan that has a CLEAN round 1 and a polish commit on top: comments only for the light case, one changed logic line for the full case. Prompt: run the review of that polish commit with the acta review skill, then stop. Graders per `plugin/evals/FACTS.md`: light case `tool_used` Agent max 0 plus a `last_message` regex for CLEAN or BLOCKED; full case `tool_used` Agent min 1. Grant the tools the case needs (`Skill`, `Read`, `Agent`) as FACTS.md says.
- [x] Run `scripts/test ./internal/plugincheck/` passes, commit; the orchestrator runs `scripts/eval` on both cases with a branch binary on PATH.

### Task 03: Bump the plugin patch version

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** All three files carry the same version, one patch above `0.1.41`, so `0.1.42`. List each file and its version.

- [x] Failing test: none new; `internal/plugincheck` already checks the three files agree.
- [x] Code: set `"version": "0.1.42"` in the three files.
- [x] Run `scripts/test ./internal/plugincheck/` passes, commit.

## State

### Findings

main was rewritten at 08:58 (8 scratch commits squashed to 7c8b360); new main tip 6d6a8c6 has the same tree as 808a9dd.
Before land: git rebase --onto main 808a9dd review-tiers in the worktree, after every task is committed.

## Polish

### Task 04: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] `plugin/skills/build/dispatch.md` `## Review`: drop the sentence about the small-change self-review; say instead that the review depth comes from `acta:review`'s `## Review tiers`.
- [x] `plugin/skills/review/SKILL.md` description: say the review runs two reviewers or a light inline review by `## Review tiers`, inside the description byte cap (trim other words if needed).
- [x] `plugin/skills/review/SKILL.md` `## Review tiers`: replace "always takes the two reviewers, except in the light tier" with plain words that say work another agent wrote takes the two reviewers in the full tier; add that a light review that finds any line outside the light tier switches to the full review.
- [x] `plugin/evals/polish-full-review/graders/reviewer-dispatched.md`: prove two reviewers ran, `min: 2` with an `input_match` the reviewer brief always carries (for example `axis`), per `plugin/evals/FACTS.md`.
- [x] `plugin/evals/polish-light-review/scaffold.sh`: the polish adds only `#` comments, no docstring.
- [x] `plugin/evals/polish-light-review/prompt.md` and `polish-full-review/prompt.md`: allow the one `## Review notes` line in the plan; every other file stays as it is.
- [x] Commit: `polish: review notes for PLN-0120`

## Review notes

- Round 1 CLEAN on both axes; six `[fix]` NOTEs went into the Task 04 polish, whose review was CLEAN on both axes. The diff changes skill text, so it took the full tier.
- `budget_test.go` was not named in the plan, but byte caps are per file, so the review SKILL.md cap had to follow the new size (11953 to 12548).
- polish-light-review's `CLEAN|BLOCKED` grader would also match a reply that only repeats the prompt; the `tool_used` grader is the one that tells the tiers apart.
- The two polish scaffolds differ only in the polish body; each case keeps its own script, like every other case.
- Eval run before the polish: 14 of 15 at 1.00, both new cases 1.00; state-resume 0.50 because its grader could not find `wt/src/invoice.py`, a case this diff does not touch.

### Task 05: Grader counts reviewer briefs in any case

**verify:** The polish-full-review grader counts a reviewer dispatch whether its brief says "Axis" or "axis", and still does not count a non-reviewer subagent. The eval case passes 3 runs out of 3.

- [x] `plugin/evals/polish-full-review/graders/`: `[Aa]xis` scored 0 on 3 runs because `input_match` is a plain substring, so the grader split into `reviewer-dispatched.md` (`input_match: "Spec"`, `min: 1`) and `standards-reviewer-dispatched.md` (`input_match: "Standards"`, `min: 1`), each comment saying why.
- [x] `plugin/evals/FACTS.md` grader table: `tool_used` `input_match` is case-sensitive; note what the orchestrator's eval run proved about it being a regex.
- [x] Commit: `fix(evals): count each reviewer by the axis it names`

### Task 06: Full review always dispatches the two reviewers

**verify:** An agent that picks the full tier dispatches the two reviewer subagents on every run, however small the diff: polish-full-review passes 3 runs of 3, and no skill line reads as letting the orchestrator read a full-tier diff itself in their place.

- [x] `plugin/skills/review/SKILL.md` `## Review tiers`: one sentence saying a full review means dispatching the two reviewer subagents, and reading the diff yourself never replaces them, however small the diff; `internal/plugincheck/skill_review_test.go` Must string for it; byte cap in `budget_test.go` follows the new size only if needed.
- [x] Run `polish-full-review` 3 times with the branch binary; 3 of 3 pass.
- [x] Commit: `fix(review): a full review always dispatches the two reviewers`
- Fix review of Tasks 05 and 06 (two reviewers, both CLEAN). Their `[fix]` NOTEs moved to `[debt]`, since a polish already ran: a third grader with `min: 2` and no `input_match`, and anchoring "Reading the diff yourself never replaces" to the full tier.
- `SKILL.md` says "Full review: the two reviewers." and then "A full review means you dispatch the two reviewer subagents.", which repeats itself; it costs bytes only.
- Wiki timestamp bumps on omp-brief-history-rules and skill-text-fits-everyone carry no content change.
