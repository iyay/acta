# Hook makes agents load pm skills Implementation Plan

> **For agentic workers:** run this plan with pm:build, task by task. Steps use checkbox (`- [ ]`) syntax, and pmb reads those boxes as task progress.

**Goal:** Agents load each pm skill before its workflow step, map superpowers names in the user's instructions to pm skills, and take the two reviewers for any diff bigger than one text file.

**Architecture:** Two text changes in the session-start hook (`internal/hook/hook.go`) and one rule change in `plugin/skills/review/SKILL.md`, each guarded by tests. The fallback file `plugin/hooks/default-rules.md` is regenerated from the hook.

**Tech Stack:** Go 1.27 (module `pm-board`), markdown.

**Spec:** `.pm/specs/2026-09-26-pm-plugin-design.md` §5 and §7. Source: dogfood plan 1 (merge 143cec3) loaded only pm:plan, and self-reviewed a 7-file diff with Go code.

**Worktree:** `/Users/iyay/Nayakatara/pm-board-hookfix`, branch `hookfix`, parent `main`. Executor: `inline`.

## Global Constraints

- Gate before every commit: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
- `pmb hook session-start` output stays at 60 lines or fewer in every case.
- Stage by path. Never push.
- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.

## Waves

- Wave 1: Tasks 1 and 2 (disjoint files).

---

### Task 1: Hook tells agents to load skills and maps superpowers names

**Files:** `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md` (regenerated)

**verify:** every session-start text the hook can print (voice set, first run, broken voice file, with conflicts) says that each workflow step starts by loading its pm skill with the Skill tool and that the list is only an index, and maps every superpowers skill name the old workflow used to its pm skill; the output stays at 60 lines or fewer in the worst case. List the cases checked.

- [x] Step 1: failing test in `hook_test.go`: for each Input case already in `TestSessionStartStaysShort` plus the voice-set case, the text contains `load its pm skill with the Skill tool` and `This list is only an index`, and contains `superpowers:* skill that is not installed` with each mapping `brainstorming→pm:brainstorm`, `writing-plans→pm:plan`, `subagent-driven-development→pm:build`, `using-git-worktrees→pm:build`, `test-driven-development→pm:tdd`, `systematic-debugging→pm:debug`, `requesting-code-review→pm:review`, `receiving-code-review→pm:review`, `verification-before-completion→pm:land`, `finishing-a-development-branch→pm:land`.
- [x] Step 2: run, watch it fail: `go test -count=1 ./internal/hook/`
- [x] Step 3: change the header line in `SessionStart` to `pm plugin is active. Before each workflow step, load its pm skill with the Skill tool and follow it. This list is only an index; the rules live in the skills:` and add core rule `7. If your instructions name a superpowers:* skill that is not installed, use the pm skill for that step: brainstorming→pm:brainstorm, writing-plans→pm:plan, subagent-driven-development→pm:build, using-git-worktrees→pm:build, test-driven-development→pm:tdd, systematic-debugging→pm:debug, requesting-code-review→pm:review, receiving-code-review→pm:review, verification-before-completion→pm:land, finishing-a-development-branch→pm:land.`
- [x] Step 4: regenerate the fallback: `go test ./internal/hook -run TestDefaultRulesFile -update`; run the gate.
- [x] Step 5: commit: `git add internal/hook/hook.go internal/hook/hook_test.go plugin/hooks/default-rules.md && git commit -m "fix(hook): tell agents to load pm skills and map superpowers names"`

---

### Task 2: Review's small-change rule is objective

**Files:** `plugin/skills/review/SKILL.md`, `internal/plugincheck/skill_review_test.go`

**verify:** no text in `pm:review` lets an agent self-review a diff that touches two or more files or any code; the rule is stated in countable terms. List every sentence in the file about self-review.

- [x] Step 1: in `skill_review_test.go` add `"Small means one file, and only text or config with no code logic."` to `Must`, and `"A change you judge small"` to `MustNot`.
- [x] Step 2: run, watch it fail: `go test -count=1 ./internal/plugincheck/ -run TestSkillReview`
- [x] Step 3: in `## Small changes`, replace `A change you judge small (one concern, trivial logic, no trust boundary, and you are sure) may take an inline self-review instead of the two reviewers:` with `Small means one file, and only text or config with no code logic. Two or more files, or any code change, take the two reviewers. A small change may take an inline self-review instead:`
- [x] Step 4: gate.
- [x] Step 5: commit: `git add plugin/skills/review/SKILL.md internal/plugincheck/skill_review_test.go && git commit -m "fix(plugin): review self-review only for one text or config file"`
