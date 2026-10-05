---
parent: scratch/2026-10-05-wiki-first-time-seed
id: SPC-0086
created: "2026-10-05 22:21:32"
hash: a5dq1dj
started: "2026-10-05 22:24:19"
finished: "2026-10-05 22:40:40"
---
# Seed the wiki for the first time in an existing project

Status: Architectural, shaped on 2026-10-05 in the session that shaped SCR-0041, on the user's explicit ruling to run it there and take every recommendation without asking. Rulings live in the SCR-0042 Log.

Why: the project wiki (SPC-0071) only grows from new plans. An existing project starts with an empty `.acta/wiki/`, while its lasting knowledge sits in ADRs, READMEs, agent rule files, commit messages and agent memory. Agents keep reading that memory, which other harnesses cannot see well. In the acta repo itself the dispatch brief still tells omp to read Claude memory (`internal/cli/dispatch_brief.go`), because no wiki page exists.

## 1. Where it lives

- `plugin/skills/migrate/SKILL.md` gets a section "Seed the wiki". It runs when the user asks, or when acta:setup offers it.
- `plugin/skills/setup/SKILL.md` offers it in one line when the repo has code and `.acta/wiki/` has no page. It never runs without a yes.

## 2. How it works

1. Say how many sources will be read, so the user knows the cost.
2. Read sources by convention, read-only: ADR folders, CHANGELOG, README, AGENTS.md, CLAUDE.md, .cursorrules, commits whose message records a decision, `WHY:` and `HACK:` comments, and the harness's own memory for this project when it has one.
3. Keep only facts the code cannot tell: decisions and why, gotchas, runbooks, reference facts, domain terms. Drop code structure, status and history that git or the plans hold.
4. Show one table: source, page type, page path, `paths`, or drop with the reason. Wait for a yes.
5. After the yes, write the pages in a worktree in the format of `plugin/skills/build/wiki.md`, pass `acta wiki check`, then review and land like any plan.

The host agent does all of it with this prompt. No third-party tool, no install, no extra model key, and nothing hosted. The page format does not change.

## 3. The acta repo itself

- Seed `.acta/wiki/` in this repo from its project memory and other sources, by the steps above. The user's ruling stands in for the table's yes.
- Then the dispatch brief stops pointing omp at Claude memory: when the repo has wiki pages, the MEMORY line becomes a line telling the recipient to run `acta wiki ls` and open the pages whose `paths` cover the plan's files. With no pages, the brief has no memory line.

## 4. Tests and version

- plugincheck: the migrate section exists and names the table and the yes; the setup line exists.
- dispatch brief: with pages, the brief names `acta wiki ls` and no Claude memory path; with none, neither.
- `acta wiki check` passes on the seeded pages.
- The last task adds 1 to the patch version in the three plugin files.

Out of scope: an OKF importer, Repowise-style decision mining, and any third-party or hosted generator.
