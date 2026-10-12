# Lean tests: one test per behaviour, at the layer that owns it

Status: Bounded, design approved by the user in chat on 2026-10-12.

## Problem

Since 2026-09-20 this repo added 26,036 lines of Go code and 58,779 lines of tests (2.3x). Two patterns cause most of it:

1. **Implementers add cases the plan never named.** PLN-kfizu27 task 01 named 8 cases in its `verify:` line; the implementer wrote 24 subtests. The skill text asks for it: the adversarial input list in `tdd/SKILL.md` applies "where the input can take them" (almost always), the checklist asks that "every new function/method has a test" and "edge cases and errors covered", the mutation check in `writing-good-tests.md` lists "missing validation for zero, empty, nil...", and the implementer self-review asks "Are tests comprehensive?".
2. **The same behaviour is tested again at every layer.** Plans split tasks by layer, and each implementer sees only its own task, so it proves the lower layer again. PLN-kfizu27 tests the merge cases in `gitc`, again in `wiki`, again in `cli`. PLN-xego2mn tests a bad `review_depth` value in `config` and again in `cli`.

`acta:review` already treats a missing test for a path the plan does not cover as a NOTE, and keeps the adversarial lens for trust boundaries. The tdd and build text do not match it.

## Goal

Each behaviour the plan names gets one test, at the layer that owns it. Layers above test only their wiring. Implementers write the plan's cases and no others. This holds for any model the implementer runs on.

Out of scope: the implementer model (set by `subagent_models`), pruning existing tests, and the review rules.

## Design

1. **`plugin/skills/slice/SKILL.md`**: each case in a `verify:` line is named once, in the task whose layer owns the behaviour. A task for a layer above says "wiring only": one success path and one error passed up from below. Both the full task template and the `minimal` task shape say this.
2. **`plugin/skills/tdd/SKILL.md`**:
   - `## Cover the plan, not the universe`: the cases to write are the plan's cases, tested at the layer that owns them; a layer above tests its wiring only.
   - The adversarial input list applies only where the input crosses a trust boundary (user input, files, network, auth, money, migrations, deletes), in the same words as the review deep lens.
   - The checklist lines "Every new function/method has a test" and "Edge cases and errors covered" become "Every case the plan names has a test, and no other".
3. **`plugin/skills/tdd/writing-good-tests.md`**: the mutation item "Missing validation for zero, empty, nil, unauthorized, or malformed input" becomes "Missing validation the plan names".
4. **`plugin/skills/build/implementer-prompt.md`**: a case the plan does not name is not written as a test; it goes in the report as one line, and the orchestrator decides. The self-review line "Are tests comprehensive?" becomes "Does every plan case have a test, and nothing beyond?".

Risk: thinner tests at upper layers can miss wiring that changes behaviour. The one passed-up error per layer and review's "goes red when reverted" check cover it.

## Tests

- `internal/plugincheck/skill_tdd_test.go`: `Must` gains the new wiring-only and trust-boundary phrases; `MustNot` gains "Every new function/method has a test" and "Edge cases and errors covered". Matching checks for the slice and build skill rules. Line and byte caps refitted.
- New eval `plugin/evals/lean-tests`: greenfield, a plan with two layers and three cases. Passes when no case is tested at both layers and no test covers a case outside the plan.
- Not a gate: the test-to-code line ratio of the next two real plans is noted in their review notes, against 2.3x today.

## Wiki

None. The rules live in the skill files.
