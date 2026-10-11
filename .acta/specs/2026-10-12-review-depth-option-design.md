---
parent: scratch/2026-10-12-review-depth-option
id: SPC-0125
created: "2026-10-12 07:01:50"
hash: pu65prv
---
# Review depth option

Status: Architectural, design approved by the user in chat on 2026-10-12 (SCR-0061).

## Problem

`acta:review` has one depth. Round 1 picks a light or full tier from the kind of lines changed, and full always means two reviewer subagents. A user cannot ask for a cheaper review on low-risk work, or a stricter one when false BLOCKERs keep wasting fix rounds.

`alibaba/open-code-review` (Apache-2.0, read at `357c3f0`) has an `--effort low|medium|high` flag, but it only changes the number of review rounds. Three of its other ideas are worth taking: a fact-check pass that may only drop a finding the diff proves wrong, earlier findings passed into later rounds with "do not repeat them", and a rule that every changed file gets its own pass.

## Goal

The user picks how deep a review goes, per user, per repo, or for one run. The level changes which review passes run, not the round count. Sensitive paths never get less than today's review.

Out of scope: per-file-type rule checklists, and picking the tier from line counts. The tier keeps coming from the kind of lines changed.

## Design

### 1. Setting (`internal/config`, `internal/cli/config_cmd.go`, `internal/setup`)

- New key `review_depth`: `quick`, `standard` or `deep`. Empty means `standard`. Any other value fails validation with the same error shape as `plan_depth`.
- It is a user key in `~/.acta/config.yaml` and a repo key in `.acta.yaml`; the repo value wins. It joins `RepoKeys`, `acta config show` (text and JSON, with the repo mark), the repo and user set flags, and the `acta setup` form as one more select question with `standard` as the default.
- The review skill picks the level in this order and does not ask when one answers: the argument of `/acta:review quick|standard|deep` (this run only, never saved), then `review_depth` from `acta config show`, then `standard`.

### 2. Levels (`plugin/skills/review/depth.md`, new)

`SKILL.md` gets a short `## Review depth` section with the pick order and a link to `depth.md`. `depth.md` holds the level rules:

- **quick:** a diff that is all light tier gets today's inline light review. Any other diff gets one reviewer subagent that answers both the Spec and the Standards axis in one report.
- **standard:** today's flow, unchanged.
- **deep:** the two reviewers, then one read-only fact-check subagent over the BLOCKERs only, on the reviewer model. It may move a BLOCKER to a NOTE on one of two grounds: (A) the code it names is not in the diff under review, or (B) one line of the diff contradicts the claim, with no further inference step. In doubt, it stays a BLOCKER. A BLOCKER about security, data loss, concurrency or a behaviour change is never moved. Each move names its ground and the line. Its prompt is `plugin/skills/review/fact-check.md` (new), and it writes its reasoning before the list of moved ids. With no BLOCKER, it does not run.

At every level:

- `code-reviewer.md` gets one rule: every changed file gets its own pass, and a file being small or secondary in the diff is no reason to skip it.
- Rounds 2 and 3 pass the reviewers the earlier rounds' findings with "do not repeat these; look for others". `code-reviewer.md` gets an `[EARLIER_FINDINGS]` placeholder, empty in round 1.

### 3. Floors

The setting can raise the depth, never lower it below these:

- A diff touching a security, auth, data migration, money or delete path runs at least `standard`, with the deep lens.
- Fix rounds run at least `standard`.
- Work the `dispatch` executor wrote runs at least `standard` in the full tier.
- Three rounds at most at every level. There is no round 4.

The orchestrator writes the level used, and any floor that raised it, into the one review line of the plan's `## Review notes`.

### 4. Checks and eval

- `internal/plugincheck`: byte caps for `review/SKILL.md`, the new `depth.md` and `fact-check.md`, and the review folder line cap, all fitted to the new text. The review skill rule requires the phrases `review_depth` and `fact-check`.
- New eval `plugin/evals/review-depth-floor`: repo config sets `review_depth: quick`, the diff changes an auth check; it passes when two reviewer subagents are dispatched.

## Tests

- Config: `review_depth` round-trips through user and repo config; repo beats user; a bad value fails with the key name in the error; `acta config show` prints it with the repo mark; the setup form defaults to `standard`.
- Plugincheck: caps and required phrases above.
- Eval: `review-depth-floor`.

## Wiki

None. The level rules live in the skill files.
