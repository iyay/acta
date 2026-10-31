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

It changes the pre-tool hook, core rule 5 (both copies: `internal/hook/hook.go`
and `plugin/hooks/default-rules.md`), the build, land and tidy skill text, and
the wiki page `omp-report-file`. The hook runs in every harness acta hooks
into (Claude Code, Codex, omp), but it blocks only inside a git repo: a Bash
call whose working folder has no repo above it goes through.

## Behavior

### The pre-tool hook blocks destructive commands

A new `hook.DestructiveBlock(dir, command string) (bool, string)` sits next to
`hook.GoTestBlock` and runs from the same pre-tool path
(`internal/cli/hook.go`). A blocked Bash call exits 2 with a short reason, as
the go test block does.

How it reads a command:

- It splits the command into pieces on `;`, `&&`, `||`, `|`, `(`, `)` and a
  newline.
- In each piece it skips `NAME=value` words and the prefixes `sudo`, `env`,
  `command`, `nohup`, `xargs` (with their flags) to find the command word.
- For `find`, the words after `-exec` or `-execdir` are read as a command too.
- Quoted words are one word, so `git commit -m "no --force here"` and
  `echo "rm -rf x"` hold no flag.
- For `git`, the global options `-C <path>`, `-c <k=v>`, `--git-dir=<path>`
  and `--work-tree=<path>` are skipped before the subcommand is matched.

A piece is blocked when its command is:

- `rm` with both a recursive flag (`-r`, `-R`, `--recursive`) and a force
  flag (`-f`, `--force`), in any form: `-rf`, `-fr`, `-Rf`, `-r -f`,
  `--recursive --force`.
- `git` with `--force` or `--force-with-lease` as a word.
- `git worktree remove`, `git push` or `git clean` with `-f` (alone or joined,
  like `-fd`, `-fdx`).
- `git reset --hard`.
- `git commit --amend` when `HEAD` of the repo the piece works in has two or
  more parents (a merge commit). That repo is the `-C <path>` of the piece,
  else the folder of the last `cd <path>` earlier in the same command, else
  `dir`. Amending one's own plain commit stays allowed.

It does not block `git branch -D`: tidy step 7 deletes the old branch that way
on purpose. It does not block `rm -r` or `rm -f` alone.

On every doubt it lets the command through, like the go test block: a command
it cannot read as words, a working folder with no repo above it, or a git it
cannot run for the amend check. A hook that stops the wrong work costs more
than one that misses.

The reason text names the blocked command, says acta never runs it, and asks
the agent to tell the user what it wanted to run and why, so the user can run
it themselves outside the agent if they mean it. It names no harness feature.

### Core rule 5

Both copies change the destructive sentence to: never run a destructive
command; the hook blocks the listed forms; say in full sentences what you
wanted to run and why, and let the user run it themselves. The tidy `git
branch -D` exemption stays as it is.

### The skills say what to do with test output

- Build (task loop step 3, where `git status` is read before the commit) and
  land (the clean-worktree precondition): untracked files that a test or a
  tool made, such as cache folders, are not uncommitted work. Never delete
  them. The orchestrator, not an implementer subagent, shows them to the user
  and offers one `.gitignore` line for them in the task commit; it adds the
  line only after the user's yes.
- Land, when they are still there at land time: show them and give the user
  three choices. Add a `.gitignore` line (one commit on the branch, then the
  gates run again only for that file, as land step 6 does for a non-planning
  path); remove them themselves; or stop the land. Never pick for the user.
- Land step 7 and tidy step 7: when `git worktree remove` refuses, stop, show
  `git status --porcelain --ignored` of that worktree, and ask with the same
  three choices. Never `--force`, never delete the files first.
- Land: never amend a commit at land time.
- Wiki `omp-report-file`: moving acta's own omp report out of the worktree is
  not a delete and stays allowed; the page says the empty folder goes with
  `rmdir`, and anything else untracked follows the rule above.

## Testing

- `DestructiveBlock`: one table test. Each blocked form above is a row that
  must block, including `git -C <path> worktree remove -f x`,
  `cd <main> && git commit --amend` with a merge `HEAD` there,
  `find . -name x -exec rm -rf {} +`, `sudo rm -rf x` and `git clean -fdx`.
  Each allowed form is a row that must pass: `git branch -D x`, `rm -r dir`,
  `rm -f file`, `git commit --amend` on a plain `HEAD`,
  `git commit -m "no --force"`, `echo "rm -rf x"`, and `rm -rf x` in a folder
  with no repo. The amend rows use a temp repo with a real merge commit.
- CLI wiring: one blocked call exits 2 with the reason on stderr; one allowed
  call exits 0.
- Core rule 5: the existing check that `default-rules.md` matches the
  generated rules still passes with the new text.
- Plugincheck guards the new build, land and tidy sentences.
- The state-resume eval gets a regex grader that fails when the test-made
  `src/__pycache__` folder is gone at the end of the run, whatever command
  removed it.
