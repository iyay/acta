---
parent: specs/2026-10-01-review-note-triage-design
depth: minimal
id: PLN-0070
created: "2026-10-01 15:10:51"
hash: wq5c6bl
---
# Review NOTE Triage Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Review NOTEs are sorted into `[fix]`, `[debt]` and `[note]`, so only `[debt]` NOTEs reach `acta debt new`.

**Spec:** `.acta/specs/2026-10-01-review-note-triage-design.md`

**Tests:** fast `scripts/test ./internal/plugincheck`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Skill text is plain English with short words. Keep every phrase the plugincheck tests already require, unless this plan replaces it.
- The bucket tags are exactly `[fix]`, `[debt]` and `[note]`. The plan section is exactly `## Review notes`.
- The three-round budget and "a round with no BLOCKER does not start" stay as they are.

## Waves

- Wave 1: Task 1, Task 2 (no shared files).

### Task 1: Review skill sorts NOTEs into three buckets

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/review/code-reviewer.md`, `internal/plugincheck/skill_review_test.go`

**verify:** No text in the review skill or the reviewer template still says that every NOTE goes to debt, or that a NOTE is never a task without the `[fix]` exception. List each place checked. The bucket rules, the polish commit on a CLEAN round (self-review, full tests, revert and move to `[debt]` on failure), the `[fix]` NOTEs riding a BLOCKER fix round, and the `## Review notes` section all appear.

- [ ] Failing test: `TestSkillReview` requires `[fix]`, `[debt]`, `[note]`, `## Review notes` and the polish commit phrase, and the template check requires the tag on NOTE lines. It fails because the skill text does not have them yet.
- [ ] Change: rewrite "Finding bar" (the "never a task" line), "Budget" and "Where findings go" in SKILL.md with the spec's bucket rules and flow. Change the NOTEs output format in code-reviewer.md so each NOTE starts with its suggested bucket tag.
- [ ] Commit: `Review sorts NOTEs into fix, debt and note buckets`

### Task 2: Hand-off text says NOTEs are sorted, not all filed

**Files:** `plugin/references/house-rules.md`, `plugin/skills/build/dispatch.md`, `internal/plugincheck/plugin_test.go`, `internal/plugincheck/build_dispatch_test.go`

**verify:** No hand-off text (house rules, dispatch brief, dispatch memory sweep) still claims that the orchestrator files every NOTE with `acta debt new`. List each line checked. "never write NOTEs to memory" stays.

- [ ] Failing test: the house-rules check and the dispatch check forbid "files them with acta debt new" and require "sorts them". They fail on the current text.
- [ ] Change: house-rules.md line 8 and dispatch.md line 262 say the orchestrator sorts NOTEs through `acta:review` once the round is CLEAN. The memory sweep line in dispatch.md points to the same sort.
- [ ] Commit: `Hand-off text says review NOTEs are sorted`
