---
parent: bugs/2026-09-29-brainstorm-skill-offers-herdr-tab
id: PLN-0035
created: "2026-09-29"
hash: roh7jn3
started: "2026-09-29"
finished: "2026-09-29"
---
# Skill Herdr Choice From Session Text Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The brainstorm skill offers a herdr tab only when the session text has the hook's `herdr:` line, and never names `HERDR_ENV`.

**Architecture:** Reword choice (b) in the "One Architectural brainstorm per session" list of `plugin/skills/brainstorm/SKILL.md`, and move `HERDR_ENV=1` from Must to MustNot in its plugincheck test.

**Tech Stack:** Markdown skill text, Go tests (`CheckSkill` in `internal/plugincheck`), `scripts/eval`.

**Spec:** `.acta/specs/2026-09-29-skill-herdr-from-session-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Do not change `plugin/evals/second-brainstorm-choices/` (prompt or grader).
- Skill text is short plain English and stays under the brainstorm `MaxLines` cap.

## File Map

- `plugin/skills/brainstorm/SKILL.md` (choice (b) only)
- `internal/plugincheck/skill_brainstorm_test.go`

## Waves

- Wave 1: Task 1.

---

### Task 1: Choice (b) reads the session text

**Files:**
- Modify: `plugin/skills/brainstorm/SKILL.md` (choice (b), "New herdr tab")
- Test: `internal/plugincheck/skill_brainstorm_test.go`

**verify:** No text in the brainstorm skill tells the agent to check an environment variable before offering the herdr tab; the only trigger is the session text's `herdr:` line, and without it herdr is not mentioned. The eval case passes 5 runs in a row with `HERDR_ENV` unset. List every place in the skill that mentions herdr and each eval run result.

**Interfaces:**
- Consumes: the session line `herdr: this session runs in a herdr tab. ...` printed by `internal/hook/hook.go` (`herdrExtra`) only when `HERDR_ENV=1`.
- Produces: none

- [x] **Step 1: Write the failing test**

In `TestSkillBrainstorm`, remove `"HERDR_ENV=1"` from `Must`, add to `MustNot`:

```go
"HERDR_ENV",
```

and add to `Must`:

```go
"only when the session text has the `herdr:` line",
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugincheck -run TestSkillBrainstorm`
Expected: FAIL (skill still names HERDR_ENV, new line missing).

- [x] **Step 3: Edit the skill**

Replace choice (b) with:

```
- (b) **New herdr tab** — offer this only when the session text has the `herdr:` line; the acta hook adds it only inside herdr. With no such line, do not mention herdr.
```

- [x] **Step 4: Run tests and gates**

Run: `go test ./... && go vet ./... && gofmt -l .`
Expected: PASS, gofmt prints nothing.

- [x] **Step 5: Run the eval five times**

Run: `for i in 1 2 3 4 5; do env -u HERDR_ENV scripts/eval --case second-brainstorm-choices 2>&1 | grep -E '^second-brainstorm-choices '; done`
Expected: five lines with score 1.00. Any FAIL: report it with the reply text; never edit the grader or prompt.

- [x] **Step 6: Commit**

```bash
git add plugin/skills/brainstorm/SKILL.md internal/plugincheck/skill_brainstorm_test.go
git commit -m "brainstorm: herdr choice follows the session text, not HERDR_ENV"
```

## Fix round 1

### Task 2: Rule 8 files the item in one call

**Files:**
- Modify: `internal/hook/hook.go` (rule 8 in `coreRules`)
- Modify: `plugin/hooks/default-rules.md` (regenerate with `go test ./internal/hook -run TestDefaultRulesFile -update`, never hand-edit)
- Test: `internal/hook/hook_test.go`

**verify:** An agent that follows rule 8 for a second big brainstorm answers the user within 3 turns on every path: it files the scratch item with one `acta scratch new` call, without loading acta:scratch first and without a follow-up `acta scratch add`, and a "written, not committed" result (exit 2) counts as filed, never as a reason to retry. `max_turns` in the eval case stays 3. List each tool call rule 8 now asks for, and each eval run result.

Root cause (traces kept from 6 runs, 2 failed with "Reached maximum number of turns (3)"): every run first loads `acta:scratch` (turn 1), then runs `acta scratch new` (turn 2). The scratch skill says "You may add your own lines below theirs", so failing runs spend turn 3 on an extra `acta scratch add` and never answer. The eval sandbox is not a git repo, so `acta scratch new` exits 2 "written, not committed"; passing runs treat that as filed and answer in turn 3.

Expected: rule 8 says: file it with one `acta scratch new` call (no skill load, no extra add); "written, not committed" still counts as filed; answer the user in the same turn. Keep the strings `file the scratch item first`, `claude --bg 'brainstorm SCRATCH-n'` and `new session`, and keep the word herdr out of rule 8.

- [x] **Step 1: Write the failing test**

In `TestSessionStartNamesSecondBrainstormChoices` (`internal/hook/hook_test.go`), add to the wanted strings: `"one acta scratch new call"` and `"written, not committed"`.

- [x] **Step 2: Run to see it fail**

Run: `go test ./internal/hook/ -run TestSessionStartNamesSecondBrainstormChoices`
Expected: FAIL on the two new strings.

- [x] **Step 3: Change rule 8**

New rule 8: `8. One Architectural brainstorm per session. A second one cannot start in this session: file the scratch item first, with one acta scratch new call (do not load acta:scratch, do not add more lines; "written, not committed" still counts as filed). In the same reply, name the two ways to open it elsewhere, a background agent (claude --bg 'brainstorm SCRATCH-n') and a new session where the user types brainstorm SCRATCH-n. When the user picks one, load acta:brainstorm for that way. Do not design it here.` Then regenerate default-rules.md.

- [x] **Step 4: Gates**

Run: `go test ./... && go vet ./... && gofmt -l .` then `for i in 1 2 3 4 5 6; do env -u HERDR_ENV scripts/eval --case second-brainstorm-choices 2>&1 | grep -E '^second-brainstorm-choices '; done`. Expected: six lines at 1.00. Never change `max_turns`, the grader or the prompt.

- [x] **Step 5: Commit**

```bash
git add internal/hook/hook.go internal/hook/hook_test.go plugin/hooks/default-rules.md
git commit -m "hook: rule 8 files the scratch item in one call and answers in the same reply"
```

## Fix round 2

### Task 3: Rule 8 short, herdr-safe, scratch index restored

**Files:**
- Modify: `internal/hook/hook.go` (rule 8 in `coreRules`; the `scratch` entry in the skill index)
- Modify: `plugin/hooks/default-rules.md` (regenerate with `go test ./internal/hook -run TestDefaultRulesFile -update`, never hand-edit)
- Test: `internal/hook/hook_test.go`

**verify:** (a) In a herdr session and outside one, no hook text forbids a choice that another hook text or `plugin/skills/brainstorm/SKILL.md` offers; list every choice each text offers or forbids. (b) The one-call, no-skill-load filing limit applies only to the second-brainstorm path in rule 8; the skill index line for acta:scratch is back to its text at bec97a4. (c) Rule 8 keeps "When the user picks one, load acta:brainstorm for that way" from the plan's Task 2 Step 3. (d) The eval passes 6 runs in a row with `HERDR_ENV` unset and `max_turns: 3`. Report each list and each run.

BLOCKER 1, `internal/hook/hook.go:54`: rule 8 says "never send the user to another terminal or tab". Inside herdr, `herdrExtra` (`hook.go:61`) and `SKILL.md:80` offer a new herdr tab; rule 8's own "new session" choice is another terminal too. Scenario: `HERDR_ENV=1`, second big brainstorm; the agent drops the herdr choice or breaks a rule. Expected: remove that clause, and every other clause added only to satisfy the grader outside herdr.

BLOCKER 2, `internal/hook/hook.go:54`: rule 8 says "they run it there: never offer to load acta:brainstorm in this session", against plan Task 2 Step 3 and `SKILL.md:77-84` (the agent runs `claude --bg` itself for (a), copies the prompt for (c)). Expected: restore "When the user picks one, load acta:brainstorm for that way."

BLOCKER 3, `internal/hook/hook.go:30`: the global acta:scratch index line now says "one acta scratch new call, body on stdin, no skill load". That reaches every scratch filing, so a side idea skips the scratch skill's "File this in Scratchpad?" ask and auto-commits on main without a yes. Expected: restore the bec97a4 text: `raw ideas ("catet", "nanti", side ideas); file with acta scratch new, never memory`.

Rule 8 target text (short plain English, one idea per sentence). Amended by the user on 2026-09-29: the text below was measured at 0/6 in the eval, because it dropped the command form (the agent guessed the `acta scratch new` shape and spent all three turns) and the id shape (it wrote `SCRATCH-1` where the grader wants `SCR-0001`). Both clauses are shapes, not prohibitions, so the property in (a) still holds; measured 10 runs at 1.00 and the four scratch cases at 1.00.

`8. One Architectural brainstorm per session. A second one cannot start in this session. File the scratch item first, with one acta scratch new call whose body is stdin: acta scratch new <slug> --title <title> < body.md: no Skill tool, no acta scratch add, and "written, not committed" still counts as filed. In the same reply, name the two ways to open it elsewhere: put the id the command printed in place of SCRATCH-n, an id like SCR-0001, never a shortened one. A background agent (claude --bg 'brainstorm SCRATCH-n') and a new session where the user types brainstorm SCRATCH-n. When the user picks one, load acta:brainstorm for that way. Do not design it here.`

- [x] **Step 1: Write the failing tests**

In `internal/hook/hook_test.go`: drop pinned strings for the removed clauses; add to `TestSessionStartNamesSecondBrainstormChoices` the wanted strings `"load acta:brainstorm for that way"` and `"one acta scratch new call"`; add a check, for `Herdr: true` and false, that the output does not contain `"another terminal or tab"` or `"never offer to load acta:brainstorm"`; add a check that the output contains `file with acta scratch new, never memory`.

- [x] **Step 2: Run to see them fail**

Run: `go test ./internal/hook/`
Expected: FAIL on the new checks.

- [x] **Step 3: Change the texts**

Set rule 8 to the target text above and restore the scratch index line. Regenerate default-rules.md.

- [x] **Step 4: Gates**

Run: `go test ./... && go vet ./... && gofmt -l .` then `for i in 1 2 3 4 5 6; do env -u HERDR_ENV scripts/eval --case second-brainstorm-choices 2>&1 | grep -E '^second-brainstorm-choices '; done`. Expected: six at 1.00. If a run fails, report its reply text and trace; do not add clauses that forbid choices another text offers. Never change `max_turns`, the grader or the prompt.

- [x] **Step 5: Commit**

```bash
git add internal/hook/hook.go internal/hook/hook_test.go plugin/hooks/default-rules.md
git commit -m "hook: rule 8 short and herdr-safe, scratch index line restored"
```
