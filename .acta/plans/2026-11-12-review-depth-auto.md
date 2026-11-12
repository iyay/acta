---
parent: specs/2026-11-12-review-depth-auto
depth: minimal
after: [PLN-0232]
id: PLN-0235
created: "2026-11-12 23:10:32"
hash: zyf9mv4
---
# Review depth auto: the review a diff needs, picked by rule and reason

**Goal:** Under `review_depth: auto` a diff gets the lowest review level its risk allows, the pick and its reason are written down, and the default moves to `auto` only when the eval cases prove risky diffs are never under-reviewed.

**Spec:** `.acta/specs/2026-11-12-review-depth-auto.md`

**Tests:** fast `scripts/test ./internal/config ./internal/cli ./internal/setup ./internal/telemetry ./internal/plugincheck`; full `scripts/test --full`; evals `scripts/eval --case <name> --runs 10 -j 8`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- `quick`, `standard` and `deep` behave exactly as before this plan, and every floor in `depth.md` holds under `auto` too.
- No line may let the orchestrator pick a lower level because a step failed or a fact is unknown: doubt about any condition means two reviewers.
- The `acta` on PATH must be the branch binary before any eval run. Build it to a folder of its own and put that folder first on PATH for the run; never overwrite the user's installed binary.
- The user's own saved setting is never rewritten by this plan.
- Skill text stays general for every acta user, in plain English. When a file passes its byte cap, raise the cap in `internal/plugincheck/budget_test.go`; never trim words an agent acts on.

## Build brief

- Land order: this plan starts after PLN-0232 is on main, because both change `plugin/skills/review/SKILL.md`, `plugin/skills/review/depth.md` and the plugincheck files.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3
- Wave 4: Task 4
- Wave 5: Task 5

### Task 1: `auto` is a valid review depth everywhere the setting is read

**Files:** `internal/config/user.go`, `internal/config/user_test.go`, `internal/config/repo_user_test.go`, `internal/cli/config_cmd.go`, `internal/cli/config_repo_test.go`, `internal/cli/help.go`, `internal/setup/form.go`, `internal/setup/form_test.go`, `internal/telemetry/store.go`, `internal/telemetry/store_test.go`
**Seam:** `cli.Run([]string{"config", "set", "--review-depth", "auto"})`, `cli.Run([]string{"config", "show"})`, and the validation of the user file and the repo file
**verify:** No place that reads, checks, lists or reports `review_depth` rejects or drops `auto`: the user file, the repo file `.acta.yaml`, `acta config set` with and without `--repo`, `acta config show` text and JSON, the setup form's choices, the help text and the telemetry allow-list. Every other value is still refused with an error that now names four values. The default is still `standard`. List every place checked and what it does with `auto` and with a bad value.
- [ ] Red: the config, cli, setup and telemetry tests fail because `auto` is refused as a review depth.
- [ ] Change: add `auto` to the allowed values, the form choices, the help text and the telemetry allow-list.
- [ ] Commit: feat(config): accept auto as a review depth

### Task 2: Three eval cases for the auto pick

**Files:** `plugin/evals/review-auto-risky-small/`, `plugin/evals/review-auto-signal/`, `plugin/evals/review-auto-safe-inline/`, `internal/plugincheck/evals_test.go`
**Seam:** `TestEvalCases` in plugincheck, and `scripts/eval --case 'review-auto-*'`
**verify:** Each case fails for the one behaviour it names and for no other reason: under `review_depth: auto`, a one-line change to a condition with no test and on no risk path must run at least one reviewer subagent; a diff on an auth path must run two; a rename covered by a test the task's record shows failing first must run none and must reply with `### Coverage`. Every case sets `review_depth: auto` in its scaffold, has a `# guards:` line for a heading the review skill has, a scaffold that refuses a folder that is not empty, and caps on turns and time. List each case and its graders.
- [ ] Red: `TestEvalCases` fails because the three case folders it now lists do not exist.
- [ ] Change: add the three case folders and their `evalCases` entries.
- [ ] Commit: test(evals): three cases for the auto review pick

### Task 3: The auto levels, their conditions and the written reason

**Files:** `plugin/skills/review/depth.md`, `plugin/skills/review/SKILL.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the review skill, and the cases `review-auto-risky-small`, `review-auto-signal`, `review-auto-safe-inline` and `review-depth-floor`
**verify:** Under `auto`, no diff with a risk signal (a security, auth, money, data migration or delete path; text an agent reads as instructions; work the `dispatch` executor wrote; a fix round) gets fewer than two reviewers, and no diff gets the inline level unless all three inline conditions hold: no risk signal, nothing changed that code outside the diff uses, and every changed logic line covered by a test the task's record shows failing first. Doubt about any condition means two reviewers. The text says size is no reason for a lower level. The summary line under `## Review notes` names the level and the reason, and the failed condition when the level is not inline. The argument `/acta:review auto` is accepted, and `quick`, `standard`, `deep` and the light tier read as before. List every path from a diff to a level, and the level each ends at.
- [ ] Red: `skill_review_test.go` fails because the review text has no `auto` level, no inline conditions and no reason on the review-notes line.
- [ ] Change: add the `auto` level with its three levels, conditions, risk signals and the reason line to `depth.md`, and name `auto` where `SKILL.md` lists the depths.
- [ ] Commit: feat(review): auto depth picks the review a diff needs and says why

### Task 4: The land report repeats the review level and its reason

**Files:** `plugin/skills/land/SKILL.md`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the land skill
**verify:** Every land report, whatever the depth, repeats the level line from the plan's `## Review notes` as it was written, so the user sees which review ran and why without opening the plan. The line is copied, never reworded or judged again at land. List every kind of land report (clean, land anyway, light review) and where the line sits in each.
- [ ] Red: `skill_land_test.go` fails because the land skill does not ask for the review level line in its report.
- [ ] Change: add one line to the land report rules that repeats the review level line.
- [ ] Commit: feat(land): report which review ran and why

### Task 5: The default moves to auto only when the cases prove it

**Files:** `internal/config/repo_user.go`, `internal/config/repo_user_test.go`, `internal/cli/config_cmd.go`, `internal/cli/config_repo_test.go`, `internal/setup/form.go`, `internal/setup/form_test.go`, `plugin/skills/review/depth.md`, `internal/plugincheck/skill_review_test.go`
**Seam:** `cli.Run([]string{"config", "show"})` with no `review_depth` set, the setup form's first value, and the plugincheck required strings for `depth.md`
**verify:** Run the three `review-auto-*` cases and `review-depth-floor`, 10 runs each, and record the scores with `acta state set <plan id> findings`. The default becomes `auto` in every place that fills a missing `review_depth` (config show, the repo file reader, the setup form, the last step of the pick order in `depth.md`) only when `review-auto-risky-small` and `review-auto-signal` each pass at least 9 of 10 and `review-depth-floor` is no worse than before; all of those places agree with each other in either outcome. Below 9 of 10 the default stays `standard` everywhere and this task ends with its boxes ticked and no code commit. When `review-auto-safe-inline` passes fewer than 7 of 10, the inline level is taken out of `depth.md` and `auto` picks between one and two reviewers. A saved setting is never changed. List the scores, the outcome, and every place that holds the default.
- [ ] Red: the config, cli and setup tests fail because a missing `review_depth` still reads as `standard`.
- [ ] Change: when the scores allow it, switch the default to `auto` in every place that fills a missing value.
- [ ] Commit: feat(config): review depth defaults to auto
