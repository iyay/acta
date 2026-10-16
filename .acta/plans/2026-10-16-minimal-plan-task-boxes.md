---
parent: specs/2026-10-16-minimal-plan-task-boxes-design
depth: minimal
id: PLN-0155
created: "2026-10-16 12:01:04"
hash: itt61cm
---
# Minimal plan tasks always count toward progress Implementation Plan

**Goal:** Minimal plans always carry three task boxes, and a finished plan's box-less tasks read as done on the board.

**Spec:** .acta/specs/2026-10-16-minimal-plan-task-boxes-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- The three box lines are exactly `- [ ] Red: <test> fails because <why>`, `- [ ] Change: <code change>`, `- [ ] Commit: <message>`.

## Waves

- Wave 1: Task 1, Task 2

### Task 1: slice skill shows the three minimal task boxes

**Files:** `plugin/skills/slice/SKILL.md`, `internal/plugincheck/skill_slice_test.go`

**Seam:** the slice SKILL.md text read by plugincheck tests

**verify:** The slice skill's minimal rule cannot lose any of the `- [ ] Red:`, `- [ ] Change:`, `- [ ] Commit:` lines without a plugincheck test failing; list each line checked.

- [ ] Red: a new test in `skill_slice_test.go` fails because the skill does not show the three box lines.
- [ ] Change: replace "three one-sentence boxes" in the minimal rule with the three literal box lines.
- [ ] Commit: `feat(slice): minimal plan tasks show their three boxes`

### Task 2: box-less tasks of a finished plan read as done

**Files:** `internal/board/board.go` (`derive`), `internal/board/board_test.go`

**Seam:** `board.Load` and the task item's `Status`

**verify:** A task with no boxes reads `done` only when its plan has `finished:`; on every other path (plan not finished, not started, started, or task with boxes) the status is what it was before. List every path checked.

- [ ] Red: a board test with a finished plan whose task has no boxes fails because the task reads `todo`.
- [ ] Change: in `derive`, a task with `Total == 0` whose plan has a `Finished` stamp gets status `done`.
- [ ] Commit: `fix(board): box-less tasks of a finished plan read as done`
