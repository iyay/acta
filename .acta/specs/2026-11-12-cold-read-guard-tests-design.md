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

The `## Cold read` section of `plugin/skills/shape/SKILL.md` keeps the "more than one area" signal. One sentence joins the end of the existing paragraph, so the file gains no line:

> Files whose only job is to check another area, such as a test over skill text or an eval case for a skill, do not count as a second area.

To count areas, drop those files first, then count what is left. In this repo that drops `internal/plugincheck/*_test.go` changes (text checks, byte and line caps, the eval table) and `plugin/evals/<case>/` folders when they only check a skill the spec changes. Go code that changes how the program behaves, plus skill text, still counts as two areas.

The other signals, the Spike and Architectural rules, and the phrases the old evals anchor on stay as they are.

## Files

- `plugin/skills/shape/SKILL.md`: the sentence above, appended to the `## Cold read` paragraph.
- `internal/plugincheck/skill_shape_test.go`: `TestShapeColdRead` adds `only job is to check another area` to its list of required phrases.
- `internal/plugincheck/budget_test.go`: set the `skills/shape/SKILL.md` byte cap to the file's exact size after the edit (today 8270), as the header comment of that file asks. The shape folder is at 178 of 180 lines, which is why the sentence goes into the existing paragraph.
- `plugin/evals/cold-read-guard-test-only/` (new), copied from `cold-read-no-signal` and changed:
  - scaffold: a small repo with `skills/greet/SKILL.md` and one Go test, `check/skill_test.go`, that reads that file and checks it holds one phrase. No CLI, no other code.
  - spec: Bounded, rewords one sentence of the greet skill and updates the phrase in the test. No trust boundary, no interface others call, no data.
  - the grader says `# guards: only job is to check another area`, like its sibling cases. The skill sentence stays generic, with no name from this repo, so it fits any user.
  - grader: the agent asks for spec review and never offers a cold read. One line tells the judge why the spec has no signal: it edits skill text and a test whose only job is to check that text.
- `internal/plugincheck/evals_test.go`: a row `{"cold-read-guard-test-only", "skills/shape/SKILL.md", "only job is to check another area"}` in `evalCases`, and an entry in the LLM-grader allowlist.

## Testing

- `scripts/test ./internal/plugincheck`.
- At land: full suite, and the eval cases `cold-read-offer`, `cold-read-no-signal`, `cold-read-guard-test-only`, `cold-read-skip`, `cold-read-accept`, `cold-read-unavailable`.
