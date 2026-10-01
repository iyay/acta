---
parent: bugs/2026-09-30-setup-never-offers-to-change-voice
id: SPC-0064
created: "2026-10-01 18:30:00"
hash: fe9now2
started: "2026-10-01 18:33:13"
finished: "2026-10-01 18:34:40"
---
# Setup offers to change values that are already set

Status: design approved by the user in chat on 2026-10-01. Bounded (text of one skill, `plugin/skills/setup/SKILL.md`, plus its Go test).

## Why

`plugin/skills/setup/SKILL.md:13` tells the first run to ask "only for what `acta config show` says is not set yet". The "Change later" section says a later run asks which part to change, but nothing says how to tell the two runs apart. A config set only in part (voice set, executor missing) takes the first-run path, so voice is never offered, even when the user says they want a new language (BUG-0021).

## Design

1. The run type comes from what `acta config show` prints:
   - **First run**: nothing is set (no file, or every value marked `(default)`). Ask every question in order, as today.
   - **Any other run**: show the current setting from `acta config show` first. When the user named a part to change, change only that part. Otherwise ask the parts that are not set yet, then ask once: "Change anything already set? (voice, executor, plan depth, subagent models, acta block)".
2. The old sentence "and only for what `acta config show` says is not set yet" goes; its job moves into the "ask the parts that are not set yet" step. The "Change later" section folds into the same rule; its `acta config set` flag list stays.
3. Everything else in the skill stays: the question texts, the save commands, the dispatch and subagent-model conditions, the acta block rules, Limits.

## Testing

`TestSkillSetup` in `internal/plugincheck/skill_setup_test.go`, red before the skill changes:
- `Must` gains the new rule sentences: the first-run condition (nothing set, every value `(default)`), "show the current setting", the user-named-part rule, and the "Change anything already set?" question.
- `MustNot` gains the old sentence "only for what `acta config show` says is not set yet".
- `MaxLines` may rise from 76 to at most 80.

The land gate runs `scripts/eval`, because the diff touches `plugin/skills/`.

## Out of scope

A new eval case for setup. Any change to `acta config` itself.
