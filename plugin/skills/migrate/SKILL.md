---
name: migrate
description: Use when the user wants docs from another workflow plugin (superpowers, gstack, a .scratch tracker, or any other format) moved into .acta/ so they can be classified and tracked. Optional - without it, acta already shows listed legacy folders read-only.
---

# Migrate

Moves planning docs written by another workflow into `.acta/`, in the acta file contract's shape. Optional: `acta` already shows folders listed under `legacy` in `.acta.yaml` as read-only items. Work in a worktree like any other change (`acta:build`, Worktree).

## superpowers docs

superpowers already writes close to the contract, so this is a move with no content change:

1. Run `acta migrate superpowers`. It is a dry run: it prints an old-path to new-path table and any name clash.
2. Show the table to the user and wait for a yes.
3. Run `acta migrate superpowers --apply`. It refuses on a dirty tree or a clash, moves with `git mv`, makes one commit, deletes nothing, drops `docs/superpowers` from `legacy`, and never pushes.

If the installed `acta` has no `migrate` command yet, use the route below for superpowers docs too, moving each file with `git mv` instead of writing a new one.

## Any other format

1. Read the source folder. Build a table with one row per source file: the source path, what it becomes (story in `specs/`, plan with tasks in `plans/`, or bug in `bugs/`), and the new path `YYYY-MM-DD-<slug>.md`. A file with no date in its name takes the date of its first commit: `git log --diff-filter=A --format=%as -- <file> | tail -1`.
2. Show the table. Change it until the user says yes. Write nothing before that.
3. Write each new file in the contract's shape:
   - a story: `# <title>`, then the source text;
   - a plan: `# <title>`, a `**Spec:** <path of its story>` line when it has one, and each task as `### Task N: <title>` with `- [ ]` or `- [x]` steps that keep the source's done state;
   - a bug: `# <symptom>`, then `## Symptom` and whichever of `## Root cause`, `## Repro` and `## Found in` the source has.
4. Leave the source files where they are. The user decides when to delete them.
5. Check with `acta list --all --json` that every new item shows and has no `problems`.
6. Commit the new files in one commit, staged by path: `acta: migrate <source> into .acta`. Never push.
