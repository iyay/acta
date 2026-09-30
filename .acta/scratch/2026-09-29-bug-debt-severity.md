---
id: SCR-0011
hash: gij3qht
title: Severity or priority for bugs and debt
status: brainstorming
created: "2026-09-29"
started: "2026-10-01 05:06:42"
finished: "2026-10-01 05:15:46"
---
User 2026-09-29, verbatim: "lalu buat bugs dan debt, kayaknya perlu ada saverity/priority deh"

Meaning: bug files (.acta/bugs/) and debt items (.acta/debt/, DEBT-n.m) should carry a severity or priority, so the most important ones can be picked first.

Open questions for the brainstorm: severity, priority, or both; which levels; who sets it (acta bug new / acta debt new flag, reviewer, acta set); per debt file or per DEBT item; how the TUI shows and sorts it. Brainstorm in a new session; likely touches the file contract, the CLI and the TUI (after SPEC-12 lands).

2026-10-01: user ruled to bundle SCR-0023 (build always uses git worktree add, drop native EnterWorktree) into this brainstorm. One spec: parent SCR-0011, closes SCR-0023. SCR-0023 design approved in chat: drop step 1a in build/SKILL.md, git worktree add is the only way, one line on why, plugincheck test fails if build/SKILL.md names EnterWorktree or native worktree.

Q1 field: user picked one field 'priority' with levels high / medium / low (no severity, no two fields).

Q2 placement: debt priority is per item, inline on the checkbox line (for example '- [ ] (high) ...'); bug priority is a frontmatter field.

Q3 when set: optional at creation (acta bug new --priority; reviewer may prefix a NOTE with (high)/(medium)/(low), acta debt new validates the level); changeable later with acta set for bugs and one command for debt items; unset stays unset, no migration of old files.

Q4 TUI: Bugs and Debt boxes sort by priority first (high, medium, low, unset), then the current id order; s still flips old/new inside each group; row shows a small colored tag H/M/L, none when unset; no new key.

Approach + section 1 approved: debt tag is '(high)'/'(medium)'/'(low)' right after the checkbox, read into Item.Priority and stripped from Title; bug uses frontmatter priority:; bad bug value goes to Problems and counts as unset; bad debt tag like (hgh) stays plain text; one command acta set <id> priority high|medium|low|none for bugs and debt items (none removes), auto-commits.

Section 2 approved: acta bug new --priority (optional, bad value refused, no file written); acta debt new keeps a leading (high)/(medium)/(low) tag per stdin line; acta set <id> priority on bugs and debt items, debt edit via a helper next to MarkItem in internal/write/mark.go, other kinds refused with 'priority is only for bugs and debt items'; review skill line 51 lets reviewers prefix a NOTE with (high); bug skill names --priority.

Section 3 approved: byPriority stable sort after ordered() in internal/tui/order.go, used only in openRows for Bugs and Debts tabs, Done pane unchanged; H/M/L tag after the id in rowText, colors from existing ANSI slots (H bright red, M yellow, L dim), no new theme field; detail meta line PRIORITY when set. Tests: board parse, write commands, tui sort/tag/meta, plugincheck for SCR-0023.
