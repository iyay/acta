---
id: SCR-0031
hash: jt3bp7q
title: setup on a partial config never offers to change voice
status: dropped
created: "2026-09-30"
schema: "1"
finished: "2026-09-30"
---
# setup on a partial config never offers to change voice

## Words

### 2026-09-30

# setup on a partial config never offers to change voice

User words (2026-09-30): "curang dong kalo harus acta config. kan user baru mana tau ada acta config"

What happened: ~/.acta/config.yaml held chat_language English, style adhd, repo_language English, no build_executor. The user ran setup in omp to set the language back to Indonesian. The skill's First run path asks "only for what acta config show says is not set yet", so omp asked only the executor and reported "voice: English, adhd — already set". The user got no chance to change the language, and the only way left was the acta config CLI, which a new user does not know about.

Gap in plugin/skills/setup/SKILL.md: nothing says how to tell a first run from a later run. A partial config takes the First run path, and the Change later path ("asks which part to change") never runs.

Idea: when the config file exists, always show the current values and ask which part to change (or keep all), then ask the unset ones. Never route the user to acta config.

### 2026-09-30

User, 2026-09-30: "ya gabisa gitu.. user baru mana tau harus /setup kedua kalinya". A new user must not need a second /setup run either. Requirement: the first setup run already offers every part, set or not, and a plain ask in chat at any time ("reply in Indonesian", in any language) changes the setting without the user knowing any command.

### 2026-09-30

User, 2026-09-30: "setup pertama kali itu harus onboard dengan bener." The first setup run is the onboarding and must walk a new user through every choice properly, in Claude Code and in omp alike.

## Context

## Log

## Open questions
