---
id: SPC-0206
created: "2026-11-12 23:10:32"
hash: fp2xn6x
started: "2026-11-13 00:27:31"
finished: "2026-11-13 00:59:00"
---
# Review depth auto: the review a diff needs, picked by rule and reason

Bounded. Worked out with the user in the terminal on 2026-11-12. The user approved the design in chat. SPC-0203 covers the evidence a review must show; this spec covers how much review a diff gets.

## Problem

Every build ends in the full review unless every changed line is a comment, docs, log text or a test name. A rename whose tests were watched going red and green still gets two reviewer subagents, and three with the fact-check at the deep level. The review costs the same for a safe diff as for a risky one.

Letting the agent choose freely is not safe either. The agent that wrote the code is the worst judge of whether it needs a review, and the wiki page `skill-escape-lines-skip-proof.md` records that a line allowing less work is read as leave to do less.

## Changes

1. A fourth value, `auto`. `review_depth` and the argument of `/acta:review` take `auto` next to `quick`, `standard` and `deep`. The three old values behave as they do today. The judgment of this spec runs only under `auto`.
2. Three levels under `auto`. The orchestrator picks the lowest level whose conditions hold.
   - Inline: the orchestrator reviews the diff itself and writes the same `### Coverage` section a reviewer writes (SPC-0203). Allowed only when no risk signal holds, the diff changes nothing that code outside the diff uses (a function, a type, a file format, a command), and every changed logic line is covered by a test that the task's record shows failing before the change.
   - One reviewer: one reviewer subagent that answers both axes. Allowed when no risk signal holds and the inline conditions do not all hold.
   - Two reviewers: the two reviewers of the standard level. Used when a risk signal holds, or when the orchestrator is in doubt about any condition above.
3. Risk signals are not up for judgment. Each of these gets two reviewers under `auto`, and the deep lens where the floors ask for it today: a security, auth, money, data migration or delete path; text an agent reads as instructions (skills, prompts, hook text); work the `dispatch` executor wrote; a fix round.
4. Risk decides, never size. The text says a small diff is no reason for a lower level, and gives a one-line change to a condition as the example of a small diff that needs a reviewer.
5. The pick is written down. The summary line under the plan's `## Review notes` names the level picked and the reason in one line, including which condition failed when the level is not inline. The land report repeats that line. A user who disagrees runs `/acta:review standard` or `/acta:review deep`, which wins for that run as it does today.
6. The light tier stays as it is at every depth: a diff of only comments, docs, log text or test names keeps the light inline review.
7. The default moves only on proof. Three eval cases are added:
   - `review-auto-risky-small`: a one-line change to a condition with no test, on no risk path. At least one reviewer subagent must run.
   - `review-auto-signal`: a diff on an auth path. Two reviewer subagents must run.
   - `review-auto-safe-inline`: a rename covered by a test the task's record shows failing first. No reviewer subagent runs, and the reply holds `### Coverage`.

   The default of `review_depth` changes from `standard` to `auto` only when `review-auto-risky-small` and `review-auto-signal` each pass at least 9 of 10 runs. Below that, `auto` ships as a value a user can set, the default stays `standard`, and the plan's `## Review notes` say so. `review-auto-safe-inline` needs 7 of 10; below that the inline level is taken out and `auto` picks between one and two reviewers.

## Files

- Go, for the value and the default: `internal/config/user.go`, `internal/config/repo_user.go`, `internal/cli/config_cmd.go`, `internal/cli/help.go`, `internal/setup/form.go`, `internal/telemetry/store.go`, and their tests.
- Skill text: `plugin/skills/review/depth.md`, `plugin/skills/review/SKILL.md`, `plugin/skills/land/SKILL.md`.
- Three new case folders in `plugin/evals/`.
- `internal/plugincheck/skill_review_test.go`, `skill_land_test.go`, `evals_test.go` and `budget_test.go`. A file that passes its byte cap gets a higher cap. No words an agent acts on are trimmed to fit.

## Testing

- Go tests for the new value in every place that reads or checks `review_depth`: validation, `acta config set`, `acta config show`, the repo file, the setup form, the telemetry allow-list.
- The three eval cases, 10 runs each, about 30 runs and about 5 minutes at `scripts/eval -j 8`. The existing `review-depth-floor` case must not get worse.
- `internal/plugincheck` keeps the required strings and the size caps.

## Order

The build starts after PLN-0232 lands, because both change `plugin/skills/review/SKILL.md`, `depth.md` and the plugincheck files. PLN-0233 and PLN-0234 move behind this plan for the same plugincheck files.

## Out of scope

A change to what `quick`, `standard` or `deep` do, the number of rounds, the fact-check, the user's own saved setting, and a per-task review.
