---
created: "2026-10-20 07:15:00"
id: SPC-0167
hash: oebb8ga
---
# Close builds without another fast suite

## Problem

Each task already runs focused tests. Build close then requires a fast suite
across the repository before review, and landing runs the full suite after
the final changes. The close suite repeats work and delays the same loop.

## Design

Bounded workflow change approved in chat on 2026-10-20.

- Remove the mandatory repository-wide fast suite from build close for every
  executor. After the wiki step, complete outstanding type checks, show their
  evidence and the task verification evidence, then start review.
- Keep each task's failing regression test and focused verification, required
  formatter checks, type checks, wiki checks, progress checks, review and the
  full landing gates. A change made after a focused test still needs the
  relevant test; this change removes only the extra repository-wide close run.
- Dispatch reply verification must not reinstate that broad fast suite. Read
  the recipient's output and use narrow checks when independent verification
  is required. Keep decisive mutation checks on sensitive changes.
- Keep the no-baseline startup rule and saved landing baseline behavior.
  The full suite runs after review and final polish, before landing, with
  existing failure and integration rerun rules unchanged.
- The fast test command remains available for manual feedback. Do not remove
  `scripts/test` or change its behavior.

## Scope and verification

Update `plugin/skills/build/SKILL.md`, `plugin/skills/build/dispatch.md`,
their plugin checks, and behavioral eval coverage for build close. Update
`.acta/wiki/skill-text-fits-everyone.md` with the new close rule.

Verify that inline, subagent and dispatch close paths do not require a broad
fast suite before review. Preserve wiki-before-review ordering, task test
evidence, required type checks and the landing gates. A behavioral eval must
detect a forbidden broad suite call without telling the agent to skip it in
the prompt. Run focused regression checks during implementation and the normal
full gates at landing. No new setting, dependency or CLI command is needed.
