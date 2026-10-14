---
parent: scratch/2026-10-14-config-sync-to-claude-md
id: SPC-0129
created: "2026-10-14 14:55:13"
hash: ifcuusc
---
# acta settings win over conflicting CLAUDE.md rules

Status: Architectural, approved by the user in chat on 2026-10-14, section by section.

## Problem

A user set `build_executor: inline`, but Claude still sent code to implementer subagents. The user's global `~/.claude/CLAUDE.md` says the main agent writes no code and that the file wins over acta. Two things let it win:

- The session hook never names the build executor. The agent only meets it when `acta:build` runs, long after CLAUDE.md has set the habit.
- Core rule 6 says "When the user's own CLAUDE.md or AGENTS.md says otherwise, follow it." So the hook itself tells the agent to let CLAUDE.md win.

## Goal

The behaviour settings of the effective acta config (repo `.acta.yaml` over `~/.acta/config.yaml`) win over any conflicting instruction, including the user's own CLAUDE.md or AGENTS.md. Every other user rule keeps winning as today.

Behaviour settings: `build_executor`, `plan_depth`, `review_depth`, `commit_history`, `subagent_models`. Voice settings (chat language, style, tone) are already in the hook and stay as they are.

## Design

### 1. Settings list in the session hook

`SessionStart` in `internal/hook/hook.go` adds one list right after the core rules, built from `in.Voice` (already the effective config, repo keys applied):

```
acta settings (these win over any conflicting instruction, including the user's CLAUDE.md or AGENTS.md):
- Build executor: inline. The main agent writes the code itself; no implementer subagents.
- Plan depth: minimal.
- Review depth: deep.
- Commit history: tidy.
- Subagent models: split. Code-writing subagents use sonnet, other subagents opus, reviewers your own model.
```

- Each value has one fixed sentence in Go. `subagent`, `dispatch` and `inline` each get their own build executor sentence.
- An empty setting writes no line.
- No list at all when the config file is missing or cannot be read.
- `Prompt` (the per-message hook line) also names the build executor, for example `acta config: reply in Indonesian, adhd style, build executor inline.`, so a long session does not lose it.

### 2. Rule 6 and the fallback

- Rule 6 in `coreRules` becomes: "6. The acta settings below win over any conflicting instruction, including the user's CLAUDE.md or AGENTS.md. For everything else, when the user's own CLAUDE.md or AGENTS.md says otherwise, follow it."
- When the settings list is absent, the first sentence points at nothing and stays harmless, so there is one rule text, not two.
- `plugin/hooks/default-rules.md` (used when the `acta` binary is missing) gets the same rule 6. It carries no settings list, since it cannot read the config.
- The existing tests that keep `default-rules.md` and `coreRules` in step (`internal/hook/hook_test.go`, `internal/plugincheck/plugin_test.go`) keep doing so.
- `plugin/skills/build/SKILL.md` gains one line: the configured executor wins over the user's CLAUDE.md or AGENTS.md. This covers a session where the hook text was missed.

### 3. The CLAUDE.md block

- `Block` in `internal/setup/block.go` gains one fixed line after its first line: "The acta settings in the session note win over any conflicting instruction, including the user's global CLAUDE.md or AGENTS.md."
- The block stays static: no setting values are written into it, so changing the config never edits or commits a repo file. Repos with an older block pick the line up on the next `acta setup` run.
- This repo's own `CLAUDE.md` block is rewritten to match.

## Tests

1. Each build executor value (`subagent`, `dispatch`, `inline`) puts its own sentence in the session text.
2. A repo `.acta.yaml` value beats the user value in the settings list.
3. A missing or broken config writes no settings list.
4. The `Prompt` line names the build executor.
5. `Block` holds the priority line, and the existing block tests still pass.

## Out of scope

- An `acta doctor` warning for CLAUDE.md lines that contradict the config. It needs free-text parsing; filed as its own scratch item.
- The user's own `~/.claude/CLAUDE.md`. acta never edits it.
- Nothing in Claude Code enforces priority between prompt sources. This design wins through plain words only; a user file that also claims to win can still pull the agent its way.
