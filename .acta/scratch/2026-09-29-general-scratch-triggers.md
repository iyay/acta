---
id: SCR-0016
hash: luodccp
title: general-scratch-triggers
status: raw
created: "2026-09-29"
finished: "2026-09-30"
---
# Scratch trigger words must be general, not one user's slang

User ruling 2026-09-29: the plugin ships to every user, so the scratch trigger examples must not be the user's own Indonesian words ("catet", "nanti", "kepikiran"). Replace them with general examples (for example "note this", "later", "idea for later"), and say the agent should recognise the same intent in any language.

Places found:
- plugin/skills/scratch/SKILL.md line 3 (description) and line 14
- plugin/hooks/default-rules.md line 10
- internal/hook/hook.go line 30 (Skills index text) and internal/hook/hook_test.go line 36
- internal/plugincheck/skill_scratch_test.go line 11 (Must list pins catet, nanti, kepikiran)

Keep: internal/write/scratch_test.go line 46 uses "catet aja dulu" as a unicode body fixture; it tests byte handling, not behaviour.

Related: PLAN-27 Task 8 renames the catet eval case to note-to-scratch. The eval's guard phrase is "note this", so this change must keep that phrase in the scratch skill.
