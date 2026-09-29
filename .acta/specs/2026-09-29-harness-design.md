---
parent: scratch/2026-09-28-harness-spec
id: SPEC-18
hash: gg2p
---

# Harness So Agents Do Not Skip the Rules

Status: design approved by the user on 2026-09-29, section by section. Covers SCRATCH-6.

## Why

Agents sometimes skip or misjudge the scratch and brainstorm rules. The
worst case is a second Architectural brainstorm in one session. Skill text
alone does not stop it, and we have no test that runs a real agent.

This spec adds three layers: hook state that remembers and blocks, text
guards in plugincheck, and a behaviour eval suite.

## Out of scope

omp. omp does not run Claude Code shell hooks, so the state, reminder and
block below do not work there, and the eval suite runs Claude Code only.
That work is SCRATCH-15.

## 1. Hook state

New subcommands in `internal/cli/hook.go`, logic in `internal/hook/`, the
same shape as `acta hook prompt` and `acta hook session-start`.

- `acta hook post-tool` reads the hook JSON on stdin (`session_id`,
  `tool_input.command`). When the command is
  `acta set scratch/<stem> status brainstorming`, it records
  `session_id -> SCRATCH-n` in `.acta/state/sessions.json`.
- `acta hook pre-tool` reads the same JSON. When the command would set a
  different scratch item to `brainstorming` and this session already has
  one, it exits 2 with a message: this session already brainstormed
  SCRATCH-n; offer the three choices from acta:brainstorm. The same item
  again passes.
- `acta hook prompt` also reads the state. When the session has
  brainstormed, it adds one reminder line naming SCRATCH-n.
- `plugin/hooks/hooks.json` gains `PreToolUse` and `PostToolUse` entries
  with matcher `Bash`. Each calls a new shell wrapper that, like
  `prompt-reminder`, never fails when acta is missing.
- `.acta/state/` is added to the gitignore through the existing
  `EnsureGitignore`.
- Any error (bad JSON, unreadable state file) is silent and exits 0, so a
  hook never jams a session. Only the block path exits 2.

Known gap: a brainstorm that never runs `acta set` is not seen by the
hook. The evals catch that.

## 2. plugincheck guards

The existing brainstorm and scratch guards stay as they are. New guards:

- `hooks.json` has `PreToolUse` and `PostToolUse` with matcher `Bash`; the
  wrappers exist, are executable, and end in `exit 0`.
- The land skill names the eval gate, `plugin/skills/`, `plugin/hooks/`
  and `scripts/eval`.
- Each eval case maps to one skill phrase that plugincheck requires, so a
  deleted rule goes red before any paid eval runs.

## 3. Eval suite

`plugin/evals/`, one case per scenario (`case.yaml` plus `graders/`):

1. A side idea during a build goes to scratch, not memory.
2. "catet aja" goes to scratch.
3. A second brainstorm in the same session offers the three choices.
4. An obvious one-file fix does not get an Architectural brainstorm.
5. A brainstorm start files a scratch item first, then sets
   `status brainstorming`.
6. User answers get appended with `acta scratch add`.

Graders check files in the scratch repo and the commands run, since those
are free. An LLM grader (`haiku`) is used only for cases 3 and 4.

`scripts/eval` wraps `claude plugin eval` with `--model sonnet
--ablation none --no-publish --allow-tools Bash`.
`--no-publish` is required: by default the report is uploaded.

The user runs on a subscription, so there is no dollar bill, only the
usage quota. No `--max-cost-usd`. Each case sets a low `max_turns` and
`timeout_seconds` instead, which caps how much quota one run can use.

The first plan task is a spike. It proves the eval sandbox does not load
the global CLAUDE.md or AGENTS.md, since those hide plugin gaps. If it
does load them, stop and bring it back to the user.

## 4. Land gate

`acta:land` runs `scripts/eval` only when both hold: the repo has
`scripts/eval`, and the branch diff touches `plugin/skills/` or
`plugin/hooks/`. A red eval stops the land the same as a red test. Other
merges skip it and cost nothing.

The land skill is shared by every project that uses acta. The
`scripts/eval` check keeps the gate off outside the acta repo, even when
another repo happens to have a `plugin/skills/` folder.

Scope: the hooks in section 1 ship in `plugin/hooks/`, so they run for
every user and project with acta in Claude Code. The eval suite and
`scripts/eval` live in the acta repo only.

## Testing

- Go tests for `post-tool`, `pre-tool` and `prompt`: match and no match,
  same item passes, different item blocks, bad JSON, missing state file,
  missing `session_id`, commands that only look alike (quoted, other
  status, other kind).
- plugincheck tests for the new guards.
- The eval suite itself, run once through `scripts/eval` before landing.
