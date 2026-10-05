---
parent: specs/2026-10-05-land-reuse-gates-planning-only
depth: minimal
id: PLN-0091
created: "2026-10-05 15:52:56"
hash: n17dmh7
---
# Land reuses gates for planning-only differences Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Land skips the gate re-run when the merge adds only planning files.

**Spec:** `.acta/specs/2026-10-05-land-reuse-gates-planning-only.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The skill names the planning root in general words (`.acta/` by default), never this repo's own setup.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: land step 6

**Files:** Modify `plugin/skills/land/SKILL.md`, and `internal/plugincheck/budget_test.go` only if the size cap needs it.
**verify:** For every merge result (same tree, only planning-root files differ, any other file differs, Markdown outside the planning root differs) the land step says exactly whether to reuse or re-run the gates, and only the planning-only case reuses them; `internal/plugincheck` passes. List each case checked against the new text.
- [ ] Failing test: `scripts/test ./internal/plugincheck` green before; read step 6 and confirm it re-runs on any tree difference.
- [ ] Code: rewrite step 6 per the spec, short.
- [ ] Commit: `skills: land reuses gates when only planning files differ`.

### Task 2: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [ ] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [ ] Code: add 1 to the patch in all three files.
- [ ] Commit: `plugin: bump patch version`.
