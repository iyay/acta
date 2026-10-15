---
type: Reference
title: acta under Codex
description: How acta hooks know they run in Codex, and each way Codex hooks differ from Claude Code
paths: [internal/hook/, plugin/hooks/, internal/doctor/codex.go, internal/setup/plan.go]
---

Codex runs the same plugin folder and `hooks/hooks.json` as Claude Code. Facts: `plugin/codex/FACTS.md`.

- Detect Codex with `hook.IsCodex()`: env `PLUGIN_ROOT` is set. Claude Code sets only `CLAUDE_PLUGIN_ROOT`. `AI_AGENT` is not a signal: Codex started from Claude Code inherits it.
- Edits come as `tool_name: "apply_patch"` with the whole patch in `tool_input.command`. Only the `*** Update File:`, `Add File:`, `Delete File:` and `Move to:` lines name files (`patchFiles`); never read the patch as Bash words. The brainstorm and test-runner guards in `internal/cli/hook.go` skip `apply_patch` too: patch text is file content, never a command.
- PreToolUse stdout is ignored by Codex; context must be `hookSpecificOutput.additionalContext` JSON. Exit 2 blocks the same way.
- No output styles: under Codex `SessionStart` appends `codexExtra` and `plugin.OutputStyle()`. The 60-line cap test does not count the style text.
- Doctor reads Codex's `config.toml` line by line (`codexHasPlugin`): comments, whitespace, CRLF and quote styles are folded away first. It is not a TOML parser; unusual header forms read as not installed.
- Plugin hooks run only after the user trusts them in `/hooks`. Doctor can only remind.
- Install is a local marketplace (`codex plugin marketplace add <plugin dir>`); a GitHub add fails because the marketplace is not at the repo root.
- Codex copies the plugin into `$CODEX_HOME/plugins/cache/acta-local/acta/<version>/` at install and loads that copy, not the folder. Hook code runs the `acta` on PATH, so a new binary is enough for hook changes; skill or hooks.json changes need `codex plugin add acta@acta-local` again.
