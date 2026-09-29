---
parent: bugs/2026-09-29-second-brainstorm-choices-not-offered
id: PLAN-33
hash: kd9t
---
# Rule 8 Names the Choices Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Session-start rule 8 names the second-brainstorm choices, so an agent that never loads acta:brainstorm still offers the right ones.

**Architecture:** Change one line in the `coreRules` constant in `internal/hook/hook.go`, then regenerate `plugin/hooks/default-rules.md` with the existing golden test's `-update` flag.

**Tech Stack:** Go, `go test`, `scripts/eval` (claude plugin eval).

**Spec:** `.acta/specs/2026-09-29-rule8-names-choices-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Do not change `plugin/evals/second-brainstorm-choices/` (prompt or grader).
- Session-start output stays within the 60-line cap in `TestSessionStartStaysShort`.
- Text is short plain English.

## File Map

- `internal/hook/hook.go` (the `coreRules` constant, rule 8)
- `internal/hook/hook_test.go` (new test)
- `plugin/hooks/default-rules.md` (generated, not hand-edited)

## Waves

- Wave 1: Task 1.

---

### Task 1: Rule 8 names the choices

**Files:**
- Modify: `internal/hook/hook.go` (rule 8 inside `coreRules`)
- Modify: `plugin/hooks/default-rules.md` (regenerated)
- Test: `internal/hook/hook_test.go`

**verify:** Every session-start output (normal, first run, broken voice) tells an agent facing a second Architectural brainstorm to load acta:brainstorm and names both portable choices (background agent `claude --bg 'brainstorm SCRATCH-n'`, new session typing `brainstorm SCRATCH-n`), and offers a herdr tab only under `HERDR_ENV=1`. `plugin/hooks/default-rules.md` matches `coreRules`. `scripts/eval --case second-brainstorm-choices` passes 3 runs in a row. List each output variant checked and each eval run result.

**Interfaces:**
- Consumes: `SessionStart(in Input) string`, `korean()` test helper, `voice.Default()`
- Produces: none

- [ ] **Step 1: Write the failing test**

Add to `internal/hook/hook_test.go`:

```go
// An agent can answer a second big brainstorm from this rule alone,
// without loading the skill. So the rule must name the real choices.
func TestSessionStartNamesSecondBrainstormChoices(t *testing.T) {
	for name, in := range map[string]Input{
		"normal":    korean(),
		"first run": {Voice: voice.Default()},
		"broken":    {Voice: voice.Default(), VoiceExists: true, VoiceErr: errors.New("x")},
	} {
		out := SessionStart(in)
		for _, want := range []string{
			"acta:brainstorm",
			"claude --bg 'brainstorm SCRATCH-n'",
			"new session",
			"HERDR_ENV=1",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: session start missing %q", name, want)
			}
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/hook -run TestSessionStartNamesSecondBrainstormChoices`
Expected: FAIL, missing `claude --bg 'brainstorm SCRATCH-n'` (and the others).

- [ ] **Step 3: Change rule 8**

In `coreRules` in `internal/hook/hook.go`, replace rule 8 with:

```
8. One Architectural brainstorm per session. A second one becomes a scratch item; load acta:brainstorm and offer the user a background agent (claude --bg 'brainstorm SCRATCH-n') and a new session where they type brainstorm SCRATCH-n. Offer a herdr tab only when HERDR_ENV=1.
```

Then regenerate the file: `go test ./internal/hook -run TestDefaultRulesFile -update`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/hook/ && go test ./... && go vet ./... && gofmt -l .`
Expected: PASS, gofmt prints nothing.

- [ ] **Step 5: Run the eval case three times**

Run: `for i in 1 2 3; do scripts/eval --case second-brainstorm-choices 2>&1 | grep -E '^second-brainstorm-choices '; done`
Expected: three lines with score 1.00. Any FAIL: report it with the reply text; never edit the grader or prompt.

- [ ] **Step 6: Commit**

```bash
git add internal/hook/hook.go internal/hook/hook_test.go plugin/hooks/default-rules.md
git commit -m "hook: rule 8 names the second-brainstorm choices"
```

## Fix round 1

### Task 2: Session hook texts agree with rule 8

**Files:**
- Modify: `internal/hook/session.go` (`blockText`, `reminderText`)
- Modify: `internal/hook/hook.go` (rule 8 in `coreRules`)
- Modify: `plugin/hooks/default-rules.md` (regenerate with `go test ./internal/hook -run TestDefaultRulesFile -update`, never hand-edit)
- Test: `internal/hook/session_test.go`, `internal/hook/hook_test.go`

**verify:** No text the hooks show an agent (session start, the pre-tool block, the per-prompt reminder) gives a count or a list of second-brainstorm choices that disagrees with rule 8, in herdr and outside herdr; and on every path the scratch item is filed before the user is asked which way to open it. List every hook string that mentions a second brainstorm and what it says.

BLOCKER, `internal/hook/session.go:25-27`: `blockText` says "offer the user the three choices from acta:brainstorm" and `reminderText` says "a second brainstorm gets the three choices", while rule 8 (`internal/hook/hook.go:54`) now says "Name the two ways" and the herdr line prints only in herdr. Scenario: a user outside herdr brainstorms SCRATCH-3; every later prompt carries "three choices"; asked for a second big change, the agent sees two in rule 8 and three in the reminder and invents a third or offers the herdr tab (BUG-6 again). Expected: both texts carry no number and point at rule 8, for example "offer the choices rule 8 names". Update the pinned strings in `internal/hook/session_test.go` (lines near 91, 108, 112).

Same line, same fix: rule 8 says "When the user picks one, file the scratch item", but `plugin/skills/brainstorm/SKILL.md` files first, then asks. A user who answers "later" loses the idea. Expected: rule 8 files the scratch item first (acta scratch new), then names the ways with the real SCRATCH id, then loads acta:brainstorm when the user picks. Keep the strings `claude --bg 'brainstorm SCRATCH-n'` and `new session` so `TestSessionStartNamesSecondBrainstormChoices` stays meaningful, and keep the word herdr out of rule 8.

- [ ] **Step 1: Write the failing tests**

In `internal/hook/session_test.go`, change the three pinned strings so they no longer hold "three choices" and instead hold "rule 8". Add a check to `TestSessionStartNamesSecondBrainstormChoices` in `internal/hook/hook_test.go` that the output holds "file the scratch item first".

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./internal/hook/`
Expected: FAIL on the new strings.

- [ ] **Step 3: Change the texts**

`session.go`: `blockText` ends "file this one as a scratch item and offer the user the choices rule 8 names." `reminderText` ends "a second brainstorm gets the choices rule 8 names." Rule 8 in `hook.go`: "... A second one cannot start in this session: file the scratch item first (acta scratch new), then name the two ways to open it elsewhere, a background agent (claude --bg 'brainstorm SCRATCH-n') and a new session where the user types brainstorm SCRATCH-n. When the user picks one, load acta:brainstorm for that way. Do not design it here." Regenerate default-rules.md.

- [ ] **Step 4: Run the gates**

Run: `go test ./... && go vet ./... && gofmt -l .` then `for i in 1 2 3; do scripts/eval --case second-brainstorm-choices 2>&1 | grep -E '^second-brainstorm-choices '; done` with `HERDR_ENV` unset (`env -u HERDR_ENV scripts/eval ...`). Report the three results; a FAIL caused by BUG-7 (the skill offering a herdr tab) is reported, not fixed here. Never edit the grader or prompt.

- [ ] **Step 5: Commit**

```bash
git add internal/hook/session.go internal/hook/hook.go internal/hook/session_test.go internal/hook/hook_test.go plugin/hooks/default-rules.md
git commit -m "hook: second-brainstorm texts point at rule 8, file the item first"
```
