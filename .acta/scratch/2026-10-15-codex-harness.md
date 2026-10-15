---
id: SCR-0072
hash: auw45fm
title: Codex support
status: brainstorming
created: "2026-10-15 13:31:24"
schema: "1"
started: "2026-10-15 13:31:27"
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

## Open questions
