# Spec line counts only .md paths Implementation Plan

> **For agentic workers:** run this plan with pm:build, task by task. Steps use checkbox (`- [ ]`) syntax, and pmb reads those boxes as task progress.

**Goal:** A plan with no spec file stops showing a false "spec ... not found" warning on the board, while a mistyped `.md` spec path still warns.

**Architecture:** `specPath` in `internal/board/parse.go` returns a path only when it ends in `.md`: the first backtick span ending in `.md`, else the first word ending in `.md`, else nothing. The plan skill header tells Bounded plans to write `**Spec:** none (Bounded, approved in chat on <date>)`.

**Tech Stack:** Go 1.27, Go tests over markdown skills in `internal/plugincheck`.

**Spec:** none (Bounded, approved in chat on 2026-09-26)

**Worktree:** created by pm:build at `../pm-board-spec-line-md`, branch `spec-line-md`, parent `main`.

## Global Constraints

- Gate before every commit: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
- Edit only the files each task names. Stage by path. Never push. Never commit the plan file.
- Comments in plain English a ten-year-old can read, saying why. No marker tags.
- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.

## Waves

- Wave 1: Tasks 1 and 2 (no shared files).

---

### Task 1: specPath takes only .md paths

**Files:**
- Modify: `internal/board/parse.go` (func `specPath` and its doc comment)
- Test: `internal/board/parse_test.go` (func `TestParseSpecLine`)

**verify:** No Spec line without a `.md` path ever yields a spec path, and every Spec line that names a `.md` path (in backticks or as a bare word, first or later) still yields it, so a mistyped `.md` path still warns. List every Spec line shape checked.

**Interfaces:**
- Consumes: nothing.
- Produces: `specPath(s string) string` keeps its signature.

- [x] **Step 1: Write the failing test**

Add these cases to the `cases` map in `TestParseSpecLine` (keep the existing four):

```go
		"**Spec:** none (Bounded, approved in chat on 2026-09-26)\n":                          "",
		"**Spec:** No spec file. Source: memory `tick-fixes-review-notes`.\n":                "",
		"**Spec:** Bounded design approved in chat on 2026-09-26 (no spec file):\n":          "",
		"**Spec:** `notes` then `.pm/specs/z.md`\n":                                          ".pm/specs/z.md",
		"**Spec:** see .pm/specs/typo-desing.md\n":                                           ".pm/specs/typo-desing.md",
		"**Spec:** (.pm/specs/w.md).\n":                                                      ".pm/specs/w.md",
```

- [x] **Step 2: Run it to see it fail**

Run: `go test ./internal/board -run TestParseSpecLine -count=1`
Expected: FAIL, for example `SpecPath = "none", want ""` and `SpecPath = "tick-fixes-review-notes", want ""`.

- [x] **Step 3: Write the implementation**

Replace `specPath` in `internal/board/parse.go`:

```go
// specPath takes the spec path out of the text after "**Spec:**". Only a
// path ending in .md counts, so a plan with no spec ("none", a note in
// backticks) shows no false warning. A path in backticks wins over a bare
// word.
func specPath(s string) string {
	for rest := s; ; {
		i := strings.Index(rest, "`")
		if i < 0 {
			break
		}
		j := strings.Index(rest[i+1:], "`")
		if j < 0 {
			break
		}
		if p := rest[i+1 : i+1+j]; strings.HasSuffix(p, ".md") {
			return p
		}
		rest = rest[i+2+j:]
	}
	for _, f := range strings.Fields(s) {
		// Trim brackets at the start and punctuation at the end, but keep a
		// leading dot so ".pm/specs/x.md" stays whole.
		if p := strings.TrimRight(strings.TrimLeft(f, "(["), ")],.:"); strings.HasSuffix(p, ".md") {
			return p
		}
	}
	return ""
}
```

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/board -count=1`
Expected: PASS.

- [x] **Step 5: Run the gate, then commit**

```bash
git add internal/board/parse.go internal/board/parse_test.go
git commit -m "fix(board): spec line counts only .md paths"
```

Then run `pmb tick plans/2026-09-26-spec-line-md#task-1 --all`.

---

### Task 2: Plan header says how to write "no spec"

**Files:**
- Modify: `plugin/skills/plan/SKILL.md` (the `**Spec:**` line of the Plan Document Header block)
- Test: `internal/plugincheck/skill_plan_test.go`

**verify:** Every place in the plan skill that tells an agent how to fill the Spec line gives the `none (Bounded, approved in chat on <date>)` form for a plan with no spec, and the test goes red if that text is removed. List every place checked.

**Interfaces:**
- Consumes: nothing.
- Produces: nothing.

- [x] **Step 1: Write the failing test**

In `TestSkillPlan`, add to `Must`: `"**Spec:** none (Bounded, approved in chat on <date>)"`.

- [x] **Step 2: Run it to see it fail**

Run: `go test ./internal/plugincheck -run TestSkillPlan -count=1`
Expected: FAIL naming the missing string.

- [x] **Step 3: Write the text**

In `plugin/skills/plan/SKILL.md`, replace:

```
**Spec:** [path to the spec/design doc this plan implements — the plan
argues from the spec, so the spec travels with it; executors read both]
```

with:

```
**Spec:** [path to the spec/design doc this plan implements — the plan
argues from the spec, so the spec travels with it; executors read both.
A Bounded plan with no spec file writes exactly
`**Spec:** none (Bounded, approved in chat on <date>)`, with no other
backticks on the line: pmb reads a .md path there as the spec]
```

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/plugincheck -count=1`
Expected: PASS.

- [x] **Step 5: Run the gate, then commit**

```bash
git add plugin/skills/plan/SKILL.md internal/plugincheck/skill_plan_test.go
git commit -m "fix(plugin): plan header shows the no-spec Spec line"
```

Then run `pmb tick plans/2026-09-26-spec-line-md#task-2 --all`.

## Fix round 1

### Task F1: A bare word counts only when it is a path

**Files:**
- Modify: `internal/board/parse.go` (func `specPath`, the bare-word loop)
- Test: `internal/board/parse_test.go` (func `TestParseSpecLine`)

**verify:** No bare word without a `/` (like `CLAUDE.md` or `README.md` in a sentence) is ever taken as the spec, every bare `.md` word that contains a `/` still is (so a mistyped spec path still warns), and backtick spans keep the rule they have now. List every Spec line shape checked, and run `go run ./cmd/pmb list` to show no plan in `.pm/plans/` has a false "spec ... not found" line. User ruling 2026-09-26: option A.

- [ ] **Step 1: Write the failing test**

Add to the `cases` map in `TestParseSpecLine`:

```go
		"**Spec:** Design approved in chat (Bounded). Rulings: (1) CLAUDE.md or AGENTS.md wins.\n": "",
		"**Spec:** see README.md and .pm/specs/v.md\n":                                               ".pm/specs/v.md",
```

- [ ] **Step 2: Run it to see it fail**

Run: `go test ./internal/board -run TestParseSpecLine -count=1`
Expected: FAIL with `SpecPath = "CLAUDE.md", want ""` and `SpecPath = "README.md", want ".pm/specs/v.md"`.

- [ ] **Step 3: Write the implementation**

In the bare-word loop of `specPath`, change the condition to `strings.HasSuffix(p, ".md") && strings.Contains(p, "/")`, and add one comment line above it: `// A bare word must look like a path, so a file named in a sentence (CLAUDE.md) is not taken as the spec.`

- [ ] **Step 4: Run the tests to see them pass, then the gate**

Run: `go test ./internal/board -count=1`, then `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`, then `go run ./cmd/pmb list` and check no `! spec` line.
Expected: PASS, no `! spec` line.

- [ ] **Step 5: Commit (gofmt folded in, never a formatting-only commit)**

```bash
git add internal/board/parse.go internal/board/parse_test.go
git commit -m "fix(board): a bare spec word must be a path"
```

Then run `pmb tick plans/2026-09-26-spec-line-md#task-F1 --all`.
