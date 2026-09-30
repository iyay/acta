---
parent: bugs/2026-09-27-flaky-fixduplicates-tempdir-cleanup
id: SPC-0060
created: "2026-10-01 15:08:47"
hash: hri519m
started: "2026-10-01 15:15:17"
finished: "2026-10-01 15:38:54"
---
Status: approved by the user on 2026-10-01 (Bounded). Fixes four bugs: BUG-0001 (`.acta/bugs/2026-09-27-flaky-fixduplicates-tempdir-cleanup.md`), BUG-0005 (`.acta/bugs/2026-09-29-doctor-stale-link-fix-hint-wrong.md`), BUG-0009 (`.acta/bugs/2026-09-29-set-cannot-change-title.md`) and BUG-0010 (`.acta/bugs/2026-09-29-wheel-notches-lost-when-a-reload-lands-before-the-frame-tick.md`).

# Fix four old bugs

The four fixes touch different files, so they can be built side by side. After landing, each bug is closed with `acta set bugs/<stem> fixed_in <merge sha>`.

## BUG-0001: test temp folders sometimes fail to delete

**Root cause, proven 2026-10-01.** With git 2.55, every `git commit` starts `git maintenance run --auto --quiet --detach` (seen with `GIT_TRACE2=1`). That child runs on after the commit returns and takes a lock file inside `.git/objects` (the git binary holds the `%s/maintenance` lock path). When `t.TempDir` cleanup runs at that moment, removing `.git/objects` fails with `unlinkat ... directory not empty`. It was seen in two different tests that commit (`TestFixDuplicatesKeepsTheFileThatReachedTheBranchFirst` and `TestNewBug`). 150 runs in a row did not repeat it, because it is a timing race.

**Fix.** `scripts/test` exports `GIT_CONFIG_COUNT=1`, `GIT_CONFIG_KEY_0=maintenance.auto` and `GIT_CONFIG_VALUE_0=false` before it runs `go test`, for both the short and the `--full` path. Every git child a test starts then skips auto maintenance. acta's own commits in a user's repo do not change.

**Test.** `scripts/test_test.go` checks that the script sets those three values.

## BUG-0005: acta doctor gives an omp command that does not exist

**Fix.** The `stale-links` hint in `internal/doctor/doctor.go` becomes `rm <full path of the dead link>` for each dead link, because removing the leftover symlink is the only thing that clears it. omp has no `unlink` action. The tests that checked for the old wrong text expect the new text.

**Test.** The doctor tests for one and for two dead links expect `rm <path>` hints, and no hint holds `omp plugin unlink`.

## BUG-0009: acta set cannot change a title

**Fix.** `acta set <id> title "<text>"` is a new field in `internal/write/ops.go`. The board reads a title from the first `# ` heading of the body, so the command rewrites that line. An item that has no heading but has `title:` in its frontmatter (old scratch items) gets that field rewritten. An item with neither gets a `# <text>` heading added at the top of the body. An empty title, or one with a line break, is refused as bad input. The unknown field message lists `title` too. It commits like every other `acta set`.

**Test.** Cases: heading rewritten; frontmatter title rewritten; heading added when there is none; empty title refused; title with a line break refused; the rest of the file unchanged in every case.

## BUG-0010: the first reload throws away the scroll of a tab

**Fix.** In `moveTo` in `internal/tui/model.go`, a stored cursor of `""` counts as the same item, so the detail offset is not reset. A tab whose cursor was never moved then keeps its scroll on reload, the same as a tab whose cursor was moved.

**Test.** The repro from the bug: a fresh tab, scroll the detail box with keys and with the wheel, reload with the same board; the detail offset stays. A reload that really changes the item on show still resets it to 0.

## Out of scope

- Turning off auto maintenance for acta's own commits in user repos.
- An `acta set` field for anything other than the title.
