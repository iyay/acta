---
id: BUG-0020
hash: sedlyuh
fixed_in: 487e5d1
finished: "2026-09-30"
---
# Setup asks about split subagent models on every run

## Symptom
Every run of acta:setup asks "split subagent models?" again, even after the user answered no before.

## Root cause
plugin/skills/setup/SKILL.md:37 says a no saves nothing. internal/voice/voice.go:177 only allows subagent_models to be "split" or empty. So "user said no" and "never asked" both leave the field empty. The first-run rule asks for every field that acta config show reports as not set, so the question comes back each time.

## Repro
1. Run /acta:setup in Claude Code and answer no to split subagent models.
2. Run acta config show: no subagent_models line.
3. Run /acta:setup again: the same question is asked.

## Found in
main at f1fb03a, found by reading after the user asked why setup keeps asking.
