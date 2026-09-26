# pm plugin three NOTEs Implementation Plan

> **For agentic workers:** run this plan with pm:build, task by task. Steps use checkbox (`- [ ]`) syntax, and pmb reads those boxes as task progress.

**Goal:** Close three review NOTEs from the follow-ups plan: the first-run voice text respects a language the user's CLAUDE.md already names, the `build` git fallback starts the branch from the recorded parent, and `dispatch` sends a BLOCKER in old code to `pm:bug` instead of the fix ticket.

**Architecture:** Text edits only. One Go string constant in `internal/hook`, two skill files. Each edit is guarded by a new required string in the existing tests.

**Tech Stack:** Go 1.27 (module `pm-board`), markdown skills.

**Spec:** Design approved in chat on 2026-09-26 (Bounded, no written spec). Rulings: (1) CLAUDE.md or AGENTS.md already naming a chat language or style means the agent offers to save those values instead of asking the three questions. (2) The fallback `git worktree add` gets the recorded parent as its start point. (3) `dispatch` follows `pm:review` "Where findings go": a defect in code already on the parent branch goes to `pm:bug`, never into the fix ticket.

**Worktree:** `/Users/iyay/Nayakatara/pm-board-notes3`, branch `notes3`, parent `main`. Executor: `subagent`.

## Global Constraints

- Go tests always with `-count=1`. Gate before every commit: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
- Edit only the lines each task names. Stage by path. Never push.
- Skill line caps still hold (checked by the existing tests).
- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.

## File map

| Path | Task |
|---|---|
| `internal/hook/hook.go` (`firstRun`), `internal/hook/hook_test.go` (`TestSessionStartFirstRun`) | 1 |
| `plugin/skills/build/SKILL.md` (Step 1b), `internal/plugincheck/skill_build_test.go` | 2 |
| `plugin/skills/dispatch/SKILL.md` ("After a review", step 1), `internal/plugincheck/skill_dispatch_test.go` | 3 |

## Waves

- Wave 1: Tasks 1, 2 and 3 (disjoint files).

---

### Task 1: First-run voice text respects CLAUDE.md

**Files:**
- Modify: `internal/hook/hook.go` (the `firstRun` constant)
- Test: `internal/hook/hook_test.go` (`TestSessionStartFirstRun`)

**verify:** when no voice file exists, the session text never tells the agent to ask the three questions, or to write in English, without first saying that a chat language or style already named in the user's CLAUDE.md or AGENTS.md wins and is only offered for saving. List every line of `firstRun` checked.

- [ ] **Step 1: Write the failing test.** In `TestSessionStartFirstRun`, add `"already names a chat language or style"` and `"or the language CLAUDE.md names"` to the `want` list:

```go
for _, want := range []string{"Voice: not set up yet.", "pmb voice set --language", "Style (ADHD reader):",
	"already names a chat language or style", "or the language CLAUDE.md names"} {
```

- [ ] **Step 2: Run it, watch it fail:** `go test -count=1 ./internal/hook/ -run TestSessionStartFirstRun` — expected FAIL: `first run missing "already names a chat language or style"`.

- [ ] **Step 3: Replace `firstRun` with:**

```go
const firstRun = `
Voice: not set up yet. If the user's CLAUDE.md or AGENTS.md already names a chat language or style, do not ask the questions below: offer once to save those values with the pmb voice set command below, and wait for a yes.
Otherwise, before other work in this session, ask the user once, in English:
1. Which language should chat use? (default English)
2. Style: adhd (answer first, short steps) or plain? (default adhd)
3. Anything about tone, in their own words? (optional)
Then save it: pmb voice set --language <full language name> --style <adhd|plain> [--tone "<text>"]
Until then, write in English (or the language CLAUDE.md names), adhd style.
`
```

- [ ] **Step 4: Run the gate** (Global Constraints). Expected: all PASS, including `cmd/pmb` hook tests that look for `Voice: not set up yet.`.

- [ ] **Step 5: Commit**

```bash
git add internal/hook/hook.go internal/hook/hook_test.go
git commit -m "fix(hook): first-run voice text offers the language CLAUDE.md already names"
```

---

### Task 2: build git fallback starts from the parent

**Files:**
- Modify: `plugin/skills/build/SKILL.md` (Step 1b code block and the line before it)
- Test: `internal/plugincheck/skill_build_test.go`

**verify:** no `git worktree add` command in `build` creates the branch without the recorded parent as its start point, and the text says where `$PARENT` comes from. List every `git worktree add` in the file.

- [ ] **Step 1: Write the failing test.** In `skill_build_test.go` `Must`, replace
`` `git worktree add "../$REPO-$SLUG" -b "$SLUG"`, `` with
`` `git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"`, "`$PARENT` is the parent branch recorded above", ``.

- [ ] **Step 2: Run it, watch it fail:** `go test -count=1 ./internal/plugincheck/ -run TestSkillBuild` — expected FAIL naming the missing `"$PARENT"` strings.

- [ ] **Step 3: Edit Step 1b.** Change the line `Create it:` to:

```markdown
Create it. `$PARENT` is the parent branch recorded above, so the branch starts from it and not from whatever is checked out:
```

and the command line to:

```bash
git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"
```

- [ ] **Step 4: Run the gate.** Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add plugin/skills/build/SKILL.md internal/plugincheck/skill_build_test.go
git commit -m "fix(plugin): build fallback worktree branches from the recorded parent"
```

---

### Task 3: dispatch sends old defects to pm:bug

**Files:**
- Modify: `plugin/skills/dispatch/SKILL.md` ("After a review", step 1)
- Test: `internal/plugincheck/skill_dispatch_test.go`

**verify:** no routing step in `dispatch` puts a defect in code already on the parent branch into a fix ticket; the text sends it to `pm:bug`. List every place in the file that routes BLOCKERs.

- [ ] **Step 1: Write the failing test.** Add `"pm:bug"` and `"already on the parent branch"` to `Must` in `skill_dispatch_test.go`.

- [ ] **Step 2: Run it, watch it fail:** `go test -count=1 ./internal/plugincheck/ -run TestSkillDispatch` — expected FAIL: missing `"pm:bug"`.

- [ ] **Step 3: Edit.** In "After a review", right after the step 1 paragraph (the one starting `1. **BLOCKERs → ONE fix ticket via `pm:plan`.**`) and before the `Exception:` line, add this line, indented three spaces like the `Exception:` line:

```markdown
   A defect in code already on the parent branch, that the diff did not bring in, stays out of the fix ticket: record it with `pm:bug` ("Where findings go" in `pm:review`).
```

- [ ] **Step 4: Run the gate.** Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add plugin/skills/dispatch/SKILL.md internal/plugincheck/skill_dispatch_test.go
git commit -m "fix(plugin): dispatch sends defects in old code to pm:bug"
```
