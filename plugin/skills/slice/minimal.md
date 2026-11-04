# Minimal plan

This file holds the rules of a `minimal` plan. The shared rules are in SKILL.md.

- `minimal`: the frontmatter holds `parent:` and `depth: minimal`. The header is `**Goal:**` (one sentence), `**Spec:**`, `**Tests:**`, `## Global Constraints` (the lean line, unless coding_guide is off, plus only what this plan needs) and `## Waves`; no Architecture, Tech Stack, File map or Interfaces. Each task has a title, `**Files:**`, `**Seam:**`, a property-shaped `**verify:**` (each case named once, at the layer that owns it; a layer above is wiring only) and exactly these three boxes, one sentence each, because acta counts them as task progress: `- [ ] Red: <test> fails because <why>`, `- [ ] Change: <code change>`, `- [ ] Commit: <message>`. No code blocks. About 7 lines a task.

Code blocks and No Placeholders apply to `full` plans only. Minimal review: coverage, verify properties, Red, Change and Commit boxes per task in each plan; fix omissions before hand-off.
