---
id: SPC-0200
created: "2026-11-12 06:04:12"
hash: hbrevv0
---
# Guard tests do not count as a second area for the cold read

Bounded. Worked out with the user on WhatsApp on 2026-11-12. Follows SPC-0198.

## Problem

Shape offers a Bounded spec a cold read when it "touches more than one area, such as Go code and skill text". In this repo almost every skill change also edits a test in `internal/plugincheck`, which only checks the skill text. So nearly every Bounded spec here counts as two areas, and the user still gets the offer on most specs.

## Design

The `## Cold read` section of `plugin/skills/shape/SKILL.md` keeps the "more than one area" signal and adds one rule: a test that only checks another area's text, such as a plugincheck test over skill text, does not count as a second area. Go code that changes how the program behaves, plus skill text, still counts.

The other signals, the Spike and Architectural rules, and the phrases the old evals anchor on stay as they are.

## Files

- `plugin/skills/shape/SKILL.md`: one sentence in `## Cold read`.
- `internal/plugincheck/skill_shape_test.go`: `TestShapeColdRead` requires the new sentence.
- `internal/plugincheck/budget_test.go`: raise the `skills/shape/SKILL.md` byte cap from 8270 to fit the sentence.
- `plugin/evals/cold-read-guard-test-only/` (new): a Bounded spec that edits a `SKILL.md` and its plugincheck test, and nothing else with a signal. The agent asks for spec review and never offers a cold read. Built on the `cold-read-no-signal` case.
- `internal/plugincheck/evals_test.go`: a row for the new case in `evalCases` and an entry in the LLM-grader allowlist.

## Testing

- `scripts/test ./internal/plugincheck`.
- At land: full suite, and the eval cases `cold-read-offer`, `cold-read-no-signal`, `cold-read-guard-test-only`, `cold-read-skip`, `cold-read-accept`, `cold-read-unavailable`.
