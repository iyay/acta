---
name: migrate
description: "acta: Use when the user wants docs from another workflow plugin (superpowers, gstack, a .scratch tracker, or any other format) moved into .acta/ so they can be classified and tracked. Optional - without it, acta already shows listed legacy folders read-only."
---

# Migrate

Moves planning docs written by another workflow into `.acta/`, in the acta file contract's shape. Optional: `acta` already shows folders listed under `legacy` in `.acta.yaml` as read-only items. Work in a worktree like any other change (`acta:build`, Worktree).

## superpowers docs

Superpowers docs take the generic route below: build the table the same way, move each file with `git mv` instead of writing a new one, and after the commit drop `docs/superpowers` from `legacy` in `.acta.yaml`.

## Any other format

1. Read the source folder. Build a table with one row per source file: the source path, what it becomes (story in `specs/`, plan with tasks in `plans/`, or bug in `bugs/`), and the new path `YYYY-MM-DD-<slug>.md`. A file with no date in its name takes the date of its first commit: `git log --diff-filter=A --format=%as -- <file> | tail -1`.
2. Show the table. Change it until the user says yes. Write nothing before that.
3. Write each new file in the contract's shape:
   - a story: `# <title>`, then the source text;
   - a plan: `# <title>`, a `**Spec:** <path of its story>` line when it has one, and each task as `### Task N: <title>` with `- [ ]` or `- [x]` steps that keep the source's done state;
   - a bug: `# <symptom>`, then `## Symptom` and whichever of `## Root cause`, `## Repro` and `## Found in` the source has.
4. Leave the source files where they are. The user decides when to delete them.
5. Check with `acta list --all --json` that every new item shows and has no `problems`.
6. Commit the new files in one commit, staged by path: `chore: migrate <source> into .acta`. Never push.

## Seed the wiki

Seeds an empty `.acta/wiki/` from an existing project. Runs when the user asks, or when `acta:setup` offers it. Never runs without a yes.

1. Say how many sources will be read, so the user sees the cost first.
2. Read sources by convention, read-only: ADR folders, CHANGELOG, README, agent rule files, commits whose message records a decision, `WHY:` and `HACK:` comments, and the harness's own memory for this project when it has one.
3. Keep only facts the code cannot tell: decisions and why, gotchas, runbooks, reference facts, domain terms. Drop what git or the plans already hold.
4. Show one table: source, page type, page path, `paths`, or drop with the reason. Wait for a yes. Write nothing before that.
5. After the yes, write the pages in a worktree in the format of `../build/wiki.md`, pass `acta wiki check`, then review and land like any plan.
