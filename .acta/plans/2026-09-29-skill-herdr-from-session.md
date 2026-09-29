---
parent: bugs/2026-09-29-brainstorm-skill-offers-herdr-tab
id: PLN-0035
created: "2026-09-29"
hash: roh7jn3
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

- [ ] **Step 1: Write the failing test**

In `TestSkillBrainstorm`, remove `"HERDR_ENV=1"` from `Must`, add to `MustNot`:

```go
"HERDR_ENV",
```

and add to `Must`:

```go
"only when the session text has the `herdr:` line",
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugincheck -run TestSkillBrainstorm`
Expected: FAIL (skill still names HERDR_ENV, new line missing).

- [ ] **Step 3: Edit the skill**

Replace choice (b) with:

```
- (b) **New herdr tab** — offer this only when the session text has the `herdr:` line; the acta hook adds it only inside herdr. With no such line, do not mention herdr.
```

- [ ] **Step 4: Run tests and gates**

Run: `go test ./... && go vet ./... && gofmt -l .`
Expected: PASS, gofmt prints nothing.

- [ ] **Step 5: Run the eval five times**

Run: `for i in 1 2 3 4 5; do env -u HERDR_ENV scripts/eval --case second-brainstorm-choices 2>&1 | grep -E '^second-brainstorm-choices '; done`
Expected: five lines with score 1.00. Any FAIL: report it with the reply text; never edit the grader or prompt.

- [ ] **Step 6: Commit**

```bash
git add plugin/skills/brainstorm/SKILL.md internal/plugincheck/skill_brainstorm_test.go
git commit -m "brainstorm: herdr choice follows the session text, not HERDR_ENV"
```
