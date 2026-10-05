---
parent: bugs/2026-10-05-failed-brainstorm-set-blocks-the-real-one
depth: minimal
id: PLN-0089
created: "2026-10-05 14:54:48"
hash: rdjobo4
started: "2026-10-05 14:57:26"
finished: "2026-10-05 15:03:23"
---
# Brainstorm hook real items Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** A mistyped brainstorm command no longer blocks the real one in the same session.

**Spec:** `.acta/specs/2026-10-05-brainstorm-hook-real-items.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The hook never blocks a real user command on a guess; when unsure, it lets the command run.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: count only existing scratch items

**Files:** Modify `internal/hook/session.go`, `internal/hook/session_test.go`.
**verify:** No brainstorm command whose stem has no `scratch/<stem>.md` is ever recorded or ever causes a block, and two different existing items in one session are still blocked. List each sequence checked: mistyped then real, real then mistyped, same real twice, two different real items.
- [x] Failing test: a temp acta root with `scratch/2026-10-05-x.md`; record `acta set scratch/x status brainstorming`, then PreTool on `acta set scratch/2026-10-05-x status brainstorming` must not block; fails because the mistyped stem was recorded.
- [x] Code: a small check that `<root>/scratch/<stem>.md` exists (Lstat, regular file) used by RecordBrainstorm before writing state and by PreTool before blocking.
- [x] Commit: `hook: count a brainstorm only for a scratch item that exists`.

### Task 2: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [x] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [x] Code: add 1 to the patch in all three files.
- [x] Commit: `plugin: bump patch version`.

## Review notes

- No test covers a directory named `<stem>.md`; the IsRegular check handles it.
- TestReminderNamesItem now accepts "one" or "SCRATCH-1"; TestBlockAndReminderNameScratchID still pins the id.
- A scratch file deleted after it was recorded no longer matters to the record or block paths; acta refuses the command itself.
