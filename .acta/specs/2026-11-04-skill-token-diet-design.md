---
parent: scratch/2026-11-04-skill-token-diet
id: SPC-0188
created: "2026-11-04 19:41:55"
hash: vtyw8gx
---
# Skill token diet

## Problem

A full feature cycle loads about 66k chars (about 16.5k tokens) of skill bodies, and each session start adds about 2.6k tokens. Much of that text is for modes the user is not in. A user on the `inline` executor still loads the subagent and dispatch text in build. A user on `minimal` plan depth still loads the full-depth text in slice. Review loads its receiving-findings text even when there are no findings. Review subagents are 10% of usage, and the Standards-axis reviewer does not need the orchestrator's model.

## Goal

Load less text per cycle and keep the agents just as good. Every user saves, whatever their settings.

## Out of scope

- `acta:frame` and `acta:lean` stay as they are. A skill that is never loaded costs only its line in the skill list.
- The Lean coding guide block stays in the SessionStart hook. It is the only way the lean rules reach the agent.
- Skill frontmatter descriptions do not change, since they drive skill matching.
- The user's own `~/.claude/CLAUDE.md` rule 9 was already changed outside this spec.

## Design

### One file per mode

When a skill has modes, its `SKILL.md` keeps only the text every mode uses. Each mode gets its own file in the skill folder. The body has a table with one row per mode: the mode, its file, and when to read it. This is the pattern build already uses for `dispatch.md`.

The agent first picks the mode by the skill's existing order, then reads exactly one mode file. It never reads the file of another mode.

- build: `subagent.md`, `dispatch.md`, `inline.md`. Mode order stays: the `/build <executor>` argument, then `build_executor` from `acta config show`, then `inline`. The current `dispatch.md` becomes the dispatch mode file.
- slice: `full.md`, `minimal.md`. Mode order stays: the `/slice` argument, then `plan_depth`, then `full`.
- review: the receiving-findings text moves to `receiving.md`. Read it when any reviewer returns a finding, when the user or an outside reviewer gives review feedback, or when a light inline review finds something.

### Standards reviewer on a cheaper tier

Review keeps the Spec-axis reviewer on the orchestrator's model. There is no new config key. The "Name no model for any subagent" line in review becomes:

> The Spec-axis reviewer runs on your own model. The Standards-axis reviewer runs on a cheaper, faster model tier than yours, for example sonnet in Claude Code. If the harness has no cheaper tier, use its default.

The same line in slice stays as it is.

### tdd trim

Common Rationalizations and Red Flags become one block of about 10 lines. Example: Bug Fix and Verification Checklist are removed. The Iron Law and the red, green, refactor steps stay. Target about 6.5k chars, down from 10.7k.

### SessionStart hook

`internal/hook` prints each plan in the work list once. It shows at most 3 plans: in-progress plans first, then the current order. When more remain, it adds one line `+N more`. Other work-list kinds (bugs, scratch) are not capped.

Core rule 8 (several brainstorms in one session) moves out of the hook, together with its move text (`claude --bg`, the herdr line). Shape already has this under "Many topics, one session"; the plan checks that it covers both move ways and adds any part that is missing.

## Plans

Each plan lands on its own, in this order.

1. review: cheaper Standards tier, `receiving.md`.
2. build: three mode files.
3. slice: two mode files.
4. tdd: trim.
5. SessionStart hook: dedup, cap 3, drop rule 8.

Each plan updates the caps in `internal/plugincheck` for the files it touches. The per-file byte caps prove the targets: the `SKILL.md` cap drops to its target and each new mode file gets its own cap. The folder line cap (`MaxLines`, which counts every markdown file in the folder) may stay or rise, since moving text does not shrink the folder.

## Targets

Chars of skill body loaded:

| Case | Today | Target |
|---|---|---|
| build, inline | 12.9k | 7k or less |
| slice, minimal | 12.7k | 7k or less |
| review, no findings | 13.4k | 10k or less |
| tdd | 10.7k | about 6.5k |
| SessionStart hook | about 1.15k tokens | about 400 tokens less |

## Testing

- Plans 1 to 4: the new per-file byte caps pass. Greenfield evals of the changed skill reach the same results as a baseline run on main. Eval transcripts show no Read of another mode's file.
- Plan 5: Go tests in `internal/hook` for the dedup, the cap of 3 plans with in-progress first, the `+N more` line, and the missing rule 8.
- Hook-only changes need no eval.
