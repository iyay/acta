---
parent: specs/2026-10-31-cleanup-command-design
depth: minimal
after: [PLN-0201]
id: PLN-0202
created: "2026-10-31 07:35:19"
hash: o05nznk
---
# Cleanup skill implementation plan

**Goal:** Add the `acta:cleanup` skill that turns `acta cleanup` candidates into proposals with checked evidence, gets the user's yes per batch, and writes them with `acta cleanup apply`.
**Spec:** `.acta/specs/2026-10-31-cleanup-command-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only their owning packages.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
The skill writes nothing before the user approves; subagents are read-only and never write; it runs only when the user asks.
Skill prose stays plain, short, and free of repo paths and user names, and fits the plugincheck size caps.
Every command the skill names exists in PLN-0201: `acta cleanup`, `--json`, `--fix`, `apply`.

## Waves

Wave 1: Task 1. Wave 2: Task 2.

### Task 1: acta:cleanup skill and its listing

**Files:** Create `plugin/skills/cleanup/SKILL.md`, `internal/plugincheck/skill_cleanup_test.go`; modify `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`, `.acta/wiki/scratch-dropped-vs-specced.md`.
**Seam:** The skill text read by plugincheck, and `hook.Skills` as printed by the session-start text and `default-rules.md`.
**verify:** The skill holds every step of the spec's skill section in order: show and confirm `--fix` first; one read-only subagent per kind with candidates, in parallel; a proposal of keep, done, dropped, wontfix or fixed with evidence and a one-line reason; the main agent checks each sha and `file:line`, and weak evidence becomes keep; one message per kind in the current channel, at most 10 numbered items, answered with ok or numbers to skip; a plan drop shows its spec drop as a line under it; approved items go to `acta cleanup apply` as one batch. `cleanup` appears in `hook.Skills`, in the session text and in `default-rules.md`, and the fallback rules still match the generated text. The wiki page says `done` means built without a spec with a reason naming the commit, and `dropped` still means not done or not valid.
- [ ] Red: `TestSkillCleanup` and the hook skill-list test fail because the skill file does not exist and `cleanup` is not in `hook.Skills`.
- [ ] Change: Write the skill, add `cleanup` to `hook.Skills` and `default-rules.md`, rewrite the wiki page, then run `scripts/test ./internal/plugincheck/ ./internal/hook/` to green.
- [ ] Commit: `feat(cleanup): acta:cleanup skill`.

### Task 2: eval for proposals before writes

**Files:** Create `plugin/evals/cleanup-proposes-first/case.yaml`, `plugin/evals/cleanup-proposes-first/prompt.md`, `plugin/evals/cleanup-proposes-first/scaffold.sh`, `plugin/evals/cleanup-proposes-first/graders/proposes-first.md`; modify `internal/plugincheck/evals_test.go`.
**Seam:** A greenfield eval case run by `scripts/eval`, and the case-layout checks in plugincheck.
**verify:** Given a scaffold with one built raw scratch item, one stale debt line and one open bug, and a prompt to clean up, the agent sends proposals with evidence and a question, and no planning file changes and no commit happens before a yes; the case passes plugincheck's layout rules.
- [ ] Red: The evals layout test fails because the new case is registered in `evals_test.go` but its files do not exist.
- [ ] Change: Add the case files, run `scripts/test ./internal/plugincheck/ -run 'Test(Eval|Scaffolds)'` to green, then `scripts/eval --case cleanup-proposes-first` with the branch binary on PATH.
- [ ] Commit: `test(cleanup): agent proposes before writing`.

## State
