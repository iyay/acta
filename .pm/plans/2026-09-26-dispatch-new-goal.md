# Dispatch sends /new and /goal as two confirmed steps Implementation Plan

> **For agentic workers:** run this plan with pm:build, task by task. Steps use checkbox (`- [ ]`) syntax, and pmb reads those boxes as task progress.

**Goal:** Every dispatch to omp waits for omp to be ready, sends `/new` alone and sees "New session started", then sends `/goal` and sees `🎯 Goal` in the status bar; a fix round sends `/goal` only; `/new` and `/goal` never go in one prompt.

**Architecture:** Text change in the dispatch skill (`SKILL.md` section on `/new` and `/goal`, and `herdr-delivery.md` "Deliver the pointer message"), guarded per file by a Go test in `internal/plugincheck`.

**Tech Stack:** Go tests over markdown skills.

**Spec:** none (Bounded, approved in chat on 2026-09-26)

**Source:** plan spec-line-md: `/new` was sent before omp's input was ready, so `/new` and `/goal` landed as one message, the goal never set, and the todo list lost the ticket ids. The old text said "Back-to-back, no settle-wait".

**Worktree:** created by pm:build at `../pm-board-dispatch-new-goal`, branch `dispatch-new-goal`, parent `main`.

## Global Constraints

- Gate before every commit: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
- Edit only the files the task names. Stage by path. Never push. Never commit the plan file. gofmt goes into the task commit, never a commit of its own.
- Plain English a ten-year-old can read. No marker tags.
- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.

## Waves

- Wave 1: Task 1.

---

### Task 1: /new and /goal are two confirmed steps

**Files:**
- Modify: `plugin/skills/dispatch/SKILL.md` (section "## `/new` and `/goal` — every dispatch, every backend")
- Modify: `plugin/skills/dispatch/herdr-delivery.md` (section "## Deliver the pointer message")
- Test: `internal/plugincheck/skill_dispatch_test.go`

**verify:** No text in the dispatch skill tells an agent to send `/new` and `/goal` together, back to back without a check, or before omp is ready; both files give the same four steps with a visible check after each; a fix round is `/goal` only; and each file has its own test that goes red if that file alone is reverted. List every place in both files that mentions `/new` or `/goal` and what each says.

**Interfaces:**
- Consumes: nothing.
- Produces: nothing.

- [x] **Step 1: Write the failing test**

Add to `internal/plugincheck/skill_dispatch_test.go`:

```go
// TestDispatchNewThenGoal reads each dispatch file on its own. /new and
// /goal once landed as one message, and the goal never set.
func TestDispatchNewThenGoal(t *testing.T) {
	wants := []string{"HARD RULE", "New session started", "🎯 Goal", "never in one prompt", "fix round: `/goal` only"}
	for _, file := range []string{"SKILL.md", "herdr-delivery.md"} {
		b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "dispatch", file))
		if err != nil {
			t.Fatal(err)
		}
		txt := string(b)
		for _, want := range wants {
			if !strings.Contains(txt, want) {
				t.Errorf("dispatch/%s missing %q", file, want)
			}
		}
		if strings.Contains(txt, "Back-to-back, no settle-wait") {
			t.Errorf("dispatch/%s still says \"Back-to-back, no settle-wait\"", file)
		}
	}
}
```

- [x] **Step 2: Run it to see it fail**

Run: `go test ./internal/plugincheck -run TestDispatchNewThenGoal -count=1`
Expected: FAIL naming missing strings in both files and the old "Back-to-back" line in SKILL.md.

- [x] **Step 3: Fix `plugin/skills/dispatch/SKILL.md`**

In the section "## `/new` and `/goal` — every dispatch, every backend", replace the line starting `Order: \`/new\` → \`/goal` (the one that ends "Back-to-back, no settle-wait. Checkpoint after the `/goal`.") with:

```markdown
**Order — HARD RULE, four steps, a check after each:**

1. Wait until omp is ready: the pane shows its empty input box (`herdr agent read <slug> --source visible`). A just-launched omp is not ready yet.
2. Send `/new` alone. Read the pane until it shows "New session started". Then rename the agent by pane id (`/new` drops the name).
3. Send `/goal` alone. Read the status bar until it shows `🎯 Goal`. `⏸ Goal` or no goal → see the held-goal block below and send it again.
4. fix round: `/goal` only — no `/new`; still check `🎯 Goal`.

`/new` and `/goal` go as two prompts, never in one prompt, and `/new` never goes before omp is ready: otherwise both land as one message and the goal never sets. These reads are part of delivery, not the comprehension checkpoint. Checkpoint after the `/goal`.
```

- [x] **Step 4: Fix `plugin/skills/dispatch/herdr-delivery.md`**

In "## Deliver the pointer message", replace the line `New session, plan, or worktree — clear first, then set the goal:` and the code block right after it with:

````markdown
New session, plan, or worktree — HARD RULE, four steps, a check after each (same as SKILL.md):

```bash
# 1. omp ready: its empty input box is on screen
herdr agent read <slug> --source visible --lines 10
# 2. /new alone, then confirm "New session started", then rename by pane id
herdr agent prompt <slug> "/new"
herdr agent read <slug> --source visible --lines 10    # must show: New session started
herdr agent rename <pane-id> <slug>
# 3. /goal alone, then confirm 🎯 Goal in the status bar
herdr agent prompt <slug> "/goal ultrathink orchestrate <one-line summary>. FIRST read the hand-off at <abs-brief-path> and obey every line — it names the plan and ticket ids, the worktree, the gates. You are the main agent here: write ZERO code yourself. Run pm:build with as MANY implementer subagents as the tickets allow: one per ticket MINIMUM, each driving pm:tdd — failing test first. Review happens on my side, not yours. Group tickets into waves by file ownership and dispatch every wave in ONE message, several subagents at once; serial only for a shared file or a real dependency. Declare the waves in your todo list before dispatching. Do NOT move, park, close, or create any pane or tab. When your last ticket is committed, run this VERBATIM: herdr agent prompt $HERDR_PANE_ID \"/pm:review <base>..<new-head> — plan <path>, round <slug>, pane \$HERDR_PANE_ID\". Each verify line is a PROPERTY that must hold on EVERY path, not the one case reported: enumerate every path, caller and surface that could break it and report that list. Line numbers in the brief are hints, never boundaries — confirm with lsp and grep. Read the hand-off, AGENTS.md and the plan BEFORE writing any todo list."
herdr agent read <slug> --source visible --lines 6     # status bar must show: 🎯 Goal
```

`/new` and `/goal` are never in one prompt, and `/new` never goes to an omp that is still starting.
````

Then change the line `Fix round on the plan the agent already holds — \`/goal\` alone, same inline tail:` to `Step 4, fix round: \`/goal\` only, on the plan the agent already holds, same inline tail; still confirm 🎯 Goal:`.

- [x] **Step 5: Run the tests to see them pass**

Run: `go test ./internal/plugincheck -count=1`
Expected: PASS (the existing `TestSkillDispatch` MaxLines 620 still holds).

- [x] **Step 6: Prove each file is guarded on its own**

In a scratch copy, revert only `plugin/skills/dispatch/SKILL.md` (`git stash push plugin/skills/dispatch/SKILL.md`), run `go test ./internal/plugincheck -run TestDispatchNewThenGoal -count=1`, see FAIL for `dispatch/SKILL.md`, then `git stash pop`. Do the same for `herdr-delivery.md`.

- [x] **Step 7: Run the gate, then commit**

```bash
git add plugin/skills/dispatch/SKILL.md plugin/skills/dispatch/herdr-delivery.md internal/plugincheck/skill_dispatch_test.go
git commit -m "fix(plugin): dispatch sends /new and /goal as two confirmed steps"
```

Then run `pmb tick plans/2026-09-26-dispatch-new-goal#task-1 --all`.

## Fix round 1

### Task F1: Put back the fix-round example

**Files:**
- Modify: `plugin/skills/dispatch/herdr-delivery.md` (the line "Step 4, fix round: ..." in "## Deliver the pointer message")
- Test: `internal/plugincheck/skill_dispatch_test.go` (func `TestDispatchNewThenGoal`)

**verify:** herdr-delivery.md gives a copyable fix-round command right under the "Step 4, fix round" line, with the `<fixed-from>..<new-head>` review range, and a test goes red if that command is removed again. List every fix-round mention in both dispatch files and what command each points to.

Review round 1 BLOCKER: commit 79e81ab deleted the fix-round bash block under that line; the plan only asked to reword the line. `grep -rn fixed-from plugin/skills` finds nothing.

- [x] **Step 1: Write the failing test**

In `TestDispatchNewThenGoal`, after the loop, add:

```go
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "dispatch", "herdr-delivery.md"))
	if err != nil {
		t.Fatal(err)
	}
	// The fix-round command was once deleted by accident. It is the only one
	// with the fix range, so an agent on a fix round needs it.
	if !strings.Contains(string(b), "/pm:review <fixed-from>..<new-head>") {
		t.Error("dispatch/herdr-delivery.md lost the fix-round command with <fixed-from>..<new-head>")
	}
```

- [x] **Step 2: Run it to see it fail**

Run: `go test ./internal/plugincheck -run TestDispatchNewThenGoal -count=1`
Expected: FAIL with "lost the fix-round command".

- [x] **Step 3: Put the block back**

Directly under the line `Step 4, fix round: \`/goal\` only, on the plan the agent already holds, same inline tail; still confirm 🎯 Goal:` insert this block (it is the one deleted in 79e81ab, word for word, plus the check line):

````markdown
```bash
herdr agent prompt <slug> "/goal ultrathink orchestrate <one-line summary>. FIRST read <abs-brief-path>. Fix the PROPERTY, not the reported case: enumerate every path that could break it. Tickets <ids> → one implementer subagent each (pm:build, pm:tdd inside), all independent ones in ONE message; declare waves first. When your last ticket is committed, run VERBATIM: herdr agent prompt $HERDR_PANE_ID \"/pm:review <fixed-from>..<new-head> — plan <path>, round <slug>, pane \$HERDR_PANE_ID\""
herdr agent read <slug> --source visible --lines 6     # status bar must show: 🎯 Goal
```
````

- [x] **Step 4: Run the tests, then the gate**

Run: `go test ./internal/plugincheck -count=1`, then `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
Expected: PASS (dispatch folder stays under MaxLines 620).

- [x] **Step 5: Commit**

```bash
git add plugin/skills/dispatch/herdr-delivery.md internal/plugincheck/skill_dispatch_test.go
git commit -m "fix(plugin): restore the fix-round dispatch command"
```

Then run `pmb tick plans/2026-09-26-dispatch-new-goal#task-F1 --all`.
