---
id: SCR-0072
hash: auw45fm
title: Codex support
status: brainstorming
created: "2026-10-15 13:31:24"
schema: "1"
started: "2026-10-15 13:31:27"
finished: "2026-10-15 16:34:08"
---
# Codex support

## Words

### 2026-10-15

User, 2026-10-15: make acta compatible with Codex (bikin acta compatible buat codex).

Known: the pm-plugin spec listed Codex as a later goal. The setup block in AGENTS.md is all a Codex agent sees today (setup-block-refresh spec).

## Context

## Log

### 2026-10-15

Q1 scope, user 2026-10-15: (3) full parity with Claude Code (skills, hooks, output style, acta doctor) plus Codex as a build executor or reviewer (links to the external-agent-reviewer scratch item).

### 2026-10-15

Split, user 2026-10-15: this brainstorm covers A only (acta runs inside Codex: skills, hooks, rules, doctor). B (acta drives Codex from outside as builder or reviewer through codex exec) goes to backlog, folded into SCR external-agent-reviewer.

### 2026-10-15

Codex 0.162.1 facts (research 2026-10-15): same SKILL.md format; plugin system with plugin.json + marketplace (.agents/plugins/marketplace.json), bundles skills + hooks/hooks.json; hooks SessionStart/UserPromptSubmit/PreToolUse/PostToolUse same JSON shape, exit 2 blocks, CLAUDE_PLUGIN_ROOT set; PreToolUse context needs hookSpecificOutput.additionalContext; edits arrive as apply_patch; no Read/MultiEdit tool; no output styles (use SessionStart stdout or AGENTS.md); plugin_hooks flag shows removed, must test.

### 2026-10-15

Q3 install path, user 2026-10-15: (3) both. Codex plugin (plugin.json + marketplace in plugin/) is the main path; acta setup writes ~/.codex/hooks.json and links skills only if plugin hooks do not load. Plan task 1 = spike proving plugin hooks load; fallback built only if they do not.

### 2026-10-15

Q4 output style, user 2026-10-15: (1) SessionStart hook prints the style text in the session note when the harness is Codex; existing AGENTS.md block stays as the minimal fallback.

### 2026-10-15

Q5 Claude-only skill text, user 2026-10-15: (1) make skill text harness-neutral; per-harness subagent and model rules move to the hook session note; plugincheck guards against Claude-only names coming back.

### 2026-10-15

Design section 1 approved 2026-10-15 (packaging + install): Codex plugin.json in plugin/, .agents/plugins/marketplace.json at repo root; shared hooks/hooks.json, add apply_patch matcher; release + plugincheck cover the Codex manifest version; task 1 spike on plugin hooks, acta setup fallback writes ~/.codex/hooks.json + links skills to ~/.agents/skills only if needed; acta doctor checks codex + plugin installed.

### 2026-10-15

Design section 2 approved 2026-10-15 (hooks): acta hook detects Codex (signal fixed in spike, default Claude Code); SessionStart adds output style text + subagent rules for Codex, subagent rules move from skills to the note for Claude Code too; PreToolUse returns wiki hints as hookSpecificOutput.additionalContext, parses apply_patch file headers, exit 2 block stays; PostToolUse/UserPromptSubmit unchanged; Go unit tests with Codex sample JSON + one manual Codex smoke test; evals stay Claude Code only.

### 2026-10-15

Design section 3 approved 2026-10-15 (skills, guards, docs): neutral wording in 5 skills; plugincheck guards no Claude-only names in skills + Codex plugin.json valid and version-matched; subagent executor means Codex subagents in Codex; README Codex install, plugin/codex/FACTS.md, wiki page codex-harness; out of scope: Codex as outside builder/reviewer (SCR-0049), Codex evals.

## Open questions
