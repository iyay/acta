---
parent: scratch/2026-10-15-codex-harness
id: SPC-0139
created: "2026-10-15 16:34:08"
hash: ks4azfr
---
# Run acta inside Codex

Status: Architectural, approved in chat on 2026-10-15, section by section.

Why: acta works in Claude Code and omp. In Codex (codex-cli 0.162.1) an agent only sees the AGENTS.md block, so it skips skills and rules. Codex now reads the same SKILL.md format, has plugins, and runs hooks with the same JSON shape as Claude Code, so acta can run there with few changes.

Goal: a user who opens Codex in an acta repo gets the acta skills, the session note, the output style, wiki hints and the second-brainstorm block, the same as in Claude Code.

Not goals: Codex as a builder or reviewer driven from outside (codex exec); that is SCR-0049. Plugin evals in Codex; evals stay Claude Code only.

## Facts this design rests on

Measured on 2026-10-15 against Codex 0.162.1 and its docs (developers.openai.com/codex: skills, plugins, hooks):

- Skills: same SKILL.md frontmatter (`name`, `description`). Plugins bundle `skills/<name>/SKILL.md`.
- Plugins: manifest `plugin.json` at the plugin root; a repo marketplace lives at `.agents/plugins/marketplace.json`. Install with `codex plugin marketplace add <repo>`.
- Hooks: SessionStart, UserPromptSubmit, PreToolUse, PostToolUse, same matcher-group JSON as Claude Code. Exit 2 blocks a tool. Codex sets `CLAUDE_PLUGIN_ROOT`.
- SessionStart and UserPromptSubmit stdout becomes context. PreToolUse stdout is ignored; context must come back as `hookSpecificOutput.additionalContext`.
- File edits arrive as `tool_name: "apply_patch"` with the patch text in `tool_input.command`. There is no Read or MultiEdit tool.
- No output styles.
- Open risk: `codex features list` shows `plugin_hooks` as removed, while the docs say plugin hooks load. Task 1 settles it.

## Section 1: packaging and install

1. Add a Codex `plugin.json` to `plugin/`. Add `.agents/plugins/marketplace.json` at the repo root, pointing at `plugin/`.
2. Keep one `plugin/hooks/hooks.json` for both harnesses. Matchers Codex never fires (`Read`, `MultiEdit`, `PowerShell`) stay. Add `apply_patch` to the PreToolUse matcher.
3. `scripts/release` bumps the Codex manifest version too. `internal/plugincheck` fails when any of the four version files disagree.
4. Task 1 is a spike: install the plugin in Codex and prove whether plugin hooks fire, and find the signal that tells `acta hook` it runs under Codex. Results go to `plugin/codex/FACTS.md`.
5. Only if plugin hooks do not fire: `acta setup` writes the acta hooks into `~/.codex/hooks.json` and links the skills into `~/.agents/skills`. If they fire, this fallback is not built.
6. `acta doctor` checks Codex when `codex` is on PATH: the acta plugin is installed (or the fallback hooks are present).

## Section 2: hooks under Codex

1. `acta hook` detects Codex from the signal task 1 finds (an env var or an input field such as `turn_id`). No signal means Claude Code, so today's behaviour does not change.
2. SessionStart under Codex adds the output style text (from `plugin/output-styles/`) to the session note, since Codex has no output styles. The note also carries the subagent rules for the harness: Codex subagents, model from `[agents]` in `config.toml`. For Claude Code the subagent model rules move from the skills into the note too.
3. PreToolUse returns wiki hints as `hookSpecificOutput.additionalContext` JSON. Claude Code accepts the same shape, so one format serves both.
4. PreToolUse with `tool_name: "apply_patch"` takes the file paths from the patch headers (`*** Update File:`, `*** Add File:`, `*** Delete File:`) and runs the wiki match on them.
5. The second-brainstorm block stays exit 2. PostToolUse and UserPromptSubmit do not change.
6. Tests: Go unit tests feed sample Codex hook input JSON. One manual smoke run in Codex, recorded in `plugin/codex/FACTS.md`.

## Section 3: skills, guards, docs

1. Skills `slice`, `shape`, `review`, `build`, `debug` lose Claude Code names (`Agent` tool, `model: "sonnet"` or `"opus"`, `subagent_type`). They say "your harness's subagents" and "the model the session note names". The rules mean the same.
2. plugincheck gains two guards: no Claude-only names in `plugin/skills/**`; the Codex `plugin.json` parses and its version matches the other three files.
3. Build executor `subagent` means Codex subagents when the harness is Codex. `dispatch` and `inline` do not change.
4. Docs: a Codex install part in `plugin/README.md`; `plugin/codex/FACTS.md`; a wiki page `codex-harness` (how acta detects Codex, how its hooks differ). The wiki page is approved with this spec; acta:build writes it.

## Success check

- In Codex, a new session in this repo shows the acta session note with the output style text.
- Editing a file covered by a wiki page shows its hint in Codex.
- A second `acta set scratch/... status brainstorming` in one Codex session is blocked.
- `scripts/test --full` passes, with the new plugincheck guards.
- Claude Code behaviour is unchanged: `scripts/eval` passes.
