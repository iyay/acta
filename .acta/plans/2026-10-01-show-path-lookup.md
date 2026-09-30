---
parent: specs/2026-10-01-show-path-lookup-design
id: PLN-0063
created: "2026-10-01 05:08:45"
hash: cmnpxkr
started: "2026-10-01 05:11:41"
finished: "2026-10-01 05:28:45"
---
# acta show --path Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

Approval: the user said "skip plan, langsung build" on 2026-10-01, so this plan runs without a separate plan review.

**Goal:** `acta show <id> --path` prints only the item's file path, and for a new-format id or a full hash it finds the file without loading the board or calling git.

**Architecture:** In `cmdShow` (`internal/cli/cli.go`), when `--path` is set, first scan the frontmatter of the `.md` files in the five planning dirs (`cfg.Root` joined with each `cfg.Dirs` field) for a line `id: <X>` or `hash: <X>`, where `X` is the input with any `.NN` sub-item part dropped. A hit prints the path relative to `cfg.RepoRoot`, slash form, the same as the `path:` line. A miss falls back to `loadAllTrees` + `b.Get` + `toJSON(...).Path`. Still no item: `unknown id X`, exit bad input.

**Tech Stack:** Go stdlib (`os`, `bufio`, `path/filepath`, `strings`).

**Spec:** `.acta/specs/2026-10-01-show-path-lookup-design.md`

**Tests:** fast `scripts/test ./internal/cli`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then one line, then the minimum; never cut validation.
- Plain `acta show` output does not change. When both `--json` and `--path` are given, `--path` wins.
- The fast scan reads only the frontmatter block (between the first two `---` lines) of each file, and matches whole values only (`id: SCR-0028`, `hash: iivf3gy`, with optional quotes). A hash prefix is not matched fast; it falls back.
- The fast path runs no git and does not call `trees.Load` or `board.Load`.
- Comments are plain English a 10-year-old can read. They say why, not what.

## File map

- Modify: `internal/cli/cli.go` (`cmdShow`, plus one small helper next to it)
- Modify: `internal/cli/cli_test.go`
- Modify: `plugin/skills/build/SKILL.md`, `plugin/skills/brainstorm/SKILL.md`

## Waves

- Wave 1: Task 1, Task 2 (different files).

---

### Task 1: acta show --path

**Files:**
- Modify: `internal/cli/cli.go`
- Test: `internal/cli/cli_test.go`

**verify:** In a temp repo, `acta show <x> --path` prints exactly one line, the path of the right file, for: a new id, a full hash, a sub-item id (`PLN-0001.01` gives the plan's path), an old id that only the board alias knows (through the fallback), a spec under a root moved with `ACTA_ROOT` or `--root`. An unknown id prints `unknown id X` and exits with bad input. A new id and a hash still resolve when PATH holds no git binary. Plain `acta show <id>` output is byte for byte what it was. List the cases checked.

- [x] **Step 1: Write the failing tests** in `internal/cli/cli_test.go`, following the setup the existing `show` tests use. One table test over the input forms above, one test that sets PATH to an empty temp dir for the fast cases, one unknown-id test.
- [x] **Step 2: Run the tests to see them fail.** Run: `scripts/test ./internal/cli -run Show`. Expected: FAIL, `--path` is an unknown flag.
- [x] **Step 3: Write the minimal code.** Add the `--path` flag and the frontmatter scan helper; wire the fallback; update the usage line to `usage: acta show <id> [--json] [--path]`.
- [x] **Step 4: Run the tests to see them pass.** Run: `scripts/test ./internal/cli`, then `go vet ./... && gofmt -l .` prints nothing. Then `go build -o /tmp/acta-path ./cmd/acta && time /tmp/acta-path show SCR-0028 --path` from the worktree and report the time (target under 0.05 s).
- [x] **Step 5: Commit** `internal/cli/cli.go internal/cli/cli_test.go` with message `acta show --path finds an item's file fast (PLN-0063)`.

### Task 2: Skills point at acta show --path

**Files:**
- Modify: `plugin/skills/build/SKILL.md`, `plugin/skills/brainstorm/SKILL.md`

**verify:** Each of the two skills holds one line saying that to find an item's file by id or hash, run `acta show <id|hash> --path`. `scripts/test ./internal/plugincheck` passes. List the two lines added.

- [x] **Step 1: Write the failing test.** Add a case to the required-rules check in `internal/plugincheck` (follow how it already requires phrases per skill) that both skills contain `acta show <id|hash> --path`.
- [x] **Step 2: Run it to see it fail.** Run: `scripts/test ./internal/plugincheck`. Expected: FAIL naming both skills.
- [x] **Step 3: Add the line** to each skill: in build, near the `## Setup` section; in brainstorm, under "Where files go". One sentence each.
- [x] **Step 4: Run it to see it pass.** Run: `scripts/test ./internal/plugincheck`.
- [x] **Step 5: Commit** the two skill files and the plugincheck test with message `Skills point agents at acta show --path (SPC-0054)`.
