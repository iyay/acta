---
id: SPC-0008
hash: wx526y6
---
# Rename pm and pmb to acta

Status: approved by the user on 2026-09-26. Built after the TUI panes work lands and the greenfield plugin spike is done.

## Why

`pm` (the plugin) and `pmb` (the CLI) are too generic: hard to search for and easy to clash with other tools. The project will be published. The user picked **acta** (Latin: records of deeds done), one name for the plugin and the CLI.

Name checks on 2026-09-26: Homebrew has no `acta`; GitHub `iyay/acta` is free; npm `acta` is taken by an unrelated React state manager (v3.0.15), which matters only if the CLI is ever published on npm (a scope would fix it).

## Goals

- One name everywhere a user or an agent sees it.
- Nothing breaks on the day of the rename: old names keep working for a while.
- Existing repos can move their root folder with one command.

## Non-goals

- Publishing (GitHub push, Homebrew tap, npm). Separate work.
- Removing the old-name fallbacks. That is a later version.
- Changing short ID prefixes (`SPEC-`, `PLAN-`, `BUG-`) or the file contract beyond the root folder name.

## 1. Names

| Today | After |
|---|---|
| plugin `pm`, skills `/pm:plan`, `pm:build`, ... | plugin `acta`, skills `/acta:plan`, `acta:build`, ... |
| CLI `pmb`, folder `cmd/pmb` | CLI `acta`, folder `cmd/acta` |
| root folder `.pm/` | `.acta/` |
| `.pm.yaml`, env `PM_ROOT` | `.acta.yaml`, env `ACTA_ROOT` |
| `~/.pm/voice.yaml` | `~/.acta/voice.yaml` |
| Go module `pm-board` | `github.com/iyay/acta` |
| `plugin/package.json` name `pm` | `acta` |

Install line in the README: `go install github.com/iyay/acta/cmd/acta@latest`.

## 2. Transition

- **CLI alias.** A `pmb` binary is still built (`cmd/pmb`, a thin wrapper). It runs the same code as `acta` and prints one line to stderr first: `pmb is now acta; this name goes away in a later version`. Exit codes and output are otherwise identical.
- **Root folder.** Lookup order becomes: `--root` flag, `ACTA_ROOT`, `PM_ROOT`, `root` in `.acta.yaml`, `root` in `.pm.yaml`, then `.acta/` if it exists, else `.pm/` if it exists, else `.acta/` (the default for a new repo).
- **Voice file.** `~/.acta/voice.yaml` is read first; if it is missing, `~/.pm/voice.yaml`. Writes (`acta voice`, `acta:setup`) always go to `~/.acta/voice.yaml`.
- **Skills.** No alias: a plugin has one name, so `pm:*` becomes `acta:*` at once. The session-start hook text names the new skills.
- **`.agents.json`, ID fields, bug files** keep their names; only the folder around them moves.

## 3. `acta migrate-root`

- Moves `.pm/` to `.acta/` with `git mv`, then commits once: `acta: move root folder .pm to .acta`.
- Also renames `.pm.yaml` to `.acta.yaml` in the same commit when it exists and only holds settings (no path rewrite needed, since `root` values are relative).
- Refuses with exit 1 and a clear message when `.acta/` already exists, when `.pm/` does not exist, or when the working tree has uncommitted changes under `.pm/`.
- Prints what it moved. Never pushes.
- Other worktrees keep their own copy until their branch merges; the board already reads both folder names (section 2), so nothing is lost in between.

## 4. What else changes

- Every skill, reference file, hook script, the omp extension, `plugin/README.md`, `plugin/.claude-plugin/plugin.json` and `marketplace.json`: `pm:` becomes `acta:`, `pmb` becomes `acta`, `.pm/` becomes `.acta/`.
- `internal/plugincheck` tests: required and forbidden strings follow the new names; a new check forbids `pm:` skill references and bare `pmb ` commands in plugin text (except the one line that explains the alias).
- Go imports move from `pm-board/...` to `github.com/iyay/acta/...`.
- This repo itself runs `acta migrate-root` as the last step, so `.pm/` becomes `.acta/` here too.
- The user renames the local folder `~/Nayakatara/pm-board` to `~/Nayakatara/acta` by hand after landing (agents never move the main checkout).

## Testing

- CLI: `acta` does everything `pmb` did; `pmb` gives the same output plus the one warning line on stderr.
- Root lookup: each step of the order in section 2, including `.pm/` only, `.acta/` only, both (`.acta/` wins), neither (new repo gets `.acta/`), and `PM_ROOT` / `.pm.yaml` still honored when the new ones are missing.
- Voice: new file first, old file as fallback, writes go to the new file.
- `migrate-root`: moves and commits once; refuses when `.acta/` exists, when `.pm/` is missing, and on uncommitted changes under `.pm/`; `.pm.yaml` renamed with it.
- Plugin text: no `pm:` skill names or bare `pmb` commands left outside the alias note; plugin manifest names are `acta`.
- Full suite green with the new module path.
