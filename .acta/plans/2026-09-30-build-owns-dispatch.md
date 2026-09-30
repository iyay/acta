---
id: PLN-0058
created: "2026-09-30"
hash: rr5so5y
---
# Build Owns Dispatch Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Remove the `acta:dispatch` skill so `acta:build` is the only skill that implements a plan, with dispatch delivery kept as a reference file inside build.

**Architecture:** The delivery part of `plugin/skills/dispatch/SKILL.md` moves to `plugin/skills/build/dispatch.md`; `herdr-delivery.md` moves next to it. The loop part (entry gate, autonomy, review, fix rounds, landing) is dropped, because build, `acta:review` and `acta:land` already hold it. Build picks its executor from a `/build <executor>` argument, then the config, then a question, and falls back to `subagent` on omp or without herdr.

**Tech Stack:** Markdown skill text, Go tests in `internal/plugincheck` and `internal/hook`.

**Spec:** `.acta/specs/2026-09-30-build-owns-dispatch-design.md`

**Tests:** fast `scripts/test`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- The CLI does not change: `acta dispatch init`, `acta reply-back` and the `.dispatch.json` record stay as they are.
- On omp, `dispatch` stays `subagent`. No dispatch from omp into another omp tab.
- No stub `acta:dispatch` skill. No file under `plugin/` names `acta:dispatch` when the plan is done.
- Build's description stays under 1024 characters.
- Every file written is in English. Comments use short, plain words and say why.
- `plugin/hooks/default-rules.md` is never edited by hand: regenerate it with `go test ./internal/hook -run TestDefaultRulesFile -update`.
- `plugin/NOTICE` needs no change: it never listed a dispatch skill folder (checked 2026-09-30).

## File map

| File | Task | Change |
|---|---|---|
| `plugin/skills/build/dispatch.md` | 1 | new: delivery part of the old dispatch skill |
| `plugin/skills/build/herdr-delivery.md` | 1 | moved from `plugin/skills/dispatch/`, pointers fixed |
| `plugin/skills/dispatch/` | 1 | deleted |
| `plugin/skills/build/SKILL.md` | 1, 2 | 1: dispatch row points at `dispatch.md`; 2: executor order, fallbacks, Close, description |
| `plugin/references/house-rules.md` | 1 | line 3 names the dispatch brief, not `acta:dispatch` |
| `internal/hook/hook.go` | 1, 2 | 1: `dispatch` leaves `Skills`; 2: build line text |
| `internal/hook/hook_test.go` | 1, 2 | 1: 11 skills; 2: build line text |
| `plugin/hooks/default-rules.md` | 1, 2 | regenerated |
| `internal/plugincheck/skill_dispatch_test.go` | 1 | deleted |
| `internal/plugincheck/build_dispatch_test.go` | 1 | new: tests over `build/dispatch.md` and `build/herdr-delivery.md`, plus the no-dispatch-skill guard |
| `internal/plugincheck/reply_back_test.go` | 1 | paths point at `skills/build/` |
| `internal/plugincheck/models_test.go` | 1 | `dispatch` leaves the list |
| `internal/plugincheck/skill_build_test.go` | 1, 2 | 1: cap and `acta:dispatch`; 2: new phrases |

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2 (shares `build/SKILL.md`, `hook.go`, `hook_test.go` and `skill_build_test.go` with Task 1).

---

### Task 1: Move dispatch delivery into build and delete the dispatch skill

**Files:**
- Create: `plugin/skills/build/dispatch.md`
- Move: `plugin/skills/dispatch/herdr-delivery.md` to `plugin/skills/build/herdr-delivery.md` (`git mv`)
- Delete: `plugin/skills/dispatch/SKILL.md` and the folder
- Modify: `plugin/skills/build/SKILL.md` (the `dispatch` row of the `## Executors` table only)
- Modify: `plugin/references/house-rules.md:3`
- Modify: `internal/hook/hook.go` (`Skills`)
- Regenerate: `plugin/hooks/default-rules.md`
- Test: `internal/plugincheck/build_dispatch_test.go` (new), `internal/plugincheck/reply_back_test.go`, `internal/plugincheck/models_test.go`, `internal/plugincheck/skill_build_test.go`, `internal/hook/hook_test.go`
- Delete: `internal/plugincheck/skill_dispatch_test.go`

**verify:** Every rule of the old dispatch skill still has exactly one home: `build/dispatch.md`, `build/herdr-delivery.md`, `build/SKILL.md`, `acta:review` or `acta:land`. List each old section with where its rules went, or which existing text already covers them. No file under `plugin/` names `acta:dispatch`, and no text path leads a model to a dispatch skill.

**Interfaces:**
- Consumes: nothing.
- Produces: `plugin/skills/build/dispatch.md` and `plugin/skills/build/herdr-delivery.md`; `hook.Skills` with 11 entries and no `dispatch`. Task 2 links to `dispatch.md` from build's executor text.

- [ ] **Step 1: Write the failing tests**

Create `internal/plugincheck/build_dispatch_test.go`:

```go
package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readBuildFile reads one file of the build skill on its own. CheckSkill joins
// the whole folder, so a revert of one file could stay green.
func readBuildFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestNoDispatchSkill keeps build the only skill that runs a plan. A model
// once loaded the dispatch skill on its own and refused with a wrong reason.
func TestNoDispatchSkill(t *testing.T) {
	if _, err := os.Stat(filepath.Join(pluginRoot(t), "skills", "dispatch")); !os.IsNotExist(err) {
		t.Error("plugin/skills/dispatch still exists; dispatch lives in skills/build/dispatch.md")
	}
	err := filepath.WalkDir(pluginRoot(t), func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "node_modules" || d.Name() == ".git") {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), "acta:dispatch") {
			rel, _ := filepath.Rel(pluginRoot(t), p)
			t.Errorf("%s still names acta:dispatch", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestBuildDispatchDelivery checks the delivery rules that came from the old
// dispatch skill. Each file is read on its own.
func TestBuildDispatchDelivery(t *testing.T) {
	for file, wants := range map[string][]string{
		"dispatch.md": {
			"/goal", "REPLY-BACK", "references/house-rules.md", "SKILL: load build", "Never wait",
			"herdr-delivery.md", "PROPERTY", "ultrathink orchestrate", ".acta/plans/",
			"GATES (from the worktree): <the plan's fast test command>",
			"exactly one read", "checkpoint unconfirmed", "gets its own one read", "outside this rule",
			"HARD RULE", "New session started", "🎯 Goal", "never in one prompt", "then `/goal` only",
			"acta:review", "acta:land", "herdr pane close", "Bugs found by recipient",
		},
		"herdr-delivery.md": {
			"exactly one read", "checkpoint unconfirmed", "HARD RULE", "New session started", "🎯 Goal",
			"never in one prompt", "then `/goal` only", "--agent omp", "acta dispatch init", "herdr pane close",
		},
	} {
		txt := readBuildFile(t, file)
		for _, want := range wants {
			if !strings.Contains(txt, want) {
				t.Errorf("build/%s missing %q", file, want)
			}
		}
		if strings.Contains(txt, "Back-to-back, no settle-wait") {
			t.Errorf("build/%s still says \"Back-to-back, no settle-wait\"", file)
		}
	}
	txt := readBuildFile(t, "dispatch.md")
	if n := strings.Count(txt, "checkpoint unconfirmed"); n < 2 {
		t.Errorf("build/dispatch.md says \"checkpoint unconfirmed\" %d times; the loop item and the section both need it", n)
	}
	// These belong to build, acta:review and acta:land now. A second copy
	// drifts, and drift is how duplicate ids once reached the parent branch.
	for _, bad := range []string{
		"acta:dispatch", "Step -3", "Refuse inside omp", "dispatch requires herdr", "## Entry gate",
		"## Autonomous loop", "## Landing", "NOTEs go to memory", "--no-ff", "git branch -d",
		"superpowers:", "git-bug", "/Users/", "<new-head-sha>", "fix round R of 5", "until herdr agent read",
		"read it again", "<one-shot test runner>",
	} {
		if strings.Contains(txt, bad) {
			t.Errorf("build/dispatch.md still carries %q", bad)
		}
	}
}
```

In `internal/plugincheck/reply_back_test.go`:
- `TestDispatchBriefLoadsBuild`: read `skills/build/dispatch.md`, error text `build/dispatch.md missing ...`.
- `TestNoHandFilledReplyBack`: loop over `skills/build` instead of `skills/dispatch`.
- `TestDispatchInitBeforeGoal`: read `skills/build/herdr-delivery.md`, error text `build/herdr-delivery.md missing ...`.

In `internal/plugincheck/models_test.go`: the list becomes `[]string{"plan", "brainstorm", "debug", "build", "review"}` and the test is renamed `TestModelsParagraphInFiveSkills`.

In `internal/plugincheck/skill_build_test.go` (`TestSkillBuild`): remove `"acta:dispatch"` from `Must`, add `"acta:dispatch"` to `MustNot`, and set `MaxLines` to `1000`. The build folder now holds `SKILL.md`, `implementer-prompt.md`, `dispatch.md` and `herdr-delivery.md`; the old caps were 682 (build) and 650 (dispatch).

In `internal/hook/hook_test.go` (`TestSessionStartListsSkillsAndRules`): `len(Skills) != 11` and the message `"%d skills, want 11"`. Also reword the comment near line 145 so it no longer says the index names the dispatch skill:

```go
		// Case-insensitive, so this also rules out HERDR_ENV. Only the rules
		// and the voice lines count: the skill index above them may name
		// dispatch as an executor, and an index is not an offer of a tab.
```

Delete `internal/plugincheck/skill_dispatch_test.go`.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/plugincheck/ ./internal/hook/`
Expected: FAIL: `TestNoDispatchSkill` (`plugin/skills/dispatch still exists`), `TestBuildDispatchDelivery` (read error on `build/dispatch.md`), the retargeted reply-back tests, and `TestSessionStartListsSkillsAndRules` (`12 skills, want 11`).

- [ ] **Step 3: Move `herdr-delivery.md`**

```bash
git mv plugin/skills/dispatch/herdr-delivery.md plugin/skills/build/herdr-delivery.md
```

Then change only its pointers. Every `SKILL.md` in it meant the old dispatch skill: point each one at `dispatch.md` and the section of the same name. The line 43 and line 221 pointers to "Step -1" point at the build worktree instead: "the worktree from `## Worktree` in SKILL.md". The line 3 note says `same as dispatch.md`. No other text changes.

- [ ] **Step 4: Write `plugin/skills/build/dispatch.md`**

No frontmatter: it is a reference file, not a skill. Start with:

```markdown
# Dispatch — the plan in another pane

Read this only when `acta:build` runs the `dispatch` executor. Build already made the worktree (`## Worktree` in SKILL.md) and picked this executor. This file says how to hand the plan to an omp agent in its own herdr tab, how to wait for it, and how to send a fix round back to it. Review and landing are build's `## Close`, `acta:review` and `acta:land`; only the steps a tab adds live here.
```

Then carry over these sections of the old `plugin/skills/dispatch/SKILL.md`, text kept except the edits named:

| Old section | In `dispatch.md` |
|---|---|
| Skill notation (paragraph under the title) | keep |
| Never wait for the recipient | keep |
| The loop at a glance | keep; Phase 1 item 1 becomes "build picked `dispatch` and made the worktree"; item 2 "Provision" keeps only the tab lookup; Phase 2 items 9-11 point at build's `## Close` and at "Fix rounds" below |
| One tab per dispatch | keep |
| Operation map | keep; the Provision row keeps only the tab part |
| `/new` and `/goal` | keep |
| Reply-back | keep |
| omp magic keywords | keep |
| omp `advisor` | keep |
| Step -1 | keep the tab half only, titled `## Provision the tab`; the worktree half is build's `## Worktree` |
| Step 0a | keep |
| Every criterion is a PROPERTY | keep |
| Stopping rule | keep |
| Spread the work | keep; the models paragraph is not repeated (build has it) |
| Comprehension checkpoint | keep |
| Idle watcher | keep |
| Hand-off brief format | keep (it carries `SKILL: load build` and `GATES (from the worktree): <the plan's fast test command>`) |
| Verify the reply | keep, with the `Bugs found by recipient` line |
| Review | keep only the rule that the small-change self-review exception of `acta:review` never applies to a dispatch; the rest is `acta:review` |
| After a review | becomes `## Fix rounds`: keep steps 2-4 (reuse the same agent by slug, `/goal` only with no `/new`, `acta dispatch init` again, same REPLY-BACK line, dispatch then stop and yield) and the 1-2 findings inline exception; the fix task itself, rounds, ranges and NOTEs follow `acta:review` |
| Landing | becomes `## Close the tab`: after `acta:land` removed the worktree, `herdr pane close <pane-id>` resolved from the slug; the report adds `Bugs found by recipient:` and `Harvested from omp:` lines |
| Memory sweep | keep only the omp memory harvest paragraph, under `## Close the tab` |
| Entry gate, Autonomous loop, Autonomy, Step -3, Step -2 | drop |

Before dropping each of the five sections in the last row, and each part cut from Review, After a review, Landing and Memory sweep, find each of its rules in `build/SKILL.md`, `plugin/skills/review/SKILL.md` or `plugin/skills/land/SKILL.md`. A rule found nowhere goes into `dispatch.md` when only a tab needs it, or into `build/SKILL.md` when every executor needs it. Write the list (rule, where it lives now) into your report for the verify line.

- [ ] **Step 5: Delete the old skill and fix the pointers**

```bash
git rm plugin/skills/dispatch/SKILL.md
```

In `plugin/skills/build/SKILL.md`, the `dispatch` row of `## Executors` becomes:

```markdown
| `dispatch` | an omp agent in its own herdr tab | read [dispatch.md](dispatch.md) and follow it. On omp, `dispatch` runs as `subagent` |
```

Also change `through acta:dispatch` in build's `description` to `through dispatch.md`, and change nothing else in the description (Task 2 rewrites it).

`plugin/references/house-rules.md:3` starts: `These apply to any agent started from a dispatch brief (build/dispatch.md).` The rest of the line stays.

In `internal/hook/hook.go`, delete the line `{"dispatch", "run build through an omp agent in its own herdr tab"},` from `Skills`.

Regenerate the rules file:

```bash
go test ./internal/hook -run TestDefaultRulesFile -update
```

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./internal/plugincheck/ ./internal/hook/`
Expected: PASS.

- [ ] **Step 7: Format, vet, commit**

```bash
gofmt -l internal/ && go vet ./internal/plugincheck/ ./internal/hook/
git add plugin/skills/build/dispatch.md plugin/skills/build/herdr-delivery.md plugin/skills/build/SKILL.md plugin/references/house-rules.md plugin/hooks/default-rules.md internal/hook/hook.go internal/hook/hook_test.go internal/plugincheck/build_dispatch_test.go internal/plugincheck/reply_back_test.go internal/plugincheck/models_test.go internal/plugincheck/skill_build_test.go
git add -u plugin/skills/dispatch internal/plugincheck/skill_dispatch_test.go
git commit -m "Move dispatch delivery into build and delete the dispatch skill"
```

### Task 2: Build picks its executor and closes the same way for all

**Files:**
- Modify: `plugin/skills/build/SKILL.md` (frontmatter `description`, `## Executors`, `## Close`)
- Modify: `internal/hook/hook.go` (the `build` entry of `Skills`)
- Regenerate: `plugin/hooks/default-rules.md`
- Test: `internal/plugincheck/skill_build_test.go`, `internal/hook/hook_test.go`

**verify:** On every harness and every setup, build reaches an executor without refusing: argument, then config, then one question. List each case (argument given; config only; neither; omp with `dispatch`; `dispatch` with no herdr) and what build does. Every executor ends through the same Close; only the fix-round delivery differs.

**Interfaces:**
- Consumes: `plugin/skills/build/dispatch.md` from Task 1, and its `## Fix rounds` and `## Close the tab` sections.
- Produces: nothing later tasks use.

- [ ] **Step 1: Write the failing tests**

In `internal/plugincheck/skill_build_test.go` (`TestSkillBuild`), replace in `Must`:
- `"the executor `acta config show` names, else asks which one"` with `"through the executor picked by `/build <executor>`, else the one `acta config show` names, else asks which one"`
- `"Dispatch is only for harnesses other than omp."` stays.

and add to `Must`:

```go
			"`/build <executor>`", "It is for this run only; never save it.",
			"HERDR_ENV=1", "Build never refuses because herdr is missing.",
			"hand it to omp in another tab or pane",
			"[dispatch.md](dispatch.md)", "## Fix rounds", "## Close the tab",
```

Add a test that reads `SKILL.md` on its own:

```go
// TestBuildExecutorOrder reads SKILL.md on its own. CheckSkill joins the
// folder, and dispatch.md also talks about executors.
func TestBuildExecutorOrder(t *testing.T) {
	txt := readBuildFile(t, "SKILL.md")
	arg := strings.Index(txt, "1. The argument of `/build <executor>`")
	cfg := strings.Index(txt, "2. `build_executor: <name>` from `acta config show`")
	ask := strings.Index(txt, "3. Neither: ask which executor to run")
	if arg < 0 || cfg < 0 || ask < 0 || !(arg < cfg && cfg < ask) {
		t.Errorf("build/SKILL.md must list the executor order argument, config, ask (got %d, %d, %d)", arg, cfg, ask)
	}
	if !strings.Contains(txt, "runs as `subagent`. Build never refuses because herdr is missing.") {
		t.Error("build/SKILL.md missing the no-herdr fallback to subagent")
	}
}
```

In `internal/hook/hook_test.go`, the expected build line becomes:

```go
		"- acta:build: run an approved plan in a worktree; executor from `/build <executor>`, else `acta config show`, else ask: subagent, dispatch or inline",
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/plugincheck/ -run 'TestSkillBuild|TestBuildExecutorOrder' && go test ./internal/hook/ -run TestSessionStartListsSkillsAndRules`
Expected: FAIL on the missing phrases and the old build line.

- [ ] **Step 3: Rewrite the build text**

`description` in `plugin/skills/build/SKILL.md`:

```yaml
description: "acta: Use to run an approved plan, and when the user asks to dispatch a plan or hand it to omp in another tab or pane. Creates the worktree without asking, then runs every task with a failing test first through the executor picked by `/build <executor>`, else the one `acta config show` names, else asks which one - subagent (the current harness's own subagents), dispatch (an omp agent in its own herdr tab, through dispatch.md) or inline (you write the code). Commits each task; review and landing follow through acta:review and acta:land."
```

In `## Executors`, keep the table and the orchestrator paragraph. Replace the two paragraphs after it ("Before you pick one, run `acta config show`..." and "On omp, `dispatch` runs as `subagent`...") with:

```markdown
Pick the executor in this order. When one of the first two gives an answer, do not ask.

1. The argument of `/build <executor>`: `subagent`, `dispatch` or `inline`. It is for this run only; never save it.
2. `build_executor: <name>` from `acta config show`.
3. Neither: ask which executor to run, as the table above describes.

Two fallbacks, each told to the user in one line:

- On omp, `dispatch` runs as `subagent`: use `agent()` with `agent="task"` and do not ask. Dispatch is only for harnesses other than omp.
- `dispatch` with no herdr (no `HERDR_ENV=1` in the environment and no `herdr` on PATH) runs as `subagent`. Build never refuses because herdr is missing.
```

At the end of the `## Close` paragraph (before `### Reply back when a dispatch record exists`), add:

```markdown
Every executor closes this way. Only the fix round is sent differently: `subagent` gives the fix task to a new implementer, `inline` fixes it yourself, and `dispatch` sends it to the same agent in its tab (`## Fix rounds` in [dispatch.md](dispatch.md)). With `dispatch`, after `acta:land` close the tab (`## Close the tab` in dispatch.md).
```

In `internal/hook/hook.go`, the `build` entry of `Skills` becomes:

```go
	{"build", "run an approved plan in a worktree; executor from `/build <executor>`, else `acta config show`, else ask: subagent, dispatch or inline"},
```

Regenerate the rules file:

```bash
go test ./internal/hook -run TestDefaultRulesFile -update
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/plugincheck/ ./internal/hook/`
Expected: PASS.

- [ ] **Step 5: Format, vet, commit**

```bash
gofmt -l internal/ && go vet ./internal/plugincheck/ ./internal/hook/
git add plugin/skills/build/SKILL.md internal/hook/hook.go internal/hook/hook_test.go plugin/hooks/default-rules.md internal/plugincheck/skill_build_test.go
git commit -m "Build picks its executor from the argument, then config, then asks"
```
