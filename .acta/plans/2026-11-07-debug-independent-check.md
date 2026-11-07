---
parent: specs/2026-11-07-debug-independent-check-design
depth: minimal
id: PLN-0214
created: "2026-11-07 17:36:17"
hash: kf977qw
---
# Debug independent check

**Goal:** When doubt is high, `acta:debug` reports a cause only after a fresh subagent reran the repro and the control and the cause held, and every Phase 3 report has the same four lines.

**Spec:** `.acta/specs/2026-11-07-debug-independent-check-design.md` (SPC-0191)

**Tests:** fast `scripts/test ./internal/plugincheck/`; full `scripts/test --full` (land only).

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Eval gate after Task 1: run `scripts/eval --case debug-independent-check --runs 3` on the current skill. Three of three pass: stop, do not start Task 2, report to the user, and close the spec as not needed. Otherwise record the results under `## Validation results` and go on.
- Scorecard after Task 2, before review: all seven debug cases, two runs each, on the old arm (main's `plugin/skills/debug/` swapped into a temp copy of this branch's `plugin/`) and on the new skill, with the flags of `scripts/eval`. Record pass counts, tokens and cost per run under `## Validation results`, next to the SPC-0157 reference numbers marked 2026-10-19. A case where the new arm passes fewer runs gets two more runs on both arms; still fewer over four runs is a blocker that sends the work back to Task 2.
- The new case must pass three of three on the new skill.
- Probes and eval temp copies live outside the repo. Never `rm -rf`; leave temp folders for the system to clean.
- Comments and skill text in plain English, short words, saying why.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2 (after the Task 1 eval gate).

## Task 1: Trap eval case

**Files:** create `plugin/evals/debug-independent-check/` (`case.yaml`, `prompt.md`, `scaffold.sh`, `graders/*.md`); modify `internal/plugincheck/evals_test.go`.
**Seam:** `go test ./internal/plugincheck/ -run 'TestEval'` and the eval runner on the case folder.
**verify:** No run can pass by naming the first suspect as the cause, and no run can pass without dispatching a subagent for the check; list each grader and the trace evidence it requires. The scaffold leaves exactly two causes that fit the symptom, the first suspect's control as written leaves the symptom in place, and the scaffold sets `PYTHONDONTWRITEBYTECODE=1`. `prompt.md` allows `[Skill, Read, Bash, Agent]` and keeps the debug tags.
- [ ] Red: `TestEval*` fails because `debug-independent-check` is registered in `evalCases` and the `debug` judge map but its folder does not exist.
- [ ] Change: add the case folder with the trap scaffold, a prompt in the style of the other debug cases, and judge graders for the outcome and the subagent dispatch; its `evalCases` phrase is `Prove the mechanism` in `skills/debug/SKILL.md`.
- [ ] Commit: `test(debug): add independent check trap eval`.

## Task 2: Independent check and four-line report

**Files:** create `plugin/skills/debug/verify.md`; modify `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/budget_test.go`, `internal/plugincheck/evals_test.go`.
**Seam:** `go test ./internal/plugincheck/ -run 'TestSkillDebug|TestBudget|TestEval'`.
**verify:** Every Phase 3 report path, with or without the check, ends in the four lines `Cause:`, `Repro:`, `Evidence:`, `Unverified:`, and no path sends a cause after two `refuted` verdicts or one `could not finish`; list each path through Phase 3 and what it reports. `SKILL.md` grows in neither bytes (6546) nor lines, the folder line cap rises by exactly the lines of `verify.md`, and the old `Must` words (including `observed`) still hold.
- [ ] Red: `TestSkillDebug` fails because its `Must` list now asks for `verify.md`, the three verdicts `holds`, `refuted`, `could not finish`, the two-refutation stop, the no-subagent fallback and the four report lines.
- [ ] Change: write `verify.md` per spec steps 1 to 5, tighten `SKILL.md` so Phase 3 requires a control when `verify.md` applies, points to it and gives the four-line report, set the new caps, and move the case phrase to `could not finish` in `skills/debug/verify.md`.
- [ ] Commit: `feat(debug): verify high-doubt causes with a fresh subagent`.

## Validation results

Filled in by the eval gate and the scorecard.
