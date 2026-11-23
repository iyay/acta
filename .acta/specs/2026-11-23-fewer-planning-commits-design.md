---
parent: scratch/2026-11-23-fewer-planning-commits
id: SPC-0217
created: "2026-11-23 20:24:12"
hash: pj3cq7u
started: "2026-11-23 20:42:48"
---
# Fewer planning commits

## Why

Every repo that uses acta gets a long stream of small planning commits. Counts from `git log` on 2026-11-23, last 30 days:

| repo | commits | one-file | planning |
|---|---|---|---|
| acta | 348 | 39% | 48% |
| walkie | 2633 | 45% | 46% |
| shopify-app-admin-dashboard | 124 | 54% | 57% |
| clever-platform (4 days) | 325 | 36% | 40% |

One feature here, cache-expiry-nudge, made 10 commits: 2 with code, 1 release, and 7 planning commits (scratch new, spec, assign ids, spec cold-read edit, plan new, spec dates, scratch done).

acta already folds a planning write into HEAD (`CommitOrFoldIn` and `CommitPathsIn` in `internal/gitc/gitc.go`), many paths included. Three checks stop most folds today:
- `canFold` takes HEAD only when its subject starts with `chore(`, or when HEAD only added exactly the same files. So a `docs(spec)` HEAD from `acta commit`, or a `chore: assign short ids` HEAD, takes no fold.
- `sameKindInRoot` needs HEAD and the new write to share a kind (`SubjectKind`). `chore(scratch)` then `chore(spec)` never fold, and `chore: ...` has no kind at all.
- `acta spec-dates` calls `gitc.Commit`, which never folds.

Land adds more commits after the merge: `fixed_in` on each fixed bug, the debt ticks (committed by hand, since `acta tick` never commits), and `acta spec-dates` on main.

The user asked for this on 2026-11-23: no stream of small planning commits, at most one planning commit per step.

## Design

### Wider fold on main

`canFold` accepts HEAD as a fold target when all of these hold:
- every file HEAD touches sits under the planning root, whatever its subject (`chore(<kind>)`, `chore: ...`, `docs(spec)`);
- every path of the new write sits under the planning root;
- HEAD is not pushed, not a merge, has no tag, and has the same author as the user;
- no other worktree stands on HEAD or on a commit built on it (`otherCheckoutHolds`, as now).

When these hold, the `chore(` prefix check and the same-kind check do not apply. When the write or HEAD has a file outside the planning root, both checks stay exactly as now.

Subject of a fold:
- one kind: as now (`FoldSubject`);
- mixed kinds, or a subject with no kind: `chore(acta): <kinds> of <stems>`, for example `chore(acta): scratch, spec, plan of cache-expiry-nudge`. Every earlier subject stays as a body line, as now.

`acta spec-dates` uses `CommitOrFoldIn` in place of `gitc.Commit`. `MarkItem` in `internal/write/mark.go` passes `cfg.Root` to `CommitPathsIn` so its writes can fold too.

A planning write never folds into a code commit on main. A pushed or tagged commit is never amended.

### Land bookkeeping

Three writes close a plan: `fixed_in` on each bug the plan fixes, the ticks of each debt item in its `closes:`, and the dates of its spec. They move into one place per land mode.

Tidy mode (the default). `acta tidy` gets a `--plan <hash>` flag. After it builds the target tree (the merge of parent and branch), it applies the three writes to the planning files of that tree:
1. `fixed_in: PLN-<hash>` on the bug named by the plan's `parent: bugs/<name>`, and on each `BUG-nnnn` in its `closes:`;
2. every line of each `DBT-nnnn` in `closes:` ticked, as `acta tick --all` does;
3. the spec dates, as `acta spec-dates` writes them.

The writes go into the last commit of the tidy chain, with one body line `Land: fixed_in, ticks and dates of PLN-<hash>`. Tidy already folds `chore` commits into the commits next to them (`foldTypes` in `internal/tidy/tidy.go`), so no new commit is made. `proof` accepts the final tree when it equals the target tree plus these writes plus the hash swaps it allows today.

Why this place:
- The writes read the target tree, so a spec that main changed while the build ran is read in its newest form and does not clash. A sibling plan that landed on main after the fork reads as done.
- Nothing is written in the worktree. A bug reads as fixed only once the parent moves to the tidy tip. A land that stops leaves `refs/acta/tidy/<branch>` unused and the bug open.
- The tidy proof and the gate reuse of tidy.md step 3 cover the writes, since they touch only planning paths.

Full mode. Right after `git merge --no-ff`, land runs one new command, `acta land-notes <plan hash>`. It makes the same three writes on main and commits them as one commit `chore(acta): land of PLN-<hash>`. `acta spec-dates` stays as a command, so older skill text still works.

`fixed_in` takes either a commit sha (7 to 40 lower-case hex, as now) or `PLN-<hash>`, the 7 character plan hash. It is the hash the `Task: PLN-<hash>#<n>` trailer uses, so it never changes on a renumber and `acta commits` finds the fix commits. `PLN-0250` (a short id) is refused, since `acta id --fix-duplicates` can renumber it. Old bugs with a sha stay valid. The board rule that reads a bug as fixed from `fixed_in` (`internal/board/board.go:654`) stays as is. The TUI bug detail shows the plan title next to a `PLN-<hash>` value.

`acta cleanup apply` keeps taking only a sha for `fixed_in` (`internal/cleanup/apply.go:130`), for fixes made outside a plan. It needs no change.

Skill lines that change:
- `plugin/skills/land/SKILL.md` step 5: the `acta spec-dates` sentence becomes `acta land-notes <plan hash>`. Steps 8 and 9 go, since `land-notes` does both.
- `plugin/skills/land/tidy.md`: step 1 passes `--plan <plan hash>` to `acta tidy`. Step 4 drops the `fixed_in` sentence. Step 5 drops the `acta spec-dates` sentence. Step 8 ("Step 8 gets the commit from 4") and the step 9 reference to debt ticks go.
- `plugin/skills/bug/SKILL.md`: the `fixed_in <merge sha>` line says land writes `fixed_in: PLN-<hash>` for a plan, and a sha is only for a fix made outside a plan.

`acta id --fix-duplicates` after a land can still write on main when it renumbers. That write folds when the rules above allow.

### Edge cases

- Two topics fold into one commit on main, for example a bug note and a new spec. Both are planning and the subject names every stem.
- Two sessions write planning on main at once. A fold amends the other session's planning commit when author and push state allow. `busy()` stops a fold during a merge, rebase, cherry-pick or revert, and the write then makes a plain commit.
- A worktree stands on HEAD (a build runs, or a stale worktree is left). The first write makes a new commit, and later writes fold into that one.
- The user commits code by hand after a planning commit. HEAD is then a code commit, so the next planning write makes a new commit.
- A plan with no bug and no debt in `closes:` gets only its spec dates. A plan with no spec gets no dates.
- A bug or debt item named in `closes:` is missing from the target tree. Tidy stops with an error that names it, and the land stops as for any tidy error.

### Files

- `internal/gitc/gitc.go`: `canFold`; `internal/gitc/subject.go`: the `chore(acta)` subject.
- `internal/cli/specdates.go`: fold. `internal/write/mark.go`: pass `cfg.Root`.
- `internal/tidy/tidy.go`: `--plan`, the three writes on the target tree, `proof`. `internal/cli` tidy command: the flag.
- New `acta land-notes` command in `internal/cli`, sharing the write code with tidy.
- `internal/write/ops.go`: `fixed_in` accepts `PLN-<hash>`.
- `internal/tui/detail.go`: plan title for a `PLN-<hash>` value.
- `plugin/skills/land/SKILL.md`, `plugin/skills/land/tidy.md`, `plugin/skills/bug/SKILL.md`.
- Wiki page `commit-subject-form.md`: the `chore(acta)` form. Wiki page `tick-stamps-spec.md`: spec dates now come from tidy or `land-notes`.

## Out of scope

- History of repos already pushed. The user asked for a rewrite of clever-platform only, as a separate job after this spec; the user runs the force push.
- A commit-count report command. The check is a `git log` count one week after land.
- Test rules, which are spec `2026-11-23-fewer-tests-design.md`.

## Verification

Red first, as new rows in the existing fold tables of `internal/gitc/gitc_test.go`:
- scratch then spec then plan on an unpushed HEAD give one commit with the `chore(acta)` subject;
- a `docs(spec)` HEAD under the root takes a fold, and so does a `chore: assign short ids` HEAD;
- a two-path write with one path outside the root does not fold;
- pushed HEAD, tagged HEAD, code HEAD and a worktree on HEAD each give a new commit.

Rows in the existing tidy tests (`internal/tidy`):
- `--plan` on a bug-fix plan: the tip tree has `fixed_in: PLN-<hash>`, the debt lines ticked and the spec dates; the commit count is the same as without `--plan`;
- main changed the spec frontmatter after the fork: no clash, and both changes are in the tip;
- a sibling plan of the same spec landed on main after the fork: `finished` is written;
- a `closes:` id missing from the tree: tidy fails and writes no ref;
- `proof` refuses a final tree with any other planning change.

Rows in the existing `set fixed_in` tests: `PLN-<hash>` accepted, `PLN-0250` and a bad sha refused. A test for `acta land-notes` in the `internal/cli` style: one commit holding all three writes.

Eval: `land-tidy-progress` gets a bug and a debt item in its plan's `closes:` in `scaffold.sh`. A new grader checks that the tidy tip holds `fixed_in: PLN-` for that bug and the ticked debt line, and that no `acta spec-dates`, `acta set ... fixed_in` or `acta tick` runs in the main checkout after the parent moves. `land-full-progress` gets a grader that `acta land-notes` runs once after the merge and no `acta spec-dates` runs.

Then `scripts/test ./internal/gitc ./internal/tidy ./internal/write ./internal/cli ./internal/tui`, `go vet ./...`, `gofmt -l .`. The diff touches `plugin/skills/land/` and `plugin/skills/bug/`, so the eval gate runs `scripts/eval --tag=land` at land. No eval case carries the `bug` tag, so the bug skill edit is covered by `internal/plugincheck` only.
