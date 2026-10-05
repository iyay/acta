---
created: "2026-10-06 05:02:12"
id: SPC-0087
hash: l3m53se
started: "2026-10-06 05:05:28"
finished: "2026-10-06 05:07:49"
---
# Refresh the acta block that setup writes

Status: Bounded, approved in chat on 2026-10-06.

Why: the block `acta:setup` writes into CLAUDE.md or AGENTS.md was last changed in ab94b15 (first-run setup). Since then acta gained the wiki, live work state and the scratch triage rule, and the block says none of it. Agents in a harness without acta hooks (Codex, for one) only see this block, so they keep putting project knowledge in their own memory. The user also moved their personal memory rules out of their global CLAUDE.md and wants the block to carry them.

## Change

`plugin/skills/setup/SKILL.md`, the block between `<!-- acta:begin -->` and `<!-- acta:end -->`, holds five rules:

1. `.acta/` holds specs, plans, bugs, debt, scratch items and the wiki.
2. Project knowledge (gotchas, runbooks, decisions with their why) goes to `.acta/wiki/`, never to agent memory. Write a page only when a fresh agent would lose time or repeat a mistake without it. When a fact changes, rewrite its page.
3. Before changing a file, run `acta wiki match <file>` and read each page it names.
4. Work in flight goes to `acta state set <plan id> next` with the text on stdin (read it back with `acta state <plan id>`), not to agent memory. Agent memory keeps only the user's own setup.
5. Raw ideas go to Scratchpad with `acta scratch new`. A finished scratch item is specced, never dropped; dropped means not done or not valid.

The first line ("This repo uses the acta plugin...") stays. Every line fits any user and any repo: no acta-repo paths, no user names.

The acta block in this repo's `CLAUDE.md` is replaced with the new block, word for word. Its hand-off file line goes away: acta has no hand-off file, and `acta state` covers work in flight.

## Out of scope

- Repos that already have a block get the new text only when setup runs again (it replaces the text between the markers).
- No hook text change; the session hook already prints its own wiki line.

## Testing

- Red first: a test in `internal/plugincheck/skill_setup_test.go` that the block in the setup skill names `.acta/wiki/`, `acta wiki match`, `acta state` and `acta scratch new`.
- Word count of the setup skill before and after, under the plugincheck size cap.
- Land gate runs `scripts/eval`, since the diff touches `plugin/skills/`.
- Version 0.1.19 to 0.1.20 in the three version files.
