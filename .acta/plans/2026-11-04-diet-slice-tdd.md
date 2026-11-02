---
parent: specs/2026-11-04-skill-token-diet-design
depth: minimal
id: PLN-0212
created: "2026-11-04 19:45:47"
hash: stx3us7
---
# Diet: slice and tdd implementation plan

**Goal:** Slice loads only the text of its plan depth, and tdd drops to about 6.5k chars.
**Spec:** `.acta/specs/2026-11-04-skill-token-diet-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only `./internal/plugincheck/`.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Skill frontmatter descriptions do not change.
Slice mode order stays: the `/slice` argument, then `plan_depth`, then `full`.
`skills/slice/SKILL.md` byte cap drops to 7000 or less; `skills/tdd/SKILL.md` to about 6500.

## Waves

Wave 1: Task 1. Wave 2: Task 2. Both edit `budget_test.go`.

### Task 1: one file per plan depth

**Files:** Create `plugin/skills/slice/full.md`, `plugin/skills/slice/minimal.md`, `plugin/evals/slice-minimal-one-file/` (case.yaml, prompt.md, scaffold.sh, graders); modify `plugin/skills/slice/SKILL.md`, `internal/plugincheck/skill_slice_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the slice skill folder as plugincheck reads it, plus `scripts/eval --case 'slice-*'`.
**verify:** Each depth gets every rule it has today and no text of the other. Cases: full-only text (Plan Document Header, Bite-Sized, full Task Structure, No Placeholders) is in `full.md`; the minimal rules are in `minimal.md`; shared text (scope, split pointer, waves, verify, test commands, hand-off) stays in the body; the body says to read only the file its depth names; byte caps hold; the new eval passes on the branch and its transcript has no Read of `full.md`.
- [ ] Red: a plugincheck test expects `full.md`, `minimal.md` and a body under 7000 bytes, and fails because neither file exists.
- [ ] Change: move the depth text, add the pointer and the eval case, set caps, run `scripts/test ./internal/plugincheck/` then the eval.
- [ ] Commit: `feat(slice): load only the plan depth's text`.

### Task 2: trim tdd

**Files:** Modify `plugin/skills/tdd/SKILL.md`, `internal/plugincheck/skill_tdd_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the tdd skill text as plugincheck reads it, plus `scripts/eval --case 'lean-tests'`.
**verify:** The Iron Law and red, green, refactor steps keep every rule; only the listed sections shrink. Cases: Common Rationalizations and Red Flags become one block of about 10 lines that still names skipping the red run and writing code before the test; Example: Bug Fix and Verification Checklist are gone; byte cap about 6500; `lean-tests` gives the same result on branch and main.
- [ ] Red: a plugincheck test caps tdd at 6500 bytes and fails at 10702.
- [ ] Change: fold the two sections, drop the two others, run `scripts/test ./internal/plugincheck/ -run 'Tdd|Budget'` then the eval on branch and main.
- [ ] Commit: `feat(tdd): trim repeated sections`.
