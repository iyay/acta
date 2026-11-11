---
parent: specs/2026-11-12-cold-read-guard-tests-design
depth: minimal
id: PLN-0227
created: "2026-11-12 06:18:11"
hash: wuu0ps6
started: "2026-11-12 06:19:00"
finished: "2026-11-12 06:20:49"
---
# Guard tests do not count as a second area for the cold read

**Goal:** Shape stops counting files whose only job is to check another area, such as a plugincheck test or an eval case for a skill, as a second area, so a Bounded skill change with its guard tests gets no cold read offer.

**Spec:** `.acta/specs/2026-11-12-cold-read-guard-tests-design.md` (SPC-0200)

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The new sentence joins the end of the existing `## Cold read` paragraph; the shape folder gains no line (178 of 180 today).
- Every other phrase `TestShapeColdRead` and the old evals anchor on stays word for word.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Guard files do not count as a second area

**Files:** `plugin/skills/shape/SKILL.md`, `internal/plugincheck/skill_shape_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** `TestShapeColdRead` and the byte cap table over the text of `plugin/skills/shape/SKILL.md`
**verify:** No reader of `## Cold read` can count a file whose only job is to check another area as a second area, and no other rule there changes. List each phrase checked: the new `only job is to check another area` sentence, that it names both a plugincheck test and an eval case, and every phrase the test already pins. The shape folder line count does not grow, and the byte cap equals the file's exact new size.
- [x] Red: `TestShapeColdRead` fails because the `## Cold read` section does not say `only job is to check another area`.
- [x] Change: append the spec's sentence to the `## Cold read` paragraph and set the `skills/shape/SKILL.md` cap in `budget_test.go` to the new exact size.
- [x] Commit: feat(shape): guard files do not count as a second area for the cold read

### Task 2: Eval for a skill change with only its guard test

**Files:** `plugin/evals/cold-read-guard-test-only/` (new: `case.yaml`, `prompt.md`, `scaffold.sh`, `graders/behavior.md`), `internal/plugincheck/evals_test.go`
**Seam:** the `evalCases` table and the LLM-grader allowlist in `evals_test.go`, and the eval case itself
**verify:** The case passes only when the agent, resuming shape after self-review on a Bounded spec that rewords one sentence of `skills/greet/SKILL.md` and updates the phrase in `check/skill_test.go`, asks for spec review with no cold read offer or hint and never loads `cold-read.md`. The grader states in one line that the spec has no signal because its test only checks the skill text. `case.yaml` says `# guards: only job is to check another area`, and `evals_test.go` pins the same phrase in shape `SKILL.md`.
- [x] Red: `evals_test.go` fails because its new `cold-read-guard-test-only` row points at a case folder that does not exist yet.
- [x] Change: copy `cold-read-no-signal` into the new folder and swap its scaffold and spec for the greet skill repo the spec describes.
- [x] Commit: test(evals): cover the cold read skip on a skill change with only its guard test

## Review notes

- Review level deep (from review_depth), no floor raised it. Round 1 CLEAN: BLOCKER 0, NOTE 5, pre-existing 0. The one [fix] (skill sentence says 'a test over skill text', spec said 'a plugincheck test') was settled by the user on 2026-11-12: the spec now uses the generic words, per the skill-text-fits-everyone wiki page; the plan Goal and Task 1 verify still say plugincheck. Open [note]s: guards line lives in the grader like the siblings, not case.yaml; only 'eval case for a skill' is pinned, not the test half; extra pinned phrase beyond the spec.

## State

### Next

Both tasks committed. Round 1 deep review running over 2a1bb41..bef626f (merge-base, main moved).
Then polish or fix round, then acta:land with full suite and the cold-read-* evals.
