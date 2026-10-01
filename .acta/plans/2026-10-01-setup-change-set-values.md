---
parent: specs/2026-10-01-setup-change-set-values-design
depth: minimal
id: PLN-0073
created: "2026-10-01 18:30:32"
hash: scaathk
started: "2026-10-01 18:33:13"
finished: "2026-10-01 18:34:40"
---
# Setup offers to change values that are already set Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Setup tells a first run from a later run by what `acta config show` prints, and a later run lets the user change values already set (BUG-0021).

**Spec:** `.acta/specs/2026-10-01-setup-change-set-values-design.md`

**Tests:** `scripts/test ./internal/plugincheck`, `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Only `plugin/skills/setup/SKILL.md` and `internal/plugincheck/skill_setup_test.go` change.
- `MaxLines` in the test stays at most 80.
- Question texts, save commands, the dispatch and subagent-model conditions, the acta block rules and Limits stay as they are; the `acta config set` flag list from "Change later" stays.
- Skill text is short plain English, the same voice as the rest of the skill.

## Waves

- Wave 1: Task 1

### Task 1: Setup run type follows what acta config show prints

**Files:** `plugin/skills/setup/SKILL.md`, `internal/plugincheck/skill_setup_test.go`

**verify:** On every config state (no file, every value `(default)`, set in part, fully set) and every user ask (none, or a named part), the skill text gives exactly one path: nothing set asks every question in order; anything set shows the current setting first, changes only the named part when the user named one, else asks the unset parts and then asks once "Change anything already set? (voice, executor, plan depth, subagent models, acta block)"; the old "only for what `acta config show` says is not set yet" rule cannot come back in any form. List each state and ask checked and the path the text gives.

- [x] Add to `TestSkillSetup` `Must` the new rule sentences (the first-run condition naming `(default)`, "show the current setting", the user-named-part rule, the "Change anything already set?" question) and to `MustNot` the old sentence "only for what `acta config show` says is not set yet"; it fails because the skill still holds the old sentence and none of the new ones.
- [x] Rewrite the First run step 2 and fold "Change later" into one run-type rule as the spec says, raising `MaxLines` to at most 80 only if needed; run `scripts/test ./internal/plugincheck` until green, plus `gofmt -l internal/plugincheck`.
- [x] Commit `setup: tell a first run from a later run and offer to change set values (BUG-0021)`.
