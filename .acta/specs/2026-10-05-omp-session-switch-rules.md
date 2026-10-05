---
parent: bugs/2026-10-05-omp-new-session-gets-no-rules
id: SPC-0076
created: "2026-10-05 12:56:11"
hash: ltwf3br
---
# omp sends the rules after /new, and doctor flags planning files acta never stamped

Status: Bounded, approved by the user in chat on 2026-10-05. Fixes BUG-0030.

Why: after /new, resume or fork, omp fires `session_switch`, not `session_start`. The acta omp extension listens to `session_start` and `session_compact` only, so the new session never gets the acta rules. In be-pmis that session wrote a spec and plan with plain git on a side branch, never ran `acta id`, and put a task output file in `.acta/plans/`. Nothing in acta pointed that out.

Design:

1. omp extension (`plugin/omp/index.ts`): `session_switch`, for every reason (`new`, `resume`, `fork`), resets the session like `session_start` does: `rulesSent` goes back to false and `acta hook session-start` runs again with source `startup`. The next agent turn then carries the rules.
2. `acta doctor` (`internal/doctor`) gets a check named `files`. It reads the planning folders of the checkout it runs in, not other branches or worktrees:
   - a file in `specs/` or `plans/` with no `id:` in its frontmatter gives `warn`, names the file, and the fix is `acta id`;
   - a file in `plans/` with no `### Task` heading gives `warn`, names the file, and the fix is to move it out of `plans/`;
   - none of these: `ok`.
3. The last task adds 1 to the patch version in the three plugin files.

Tests:
- omp: session start, a turn, then `session_switch` with each reason, then a turn: that second turn carries the rules.
- doctor: a stamped plan with tasks (ok), a plan with no id (warn), a spec with no id (warn), a plans file with no task heading (warn).

Out of scope: fixing the files in be-pmis.
