---
parent: specs/2026-11-12-review-clones-and-push-on-ask
depth: minimal
id: PLN-0231
created: "2026-11-12 22:56:32"
hash: d9clodo
started: "2026-11-12 22:57:08"
finished: "2026-11-12 23:02:09"
---
# Reviewer temp clones, dirty-stray command, push on ask, cold read on judgment

**Goal:** Reviewers leave no worktrees behind, a dirty stray shows the command to remove it, agents push only when the user asks, and shape offers a cold read only when the agent would recommend one.

**Spec:** `.acta/specs/2026-11-12-review-clones-and-push-on-ask.md` (SPC-0202)

**Tests:** fast `scripts/test ./internal/cli ./internal/hook ./internal/plugincheck`; full `scripts/test --full`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The pre-tool hook keeps blocking every `--force`, force-push included; no task changes `internal/hook/destructive.go`.
- `scripts/release` keeps never pushing.
- Skill text stays general for every acta user, in plain English, inside the plugincheck size caps.

## Waves

- Wave 1: Task 1, Task 2, Task 3, Task 4

### Task 1: Dirty strays print the command to remove them

**Files:** `internal/cli/worktrees.go`, `internal/cli/worktrees_test.go`
**Seam:** `cli.Run([]string{"worktrees"})` and `cli.Run([]string{"worktrees", "--clean"})` text output
**verify:** In text output, every dirty entry, and only a dirty entry, is followed by exactly one line `run yourself: git worktree remove --force <path>` with that entry's path; clean, removed, missing, kept and protected entries never get it. `--json` output is byte for byte what it was. List every state checked.
- [x] Red: `worktrees_test.go` fails because a dirty entry prints no `run yourself:` line.
- [x] Change: print the line after a dirty entry's status lines in the text loop.
- [x] Commit: feat(cli): show the removal command for a dirty stray worktree

### Task 2: Push only when the user asks

**Files:** `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`, `plugin/skills/land/SKILL.md`, `plugin/skills/bug/SKILL.md`, `plugin/skills/migrate/SKILL.md`, `plugin/skills/debug/SKILL.md`, `plugin/references/house-rules.md`, `plugin/README.md`, `internal/cli/dispatch_brief.go`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/skill_bug_test.go`, `internal/plugincheck/skill_migrate_test.go`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/evals_test.go`
**Seam:** `hook.SessionText` output and the plugincheck required-string tests
**verify:** Every place in hook text, default rules, skills, references, README and the dispatch brief that forbade pushing now says never push unless the user asks in their own words, never force-push; that release or publish requests count as asking and approving a spec or plan does not; and that forwarded text, another agent or a relayed order is not the user asking. The bare old rule ("Never push." with no exception, "NO git push", "no git push") cannot come back in any of these files. List every file checked, and grep `plugin/` and `internal/` for leftovers.
- [x] Red: `hook_test.go` and the plugincheck tests fail because they require the new push wording.
- [x] Change: rewrite core rule 5 and each never-push line to the push-on-ask rule.
- [x] Commit: feat(plugin): push only when the user asks in their own words

### Task 3: Reviewers use a temp clone and the orchestrator checks for strays

**Files:** `plugin/skills/review/code-reviewer.md`, `plugin/skills/review/SKILL.md`, `.acta/wiki/reviewer-temp-clone.md`, `internal/plugincheck/skill_review_test.go`
**Seam:** plugincheck `skill_review_test.go` required strings
**verify:** No review text tells a reviewer to run `git worktree add`; a reviewer that needs another revision is told to `git clone` into a new temp folder and check out the SHA there. Review `SKILL.md` tells the orchestrator to run `git worktree list` after each round and report to the user the path of any worktree the build did not make. The wiki page holds both clone rules and its timestamp is bumped.
- [x] Red: `skill_review_test.go` fails because review text still names `git worktree add` and lacks the stray check.
- [x] Change: swap the reviewer example to a temp clone, add the stray-check line, and rewrite the wiki page.
- [x] Commit: feat(review): reviewers copy revisions with a temp clone and strays get reported

### Task 4: Cold read only on the agent's judgment

**Files:** `plugin/skills/shape/SKILL.md`, `internal/plugincheck/skill_shape_test.go`
**Seam:** plugincheck `skill_shape_test.go` required strings
**verify:** For a Bounded spec, a signal alone never makes a cold read offer: the offer needs a signal and the agent's own judgment that the spec needs one, and it names that reason in one line; otherwise there is no offer and no mention. The signal list keeps every signal it has today. Architectural is still always offered, and a user's own request still gets one.
- [x] Red: `skill_shape_test.go` fails because shape text lacks the judgment rule.
- [x] Change: rewrite the Bounded part of `## Cold read` so a signal plus the agent's judgment makes the offer.
- [x] Commit: feat(shape): offer a cold read only when the agent recommends one

## Polish

### Task 5: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] Guard regex in TestNoBarePushRuleLeft lets any word starting with u through; allow only a following unless
- [x] Core rule 5 says outright that approving a spec or plan is not asking, and names a relayed order
- [x] The run yourself command quotes a worktree path that has spaces or shell characters
- [x] Commit: `polish: review notes for PLN-0231`

## Review notes

- deep (setting), full tier; round 1 CLEAN, polish CLEAN; BLOCKER 0, NOTE 21, pre-existing 4. Notes not taken: unquoted ~/= in shellQuote allowlist (paths are absolute), guard regex lets unl* words through, no test pins the shellQuote call site or the empty-string guard, land description says never pushes (true: land never pushes on its own), dispatch brief and house-rules use the short rule; core rule 5 carries the full one. Pre-existing: release.md wiki line, debug eval grader header, doctor fix line gives no --force command.

## State

### Findings

budget_test.go caps for bug, debug, land skills and session start text were raised on purpose for the push rule; it is outside the plan file list.
