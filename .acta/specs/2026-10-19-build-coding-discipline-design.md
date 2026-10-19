---
id: SPC-0159
created: "2026-10-19 11:32:17"
hash: rnpuupq
started: "2026-10-19 13:37:40"
finished: "2026-10-19 13:53:36"
---
# Build executor coding discipline

## Scope
Bounded follow-up to SPC-0158 / PLN-0170, kept in a separate spec and plan.
Do not change that plan or its in-flight branch. Build this follow-up only
after PLN-0170 lands, adapting the final build instructions rather than
restoring their earlier structure.

## Design
Make three coding practices explicit for inline, subagent, and dispatch
executors, using existing rules rather than introducing a new workflow:

- Before editing, trace the changed behavior through its relevant callers,
  shared implementation, and existing tests. Read enough to choose the
  smallest complete change; do not scan the whole repository by default.
  Fix shared causes where affected callers pass through, within approved
  scope. Surface an incompatible contract or scope expansion before coding it.
- Before committing, inspect the task diff for unnecessary code, accidental
  contract changes, lost validation/error handling, and redundant tests.
  Reuse the worker's existing self-check and give inline the same contract.
  This is not an independent review, additional report, or extra fix round.
- When execution fails, use evidence to distinguish an implementation
  defect, an environment/setup failure, and a contradiction or missing
  decision in the plan. Route defects through acta:debug, correct routine
  setup within existing authority, and escalate unresolved scope/contract
  decisions. Do not repeat the same failed attempt without new evidence,
  weaken checks, or treat an environment error as meaningful TDD RED.

Keep common guidance in the build entrypoint or its existing shared rules;
the implementer template carries only the handoff needed to apply it.
Preserve executor selection, gates, state, commits, final review, and land.
Keep test selection risk-based: reuse coverage, add only a distinct approved
behavior or concrete regression, and impose no line or assertion quota.

## Files and validation
After PLN-0170 lands, inspect `plugin/skills/build/` and existing references
to select the smallest edit. Update affected checks in
`internal/plugincheck/` and reuse behavioral cases in `plugin/evals/`.
No CLI feature, model setting, framework, ledger, or implementation work now.

Use isolated before/after scenarios for missed shared callers, a meaningful
pre-commit self-check, and failure triage. Inspect actual tool traces, code,
and tests rather than self-reported compliance; reuse or extend an existing
case whenever it protects the same behavior. New cases require a real gap.
Observe the relevant failure before editing guidance, preserve working
coverage, and run focused checks plus the existing review/land gates.
Acceptance is more complete minimal changes and correct failure handling
without redundant tests, per-task reviewers, or duplicate instruction blocks.
