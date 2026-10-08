---
parent: specs/2026-10-08-fold-all-planning-commits
closes: [SCR-0056]
depth: minimal
id: PLN-0121
created: "2026-10-08 09:32:47"
hash: h7t7n68
started: "2026-10-08 09:34:23"
finished: "2026-10-08 09:46:09"
---
# Every one-file planning commit folds

**Goal:** Any acta commit of one planning file folds into HEAD when HEAD is an unshared acta planning commit of that same file, keeping every subject, and agents commit planning files with a new `acta commit` instead of `git commit`.

**Spec:** .acta/specs/2026-10-08-fold-all-planning-commits.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Code commits (`feat`, `fix`, `test`, `polish` and any subject not starting with `chore(`) never fold, and are never folded into.
- A failed fold never loses text: any failed check or failed amend ends in the plain commit.
- Every SPC-0110 guard stays: one parent and only `path` in HEAD, HEAD on a branch no remote holds, no other branch or worktree holds HEAD, same author, the file really changed.
- Tests use temp repos only. acta write commands auto-commit, so never run them in this repo.
- Skill and eval text is plain English that fits every repo: no repo paths, no user names.
- Exact phrase later tasks rely on, written verbatim in the skills: `acta commit <path> -m "<message>"`.

## Waves

- Wave 1: Task 01, Task 05
- Wave 2: Task 02, Task 03
- Wave 3: Task 04

### Task 01: Fold any one-file planning commit in gitc

**Files:**
- Modify: `internal/gitc/gitc.go`
- Test: `internal/gitc/gitc_test.go`

**verify:** `CommitOrFold` folds exactly when HEAD's subject starts with `chore(`, HEAD touches only `path`, and every SPC-0110 guard holds; the folded commit keeps HEAD's subject and body and adds the new subject as one body line when it differs, so no subject is ever lost; `CommitPaths` with exactly one path behaves like `CommitOrFold`, with more paths like today. List each path checked and its result.

- [x] Failing test: cases in `TestCommitOrFold`: a different `chore(spec): ...` subject on the same file folds and `git log -1 --format=%B` holds both subjects; a third different subject adds a third line; a HEAD subject `feat: x` never folds; the old SPC-0110 cases still pass with the rule-1 case now meaning "HEAD is not a `chore(` commit". `TestCommitPaths`: one path folds, two paths make a new commit. Run `scripts/test ./internal/gitc/ -run 'TestCommitOrFold|TestCommitPaths'` and see it fail.
- [x] Code: in `canFold` replace the `f[0] != msg` check with "HEAD subject starts with `chore(`" (read `%s`, `%an`, `%P` as today); in `CommitOrFold` read HEAD's full message (`%B`) and amend with `-m <HEAD subject> -m <rest of HEAD body plus the new subject when it differs from every line already there>`; `CommitPaths` sends a one-path call to `CommitOrFold` with `wasDirty` false. Plain-English comments saying why.
- [x] Run `scripts/test ./internal/gitc/` passes, vet and gofmt clean, commit `feat(gitc): fold any one-file planning commit and keep every subject`.

### Task 02: Every one-file write folds

**Files:**
- Modify: `internal/write/ops.go`, `internal/write/scratch.go`
- Test: `internal/write/scratch_test.go`, `internal/write/ops_test.go` (or the test file that already covers `SetValue`)

**verify:** Every one-file write command (`set`, `bug`, `debt`, `state`, `priority`, `scratch new`, `scratch add`) run twice on one file in a row leaves one commit holding both changes and both subjects, and a write on another file in between makes a new commit. List each command checked and its commit count.

- [x] Failing test: two `SetValue` calls on one bug make one commit with both subjects in `%B`; `NewScratch` then `AppendScratch` on the same item make one commit; a write to another file between them makes a new commit. Run `scripts/test ./internal/write/ -run 'Fold|Scratch|SetValue'` and see it fail.
- [x] Code: `finish` calls `gitc.CommitOrFold`; delete `finishFold` and point `scratch.go` back at `finish`. Fix older tests that counted one commit per write on the same file.
- [x] Run `scripts/test ./internal/write/` passes, vet and gofmt clean, commit `feat(write): every one-file planning write folds into its own last commit`.

### Task 03: acta commit command

**Files:**
- Create: `internal/cli/commit.go`, `internal/cli/commit_test.go`
- Modify: `internal/cli/cli.go` (dispatch and usage text)

**verify:** `acta commit <path> -m <msg>` commits only a file under the planning root, through `gitc.CommitOrFold`, and refuses with a clear error and no commit when the path is outside the planning root, missing, or the message is empty. List each input checked and its exit code and commit count.

- [x] Failing test: in a temp repo, `acta id`-style first commit then `acta commit .acta/specs/x.md -m "chore(spec): fix wording"` after an edit gives one commit with both subjects; a path outside the planning root, a missing path, and an empty `-m` each exit non-zero with no new commit. Run `scripts/test ./internal/cli/ -run TestCommit` and see it fail.
- [x] Code: `commit` case in `cli.go` calling a small `runCommit` in `commit.go`: resolve the planning root the same way other write commands do (`config`), check the path is inside it, call `gitc.CommitOrFold(repoRoot, path, msg, false)`, print the path like other write commands and use their skipped exit code when nothing was committed; add the command to the usage text.
- [x] Run `scripts/test ./internal/cli/` passes, vet and gofmt clean, commit `feat(cli): acta commit commits one planning file and folds it`.

### Task 04: Skills send planning commits through acta commit

**Files:**
- Modify: `plugin/skills/shape/SKILL.md`, `plugin/skills/slice/SKILL.md`, `plugin/skills/build/SKILL.md`, `plugin/skills/land/SKILL.md`, `internal/plugincheck/` skill tests and `budget_test.go` caps as needed, `internal/plugincheck/evals_test.go`
- Create: `plugin/evals/planning-commit-folds/` (`case.yaml`, `prompt.md`, `scaffold.sh`, `graders/`)

**verify:** No skill tells an agent to commit a spec or plan file with `git commit`; every place that commits a spec, a plan or plan ticks names `acta commit <path> -m "<message>"`; code commits by implementers still use `git commit`. The eval fails a run that commits a spec edit with `git commit` or leaves two commits for that spec. List each skill line changed and each grader with the wrong run it rejects.

- [x] Failing test: Must strings for `acta commit <path> -m "<message>"` in the shape, slice and build skill tests; `{"planning-commit-folds", "skills/shape/SKILL.md", "acta commit <path> -m \"<message>\""}` in `evalCases`. Run `scripts/test ./internal/plugincheck/` and see it fail.
- [x] Code: shape (spec edit after `acta id`), slice (plan commit and plan edits), build (wave tick commit, plan or spec stamp commits) and land (plan tick commit before the merge) say to commit with `acta commit <path> -m "<message>"`; keep slice's implementer example `git commit -m "feat: ..."` as it is. Eval case copying `plugin/evals/repo-override-ask`'s safe scaffold shape: a repo with one spec committed as `chore(spec): assign short ids`, prompt asks to fix one word in that spec and commit it, then stop. Graders per `plugin/evals/FACTS.md` (`input_match` is a case-sensitive plain substring): `tool_used` Bash `input_match: "git commit"` max 0; `tool_used` Bash `input_match: "acta commit"` min 1. Raise a cap only to the new file size.
- [x] Run `scripts/test ./internal/plugincheck/ ./plugin/` passes, commit `feat(plugin): skills commit planning files with acta commit`; the orchestrator runs the eval case 3 times.

### Task 05: Wiki line and version bump

**Files:**
- Modify: `.acta/wiki/commit-subject-form.md`, `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The wiki fold line states the wider rule (any one-file `chore(` planning commit of the same file, subjects kept as body lines, `acta commit` for agent edits) and nothing still says only `scratch add` folds; all three version files read `0.1.43`. List each file and what it says.

- [x] Failing test: none new; `internal/plugincheck` checks the three version files agree.
- [x] Code: rewrite the fold line in `commit-subject-form.md` and bump its `timestamp`; set `"version": "0.1.43"` in the three files.
- [x] Run `scripts/test ./internal/plugincheck/` passes, commit `chore(plugin): bump version to 0.1.43`.

## Polish

### Task 06: Review polish

**verify:** every NOTE below is applied, and nothing else changes; the planning-commit-folds eval passes 3 runs of 3.

- [x] Session note (`internal/hook`, next to the other always-on rules, and the hook fallback text in `plugin/hooks/` if it carries those rules): one plain sentence telling the agent to commit spec, plan and other planning files with `acta commit <path> -m "<message>"`, never `git commit`; a hook test pins it.
- [x] `plugin/evals/planning-commit-folds/scaffold.sh`: copy the real git into `bin/git` the way `plugin/evals/state-resume/scaffold.sh` does, so a commit can run in the eval sandbox.
- [x] `internal/gitc/gitc.go` fold message: keep HEAD's body exactly as written (blank lines and indents stay); add the new subject as one line before HEAD's trailer block, so trailers such as `Co-Authored-By` stay trailers; a gitc test pins both.
- [x] `internal/cli/commit_test.go`: a symlink inside the planning root that points outside it is refused with no commit.
- [x] Commit: `polish: review notes for PLN-0121`

## Review notes

- Round 1 CLEAN on both axes. `[fix]` NOTEs went into the Task 06 polish: trailers kept through a fold, symlink escape test, and, after the planning-commit-folds eval failed because no acta skill loads for a plain "commit this spec" ask, a session-note rule that sends planning commits through `acta commit`. Polish review CLEAN on both axes.
- The eval prompt gained "File edit tools are off here, so make the change with a shell command.", since `scripts/eval` grants only Bash and Read; it names neither commit command.
- `acta doctor --fix` commits one path through `CommitPaths`, so it can fold into the user's own unpushed `chore(` commit of that file; `chore: doctor fix` has no `(`, so nothing folds into it.
- `acta commit` with a `feat(...)` message on a planning file whose HEAD is a `chore(` commit folds, and the feat subject becomes a body line; that follows the spec.
- `acta commit --root . <file>` treats the whole repo as the planning root; that only happens when the user picks that root.
- The cli test for "acta id then acta commit" makes its first commit with git; a hand run with the built binary gave one commit holding both subjects.
- A multi-line `-m` is matched one line at a time, so a repeat fold can add it again; `git commit -m` cleanup drops trailing spaces in HEAD's body.
- The symlink test checks refusal and commit count, not the error wording.
- Byte caps for build and slice keep slack from before this plan.
