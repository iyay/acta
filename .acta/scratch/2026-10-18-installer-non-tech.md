---
id: SCR-0075
hash: fcc35yb
title: Installer for non-technical users
status: raw
created: "2026-10-18 21:31:16"
schema: "1"
started: "2026-10-18 21:31:21"
---
# Installer for non-technical users

## Words

### 2026-10-18

User ask (2026-10-18): "gw perlu installer deh biar orang2 non tech tetep bisa pake" - an installer so non-technical people can use acta.

## Context

## Log

### 2026-10-18

Q1 blocker: user picked (1) - non-tech users will not open Terminal at all; they need a double-click installer (.pkg/.dmg style). Today only curl|sh install.sh exists.

### 2026-10-18

Q2 signing: user picked (2) - no Apple Developer ID; ship unsigned, plus an illustrated guide past Gatekeeper. Note: macOS 15+ removed right-click Open; path is System Settings > Privacy & Security > Open Anyway.

### 2026-10-18

Q3 setup questions: user picked (1) - no questions at install; safe defaults, changed later through chat.

### 2026-10-18

Q4 target: users of the Claude desktop app and Codex desktop app. They cannot add skills/plugins today because it needs terminal setup. Installer must make acta work inside those desktop apps.

### 2026-10-18

Research (docs only, apps not installed here): Claude desktop Code tab shares ~/.claude settings+plugins with CLI; Cowork uses account-synced plugins, not ~/.claude. Codex desktop shares ~/.codex config.toml + plugins. GUI PATH for hooks unconfirmed. acta setup needs a terminal (setup_cmd.go:58-62) and finds harnesses only via claude/codex CLI on PATH (setup_cmd.go:105-112), so desktop-only users get nothing today.

### 2026-10-18

User has no Claude desktop app (earlier "muncul" was the CLI). Desktop behaviour stays untested here; design must not depend on GUI PATH.

### 2026-10-18

Parked 2026-10-18 by user before design. Open question when resumed: how to test on a real desktop app (install Claude desktop here / non-tech tester / ship defensive). Leaning: hooks call $HOME/.local/bin/acta by full path, then PATH.

## Open questions
