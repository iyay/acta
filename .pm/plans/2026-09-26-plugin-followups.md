---
id: PLAN-5
hash: swyd
---
# pm plugin follow-ups Implementation Plan

> **For agentic workers:** run this plan with pm:build, task by task. Steps use checkbox (`- [ ]`) syntax, and pmb reads those boxes as task progress.

**Goal:** Fix four things found in the pm plugin review: when a finding becomes a bug file, stale text in `dispatch` and the house rules, the in-repo worktree fallback in `build`, and branches dropped from the board when a tag has the same name.

**Architecture:** Skill text edits guarded by the existing `internal/plugincheck` tests (new required and forbidden strings), plus a one-line change in `internal/gitc` with a git-backed test.

**Tech Stack:** Go 1.27 (module `pm-board`), markdown skills.

**Spec:** `.pm/specs/2026-09-26-pm-plugin-design.md`. Ruling from the user on 2026-09-26: a review BLOCKER in the diff under review is not a bug; it is unfinished work of that story and lives as a `## Fix round` task in the plan. Only a defect in code already on the parent branch becomes a bug file through `pm:bug`.

**Worktree:** `/Users/iyay/Nayakatara/pm-board-followups`, branch `followups`, parent `main`. Executor: `inline`.

## Global Constraints

- Go tests always with `-count=1`. Gate before every commit: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
- Edit only the lines each task names. Stage by path. Never push.
- Skill line caps from the plugin plan still hold (checked by the existing tests).
- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.

## File map

| Path | Task |
|---|---|
| `plugin/skills/bug/SKILL.md`, `internal/plugincheck/skill_bug_test.go` | 1 |
| `plugin/skills/review/SKILL.md`, `internal/plugincheck/skill_review_test.go` | 1 |
| `plugin/skills/dispatch/SKILL.md`, `internal/plugincheck/skill_dispatch_test.go` | 1 |
| `plugin/references/house-rules.md`, `internal/plugincheck/plugin_test.go` (`TestHouseRules`) | 1 |
| `plugin/skills/build/SKILL.md`, `internal/plugincheck/skill_build_test.go` | 1 |
| `internal/gitc/gitc.go`, `internal/gitc/worktrees_test.go` | 2 |

## Waves

- Wave 1: Tasks 1 and 2 (disjoint files).

---

### Task 1: Skill text matches the rules

**Files:** the five skill/reference files and their tests listed in the file map.

**verify:** no skill or reference file tells an agent to record a BLOCKER in the diff under review as a bug file, to skip a per-task reviewer or fix loop that no longer exists, to use Critical / Important / Minor, to follow a CLAUDE.md section name, or to create a worktree inside the repo; and `pm:review` says where each kind of finding goes (diff BLOCKER to a fix round, pre-existing defect to `pm:bug`). List the files checked.

- [x] **Step 1: Tighten the tests** so each stale phrase is forbidden and each new rule is required:

`skill_bug_test.go`: add `"already on the parent branch"` to `Must`; add `"returned a BLOCKER"` and `"a review BLOCKER"` to `MustNot`.

`skill_review_test.go`: add `"## Where findings go"`, `"pm:bug"` and `"already on the parent branch"` to `Must`.

`skill_dispatch_test.go`: add to `MustNot`: `"Important/Minor"`, `"per-task reviewer"`, `"fix round R of 5"`, `"WORKTREE LANDING"`, `"--Users-"`.

`plugin_test.go` `TestHouseRules`: add `"per-task reviewer"` to the `bad` list.

`skill_build_test.go`: add to `MustNot`: `"default to \`.worktrees/\`"`, `"Step 0 consent"`, `"ls -d .worktrees"`.

- [x] **Step 2: Run, watch them fail** for the stale phrases: `go test -count=1 ./internal/plugincheck/`

- [x] **Step 3: Edit the text.**

`plugin/skills/bug/SKILL.md`:
- In the frontmatter `description`, replace `a review BLOCKER, ` with `a defect found in code already on the parent branch, `.
- In `## When`, replace the line `- \`pm:review\` returned a BLOCKER.` with `- \`pm:review\` found a defect in code already on the parent branch (not one the diff under review brought in).`
- In the `Not a bug file:` line, after `a review NOTE,` add ` a BLOCKER in the diff under review (it is unfinished work of that story and goes into the plan's fix round),`.

`plugin/skills/review/SKILL.md`: add this section right before `## Small changes`:

```markdown
## Where findings go

- A BLOCKER in the diff under review is unfinished work of that story, not a bug. It goes into the one fix task of this round (`## Fix round <n>` in the same plan), where it already shows on the board under the story.
- A defect a reviewer finds in code already on the parent branch, that the diff did not bring in, is a bug. Record it with `pm:bug`, and fix it through its own plan with `parent: bugs/<file>`. It never widens this plan.
- A NOTE stays a NOTE: one line in memory, no file, no task.
```

`plugin/skills/dispatch/SKILL.md`:
- Delete the bullet that starts `- **OVERRIDE on the skill:** the recipient skips` (`pm:build` has no per-task reviewer or fix loop to skip).
- In the line starting `**Clean** = Spec axis matches AND zero BLOCKERs`, replace the parenthesis `(the skill's Critical with a reproducible scenario; Important/Minor = NOTE)` with `(the finding bar in \`pm:review\`)`.
- Rename the heading `## Landing — WORKTREE LANDING, from the main checkout` to `## Landing — from the main checkout`.
- In the omp memory harvest paragraph, replace the example `--Users-iyay-Nayakatara-PMIS-codes-febe-be-pmis-.claude-worktrees-qty-fix--` with `--home-me-code-app-worktree--`.

`plugin/references/house-rules.md`: in the SKILL line, delete the sentence `Skip the skill's per-task reviewer; review happens on the orchestrator's side.` and put in its place `Review happens on the orchestrator's side, never yours.`

`plugin/skills/build/SKILL.md`:
- In 1a, replace `The user has asked for an isolated workspace (Step 0 consent). ` with nothing (the Worktree section already says not to ask).
- In 1b, replace items 2 and 3 of the directory list with one item: `2. Otherwise use \`../<repo>-<slug>\`, next to the repo, never inside it.`
- Delete the paragraph that starts `Safety check for project-local directories only:` (there are no project-local worktree folders any more).

- [x] **Step 4: Run, watch them pass:** `go test -count=1 ./internal/plugincheck/`, then the full gate.

- [x] **Step 5: Commit**

```bash
git add plugin/skills/bug/SKILL.md plugin/skills/review/SKILL.md plugin/skills/dispatch/SKILL.md plugin/references/house-rules.md plugin/skills/build/SKILL.md internal/plugincheck/skill_bug_test.go internal/plugincheck/skill_review_test.go internal/plugincheck/skill_dispatch_test.go internal/plugincheck/plugin_test.go internal/plugincheck/skill_build_test.go
git commit -m "fix(plugin): review BLOCKERs stay in the fix round; drop stale skill text"
```

---

### Task 2: Branches with the same name as a tag stay on the board

**Files:** `internal/gitc/gitc.go` (`UnmergedBranches`), `internal/gitc/worktrees_test.go`

**verify:** every unmerged local branch shows by its own name, whether or not a tag, remote-tracking ref or another ref shares its short name, and names with slashes keep them. List the name clashes tested.

- [x] **Step 1: Failing test** (append to `worktrees_test.go`):

```go
func TestUnmergedBranchesWithSameNameTag(t *testing.T) {
	repo := setupRepo(t)
	git(t, repo, "checkout", "-q", "-b", "feat/x")
	writeFile(t, filepath.Join(repo, "b.md"), "changed on feat/x\n")
	git(t, repo, "commit", "-q", "-am", "feat/x work")
	git(t, repo, "tag", "feat/x")
	git(t, repo, "checkout", "-q", "-b", "dup")
	writeFile(t, filepath.Join(repo, "a.md"), "changed on dup\n")
	git(t, repo, "commit", "-q", "-am", "dup work")
	git(t, repo, "tag", "dup")
	git(t, repo, "checkout", "-q", "main")

	got, err := UnmergedBranches(repo)
	if err != nil || !reflect.DeepEqual(got, []string{"dup", "feat/x"}) {
		t.Fatalf("got %v %v, want [dup feat/x]", got, err)
	}
	if _, err := BranchFiles(repo, "dup", "."); err != nil {
		t.Fatalf("BranchFiles on a branch that shares its name with a tag: %v", err)
	}
}
```

- [x] **Step 2: Run, watch it fail** (`heads/dup` and `heads/feat/x` come back): `go test -count=1 ./internal/gitc/ -run SameNameTag`

- [x] **Step 3: Fix:** in `UnmergedBranches`, change `--format=%(refname:short)` to `--format=%(refname:lstrip=2)`.

- [x] **Step 4: Run, watch it pass,** then the full gate.

- [x] **Step 5: Commit**

```bash
git add internal/gitc/gitc.go internal/gitc/worktrees_test.go
git commit -m "fix(gitc): list unmerged branches by full name when a tag shares it"
```
