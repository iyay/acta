---
parent: scratch/2026-10-19-land-agent-forces-cleanup
id: SPC-0186
created: "2026-10-31 20:19:51"
hash: sem7ds7
---
# Block destructive commands and keep test output

## Problem

In eval runs agents ran `rm -rf src/__pycache__`, `git worktree remove --force`
and `git commit --amend` on a merge commit, though no skill asks for them. A
spike on 2026-10-31 (state-resume, one run, trace kept) found the cause: after
the task tests ran, `git status` listed `?? src/__pycache__/`. The build wanted
a clean tree before it committed the plan file, so the agent deleted the
folder. `git worktree remove` refuses a tree with untracked files, which is the
same push toward `--force` at land. The skills say what a clean tree is, but
not what to do with files a test made.

## Scope

Bounded. It changes the existing pre-tool hook and the build and land skill
text. It applies in every repo that uses acta.

## Behavior

### The pre-tool hook blocks destructive commands

A new `hook.DestructiveBlock(dir, command string) (bool, string)` sits next to
`hook.GoTestBlock` and runs from the same pre-tool path
(`internal/cli/hook.go`). A blocked Bash call exits 2 with a short reason, as
the go test block does. It blocks a piece of the command (split on `;`, `&&`,
`||`, `|`, like the go test block) when it is:

- `rm` with both a recursive flag (`-r`, `-R`, `--recursive`) and a force
  flag (`-f`, `--force`), in any form: `-rf`, `-fr`, `-Rf`, `-r -f`,
  `--recursive --force`.
- a `git` command with `--force` or `--force-with-lease` anywhere in it.
- `git worktree remove` or `git push` with `-f`.
- `git reset --hard`.
- `git commit --amend` when `HEAD` in `dir` has two or more parents (a merge
  commit). Amending one's own plain commit stays allowed.

It does not block `git branch -D`: tidy step 7 deletes the old branch that way
on purpose.

On every doubt it lets the command through, like the go test block: a command
it cannot read as words, a `dir` with no repo above it for the amend check, or
a git it cannot run. A hook that stops the wrong work costs more than one that
misses.

The reason text names the command, says acta never runs it on its own, and
says the user can run it with `!` in the prompt if they mean it.

### The skills say what to do with test output

- Build (task loop step 3, where `git status` is read before the commit) and
  land (the clean-worktree precondition): untracked files that a test or a tool
  made, such as cache folders, are not uncommitted work. Never delete them.
  Show them to the user and offer one `.gitignore` line for them in the task
  commit; add it only after the user's yes.
- Land step 7 and tidy step 7: when `git worktree remove` refuses, stop, show
  `git status --porcelain --ignored` of that worktree, and ask the user. Never
  `--force`, never delete the files first.
- Land: never amend a commit at land time.

## Testing

- `DestructiveBlock`: one table test. Each blocked form above is a row that
  must block; each allowed form is a row that must pass: `git branch -D x`,
  `rm -r dir`, `rm -f file`, `git commit --amend` on a plain `HEAD`, a quoted
  string that only mentions `rm -rf`, and a folder with no repo. The amend rows
  use a temp repo with a real merge commit.
- CLI wiring: one blocked call exits 2 with the reason on stderr; one allowed
  call exits 0.
- Plugincheck guards the new build, land and tidy sentences.
- The state-resume eval gets a regex grader that fails when any Bash call holds
  `rm -rf` or `rm -fr`.
