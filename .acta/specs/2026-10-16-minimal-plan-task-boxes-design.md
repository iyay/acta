---
id: SPC-0143
created: "2026-10-16 12:00:30"
hash: dnpjwst
started: "2026-10-16 12:04:58"
finished: "2026-10-16 12:05:50"
---
# Minimal plan tasks always count toward progress

Status: Bounded, approved by the user in chat on 2026-10-16.

## Why

The board counts a plan's progress from the `- [ ]` boxes under each task. Many minimal plans have tasks with no box at all (PLN-0145 has none), or one box where three were meant. A task with no box never reads as done, so the plan never shows as finished, even after it landed. The slice skill only says "three one-sentence boxes" and shows no example, so agents skip the boxes or merge them into one.

## Part 1: slice skill

File: `plugin/skills/slice/SKILL.md`, the `minimal` line under Plan depth.

- Replace "three one-sentence boxes" with the three literal lines a minimal task must hold:
  - `- [ ] Red: <test> fails because <why>`
  - `- [ ] Change: <code change>`
  - `- [ ] Commit: <message>`
- A test in `internal/plugincheck` fails when the slice skill's minimal rule no longer shows the `- [ ] Red:`, `- [ ] Change:` and `- [ ] Commit:` lines.

## Part 2: board

File: `internal/board/board.go`, `derive()`.

- A task with no boxes whose plan has `finished:` in its frontmatter reads as `done`.
- A task with no boxes whose plan has no `finished:` keeps its current status (`todo`, or `in-progress` when started).
- Tasks with boxes are unchanged.

## Out of scope

- No backfill of old plan files. The board reads them differently; the files stay as they are.
