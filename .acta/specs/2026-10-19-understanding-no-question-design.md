---
id: SPC-0156
created: "2026-10-19 09:35:11"
hash: c92eb9g
started: "2026-10-19 09:46:01"
finished: "2026-10-19 09:48:52"
---
# Landing explanation without a question

## Problem

With `review_understanding: true`, the `acta:land` report ends with one
optional question: a scenario where the user predicts the outcome. The user
does not want the question. The short explanation stays.

## Design

Bounded change approved in chat on 2026-10-19.

- With `review_understanding: true`, the land report adds only the short
  explanation: main behavior, one concrete example, the main risk or
  boundary, what the tests do and do not prove.
- No question, no follow-up check of an answer, no extra scenario.
- The setting keeps its name, so existing configs keep working.
- With the setting false or missing, nothing changes.

## Files and verification

- Skills: `plugin/skills/land/SKILL.md` (the question, the "land already
  done" rule about answers, and the answer-check paragraph go),
  `plugin/skills/review/depth.md` (says explanation only).
- Setting text: `internal/setup/form.go`, `internal/cli/config_cmd.go`,
  `internal/cli/help.go`, the comment in `internal/config/user.go`. Each says
  the landing is explained, with no question.
- Tests: `internal/plugincheck/skill_land_test.go`, `skill_review_test.go`,
  `evals_test.go` require the explanation and forbid the question.
- Evals: delete `plugin/evals/review-understanding-followup`. Change the
  `review-understanding-report` grader to want the explanation and no
  question. `review-understanding-disabled` stays.
- Verify: `scripts/test` green; `grep -rn "optional question"` over
  `plugin/` and `internal/` finds nothing. Evals are not run unless asked.
