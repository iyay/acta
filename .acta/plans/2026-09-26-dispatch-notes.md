---
id: PLN-0002
hash: z2gsoqx
---
# Dispatch checkpoint and closing ticks Implementation Plan

> **For agentic workers:** run this plan with pm:build, task by task. Steps use checkbox (`- [ ]`) syntax, and pmb reads those boxes as task progress.

**Goal:** The dispatching agent reads the recipient exactly once at the checkpoint and then yields, and every plan box is ticked before the reply-back so omp's goal audit never reopens.

**Architecture:** Text changes in `plugin/skills/dispatch/SKILL.md`, `plugin/references/house-rules.md` and `plugin/skills/build/SKILL.md`, guarded by `internal/plugincheck` tests.

**Tech Stack:** Go 1.27 tests over markdown skills.

**Spec:** `.pm/specs/2026-09-26-pm-plugin-design.md` §2 (build, dispatch). Source: dogfood plan 2 (merge ea30e32): the dispatcher polled omp with sleep and until-loops after the checkpoint; omp's /goal audit found 6 unticked boxes after the reply-back and reopened its goal.

**Worktree:** `/Users/iyay/Nayakatara/pm-board-notefix`, branch `notefix`, parent `main`. Executor: `inline`.

## Global Constraints

- Gate before every commit: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
- Edit only the lines each task names. Stage by path. Never push.
- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.

## Waves

- Wave 1: Tasks 1 and 2 (disjoint files).

---

### Task 1: Checkpoint is one read, then yield

**Files:** `plugin/skills/dispatch/SKILL.md` (section `## Comprehension checkpoint`), `internal/plugincheck/skill_dispatch_test.go`

**verify:** no text in the dispatch skill folder lets the dispatcher read the recipient's pane more than once between the `/goal` and the reply-back, or wait in a loop, sleep or poll for it; when the todo list is not visible at that one read, the text says to report the checkpoint as unconfirmed and yield. List every sentence in the folder about reading or waiting for the recipient.

- [x] Step 1: in `skill_dispatch_test.go` add to `Must`: `"exactly one read"`, `"checkpoint unconfirmed"`; add to `MustNot`: `"until herdr agent read"`, `"read it again"`.
- [x] Step 2: run, watch it fail: `go test -count=1 ./internal/plugincheck/ -run TestSkillDispatch`
- [x] Step 3: in `## Comprehension checkpoint`, replace the first line `~20s after the \`/goal\`, one read. Two things must be true in the todo list:` with `About 20s after the \`/goal\`, exactly one read (\`herdr agent read <slug> --source recent-unwrapped --lines 60\`). No loops, no sleeps, no second read. If the todo list is not on screen yet, report "checkpoint unconfirmed" in the Phase 1 report and yield anyway: the reply-back is the real signal. When the todo list is on screen, two things must be true:`. Check the rest of the folder (`herdr-delivery.md`) for any sentence that allows a second read or a wait loop and bring it in line.
- [x] Step 4: gate.
- [x] Step 5: commit: `git add plugin/skills/dispatch internal/plugincheck/skill_dispatch_test.go && git commit -m "fix(plugin): dispatch checkpoint is one read, then yield"`

---

### Task 2: Every box ticked before the reply-back

**Files:** `plugin/references/house-rules.md` (PROGRESS line), `plugin/skills/build/SKILL.md` (hand-off bullets and `## Close`), `internal/plugincheck/plugin_test.go` (`TestHouseRules`), `internal/plugincheck/skill_build_test.go`

**verify:** every place that tells an implementer or a recipient how to tick boxes also says to run `pmb tick <task-id> --all` right after the task's commit, and the recipient is told to check with `pmb show` that every task of the plan shows all boxes done before it sends the reply-back. List every `pmb tick` instruction in the plugin folder checked.

- [x] Step 1: in `TestHouseRules` add `"--all right after"` and `"before the reply-back"` to the required list; in `skill_build_test.go` add `"--all right after"` to `Must`.
- [x] Step 2: run, watch it fail: `go test -count=1 ./internal/plugincheck/ -run 'TestHouseRules|TestSkillBuild'`
- [x] Step 3: append to the PROGRESS line in `house-rules.md`: ` Right after a ticket's commit, run pmb tick <its task id> --all so no box is left open. As the recipient, check with pmb show <plan id> --json that every task shows all boxes done before the reply-back.` In `build/SKILL.md`, add to the hand-off bullet that names `pmb tick`: `; right after the task's commit, run it again with --all right after the commit so no box stays open`, and in `## Close` add `Before the review, every task of the plan shows all its boxes done (\`pmb show\`).`
- [x] Step 4: gate.
- [x] Step 5: commit: `git add plugin/references/house-rules.md plugin/skills/build/SKILL.md internal/plugincheck/plugin_test.go internal/plugincheck/skill_build_test.go && git commit -m "fix(plugin): tick every box with --all after each task commit"`
