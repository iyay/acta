---
id: SCR-0035
hash: rm4528t
title: Rename brainstorm and plan to shape and slice, version policy, token budgets
status: brainstorming
created: "2026-10-03 12:59:27"
schema: "1"
started: "2026-10-03 12:59:31"
finished: "2026-10-03 13:04:49"
---
# Rename brainstorm and plan to shape and slice, version policy, token budgets

## Words

### 2026-10-03

User asked (2026-10-03). Step 1 of a five-step roadmap to make acta leaner than superpowers, mattpocock/skills and gstack, without losing quality.

1. Rename the brainstorm and plan skills: the names are too close to superpowers. Picked names: shape (was brainstorm) and slice (was plan). A later optional skill, frame, will come before shape for new products.
2. Rename the skills only. The scratch status `brainstorming` stays as it is.
3. Version: keep 0.1.0 for now. Until the first release, the patch number goes up as changes land.
4. Lock a token budget per skill in plugincheck, so the plugin cannot quietly grow back while later specs make each skill leaner.

## Context

## Log

### 2026-10-03

Answers given in chat before this item was filed:
- Names: shape for brainstorm, slice for plan (frame comes later, optional). Approved.
- Scope: rename the skills only; status `brainstorming` stays.
- Version: stay on 0.1.0 now; the patch goes up as changes land, until release.
- Sonnet runs implementers only; reviewers stay on the orchestrator model.
- Adapted skills are rewritten, never copied 1:1, and must cost fewer tokens without losing quality.
- Goal: acta more efficient than superpowers, mattpocock/skills and gstack; lock per-skill token budgets so it cannot grow back.

### 2026-10-03

Design approved section by section (user: "1 oke, 2 oke, 3 oke"):
1. Rename: skills/brainstorm -> skills/shape, skills/plan -> skills/slice; all live references move; "brainstorm" stays as the activity word (rule 8, status, hook regex, TUI), "plan" stays as the file kind; shape's description keeps "brainstorm"; no alias skills; no_old_names_test guards the old ids.
2. Version: 0.1.x until release; a plan that changes plugin/, cmd/ or internal/ ends with a task that bumps the patch in the three manifests; plugin_test checks the three agree and look like x.y.z; this plan bumps to 0.1.1.
3. Budgets: plugincheck test with byte caps per .md file under plugin/skills and plugin/references, per skill description, and for the SessionStart text; caps start at today's sizes; later specs lower them; the 4240-line total cap goes away.

### 2026-10-03

User approved spec SPC-0067 ("ooke") and gave permission to update ~/.claude/CLAUDE.md. Ruling: the orchestrator edits it right after land, with a backup, so it never names a skill that is not installed yet. ~/.claude/AGENTS.md is an older copy and stays as it is.

## Open questions
