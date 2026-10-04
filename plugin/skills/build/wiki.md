# Wiki pages

Read this before you write, change or propose a wiki page. A page keeps project knowledge in git, not in agent memory.

## Format

Pages live in `.acta/wiki/` under the acta root, next to the specs. One concept per file. Subfolders are fine. A page id is its path minus `.md`. A page has no acta id and no status, and the board skips the wiki. A repo with no `.acta/wiki/` has the feature off and pays 0 tokens; the first page turns it on.

```yaml
---
type: Gotcha
title: x/ansi Wrap overflows
description: Wrap can return lines wider than the limit after " -"; use Hardwrap(Wordwrap(...))
paths: [internal/tui/]
timestamp: 2026-10-04T14:05:00+07:00
---
```

- `type` is one of five. `Decision`: a choice and why; conventions too. `Gotcha`: a trap that cost time, and its fix. `Runbook`: steps for a job; playbooks too. `Reference`: facts to look up; architecture too. `Glossary`: one page, `glossary.md`, one domain term per line.
- `description` is one line, 120 characters at most. It is the hint text, so it carries the key fact.
- `paths` lists files or folders from the repo root. A folder ends with `/`. Matching is by prefix, with no globs. It may be empty.
- `timestamp` is ISO 8601 with a time. Bump it on every change and on every re-check. `date -u +%Y-%m-%dT%H:%M:%SZ` prints one.
- The body is 250 words at most. Split a longer page and link the parts.
- There is no `index.md` and no `log.md`. `acta wiki ls` builds the list from the frontmatter and git keeps the history. Both files would clash between parallel worktrees.

## Writing

Pages are written in the last step of `acta:build`, before review, so review sees the wiki diff. The agent that ran the plan writes them: the main agent, or omp under `dispatch`. A plan has no wiki task. All of it is one commit in the worktree, and none when nothing changed.

1. Run `acta wiki check <parent>..HEAD`. It checks the pages whose `paths` cover a file the branch changed, and prints one line per problem: a stale page (a commit changed a file under its `paths` after its `timestamp`), or a rule above broken.
2. Fix each touched page: read it, read the branch diff under its `paths`, and fix every line that is now wrong. Bump `timestamp` even when nothing was wrong.
3. Add a page only for a lesson a fresh agent would lose time without.
4. Run the check again until it prints nothing, then commit what changed.

A page is written at most once per plan. An update changes lines. It does not rewrite the page. Nothing that git, plans or debt already hold gets a page, so no LANDED status and no review NOTEs.
