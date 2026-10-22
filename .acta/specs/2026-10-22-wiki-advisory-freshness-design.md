---
created: "2026-10-22 22:54:50"
id: SPC-0172
hash: trbirsx
started: "2026-10-22 23:10:46"
---

# Wiki freshness warnings and narrower paths

## Goal

Reduce mandatory wiki maintenance while keeping useful project notes and visible
reminders to review them.

## Check behavior and ownership

`acta wiki check [<range>]` reports stale pages as warnings and exits 0 when
there are no errors. Unreadable pages, invalid metadata or paths, invalid ranges,
and failures to read git history remain errors with a nonzero exit. Mixed runs
print both warnings and errors and exit nonzero.

Replace the stale message's instruction to fix the page or run `wiki ok` with:
`warning: stale: <sha> "<subject>" changed its paths after the page was last checked; review the page against relevant changes and update any incorrect facts`.
A warning does not require a page edit or a confirmation commit.

Keep the existing ancestry calculation, range filtering, `.checked`, and
`acta wiki ok`. A range selects affected pages; freshness still uses all history
reachable from HEAD. A covered branch commit can therefore produce a warning
even when it changes no fact in the page. This is a review reminder, not proof
that the page is incorrect.

The agent doing the build reads the affected pages and relevant diffs, and fixes
incorrect facts before review. If a page is still correct after that review,
the agent may run `wiki ok` to record the review, but is not required to do so.
A maintainer or agent reviewing existing warnings can use the same command.
No automatic background checking, blanket confirmations, or timestamp refresh
is added. Without a recorded review, no-range warnings may persist across
branches; that is an accepted cost of keeping optional review records. Land
blocks on errors, not on the presence of warnings or nonempty check output.

## Paths and review before page edits

Limit path narrowing to these three known mismatches:

| Page | Replacement paths |
| --- | --- |
| `dispatch-plan-path` | `internal/cli/dispatch.go` |
| `fix-duplicate-ids` | `internal/cli/id.go`, `internal/write/ids.go` |
| `tick-stamps-spec` | `internal/cli/tick.go`, `internal/cli/specdates.go`, `internal/write/mark.go`, `internal/write/specdates.go` |

These pages must stop matching `internal/cli/state.go` and must still match each
listed file. Other pages retain their paths. In particular, preserve the five
broad TUI pages and the four pages covering `plugin/omp/`. Additional narrowing
requires a separate proposal rather than expanding this change to all 43 pages.

A page edit, including frontmatter alone, counts as a check in the existing
freshness engine. Before the first wiki edit in the worktree, capture the
no-range check output. Read the full body of every page to be edited and compare
its claims with the relevant code and changes before editing its paths or body.
Record those checks in the plan's verification evidence. Do not treat the
resulting disappearance of a warning as evidence that this review happened.

At spec revision time, the two stale pages are `plugin-loads-from-repo` and
`plugin-ships-in-binary`. They are outside the path changes and must not be edited
or marked checked just to clear the baseline. If implementation needs to edit
either, review its claims against the outstanding relevant changes first and
record that evidence before committing. Re-capture the baseline at build time;
apply the same rule to any newly stale page that needs an edit.

## Documentation and workflow

Update all six existing surfaces together:

- `plugin/skills/land/SKILL.md`: errors block landing; remove the routine
  instruction that a stale page only needs `wiki ok`.
- `plugin/skills/build/SKILL.md`: review affected notes and correct wrong facts;
  confirmation of unchanged notes is optional.
- `plugin/skills/build/wiki.md`: distinguish errors from warnings, explain
  optional `wiki ok`, and replace "until it prints nothing" with "until there
  are no errors and any incorrect facts found in review have been fixed".
- `plugin/skills/migrate/SKILL.md`: seeding must resolve check errors and correct
  wrong facts; warnings alone do not fail the seed gate.
- `internal/cli/help.go`: describe warnings and optional review recording.
- `.acta/wiki/wiki-freshness-from-git.md`: explain advisory freshness, retained
  ancestry behavior, optional confirmations, and page-edit review semantics.

## Verification

Use focused regression tests for stale-only success, mixed warning/error
failure, unchanged hard errors, range filtering, and the revised message with
no instruction to run `wiki ok`. Verify all positive and negative path matches
listed above and the retained optional confirmation behavior.

Extend the existing `wiki-close` eval rather than adding a new case. Keep its
incorrect page, which must be corrected, and add a still-correct stale page
that may remain unchanged and unconfirmed. Its graders must accept optional
confirmation, require a successful check and correction of wrong facts, and
allow a nonempty warning result before review.

Changes to skills trigger the existing eval gate at land. Run that required
gate and report its result; a focused eval during development does not replace
it. Keep existing wiki, CLI, hook, and relevant plugin checks green.

## Boundaries

No new memory subsystem, search command, session summary, dependency, page type,
configuration, freshness algorithm, or migration. This change reduces mandatory
maintenance; it does not prove page accuracy or speed up history checks.
