---
parent: specs/2026-10-31-block-destructive-commands-design
depth: minimal
id: PLN-0206
created: "2026-10-31 20:24:08"
hash: uageial
started: "2026-10-31 20:25:30"
---
# Destructive command block hook implementation plan

**Goal:** Add `hook.DestructiveBlock`, run it from the pre-tool hook so a destructive Bash call exits 2 with a reason, and change core rule 5 to match.
**Spec:** `.acta/specs/2026-10-31-block-destructive-commands-design.md` (The pre-tool hook blocks destructive commands, Core rule 5)
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only their owning packages.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
On every doubt the hook lets the command through: a command it cannot read as words, a working folder with no repo above it, or a git it cannot run for the amend check.
`git branch -D`, `rm -r` alone, `rm -f` alone and `git commit --amend` on a plain `HEAD` are never blocked.
The reason text names the blocked command, says acta never runs it, and asks the agent to tell the user what it wanted to run and why; it names no harness feature.
Reuse the command reading of `internal/hook/gotest.go` (pieces, words, `NAME=value`), extended where the spec needs it; do not write a second parser.

## Waves

Wave 1: Task 1. Wave 2: Task 2, Task 3.

### Task 1: DestructiveBlock reads and judges a command

**Files:** Create `internal/hook/destructive.go`, `internal/hook/destructive_test.go`; modify `internal/hook/gotest.go` only where its piece and word readers must learn the newline separator and quoting the spec names.
**Seam:** `hook.DestructiveBlock(dir, command string) (bool, string)`.
**verify:** No blocked form in the spec passes and no allowed form is blocked. Blocked rows: every `rm` recursive plus force spelling; `git` with `--force` or `--force-with-lease`; `git worktree remove`, `git push`, `git clean` with `-f` alone or joined (`-fdx`); `git reset --hard`; `git commit --amend` on a merge `HEAD`, found through `-C <path>`, through an earlier `cd <path>` in the same command, and through `dir`; `git -C <path> worktree remove -f x`; `find . -exec rm -rf {} +`; `sudo rm -rf x`; a newline-split command. Allowed rows: `git branch -D x`, `rm -r dir`, `rm -f file`, amend on a plain `HEAD`, `git commit -m "no --force"`, `echo "rm -rf x"`, `rm -rf x` with no repo above `dir`, and an unreadable command. GoTestBlock's own tests stay green. The amend rows use a temp repo with a real merge commit.
- [x] Red: the table test fails because `DestructiveBlock` does not exist.
- [x] Change: Write `DestructiveBlock` over the shared readers, then run `scripts/test ./internal/hook/` to green.
- [x] Commit: `feat(hook): block destructive commands`.

### Task 2: pre-tool runs the block

**Files:** Modify `internal/cli/hook.go`, `internal/cli/hook_test.go`.
**Seam:** `cli.Run` with `hook pre-tool` and a Bash tool event on stdin.
**verify:** Wiring only: one blocked Bash call exits 2 with the reason on stderr; one allowed Bash call exits 0; a non-Bash tool event is never checked; the go test block still works next to it.
- [ ] Red: the CLI test fails because a `rm -rf x` Bash event in a temp repo exits 0.
- [ ] Change: Call `hook.DestructiveBlock` next to `hook.GoTestBlock` in the pre-tool path, then run `scripts/test ./internal/cli/ -run TestHook` to green.
- [ ] Commit: `feat(hook): pre-tool blocks destructive commands`.

### Task 3: core rule 5 says the user runs it

**Files:** Modify `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`.
**Seam:** The session-start rules text from `internal/hook` and its copy in `plugin/hooks/default-rules.md`.
**verify:** Both copies say: never run a destructive command; the hook blocks the listed forms; say in full sentences what you wanted to run and why, and let the user run it themselves. The tidy `git branch -D` exemption is unchanged, and the check that `default-rules.md` matches the generated rules still passes.
- [ ] Red: a hook test that wants the new rule 5 sentence fails on the old text.
- [ ] Change: Rewrite rule 5 in `coreRules`, regenerate `default-rules.md` the way `hook_test.go` does (`-update`), then run `scripts/test ./internal/hook/` to green.
- [ ] Commit: `feat(hook): core rule 5 lets the user run destructive commands`.

## Build brief

Land order: PLN-0206 and PLN-0207 land on main in either order; they share no file.

## State

### Next

Task 2: call hook.DestructiveBlock(dir, command string) (bool, string) next to hook.GoTestBlock in the pre-tool path; on true exit 2 with the second return on stderr. It already returns false outside a git repo.
Task 3: reason text shape is "acta never runs destructive commands, so it blocked: <command words joined by spaces>. Tell the user in full sentences what you wanted to run and why. They can run it themselves outside the agent if they mean it." Rule 5 should match that wording.
