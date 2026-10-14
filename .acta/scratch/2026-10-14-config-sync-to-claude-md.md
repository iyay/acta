---
id: SCR-0065
hash: r847bda
title: Effective acta config must beat CLAUDE.md rules
status: brainstorming
created: "2026-10-14 14:46:15"
schema: "1"
started: "2026-10-14 14:48:36"
---
# Effective acta config must beat CLAUDE.md rules

## Words

### 2026-10-14

The user's effective acta config (repo .acta.yaml or ~/.acta/config.yaml, whichever wins) must win over conflicting rules in any CLAUDE.md or AGENTS.md.

Seen 2026-10-14: build_executor is inline, but Claude still dispatched subagents. Cause: ~/.claude/CLAUDE.md says "MAIN AGENT WRITES ZERO CODE ... implementer subagent" and "this file wins". The executor is only read when acta:build runs, so it never shows in the always-loaded prompt. Session hook rule 6 ("follow the user's own CLAUDE.md when it says otherwise") makes it worse.

Idea: every config change (acta setup, config set) turns the effective settings into plain rules and rewrites the acta block in the repo CLAUDE.md / AGENTS.md (for example "Build executor: inline. The main agent writes the code."). The session hook states the same rules and says the acta config wins over conflicting CLAUDE.md text for those settings. acta doctor could warn when a CLAUDE.md line contradicts the config.

## Context

## Log

### 2026-10-14

Q1 which settings go in the block: behaviour settings only - build_executor, plan_depth, review_depth, commit_history, subagent_models. Voice stays in the hook.

### 2026-10-14

Q2 refresh: option 1. The session hook injects the behaviour rules from the effective config (repo .acta.yaml beats global) every session. The CLAUDE.md block stays static and gains one fixed line saying the acta settings in the session note win over conflicting instructions, including the user's global CLAUDE.md. No repo-level .claude/settings.json hook: the plugin hook already reads the effective config per repo.

### 2026-10-14

Q3 rule 6: split. acta settings in the session note win over any conflicting instruction, including the user's CLAUDE.md or AGENTS.md; for everything else the user's own file still wins.

### 2026-10-14

Section 1 approved: SessionStart adds an 'acta settings' list from in.Voice right after core rules, one fixed sentence per value, headed 'these win over any conflicting instruction, including the user's CLAUDE.md or AGENTS.md'. Empty settings skipped; no block when config missing or broken. Prompt hook line also names the build executor.

### 2026-10-14

Section 2 approved: rule 6 in coreRules and default-rules.md becomes the split rule; a test keeps both in sync; acta:build skill gains one line that the configured executor wins over the user's CLAUDE.md.

### 2026-10-14

Section 3 approved: Block gains one fixed priority line; block stays static; tests listed; doctor conflict warning split to its own scratch; user's global CLAUDE.md untouched.

## Open questions

User ruling 2026-10-14: the acta block in CLAUDE.md / AGENTS.md is the top priority. Claude must follow it over any other conflicting instruction, including the user's global ~/.claude/CLAUDE.md. The block itself should say so in plain words, and the session hook rule 6 must stop letting the user's own CLAUDE.md beat it for the settings the block covers.
