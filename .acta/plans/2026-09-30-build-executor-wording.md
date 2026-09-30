---
parent: debt/2026-09-30-config-file-name
closes: [DBT-0032.01]
created: "2026-09-30"
id: PLN-0043
hash: gh92td7
started: "2026-09-30"
finished: "2026-09-30"
---
# Build Executor Wording Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The build skill stops calling `subagent` the default executor and says the executor comes from `acta voice show`, else ask.

**Architecture:** Text change in `plugin/skills/build/SKILL.md` (frontmatter `description` and one Executors table row), guarded by `TestSkillBuild` in `internal/plugincheck`.

**Tech Stack:** Skill markdown checked by Go tests (`CheckSkill` with `Must` / `MustNot`).

**Spec:** `.acta/specs/2026-09-30-build-executor-wording-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- `plugin/skills/setup/SKILL.md` is not touched.
- The "Before you pick one, run `acta voice show`" paragraph in the build skill stays as it is.
- The `description` must still start with `acta: ` and stay under 1024 characters.
- Before the commit, run `gofmt -l cmd internal` (it must print nothing), `go vet ./...` and `go test ./...`.

## File Map

- `plugin/skills/build/SKILL.md`: line 3 `description`, line 16 table row.
- `internal/plugincheck/skill_build_test.go`: `TestSkillBuild` `Must` and `MustNot`.

## Waves

- Wave 1: Task 1.

---

### Task 1: Build skill takes the executor from the config

**Files:**
- Modify: `plugin/skills/build/SKILL.md` (frontmatter `description`, the `subagent` row of the Executors table)
- Test: `internal/plugincheck/skill_build_test.go`

**verify:** No text in `plugin/skills/build/` calls `subagent` the default executor in any form: reworded, in the description, or in the table. List every place checked. `grep -rn 'default' plugin/skills/build/` shows no line that ties "default" to an executor.

**Interfaces:**
- Consumes: `CheckSkill(t, SkillRule{Name, MaxLines, Must, MustNot})` from `internal/plugincheck`.
- Produces: nothing other tasks use.

- [x] **Step 1: Write the failing test**

In `internal/plugincheck/skill_build_test.go`, `TestSkillBuild`, add to `Must`:

```go
			"the executor `acta voice show` names, else asks which one",
```

and add to `MustNot`:

```go
			"subagent (default", "`subagent` (default)",
```

- [x] **Step 2: Run the test to see it fail**

Run: `go test ./internal/plugincheck/ -run TestSkillBuild`
Expected: FAIL, naming the missing "the executor `acta voice show` names" phrase and both banned "(default" phrases.

- [x] **Step 3: Change the skill text**

In `plugin/skills/build/SKILL.md`, line 3 becomes:

```yaml
description: "acta: Use to run an approved plan. Creates the worktree without asking, then runs every task with a failing test first through the executor `acta voice show` names, else asks which one - subagent (the current harness's own subagents), dispatch (an omp agent in its own herdr tab, through acta:dispatch) or inline (you write the code). Commits each task; review and landing follow through acta:review and acta:land."
```

The first row of the Executors table becomes:

```markdown
| `subagent` | a fresh subagent per task | Claude Code: the Agent tool, with the model ## Models names. omp: `agent()` with `agent="task"` (omp has no model argument; its role config picks the model). |
```

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/plugincheck/ && go test ./...`
Expected: PASS.

Run: `grep -rn 'default' plugin/skills/build/`
Expected: no line that ties "default" to an executor.

- [x] **Step 5: Commit**

```bash
gofmt -l cmd internal && go vet ./... && go test ./...
git add plugin/skills/build/SKILL.md internal/plugincheck/skill_build_test.go
git commit -m "Build skill takes the executor from the config, not a default"
```
