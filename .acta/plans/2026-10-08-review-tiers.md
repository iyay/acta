---
parent: specs/2026-10-08-review-tiers
depth: minimal
id: PLN-0120
created: "2026-10-08 08:56:57"
hash: nxy7zjg
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

- [ ] Failing test: in `skill_review_test.go` swap the Must strings `Small means one file, and only text or config with no code logic.` and `The polish commit counts as no round. Every polish commit gets the two reviewers.` and `then the two reviewers review the polish range.` for the new phrases (the three from Global Constraints plus `The polish commit counts as no round.`), and add both old sentences to MustNot; run `scripts/test ./internal/plugincheck/ -run Review` and see it fail.
- [ ] Code: replace `## Small changes` with `## Review tiers` holding the spec's light tier, full tier, light review and where-it-applies rules; in `## After a CLEAN round` step 2, the polish gets the review its tier picks (`## Review tiers`) in place of the two reviewers, keeping the full test run and every revert and `[debt]` rule; adjust `dispatch.md` only if it restates the old rule; stay under the review `MaxLines` cap.
- [ ] Run `scripts/test ./internal/plugincheck/` passes, commit.

### Task 02: Eval cases for both tiers

**Files:**
- Create: `plugin/evals/polish-light-review/` and `plugin/evals/polish-full-review/` (each `case.yaml`, `prompt.md`, `scaffold.sh`, `graders/`)
- Modify: `internal/plugincheck/evals_test.go`

**verify:** Each case fails a run that picks the wrong tier and passes a run that picks the right one: the light case rejects any reviewer subagent dispatch and any reply without a verdict, and the full case rejects a run that dispatches no reviewer. List each grader and the wrong run it rejects.

- [ ] Failing test: add `{"polish-light-review", "skills/review/SKILL.md", "Light review: the orchestrator reads the full diff"}` and `{"polish-full-review", "skills/review/SKILL.md", "## Review tiers"}` to `evalCases`, and run `scripts/test ./internal/plugincheck/ -run Eval` and see it fail on the missing folders.
- [ ] Code: copy the shape of `plugin/evals/repo-override-ask` (scaffold guards an empty folder, writes the config only when missing, copies acta into `bin/`, git init with a local user). Each scaffold makes a repo with a plan that has a CLEAN round 1 and a polish commit on top: comments only for the light case, one changed logic line for the full case. Prompt: run the review of that polish commit with the acta review skill, then stop. Graders per `plugin/evals/FACTS.md`: light case `tool_used` Agent max 0 plus a `last_message` regex for CLEAN or BLOCKED; full case `tool_used` Agent min 1. Grant the tools the case needs (`Skill`, `Read`, `Agent`) as FACTS.md says.
- [ ] Run `scripts/test ./internal/plugincheck/` passes, commit; the orchestrator runs `scripts/eval` on both cases with a branch binary on PATH.

### Task 03: Bump the plugin patch version

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** All three files carry the same version, one patch above `0.1.41`, so `0.1.42`. List each file and its version.

- [ ] Failing test: none new; `internal/plugincheck` already checks the three files agree.
- [ ] Code: set `"version": "0.1.42"` in the three files.
- [ ] Run `scripts/test ./internal/plugincheck/` passes, commit.
