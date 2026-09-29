---
id: PLN-0024
hash: hjsqd1o
---
# Skill Paths Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Bounded brainstorms write a short spec and go through acta:plan, and the dispatch skill lands only through acta:land.

**Architecture:** Two skill text changes, each guarded by a plugincheck test written red first. No Go code outside `internal/plugincheck`.

**Tech Stack:** Markdown skill files, Go tests in `internal/plugincheck` (`CheckSkill`, `SkillRule`).

**Spec:** `.acta/specs/2026-09-29-skill-paths-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- No change to `plugin/skills/land/SKILL.md`; it already holds the landing rule, `acta id --fix-duplicates` included.
- No Go code change to `acta id` or `AssignIDs`.
- Skill text and comments in plain English a 10-year-old can read; say why, not what.
- Gates: `gofmt -l .` prints nothing, `go vet ./...` is ok, `go test ./...` is green.

---

## File map

- Modify: `plugin/skills/brainstorm/SKILL.md` (Three Paths Bounded entry, Bounded checklist, flow graph, terminal states paragraph, "The Process" intro)
- Modify: `internal/plugincheck/skill_brainstorm_test.go`
- Modify: `plugin/skills/dispatch/SKILL.md` (Autonomous loop item 3, loop-at-a-glance item 11, the Landing section)
- Modify: `internal/plugincheck/skill_dispatch_test.go`

## Waves

- Wave 1: Task 1 and Task 2 in parallel (disjoint files).

### Task 1: Bounded brainstorms write a short spec

**Files:**
- Modify: `plugin/skills/brainstorm/SKILL.md`
- Test: `internal/plugincheck/skill_brainstorm_test.go`

**verify:** No text in `plugin/skills/brainstorm/` still tells an agent that a Bounded brainstorm ends without a spec file or without `acta:plan`, in any form: the Three Paths entry, the checklist, the flow graph, the terminal-states paragraph, the Process intro, or a reworded copy of any of them. Spike still writes no spec, and Architectural is unchanged. Enumerate every place you found that said Bounded has no spec, no plan, or a chat-only design, with what replaced each, and report the list.

**Interfaces:**
- Consumes: `CheckSkill(t, SkillRule{Name, MaxLines, Must, MustNot})` from `internal/plugincheck/check.go`. It joins every `.md` in the skill folder and checks the text.
- Produces: nothing later tasks use.

- [x] **Step 1: Write the failing test.** In `internal/plugincheck/skill_brainstorm_test.go`, keep `TestSkillBrainstorm` as it is and add a second test that reads `plugin/skills/brainstorm/SKILL.md` on its own:

```go
// TestBrainstormBoundedWritesSpec reads the brainstorm skill on its own.
// A Bounded design once lived only in chat, so the user never saw a spec.
func TestBrainstormBoundedWritesSpec(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "brainstorm", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, want := range []string{"Write short spec", "short spec (about half a page)"} {
		if !strings.Contains(txt, want) {
			t.Errorf("brainstorm/SKILL.md missing %q", want)
		}
	}
	for _, old := range []string{
		"No spec file, no implementation plan document",
		"no plan document",
		"no plan doc",
		"Implement via normal workflow",
		"implementation proceeds directly",
		"short in-chat design",
	} {
		if strings.Contains(txt, old) {
			t.Errorf("brainstorm/SKILL.md still says %q", old)
		}
	}
}
```

Add `"os"`, `"path/filepath"` and `"strings"` to the imports.

- [x] **Step 2: Run** `go test ./internal/plugincheck/ -run TestBrainstormBoundedWritesSpec -v`. Expected: FAIL, missing "Write short spec" and still says "No spec file, no implementation plan document".

- [x] **Step 3: Edit `plugin/skills/brainstorm/SKILL.md`.**
  - Three Paths, Bounded entry: replace "present a short design IN CHAT (a few sentences to a few short paragraphs), and STOP. ... No spec file, no implementation plan document." with: present one recommended design in chat and STOP until the user says yes; on yes, write a short spec (about half a page) to `.acta/specs/YYYY-MM-DD-<topic>-design.md` in the worktree, run `acta id`, commit, ask the user to review the file, and on approval invoke acta:plan. No 2-3 approaches, no per-section approval, no one-per-session limit. Keep the lines about "bounded means the flow you are changing is already here to read" and "a bounded task's approval is as hard a gate as an architectural one".
  - Bounded checklist: step 5 becomes "**Write short spec** — about half a page in `.acta/specs/`, run `acta id`, commit in the worktree"; add step 6 "**User reviews spec**" and step 7 "**Transition to implementation** — invoke acta:plan skill".
  - Flow graph: replace the node "Implement via normal workflow (no plan doc)" with "Write short spec" and add the edges `"Human approves?" -> "Write short spec" [label="bounded: yes"]` and `"Write short spec" -> "User reviews spec?"`, so Bounded joins the Architectural review and plan nodes.
  - Terminal states paragraph: Bounded ends at acta:plan too, after its short spec is approved.
  - "The Process" intro: bounded work is context, a few questions, one design in chat, then a short spec.

- [x] **Step 4: Run** `go test ./internal/plugincheck/ -v -run 'TestSkillBrainstorm|TestBrainstormBoundedWritesSpec'`. Expected: PASS. If `TestSkillBrainstorm` fails on MaxLines, the edit grew too much; trim the new text instead of raising the cap.

- [x] **Step 5: Gates and commit.**

```bash
gofmt -l . && go vet ./... && go test ./...
git add plugin/skills/brainstorm/SKILL.md internal/plugincheck/skill_brainstorm_test.go
git commit -m "feat(plugin): bounded brainstorms write a short spec and go through acta:plan"
```

### Task 2: Dispatch lands through acta:land

**Files:**
- Modify: `plugin/skills/dispatch/SKILL.md`
- Test: `internal/plugincheck/skill_dispatch_test.go`

**verify:** No text in `plugin/skills/dispatch/` holds its own merge or landing steps that `acta:land` already owns, in any form: the merge command, `--no-ff`, the stray-file check, the post-merge gate re-run, the worktree and branch cleanup, in the Landing section, the Autonomous loop, the loop at a glance, or `herdr-delivery.md`. Every place that lands points at `acta:land`. The steps `acta:land` lacks stay: closing the pane by slug after the worktree is removed, the `Bugs found by recipient:` report line, and the omp memory harvest. Enumerate every landing restate you found and what replaced it, and report the list.

**Interfaces:**
- Consumes: `CheckSkill` and `SkillRule` from `internal/plugincheck/check.go`.
- Produces: nothing later tasks use.

- [x] **Step 1: Write the failing test.** In `internal/plugincheck/skill_dispatch_test.go`, add `"--no-ff"`, `"Contamination check"` and `"git branch -d"` to the `MustNot` list of `TestSkillDispatch`, and add this test:

```go
// TestDispatchLandsThroughLand reads SKILL.md on its own. Dispatch once kept
// its own merge steps, missed acta id --fix-duplicates, and landed duplicate ids.
func TestDispatchLandsThroughLand(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "dispatch", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	i := strings.Index(txt, "## Landing")
	if i < 0 {
		t.Fatal("dispatch/SKILL.md has no Landing section")
	}
	landing := txt[i:]
	if j := strings.Index(landing[3:], "\n## "); j >= 0 {
		landing = landing[:j+3]
	}
	for _, want := range []string{"acta:land", "herdr pane close", "Bugs found by recipient"} {
		if !strings.Contains(landing, want) {
			t.Errorf("dispatch Landing section missing %q", want)
		}
	}
}
```

- [x] **Step 2: Run** `go test ./internal/plugincheck/ -run 'TestSkillDispatch|TestDispatchLandsThroughLand' -v`. Expected: FAIL, `--no-ff` and `Contamination check` still present.

- [x] **Step 3: Edit `plugin/skills/dispatch/SKILL.md`.**
  - Autonomous loop item 3: "Review clean + every ticket done → run `acta:land` in that same turn, then close the tab. No 'ready to merge, shall I?'." Drop the inline merge steps.
  - Loop at a glance item 11: "Clean AND complete → `acta:land`, then close the tab. Never leave a reviewed branch parked."
  - Landing section: keep the heading and the "same turn, never ask" rule. Replace steps 1-7 with: run `acta:land` for the whole landing (it holds preconditions, stray files, parent, merge, `acta id --fix-duplicates`, post-merge gates, cleanup, never push); then, after the worktree is removed, `herdr pane close <pane-id>` resolved from the slug, the one place a dispatch tab is closed. Keep the stop cases paragraph. The landing report adds one line `Bugs found by recipient:` naming each bug under `.acta/bugs` in the diff, or `none`, then the omp harvest line from "Memory sweep before done".
  - Check `herdr-delivery.md` for landing restates; its line saying you close the pane in Landing after merge and worktree removal stays.

- [x] **Step 4: Lower the cap and run.** Count the folder: `cat plugin/skills/dispatch/*.md | wc -l`. Set `MaxLines` in `TestSkillDispatch` to that number rounded up to the next multiple of 5. Run `go test ./internal/plugincheck/ -v -run 'Dispatch'`. Expected: PASS.

- [x] **Step 5: Gates and commit.**

```bash
gofmt -l . && go vet ./... && go test ./...
git add plugin/skills/dispatch/SKILL.md internal/plugincheck/skill_dispatch_test.go
git commit -m "feat(plugin): dispatch lands through acta:land"
```
