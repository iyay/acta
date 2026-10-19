---
id: SPC-0158
created: "2026-10-19 11:00:25"
hash: qtnyr4i
started: "2026-10-19 11:04:23"
finished: "2026-10-19 11:13:46"
---
# Build efficiency and implementation quality

## Problem
The build skill supports inline execution, but its task loop and waves are
written around dispatch. Generic dependency setup and repeated instructions
spend work and context without proving better implementation quality.

## Approved design
This is a Bounded change to the existing build workflow.

- Keep the shared contract in the entrypoint. Give inline execution a direct
  task loop; load worker and dispatch details only for the chosen executor.
  Preserve executor precedence, fallbacks, worktree isolation, readiness,
  approval gates, progress ticks, task trailers, wiki updates, closing review,
  and landing. Do not change user settings or add per-task reviewers.
- Follow the repository's setup commands and package manager. Install only
  what is needed for the required checks; preserve the clean-baseline gate.
- Read the plan and spec once, then use the current task and relevant code.
  Reuse acta state and commit history for recovery; add no parallel ledger.
- Use the smallest implementation that satisfies the approved behavior.
  Reuse existing code before adding dependencies or abstractions. Preserve
  validation, data-loss protection, security, and accessibility.
- Select tests by the behavior or concrete failure they protect. Reuse or
  extend an existing test before adding another. Test observable behavior at
  the most useful boundary. Avoid duplicate coverage across layers, internal
  implementation assertions, tautologies, large fixtures, and excess mocks.
  Assertions cover the essential result and any meaningful side effects.
  Additional cases outside the plan remain scope questions, not automatic work.
- Keep failing-test-first execution. RED must expose the missing or broken
  behavior, not an unrelated setup error. Work one behavior at a time.
  Preserve repository coverage requirements; 80% is an example, not a new
  global threshold. Neither assertion count nor test-to-code line ratio is a
  hard limit. A large test diff prompts a check for avoidable maintenance cost.
- Run focused tests during iteration and the task's relevant checks before
  commit. Keep the full suite at the existing land gate. Reuse valid evidence
  only while changes and findings have not invalidated it. Required checks
  remain mandatory; report unavailable checks without claiming a pass.

## Scope
Change `plugin/skills/build/`, related checks in `internal/plugincheck/`,
and the smallest necessary set of behavioral cases in `plugin/evals/`.
Update affected wiki facts when necessary. Keep `tdd` and `lean` as the
sources of their rules rather than copying their full instructions into build.
No CLI features, new test framework, model defaults, or shipping automation.

## Validation and acceptance
Compare the old and revised skill with the same realistic requests in the
existing isolated eval harness. Reuse existing cases; add cases only for gaps.
Cover inline execution, executor routing, recovery from recorded progress,
and pressure to overproduce tests. Inspect artifacts and tool traces, not
only agent claims. Record instruction size and observed tool/check overhead
alongside correctness and workflow compliance. Do not claim improved runtime
or implementation quality from shorter text alone.

The revised instructions must be smaller without moving duplicates into
companion files. Required gates and important behavior must remain covered.
New tests must protect distinct behavior or a demonstrated regression, not
merely enforce new wording. Maintain packaging and skill budgets; run the
relevant checks and existing review/land gates before claiming completion.

## Source adaptations
- Ponytail: reuse first and the smallest complete change.
  https://github.com/dietrichgebert/ponytail/blob/main/skills/ponytail/SKILL.md
- Matt Pocock: public-boundary tests and one behavior per RED/GREEN cycle.
  https://github.com/mattpocock/skills/blob/main/skills/engineering/tdd/SKILL.md
- Superpowers: direct inline execution, recoverable progress, completion evidence.
  https://github.com/obra/superpowers/blob/main/skills/executing-plans/SKILL.md
- pstack: compare observable agent behavior under the same scenarios.
  https://github.com/cursor/plugins/blob/main/pstack/skills/poteto-mode/playbooks/eval.md
- gstack: verify the affected surface and disclose untested behavior.
  https://github.com/garrytan/gstack

Adapt these practices to acta; do not import their tooling, model assignments,
publication actions, or conflicting workflow rules.
