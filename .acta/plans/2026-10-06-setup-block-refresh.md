---
parent: specs/2026-10-06-setup-block-refresh-design
depth: minimal
id: PLN-0097
created: "2026-10-06 05:03:29"
hash: z2a2jjb
started: "2026-10-06 05:05:28"
---
# Setup block refresh Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The acta block that `acta:setup` writes names the wiki, `acta wiki match`, `acta state` and the scratch triage rule, and this repo's block matches it word for word.

**Spec:** `.acta/specs/2026-10-06-setup-block-refresh-design.md`

**Tests:** `scripts/test ./internal/plugincheck`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Block text fits any user and any repo: no acta-repo paths, no user names, no hand-off file line.
- `plugin/skills/setup/SKILL.md` is at its 80-line cap now; the block adds four lines, so the cap in `skill_setup_test.go` goes to 84 and no other line of the skill changes.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: New block text in the setup skill

**Files:** `plugin/skills/setup/SKILL.md`, `internal/plugincheck/skill_setup_test.go`, `CLAUDE.md`

**verify:** No copy of the acta block in the repo (setup skill template, this repo's `CLAUDE.md`) can differ from the other or miss any of the five spec rules; list every place the block text lives and what each holds.

- [x] Failing test: add `.acta/` holds-the-wiki line, `acta wiki match`, `acta state`, `acta scratch new` and "never dropped" to the setup skill's `Must` list; it fails because the template has none of them but `acta scratch new`.
- [x] Code: replace the lines between `<!-- acta:begin -->` and `<!-- acta:end -->` in the setup skill with the first line plus the five spec rules, raise `MaxLines` to 84, and replace this repo's `CLAUDE.md` acta section with the same text word for word (drop its hand-off line).
- [x] Commit: `feat(setup): acta block names the wiki, live state and scratch triage (SPC-0087)`

### Task 2: Version bump

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three version files agree on 0.1.20 and `internal/plugincheck` passes; list each file and its version.

- [ ] Failing test: none new; `scripts/test ./internal/plugincheck` already fails if the three files disagree, so bump all three in one edit and run it.
- [ ] Code: change 0.1.19 to 0.1.20 in all three files.
- [ ] Commit: `chore(plugin): version 0.1.20`
