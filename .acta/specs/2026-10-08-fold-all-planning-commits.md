---
parent: scratch/2026-10-08-fold-all-planning-commits
created: "2026-10-08 09:01:33"
id: SPC-0112
hash: wg0p3rt
---
# Every one-file planning commit folds, not only scratch add

Status: Bounded, approved by the user in chat on 2026-10-08. Builds on SPC-0110.

Why: SPC-0110 folds only a repeat `acta scratch add` with the same subject. Other planning commits still pile up. On 2026-10-08 one spec got `83ab0a5 chore(spec): assign short ids` from `acta id` and then `6c3a389 chore(spec): split the review tier eval into two cases` from the agent's own `git commit`: two commits for one file, minutes apart. The user wants every one-file planning commit to fold the same way.

Design:
- `gitc.CommitOrFold` rule 1 changes. HEAD no longer needs the same subject. It needs to be an acta planning commit, which means its subject starts with `chore(` and its one changed file is the same `path`. The other rules from SPC-0110 stay as they are: the commit has one parent and touches only `path`; HEAD is on a branch that no remote holds; no other branch or worktree holds HEAD; the author is the same; and the file really changed.
- Message: when the new subject differs from HEAD's subject, the folded commit keeps HEAD's subject as its subject. HEAD's body lines are kept, and the new subject is added as one more body line. When the subjects are the same, nothing is added. No subject is ever lost.
- Every one-file write in `internal/write` folds: `finish` calls `CommitOrFold`, so `set`, `bug`, `debt`, `state`, `priority`, `scratch new` and `scratch add` all fold. `finishFold` goes away. `acta id` and `mark` fold only when they commit exactly one file. With more files they commit as today.
- New command `acta commit <path> -m <msg>`. It refuses a path outside the planning root and a missing message. It commits through `CommitOrFold` and prints the path, the same way other write commands do. `--json` is not needed.
- Skill text tells agents to commit planning files with `acta commit`, never `git commit`. In shape, that is a spec edit after `acta id`. In slice, it is a plan edit. In build, it is the wave tick commit (`chore(plan): tick wave <n>`), plus any plan or spec stamp commit before landing. Code commits by implementers stay plain `git commit`.
- Tests use temp repos only:
  - `acta id`, then `acta commit` on the same spec, gives one commit with both subjects.
  - Two `acta set` calls on one bug give one commit.
  - A HEAD that is a `feat(...)` commit never folds.
  - A two-file `acta id` gives a new commit.
  - `acta commit` outside the planning root is refused.
  - The SPC-0110 guard cases still pass.
- An eval case `planning-commit-folds`, listed in `evalCases`, checks two things. The agent edits a committed spec with `acta commit`, so no Bash call matches `git commit`. And the repo ends with one commit for that spec.
- `.acta/wiki/commit-subject-form.md` gets its fold line updated to the wider rule.
- The last task adds 1 to the patch version in the three plugin files.
