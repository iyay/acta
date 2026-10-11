---
parent: specs/2026-10-11-wiki-freshness-from-git-design
depth: minimal
closes: [SCR-0060]
id: PLN-0134
created: "2026-10-11 21:09:19"
hash: kfizu27
started: "2026-10-11 21:11:46"
finished: "2026-10-11 21:18:32"
---
# Wiki freshness from git

**Goal:** `acta wiki check` judges a page stale by git ancestry instead of a `timestamp` field, `acta wiki ok` records a no-change re-check as an empty commit with a `Wiki-Checked:` trailer, tidy keeps those trailers, and the plugin text, evals and this repo's pages stop using `timestamp`.

**Spec:** `.acta/specs/2026-10-11-wiki-freshness-from-git-design.md`

**Tests:** fast `scripts/test ./internal/gitc ./internal/wiki ./internal/cli ./internal/tidy ./internal/plugincheck`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Trailer form, verbatim: `Wiki-Checked: <page id>`, one line per page. The page id is `wiki.Page.ID` (path under the wiki folder, slashes, no `.md`).
- Re-check commit subject, verbatim: `chore(wiki): re-check <id>[, <id>...]`, ids in the order given.
- Stale message, verbatim: `stale: <short sha> "<subject>" changed its paths after the page was last checked; fix the page, or run acta wiki ok <id> when nothing is wrong`.
- git runs through the git binary with `exec`, the way `internal/gitc` does. No go-git, no new dependency. Read trailers with git's own `%(trailers:key=Wiki-Checked,valueonly)` format, not a hand parser.
- Freshness uses ancestry only: no commit dates anywhere in the stale rule.
- No version bump.
- Tests build git repos under `t.TempDir()` and never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01, Task 02, Task 03
- Wave 2: Task 04, Task 05
- Wave 3: Task 06

### Task 01: git queries for check commits and unseen changes

**Files:**
- Modify: `internal/gitc/gitc.go`
- Test: `internal/gitc/gitc_test.go`

**verify:** `Unseen` returns a commit exactly when some commit reachable from HEAD changed a file under the paths and is not an ancestor of (or equal to) any check commit, and `CheckCommits` returns every commit reachable from HEAD that changed the page file or carries a `Wiki-Checked:` trailer naming exactly that id. List every case checked: no check commits, the check commit is the path commit itself, a check on one side of a merge and the path change on the other, checks on both sides, a trailer naming a different id, a trailer id that is a prefix of the asked id, a trailer-shaped line in the body above the trailer block, a path commit dated earlier than the check but not its ancestor.

- [x] **Failing test:** temp repos with real commits (merges built with `git merge --no-ff`) call `CheckCommits(repo, pageFile, id string) ([]string, error)`, `Unseen(repo string, checks, paths []string) (sha, subject string, err error)` and `CommitEmpty(repo, msg string) (string, error)`; fails because the functions do not exist.
- [x] **Change:** `CheckCommits` unions `git log --format=%H -- <pageFile>` with the shas from `git log --grep=^Wiki-Checked: --format=%H%x00%(trailers:key=Wiki-Checked,valueonly,separator=%x1f)` whose values equal `id`. `Unseen` runs `git --literal-pathspecs log -1 --format=%h%x00%s HEAD --not <checks...> -- <paths...>`; empty output gives an empty sha. `CommitEmpty` runs `git commit --only --allow-empty -m <msg>` (staged changes stay staged and out of the commit) and returns the short sha of HEAD. Leave `LastChange` in place; Task 04 removes it.
- [x] **Commit:** `feat(gitc): find a page's check commits and unseen path changes`

### Task 02: tidy keeps Wiki-Checked trailers

**Files:**
- Modify: `internal/tidy/tidy.go`
- Test: `internal/tidy/tidy_test.go`

**verify:** After tidy, every `Wiki-Checked:` value carried by any commit of the old branch appears on exactly one commit of the new chain: the commit its old commit folded into, once, even when two folded commits name the same page. Messages of commits that fold nothing come out byte for byte as before. List every case checked.

- [x] **Failing test:** a branch with a code commit, then a `chore(wiki): re-check a` empty commit with `Wiki-Checked: a`, then a `chore(wiki): re-check a, b` commit; after `Run`, the code commit's new message ends with `Wiki-Checked: a` and `Wiki-Checked: b`, each once; a branch with no such trailers keeps its messages unchanged. Fails because `commitNode` keeps only the source message.
- [x] **Change:** in `commitNode`, read `%(trailers:key=Wiki-Checked,valueonly)` of every sha in `n.olds`, drop values the source message already carries and repeats, and append the rest as `Wiki-Checked: <id>` lines to the source message's trailer block (a new block after a blank line when it has none).
- [x] **Commit:** `feat(tidy): keep Wiki-Checked trailers of folded commits`

### Task 03: plugin text and evals drop timestamp

**Files:**
- Modify: `plugin/skills/build/wiki.md`, `plugin/skills/build/SKILL.md`, `plugin/skills/land/SKILL.md`, `plugin/evals/wiki-close/scaffold.sh`, `plugin/evals/wiki-close/prompt.md`, `plugin/evals/wiki-hint/scaffold.sh`
- Delete: `plugin/evals/wiki-close/graders/timestamp-bumped.md`
- Create: `plugin/evals/wiki-close/graders/page-fixed.md`
- Test: `internal/plugincheck/skill_build_test.go`

**verify:** No plugin text or eval file tells an agent to write, bump or read a page `timestamp`, and every place that said to bump it now says to fix the page or run `acta wiki ok <id>`. List every file under `plugin/` that mentions a page timestamp before and after (grep `timestamp` across `plugin/`), and every rule line checked.

- [x] **Failing test:** in `skill_build_test.go`, replace the four timestamp expectations (lines near 243, 283, 294, 295, 304) with: wiki.md has no `timestamp`, wiki.md and build SKILL.md contain `acta wiki ok`, wiki.md contains a rule that a whole top-level folder in `paths` makes the page stale on nearly every branch, and land SKILL.md has no `timestamp` bump; fails on the current text.
- [x] **Change:** wiki.md: drop `timestamp` from the example and the field list; check step says a page is stale when a commit changed a file under its `paths` after the page was last checked; fix step says fix every wrong line, or run `acta wiki ok <id>` when the page is still right; add the narrow `paths` rule. build SKILL.md wiki step and land SKILL.md wiki gate: replace the bump with `acta wiki ok`. Both scaffolds drop the `timestamp:` line and its comment, and wiki-close commits the page before the code change so the page starts stale. New grader `page-fixed.md`: `type: regex`, pattern `note`, target file `.acta/wiki/parse-output.md`. Prompt description: "re-checks the wiki page that covers a file the branch changed and fixes it before review". Keep every skill under its plugincheck size cap.
- [x] **Commit:** `feat(plugin): wiki re-check uses acta wiki ok, not a timestamp`

### Task 04: check uses ancestry, timestamp dropped

**Files:**
- Modify: `internal/wiki/check.go`, `internal/wiki/wiki.go`, `internal/gitc/gitc.go` (remove `LastChange` once nothing calls it), `internal/gitc/gitc_test.go` (its `LastChange` tests)
- Test: `internal/wiki/check_test.go`, `internal/wiki/wiki_test.go`

**verify:** No page is ever called stale while every path change reachable from HEAD is seen by one of its check commits, and no page with an unseen path change passes, with or without a range; the `timestamp` field changes no outcome. List every path through `checkPage` and what each returns: no usable paths, no check commits, a seen change, an unseen change, an uncommitted or untracked page file, a git error, a page with a `timestamp` that does not parse, a page with none.

- [x] **Failing test:** rewrite the date table in `check_test.go` as ancestry cases (path change after the last page edit is stale with the verbatim message; after `CommitEmpty` with a `Wiki-Checked:` trailer it is not; a later path change makes it stale again; the two-branch merge cases; an uncommitted edit to the page file is not stale); "timestamp missing" is no longer a problem; in `wiki_test.go`, a page whose `timestamp` does not parse loads with no error. Fails on the date rule.
- [x] **Change:** drop `Timestamp` from `Page` and `front` and its parse; drop the missing-field message; in `checkPage`, when `gitc.IsDirty` says the page file has changes, skip the stale test; else `CheckCommits` then `Unseen`, and a non-empty sha gives the verbatim stale message. A git error gives `cannot read the git history: <err>`. Remove `LastChange` and its tests.
- [x] **Commit:** `feat(wiki): judge stale pages by git ancestry, not timestamp`

### Task 05: acta wiki ok

**Files:**
- Modify: `internal/cli/wiki.go`, `internal/cli/cli.go` (the unknown command usage list)
- Test: `internal/cli/wiki_test.go` (create if missing; else the file that tests `cmdWiki`)

**verify:** `acta wiki ok` makes a commit only when every id names a loaded page, and that commit changes no file and carries one `Wiki-Checked:` trailer per id. List every path: no ids, one unknown id among good ones, duplicate ids, a staged file present, not a git repo, a commit that git refuses. Each failure exits non-zero, prints why, and leaves HEAD where it was.

- [x] **Failing test:** in a temp repo with pages `a` and `sub/b`, `ok a sub/b` exits 0, prints the short sha, HEAD's tree equals its parent's, the message is `chore(wiki): re-check a, sub/b` plus both trailers; `ok a nope` exits non-zero with HEAD unchanged; a staged file stays staged and outside the commit; two branches that each ran `ok a` merge with no conflict. Fails because `ok` is not a subcommand.
- [x] **Change:** add `ok` to `cmdWiki`, `wikiFits` (one id or more) and `wikiUsage`; drop duplicate ids keeping first order; refuse unknown ids before any git call; build the message and call `gitc.CommitEmpty`. Update the `cmdWiki` comment, which says the subcommands only read.
- [x] **Commit:** `feat(cli): acta wiki ok records a no-change re-check`

### Task 06: this repo's pages drop timestamp

**Files:**
- Modify: every page under `.acta/wiki/` that has a `timestamp:` line

**verify:** No page under `.acta/wiki/` keeps a `timestamp:` line, no other line of any page changed, and `acta wiki check` with the branch binary prints nothing new about field rules. Show `git diff --stat` and confirm every changed line is a removed `timestamp:` line.

- [x] **Failing test:** `grep -rl '^timestamp:' .acta/wiki` lists pages (red).
- [x] **Change:** delete the `timestamp:` line from each page with `sed`; nothing else. Run `go run ./cmd/acta wiki check` and show the output.
- [x] **Commit:** `chore(wiki): drop timestamp from pages`

## State

### Findings

gitc.CheckCommits and gitc.Unseen run git log with --full-history --no-merges: without full history git drops same-content side commits; without no-merges a merge commit counts as a path change forever. A conflict fix made only inside a merge commit is not seen as a path change.
