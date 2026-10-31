---
parent: specs/2026-10-31-cleanup-command-design
depth: minimal
after: [PLN-0201]
id: PLN-0202
created: "2026-10-31 07:35:19"
hash: o05nznk
started: "2026-10-31 12:03:53"
finished: "2026-10-31 12:08:17"
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
- [x] Red: `TestSkillCleanup` and the hook skill-list test fail because the skill file does not exist and `cleanup` is not in `hook.Skills`.
- [x] Change: Write the skill, add `cleanup` to `hook.Skills` and `default-rules.md`, rewrite the wiki page, then run `scripts/test ./internal/plugincheck/ ./internal/hook/` to green.
- [x] Commit: `feat(cleanup): acta:cleanup skill`.

### Task 2: eval for proposals before writes

**Files:** Create `plugin/evals/cleanup-proposes-first/case.yaml`, `plugin/evals/cleanup-proposes-first/prompt.md`, `plugin/evals/cleanup-proposes-first/scaffold.sh`, `plugin/evals/cleanup-proposes-first/graders/proposes-first.md`; modify `internal/plugincheck/evals_test.go`.
**Seam:** A greenfield eval case run by `scripts/eval`, and the case-layout checks in plugincheck.
**verify:** Given a scaffold with one built raw scratch item, one stale debt line and one open bug, and a prompt to clean up, the agent sends proposals with evidence and a question, and no planning file changes and no commit happens before a yes; the case passes plugincheck's layout rules.
- [x] Red: The evals layout test fails because the new case is registered in `evals_test.go` but its files do not exist.
- [x] Change: Add the case files, run `scripts/test ./internal/plugincheck/ -run 'Test(Eval|Scaffolds)'` to green, then `scripts/eval --case cleanup-proposes-first` with the branch binary on PATH.
- [x] Commit: `test(cleanup): agent proposes before writing`.

## Polish

### Task 3: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] skill_cleanup_test.go: drop the two-letter Must token "ok", which matches ordinary words; "numbers to skip" carries the check.
- [x] evals_test.go: keep the cleanup case's fixed graders type-checked by folding the exception into mustJudge, not a continue.
- [x] cleanup-proposes-first/prompt.md: the max_turns comment says one or two subagents; the skill starts one per kind, three here.
- [x] cleanup SKILL.md: the specced-scratch advice can never fire, since cleanup lists only raw and brainstorming scratch; say to reuse each id from --json in the apply lines instead.
- [x] Commit: `polish: review notes for PLN-0202`

## Fix round 1

### Task 4: Warn about the cost before subagents start

**verify:** With more than 10 candidates, no subagent starts and nothing is written before the user has seen the count per kind, the subagent and message count, and picked all, some kinds, or the N oldest; with 10 or fewer the flow goes straight to proposals

- [x] User ask 2026-10-31 12:18: cleanup can cost many tokens (128 candidates on this repo), so warn before running. Spec step 2 and the skill: with more than 10 candidates, say the count per kind, the subagents it will start and the approval messages it needs, then ask all, some kinds, or the N oldest; start nothing before the answer.
- [x] Test: skill Must tokens for the new step, and a new eval case cleanup-warns-cost with more than 10 candidates where no Agent call and no write happens before the answer.
- [x] Commit with the trailer `Task: PLN-o05nznk#4`

## Fix round 2

### Task 5: Round 2 review notes

**verify:** The warning threshold reads as a total, the spec names both eval cases, and the warns-cost judge asks for the count per kind; nothing else changes

- [x] Skill and spec step 2: say "more than 10 candidates in total", and fix the mixed "you would start ... it needs" wording.
- [x] Spec Testing list still says one eval case; name both cleanup cases.
- [x] cleanup-warns-cost judge: require the count per kind, not "by kind or in total"; drop input_match "." from no-subagent; raise max_turns from 10 to 14.
- [x] Commit with the trailer `Task: PLN-o05nznk#5`

## Review notes

- Round 1 deep (Spec, Standards): CLEAN, no BLOCKER. Polish: four notes in one commit, full tier, both axes CLEAN, eval cleanup-proposes-first 1.00 before and after. Notes kept, not acted on: skill step order is not tested, only presence; no-apply grader misses other write commands; four graders beyond the plan's file list; allowed_tools omits Bash like no sibling; skill does not name the dirty-file refusal; debt-line examples use short ids; budget caps sit at exact size; dropping the ok token leaves the ok half of the answer unguarded; lean-tests and debug-race exceptions still skip their fixed graders; the proposes-first judge is pinned by filename only.
- Cost warning (user ask 12:18): task 4, round 2 both axes CLEAN; task 5 round 2 notes, round 3 both axes CLEAN; both cleanup evals 1.00 after each change. Notes kept: spec line 145 not rewrapped; spec Testing does not name the case folders; no-subagent grader has no guards line; max_turns comment does not say why 14; spec commit a7382a8 of task 5 has no trailer; judge map built inside the loop; task 4 and 5 were added as fix-round tasks because the plan already had its one polish.

## State

### Next

Review CLEAN and polish done; next is acta:land.

### Findings

Full suite at land went red once on TestWheelNotchesBelongToTheScreenTheyWereGatheredOn (internal/tui scroll_test.go:918); the branch touches no TUI file and the test passed 5 of 5 alone. A TUI flake like the known click one.
