---
parent: specs/2026-11-04-skill-token-diet-design
depth: minimal
id: PLN-0212
created: "2026-11-04 19:45:47"
hash: stx3us7
started: "2026-11-04 19:49:19"
finished: "2026-11-04 19:55:45"
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

Wave 1: Task 1. Wave 2: Task 2. Wave 3: Task 3. All edit `budget_test.go`; Tasks 2 and 3 share the tdd skill.

### Task 1: one file per plan depth

**Files:** Create `plugin/skills/slice/full.md`, `plugin/skills/slice/minimal.md`, `plugin/evals/slice-minimal-one-file/` (case.yaml, prompt.md, scaffold.sh, graders); modify `plugin/skills/slice/SKILL.md`, `internal/plugincheck/skill_slice_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the slice skill folder as plugincheck reads it, plus `scripts/eval --case 'slice-*'`.
**verify:** Each depth gets every rule it has today and no text of the other. Cases: full-only text (Plan Document Header, Bite-Sized, full Task Structure, No Placeholders) is in `full.md`; the minimal rules are in `minimal.md`; shared text (scope, split pointer, waves, verify, test commands, hand-off) stays in the body; the body says to read only the file its depth names; byte caps hold; the new eval passes on the branch and its transcript has no Read of `full.md`.
- [x] Red: a plugincheck test expects `full.md`, `minimal.md` and a body under 7000 bytes, and fails because neither file exists.
- [x] Change: move the depth text, add the pointer and the eval case, set caps, run `scripts/test ./internal/plugincheck/` then the eval.
- [x] Commit: `feat(slice): load only the plan depth's text`.

### Task 2: trim tdd

**Files:** Modify `plugin/skills/tdd/SKILL.md`, `internal/plugincheck/skill_tdd_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the tdd skill text as plugincheck reads it, plus `scripts/eval --case 'lean-tests'`.
**verify:** The Iron Law and red, green, refactor steps keep every rule; only the listed sections shrink. Cases: Common Rationalizations and Red Flags become one block of about 10 lines that still names skipping the red run and writing code before the test; Example: Bug Fix and Verification Checklist are gone; byte cap about 6500; `lean-tests` gives the same result on branch and main.
- [x] Red: a plugincheck test caps tdd at 6500 bytes and fails at 10702.
- [x] Change: fold the two sections, drop the two others, run `scripts/test ./internal/plugincheck/ -run 'Tdd|Budget'` then the eval on branch and main.
- [x] Commit: `feat(tdd): trim repeated sections`.

### Task 3: tighten tdd wording to the target

**Files:** Modify `plugin/skills/tdd/SKILL.md`, `internal/plugincheck/skill_tdd_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the tdd skill text as plugincheck reads it, plus `scripts/eval --case 'lean-tests'`.
**verify:** Every rule tdd holds today still holds, in fewer words. Cases: the Iron Law and the red, green, refactor rules keep their meaning; the Good/Bad examples may shrink but each still shows one good and one bad case; the line that the full suite runs only in `acta:land` is back; any reworded pinned sentence has its assertion retargeted to the same rule; byte cap about 6500; `lean-tests` gives the same result on branch and main.
- [x] Red: the tdd cap drops to 6500 and fails at 8105.
- [x] Change: tighten wording and examples, restore the full-suite line, retarget assertions, run `scripts/test ./internal/plugincheck/ -run 'TDD|Budget'` to green.
- [x] Commit: `feat(tdd): tighten wording to the size target`.

## Fix round 1

### Task 4: Restore shared File Structure and tidy tdd pins

**verify:** Each slice depth gets every rule it had before the split; every tdd rule is pinned by its current wording; slice body at or under 7000 and tdd at or under 6500 bytes

- [x] slice File Structure rules moved to full.md, so a minimal-depth agent lost them; put a short File Structure back in the shared body, keep the body at or under 7000 bytes
- [x] [fix] tdd 'Each means: delete the code, start over' covers items whose answer is not delete; scope it to the items that mean delete
- [x] [fix] pin 'The full suite runs only in `acta:land`' in skill_tdd_test.go
- [x] [fix] lean-tests graders guards: lines quote a removed tdd sentence; point them at the current one
- [x] [fix] add slice-minimal-one-file row to evalCases in evals_test.go
- [x] Commit with the trailer `Task: PLN-stx3us7#4`

## Polish

### Task 5: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] slice File Structure lost the rule 'Design units with clear boundaries and well-defined interfaces'; add a short clause back to the body within 7000 bytes
- [x] skill_tdd_test.go pins the delete header and Other answers by presence only; check their order so non-delete items stay under Other answers
- [x] tdd explore item reads 'Explore in a throwaway, then test first.'; restore 'Throw the exploration away, then write the test.' and pin it
- [x] Commit: `polish: review notes for PLN-0212`

## Review notes

- Deep depth, full tier; Spec axis opus, Standards axis sonnet. Round 1 BLOCKED (File Structure lost for minimal, confirmed by fact-check), fixed, round 2 CLEAN. Polish restored the interface clause and the exploration line and pinned the tdd order; polish review CLEAN. Notes [note]: order test anchors on the sunk-cost item, not the last delete item; 'clear boundaries' is implied, not stated; minimal run loads body plus minimal.md (about 8k).

## State

### Next

Paused 2026-11-04 by user. Review CLEAN, polish done, full suite green at f31fc8e (4672 pass).
Next: run scripts/eval (include slice-minimal-one-file and lean-tests), rerun red cases alone up to 2 times, then land per tidy.md.
