---
id: SCR-0026
hash: f9qb4dr
title: Plan skill ignores build_executor config
status: raw
created: "2026-09-29"
schema: "1"
---
# Plan skill ignores build_executor config

## Words

### 2026-09-29

kok masih ke subagent sih default-nya? di config udah pake dispath padahal

fix skill catet di scratch

## Context

After PLN-0036 was written, acta:plan asked which executor to use and called subagent the default, while ~/.acta/voice.yaml has `build_executor: dispatch`.
- plugin/skills/plan/SKILL.md:200 asks for the executor and names subagent as default. It never runs `acta voice show`.
- plugin/skills/build/SKILL.md:22 already reads the config, so build itself picks dispatch.
- internal/hook/hook.go:24 session-start index text says "executor subagent (default)".
Proposed fix: plan skill runs `acta voice show` and names the configured executor without asking; hook text says the executor comes from config, else ask.
Source: reading the files, 2026-09-29.

## Log

## Open questions
