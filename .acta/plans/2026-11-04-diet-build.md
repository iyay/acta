---
parent: specs/2026-11-04-skill-token-diet-design
depth: minimal
id: PLN-0209
created: "2026-11-04 19:45:47"
hash: boe2a2o
---
# Diet: build skill implementation plan

**Goal:** Build loads only the text of the executor it runs.
**Spec:** `.acta/specs/2026-11-04-skill-token-diet-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only `./internal/plugincheck/`.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Skill frontmatter descriptions do not change.
Executor order stays: the `/build <executor>` argument, then `build_executor` from `acta config show`, then `inline`.
`skills/build/SKILL.md` byte cap drops to 7000 or less.

## Waves

Wave 1: Task 1.

### Task 1: one file per executor

**Files:** Create `plugin/skills/build/subagent.md`, `plugin/skills/build/inline.md`; modify `plugin/skills/build/SKILL.md`, `plugin/skills/build/dispatch.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/build_dispatch_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the build skill folder as plugincheck reads it, plus `scripts/eval --case 'build-*'`.
**verify:** Each executor gets every rule it has today and no text of another executor. Cases: every executor-only sentence of today's body (Models, Inline task loop, Worker hand-off, the per-executor fix round) sits in exactly one mode file; text all executors use (worktree, plan, task loop core, close) stays in the body; the body's executor table points each row at its file and says to read only that one; existing skill_build and build_dispatch assertions still find their text, now in the mode file; byte caps hold; `build-*` evals give the same results on branch and main, and the inline-run transcript has no Read of `subagent.md` or `dispatch.md`.
- [ ] Red: a plugincheck test expects `subagent.md` and `inline.md` and a body under 7000 bytes, and fails because neither file exists.
- [ ] Change: move the executor text, point the table, retarget the moved assertions, set caps, run `scripts/test ./internal/plugincheck/` then the evals on branch and main.
- [ ] Commit: `feat(build): load only the running executor's text`.
