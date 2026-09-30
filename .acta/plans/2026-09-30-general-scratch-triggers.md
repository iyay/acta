---
parent: specs/2026-09-30-general-scratch-triggers-design
created: "2026-09-30"
id: PLN-0049
hash: ysib7t0
---
# General scratch trigger words Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The scratch trigger examples the plugin ships are general words ("note this", "later", "idea for later", any language), not one user's Indonesian slang.

**Architecture:** Two text edits. The scratch skill gets new example words, pinned by its plugincheck test. The hook skills index line gets new words in `internal/hook/hook.go`, and `plugin/hooks/default-rules.md` is regenerated from it with the existing `-update` flag.

**Tech Stack:** Go, standard `testing`, markdown.

**Spec:** `.acta/specs/2026-09-30-general-scratch-triggers-design.md`

**Tests:** fast `scripts/test ./internal/plugincheck/ ./internal/hook/`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- The phrase "note this" must stay in `plugin/skills/scratch/SKILL.md`. The `note-to-scratch` eval guards on it (`internal/plugincheck/evals_test.go`).
- The hook scratch line is exactly: `raw ideas ("note this", "later", side ideas, any language); file with acta scratch new, never memory`
- Do not touch test fixtures that use "catet" as body text: `internal/write/ops_test.go`, `internal/write/scratch_test.go`, `internal/board/testdata/basic/.acta/scratch/`, `internal/tui/sidebar_test.go`.
- Comments and commits in plain English.

## File map

- `plugin/skills/scratch/SKILL.md`: description line and the first "When to file" bullet. (Task 1)
- `internal/plugincheck/skill_scratch_test.go`: `Must` and `MustNot` lists. (Task 1)
- `internal/hook/hook.go`: the `scratch` entry of the skills index. (Task 2)
- `internal/hook/hook_test.go`: the two expected scratch lines. (Task 2)
- `plugin/hooks/default-rules.md`: regenerated, never edited by hand. (Task 2)

## Waves

- Wave 1: Task 1, Task 2. No shared files.

---

### Task 1: Scratch skill uses general trigger words

**Files:**
- Modify: `plugin/skills/scratch/SKILL.md` (line 3 description, line 12 bullet)
- Test: `internal/plugincheck/skill_scratch_test.go`

**verify:** The words "catet", "nanti" and "kepikiran" cannot appear anywhere in the scratch skill, in any spot: description, bullets or examples. The skill still names "note this", "idea for later" and "any language". List every place in the skill that names trigger words and what each now says.

**Interfaces:**
- Consumes: `CheckSkill`, `SkillRule` (existing, `internal/plugincheck`).
- Produces: nothing new.

- [ ] **Step 1: Write the failing test**

In `internal/plugincheck/skill_scratch_test.go`, change the `Must` line and the `MustNot` list:

```go
		Must: []string{
			"acta scratch new", "acta scratch add", "verbatim", "images", "Filed SCR-0001",
			"note this", "idea for later", "any language", "File this in Scratchpad?", "never", "memory",
			"new session", "main branch", "status dropped",
			"--section context", "--section questions", "right after",
			"what work was going on", "what already exists", "file:line",
			"where the facts came from",
		},
		MustNot: []string{"superpowers:", "You may add your own lines", "below theirs", "catet", "nanti", "kepikiran"},
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugincheck/ -run 'TestSkillScratch$' -v`
Expected: FAIL. It reports "idea for later" and "any language" missing and "catet", "nanti", "kepikiran" present.

- [ ] **Step 3: Write minimal implementation**

In `plugin/skills/scratch/SKILL.md`, line 3 becomes:

```
description: "acta: Use when the user drops a raw idea (note this, later, idea for later, or the same intent in any language) or a side idea shows up during other work. Files it in Scratchpad with acta scratch new; never in agent memory."
```

Line 12 becomes:

```
- The user drops a raw idea: "note this", "later", "idea for later", or the same intent in any language. File it at once. Do not ask first; the words are the idea.
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/plugincheck/ -v -run 'TestSkillScratch|TestEval'`
Expected: PASS. The eval guard test still finds "note this".

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/plugincheck && go vet ./internal/plugincheck/
git add plugin/skills/scratch/SKILL.md internal/plugincheck/skill_scratch_test.go
git commit -m "Use general scratch trigger words in the scratch skill"
```

### Task 2: Hook skills index uses general trigger words

**Files:**
- Modify: `internal/hook/hook.go:30`
- Modify: `plugin/hooks/default-rules.md` (regenerated)
- Test: `internal/hook/hook_test.go` (lines 36 and 140)

**verify:** No text the session-start hook prints, from Go or from `plugin/hooks/default-rules.md`, names "catet", "nanti" or "kepikiran" in the scratch line, and the two sources print the same line. List every surface that carries the scratch index line and what each now says.

**Interfaces:**
- Consumes: nothing from Task 1.
- Produces: nothing new.

- [ ] **Step 1: Write the failing test**

In `internal/hook/hook_test.go`, replace the expected line near line 36 with:

```go
		`- acta:scratch: raw ideas ("note this", "later", side ideas, any language); file with acta scratch new, never memory`,
```

and the check near line 140 with:

```go
		if !strings.Contains(out, "- acta:scratch: raw ideas (\"note this\", \"later\", side ideas, any language); file with acta scratch new, never memory") {
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/hook/ -v`
Expected: FAIL. Both checks report the new scratch line missing.

- [ ] **Step 3: Write minimal implementation**

In `internal/hook/hook.go` line 30:

```go
	{"scratch", `raw ideas ("note this", "later", side ideas, any language); file with acta scratch new, never memory`},
```

Then regenerate the file the shell hook and omp read:

```bash
go test ./internal/hook -run TestDefaultRulesFile -update
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/hook/ -v && grep -n 'acta:scratch' plugin/hooks/default-rules.md`
Expected: PASS, and the grep shows the new line.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/hook && go vet ./internal/hook/
git add internal/hook/hook.go internal/hook/hook_test.go plugin/hooks/default-rules.md
git commit -m "Use general scratch trigger words in the hook skills index"
```
