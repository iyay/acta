---
id: SPC-0148
created: "2026-10-18 20:33:10"
hash: s0x5xxp
started: "2026-10-18 20:34:48"
finished: "2026-10-18 20:38:05"
---
# Complete the understanding report grader

## Problem

The grader in `plugin/evals/review-understanding-report/graders/behavior.md`
can pass a landing report without the required risk or boundary, gate results,
cleanup, wiki pages, next action, or trace evidence of reading the final code.
Its six current criteria do not enforce those existing landing requirements.

## Design

Bounded change, approved by the user in chat. Keep criteria 1–6 and append
three required criteria to the existing trace-focused LLM grader:

7. The report states the most important risk or boundary supported by the
   landed code, or says no risk is visible when justified by that code.
   Unsupported or invented claims fail this criterion.
8. The report includes gate results, cleanup, changed wiki pages or an
   explicit none, and a next action. The fixture has no tests, so an honest
   statement that no tests ran is valid; do not require invented test counts.
9. The trace shows a successful read of `src/retry.py` after the successful
   merge, from the final parent checkout or the exact landed commit. Accept
   Read, cat, git show, or another equivalent read. An earlier worktree read,
   a read of the stale review notes, a read of a different revision, a failed
   read, or the final reply merely claiming a read does not count.

Every criterion must hold for PASS. Keep the existing optional-question
contract and the requirement that merging happens before the question.
Judge equivalent wording by meaning; do not require literal section labels.
The change tightens the grader only; it does not change the landing skill,
fixture, prompt, config, other evals, or the eval runner.

## Validation

Use a complete synthetic trace for the current retry fixture as a positive
control. Create three negative versions, each failing exactly one of the
new criteria while preserving the other requirements. Check the old grader
first to expose its permissive result, then verify the tightened grader
passes the complete trace and rejects each incomplete trace. Keep trace
fixtures and replay output outside the repo; use existing grading tools
without adding a runner or dependencies.

Cover equivalent successful source-read tools and revisions in the trace
criteria. Run the existing `review-understanding-report` greenfield eval and
plugin checks, then the required landing gates when build is authorized.

## Scope and workflow

Implementation changes only
`plugin/evals/review-understanding-report/graders/behavior.md`, through a new
minimal one-task plan and an isolated worktree. Build remains on hold until
the user requests it. No wiki page is needed; the contract lives in the grader.
