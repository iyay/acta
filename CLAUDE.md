# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

Two things that ship together:

- `plugin/`: the acta workflow plugin (skills, hooks, evals) for Claude Code and omp. See `plugin/README.md`.
- `cmd/acta` + `internal/`: the `acta` Go CLI the plugin calls. It reads and writes the planning files in `.acta/` (specs, plans, bugs, debt, scratch) and commits each change. With no arguments it opens a lazygit-style TUI over them. `cmd/pmb` is an old alias that calls the same `cli.Run`.

This repo uses its own plugin, so `.acta/` here holds the real planning history of acta itself.

## Commands

```bash
scripts/test                      # go test -short ./... (skips slow screen-size sweeps)
scripts/test --full               # every test; waits for other full runs on this machine (acta run-one)
scripts/test ./internal/tui -run TestName   # one package, one test
go vet ./... && gofmt -l .        # vet and format check
scripts/eval                      # plugin behaviour evals (claude plugin eval, costs quota)
go install ./cmd/acta             # rebuild the acta on PATH; hooks and the TUI use that binary
```

Agents run `scripts/test --full` as the gate, not `-short`. The eval hooks call the `acta` on PATH, so put the branch binary on PATH before `scripts/eval`.

## Version

Until the first release the version stays on 0.1.x. Every plan that lands ends with a task that adds 1 to the patch in `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`, with no exceptions. `internal/plugincheck` fails when the three files disagree or the version is not `x.y.z`.

## Architecture

- `internal/cli`: command dispatch (`list`, `show`, `set`, `bug`, `debt`, `scratch`, `voice`, `hook`, `tick`, `id`, `doctor`, `dispatch`, `reply-back`, `run-one`). Both binaries call `cli.Run`.
- `internal/board`: loads every planning file of a repo into one `Board`. `internal/trees` adds the other git worktrees so the board shows work before it lands. `internal/config` finds where the planning files live, and also holds the user setting.
- `internal/write` changes planning files; `internal/gitc` commits them with the real git binary. Write commands auto-commit, so run them in a temp clone when only testing.
- `internal/tui` (Bubble Tea + lipgloss) draws the board and watches the planning dirs with fsnotify. `internal/theme` holds its colors.
- `internal/config/user.go`: the user setting (chat language, style, tone, repo language, build executor, subagent models), stored in `~/.acta/config.yaml` (or `PM_VOICE_FILE`).
- `internal/hook`: builds the text the plugin hooks inject. `plugin/hooks/*` are thin shell scripts that call `acta hook ...` and fall back to `default-rules.md` when `acta` is missing. `pre-tool` exits 2 to block a second brainstorm in one session.
- `internal/doctor`: install checks behind `acta doctor` (`--fix` touches the repo only).
- `internal/plugincheck`: tests that keep `plugin/` valid: skill frontmatter, size caps, required rules, eval case layout, no absolute user paths. Editing a skill can fail Go tests.
- `plugin/omp/`: the omp extension (TypeScript). `plugin/omp/FACTS.md` and `plugin/evals/FACTS.md` record measured facts about omp and the eval sandbox.

Skill text is cached per session: edits to `plugin/skills/*/SKILL.md` show up only after a restart.

<!-- acta:begin -->
## acta
This repo uses the acta plugin. Before each workflow step, load the matching acta skill and follow it.
Specs, plans, bugs, debt, scratch items and the wiki live in `.acta/`.
Project knowledge (gotchas, runbooks, decisions with their why) goes to `.acta/wiki/`, never to agent memory. Write a page only when a fresh agent would lose time or repeat a mistake without it. When a fact changes, rewrite its page.
Before changing a file, run `acta wiki match <file>` and read each page it names.
Work in flight goes to `acta state set <plan id> next` with the text on stdin (read it back with `acta state <plan id>`), not to agent memory. Agent memory keeps only the user's own setup.
Raw ideas go to Scratchpad with `acta scratch new`. A finished scratch item is specced, never dropped; dropped means not done or not valid.
<!-- acta:end -->
