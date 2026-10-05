---
parent: specs/2026-10-05-wiki-first-time-seed-design
depth: minimal
id: PLN-0096
created: "2026-10-05 22:22:00"
hash: rg4vtp6
started: "2026-10-05 22:24:19"
finished: "2026-10-05 22:40:40"
---
# Wiki first-time seed Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** acta can seed an empty wiki from an existing project, the acta repo gets its first pages, and the dispatch brief points at the wiki instead of Claude memory.

**Spec:** `.acta/specs/2026-10-05-wiki-first-time-seed-design.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Skill text fits every user and harness: no repo-specific paths, no user names, short plain words.
- Pages follow `plugin/skills/build/wiki.md` exactly: five types, one-line description of at most 120 characters, body of at most 250 words, `paths` by prefix, ISO `timestamp`.
- No third-party tool, no install, nothing hosted.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3
- Wave 3: Task 4

### Task 1: migrate section and setup offer

**Files:** Modify `plugin/skills/migrate/SKILL.md`, `plugin/skills/setup/SKILL.md`, a test file under `internal/plugincheck/`, and `internal/plugincheck/budget_test.go` only if a cap needs it.
**verify:** The migrate skill holds a "Seed the wiki" section with the five steps of the spec (cost first, sources by convention read-only, keep only facts the code cannot tell, one table then a yes, pages in a worktree that pass acta wiki check); setup offers it in one line only when the repo has code and no wiki page; plugincheck fails when the section, the table or the yes is gone. List each spec step and where the text says it.
- [x] Failing test: a plugincheck test that wants the migrate section, the table and the yes, and the setup line; fails on today's text.
- [x] Code: write the section and the setup line, short.
- [x] Commit: `skills: migrate can seed the wiki in an existing project`.

### Task 2: seed the acta wiki

**Files:** Create pages under `.acta/wiki/` only.
**verify:** Every page passes `acta wiki check` and the format in `plugin/skills/build/wiki.md`; every page holds a fact the code cannot tell (a decision and why, a gotcha and its fix, a runbook, a reference fact, or a domain term); no page repeats status, review NOTEs or history that git, plans or debt hold. List each page with its type, its `paths` and the source it came from, and each source item dropped with the reason.
- [x] Failing test: `acta wiki ls` prints nothing today.
- [x] Code: follow the new migrate steps on this repo. Sources: the project memory named in the brief's MEMORY line (read-only), CLAUDE.md, plugin/evals/FACTS.md, plugin/omp/FACTS.md. The user's ruling stands in for the table's yes; put the table in the reply.
- [x] Commit: `wiki: seed the acta wiki from memory and repo docs`.

### Task 3: brief reads the wiki

**Files:** Modify `internal/cli/dispatch_brief.go`, `internal/cli/dispatch_brief_test.go`.
**verify:** For every repo (with wiki pages, without) the brief never names a Claude memory path; with pages it tells the recipient to run `acta wiki ls` and open the pages whose `paths` cover the plan's files; without pages it has no memory or wiki line. List both cases checked.
- [x] Failing test: a brief built for a repo with one wiki page must name `acta wiki ls` and no `.claude` path; fails because the MEMORY line names Claude memory today.
- [x] Code: replace the MEMORY line and `memoryPaths` with a wiki line written only when the repo has pages; drop what becomes unused.
- [x] Commit: `dispatch: the brief points at the wiki, not Claude memory`.

### Task 4: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [x] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [x] Code: add 1 to the patch in all three files.
- [x] Commit: `plugin: bump patch version`.

## Polish

### Task 5: Review polish

**verify:** every NOTE below is applied, and nothing else changes; `acta wiki check` stays clean and each touched page gets its timestamp bumped.

- [x] .acta/wiki/dispatch-plan-path.md: quote the real error text from internal/cli/dispatch.go (`plan %q is outside %s`), not "outside the worktree .acta".
- [x] .acta/wiki/omp-skill-names.md: use current skill names in the examples (shape, slice), not brainstorm.
- [x] .acta/wiki/eval-red-case-debug.md: say the symlink check in general words (a symlink under the home folder that the sandbox cannot read), with no "on this machine" or Docker-specific setup.
- [x] .acta/wiki/ghostty-minimum-contrast.md: state the rule for any terminal with a minimum-contrast setting; keep 1.6 as the safe ratio, drop the one config value as a fact about a person.
- [x] plugin/references/house-rules.md: the MEMORY line names Claude memory paths the brief no longer has; rewrite it to say wiki pages named by `acta wiki ls` are read-only for the recipient, or drop it.
- [x] Commit: `polish: review notes for PLN-0096`

## Review notes

- SCR-0042 was shaped in the same session as SCR-0041 on the user's ruling; its status stays raw because the hook blocks a second brainstorming set.
- The brief adds its WIKI line only when wiki.Load finds pages, and drops the load error; fine for a hint line.
- dispatch-plan-path quotes the raw format string `plan %q is outside %s`, good for grep, less plain to read.
- The polish also bumped parallel-implementers-one-worktree's timestamp, since its paths cover house-rules.md.
