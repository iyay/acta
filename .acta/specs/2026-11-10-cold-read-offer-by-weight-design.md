---
id: SPC-0198
created: "2026-11-10 18:16:09"
hash: zms9iv8
started: "2026-11-10 19:34:23"
finished: "2026-11-10 19:59:48"
---
# Offer the cold read only when the spec is heavy

Bounded. Worked out with the user on WhatsApp on 2026-11-10, with cold read fixes F1 to F4.

## Problem

`plugin/skills/shape/SKILL.md` offers a cold read after self-review for every Bounded and Architectural spec. A short, simple spec still gets the question, so the user has to say no every time.

## Design

The `## Cold read` section of shape `SKILL.md` decides the offer by path and by concrete signals in the spec, not by a general feeling that the work is big:

- Spike: never offered (unchanged).
- Architectural: always offered (unchanged).
- Bounded: offered only when the spec shows at least one signal:
  - it touches a trust boundary, auth, money or a data migration;
  - it changes an interface, file format or command that other code or users depend on;
  - it touches files in more than one area, for example Go code and skill text.
- Bounded with no signal: skip the offer without a word and go straight to the user review gate (step 8). No extra line hints that a cold read exists, since that would push the user to answer.
- The signals are judged once, at step 7, on the spec as written then. Later edits at step 8 do not bring the offer back.
- When the user asks for a cold read on their own, at any point before the plan, run it from `cold-read.md` as today, whatever the path or signals.

Step 7 of both checklists keeps its place between self-review and user review, so `TestShapeColdRead` still holds. The Bounded step 7 text says the offer depends on the signals. The `## Cold read` section keeps the phrase "Skip goes to final spec approval", which the existing `cold-read-skip` eval anchors on. `cold-read.md` does not change: it is read only after the user accepts, so the signals must live in `SKILL.md`.

## Files

- `plugin/skills/shape/SKILL.md`: the `## Cold read` section and step 7 of the Bounded checklist.
- `internal/plugincheck/skill_shape_test.go`: a check that the `## Cold read` section names the Bounded signals and the silent skip.
- `internal/plugincheck/budget_test.go`: raise the `skills/shape/SKILL.md` byte cap from 7841 by about 400, to fit the new text. Today the file is 7812 bytes.
- `plugin/evals/cold-read-no-signal/` (new): a Bounded spec with no signal; the agent asks for spec review without offering a cold read and never loads `cold-read.md`.
- `internal/plugincheck/evals_test.go`: a row for `cold-read-no-signal` in `evalCases` and an entry in the LLM-grader allowlist.
- `plugin/evals/cold-read-skip/`: kept as it is. It tests a user who turns down an offered cold read.
- `plugin/evals/cold-read-offer/`: kept. Its export spec touches customer data and access rules, so it carries a signal and the offer must still fire. Its grader gains one line saying the spec carries a sensitive signal.

## Testing

- `scripts/test ./internal/plugincheck`.
- At land: full suite, and the eval cases `cold-read-offer`, `cold-read-no-signal`, `cold-read-skip`, `cold-read-accept`, `cold-read-unavailable`.
