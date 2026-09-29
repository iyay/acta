---
id: PLAN-30
hash: a7ks
---
# Specs and Plans on Main Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Specs and plans are written and committed on main; the worktree is made only when build starts, and plan ticks get committed at land.

**Architecture:** Skill text only. Each skill file gets new wording, and its existing `internal/plugincheck/skill_<name>_test.go` gets Must / MustNot strings that pin the new rule.

**Tech Stack:** Markdown skill files, Go tests (`CheckSkill` in `internal/plugincheck/check.go`).

**Spec:** `.acta/specs/2026-09-29-docs-on-main-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Skill text is English, short words, short sentences.
- Keep every skill under its `MaxLines` in its test.
- `plugin/skills/dispatch/SKILL.md` does not change.
- No Go code outside `internal/plugincheck/*_test.go`.

## File Map

- `plugin/skills/brainstorm/SKILL.md` + `internal/plugincheck/skill_brainstorm_test.go` (Task 1)
- `plugin/skills/plan/SKILL.md` + `internal/plugincheck/skill_plan_test.go` (Task 2)
- `plugin/skills/build/SKILL.md` + `internal/plugincheck/skill_build_test.go` (Task 3)
- `plugin/skills/land/SKILL.md` + `internal/plugincheck/skill_land_test.go` (Task 4)

## Waves

- Wave 1: Task 1, Task 2, Task 3, Task 4 (no shared files).

---

### Task 1: Brainstorm commits the spec on main

**Files:**
- Modify: `plugin/skills/brainstorm/SKILL.md` (Bounded path paragraph, Bounded checklist step 5, "Where files go" paragraph)
- Test: `internal/plugincheck/skill_brainstorm_test.go`

**verify:** No place in the brainstorm skill tells the agent to create a worktree or to commit the spec in one; every place that says where the spec is committed says the checked-out branch (usually main). List every spot that names a commit location.

**Interfaces:**
- Consumes: none
- Produces: phrase `commit it on the branch that is checked out (usually main)`

- [ ] **Step 1: Write the failing test**

In `TestSkillBrainstorm`, add to `Must`:

```go
"commit it on the branch that is checked out (usually main)",
"`acta:build` makes the worktree",
```

and add to `MustNot`:

```go
"in the worktree", "create its worktree now", "as the first commit",
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugincheck -run TestSkillBrainstorm`
Expected: FAIL naming the missing Must strings and the found MustNot strings.

- [ ] **Step 3: Edit the skill**

Bounded paragraph: replace
`` `.acta/specs/YYYY-MM-DD-<topic>-design.md` in the worktree, run `` with
`` `.acta/specs/YYYY-MM-DD-<topic>-design.md`, run ``.

Checklist step 5: replace `commit in the worktree` with `commit on main`.

"Where files go": replace the sentence starting "Commit the spec on the branch the work will use" with:

```
Write the spec and commit it on the branch that is checked out (usually main). No worktree yet: `acta:build` makes the worktree once the plan is approved.
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/plugincheck -run TestSkillBrainstorm`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add plugin/skills/brainstorm/SKILL.md internal/plugincheck/skill_brainstorm_test.go
git commit -m "brainstorm: commit the spec on main, no worktree"
```

### Task 2: Plan commits the plan on main

**Files:**
- Modify: `plugin/skills/plan/SKILL.md` (the **Context:** line and the **Save plans to:** list)
- Test: `internal/plugincheck/skill_plan_test.go`

**verify:** The plan skill never tells the agent to write or commit the plan in a worktree, and it says the worktree comes from `acta:build` only after the plan is approved. List every line that names where the plan lives.

**Interfaces:**
- Consumes: none
- Produces: phrase `Commit the plan on main`

- [ ] **Step 1: Write the failing test**

In `TestSkillPlan`, add to `Must`:

```go
"Commit the plan on main", "only after the plan is approved",
```

and add to `MustNot`:

```go
"If working in an isolated worktree",
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugincheck -run TestSkillPlan`
Expected: FAIL

- [ ] **Step 3: Edit the skill**

Replace the **Context:** line with:

```
**Context:** Write and commit the plan on main, next to its spec. `acta:build` makes the worktree, only after the plan is approved.
```

Add a bullet under **Save plans to:**, after the `acta id` bullet:

```
- Commit the plan on main. The worktree branch starts from that commit, so it carries the plan.
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/plugincheck -run TestSkillPlan`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add plugin/skills/plan/SKILL.md internal/plugincheck/skill_plan_test.go
git commit -m "plan: commit the plan on main, worktree comes at build"
```

### Task 3: Build exempts spec and plan, guards the plan on main

**Files:**
- Modify: `plugin/skills/build/SKILL.md` (the `## Worktree` section, first paragraph)
- Test: `internal/plugincheck/skill_build_test.go`

**verify:** The build skill still gives every code change a worktree, names the spec and plan as the only files written on main, and forbids editing a running build's plan on main. List every rule in `## Worktree` and which case it covers.

**Interfaces:**
- Consumes: none
- Produces: phrases `stay on main` and `do not edit its plan on main`

- [ ] **Step 1: Write the failing test**

In `TestSkillBuild`, add to `Must`:

```go
"The spec and plan from acta:brainstorm and acta:plan stay on main",
"do not edit its plan on main",
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugincheck -run TestSkillBuild`
Expected: FAIL

- [ ] **Step 3: Edit the skill**

After the paragraph starting "Do not ask whether to create a worktree", add:

```
The spec and plan from acta:brainstorm and acta:plan stay on main; they are already committed when the worktree is made. While a build runs, do not edit its plan on main: the ticks live in the worktree copy and would clash at merge. A plan change goes in the worktree.
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/plugincheck -run TestSkillBuild`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add plugin/skills/build/SKILL.md internal/plugincheck/skill_build_test.go
git commit -m "build: spec and plan stay on main, guard plan edits during build"
```

### Task 4: Land commits plan ticks before merge

**Files:**
- Modify: `plugin/skills/land/SKILL.md` (step 1, Preconditions)
- Test: `internal/plugincheck/skill_land_test.go`

**verify:** On every land, the worktree's plan ticks are committed before the clean-worktree check and the merge, so no tick is lost when the worktree is removed. List each land step that touches the plan file and its order against the merge.

**Interfaces:**
- Consumes: none
- Produces: commit message `acta: tick <plan>`

- [ ] **Step 1: Write the failing test**

In `TestSkillLand`, add to `Must`:

```go
"acta: tick <plan>", "before the merge, so the ticks reach main",
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugincheck -run TestSkillLand`
Expected: FAIL

- [ ] **Step 3: Edit the skill**

At the start of step 1 (Preconditions), before "Preconditions, all of them", insert:

```
First commit the plan file in the worktree as one commit `acta: tick <plan>`, before the merge, so the ticks reach main. Build never commits the plan file, so without this the worktree is never clean and the ticks die with it.
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/plugincheck -run TestSkillLand`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add plugin/skills/land/SKILL.md internal/plugincheck/skill_land_test.go
git commit -m "land: commit plan ticks before merge"
```
