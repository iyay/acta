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
---
```

- `type` is one of five. `Decision`: a choice and why; conventions too. `Gotcha`: a trap that cost time, and its fix. `Runbook`: steps for a job; playbooks too. `Reference`: facts to look up; architecture too. `Glossary`: one page, `glossary.md`, one domain term per line.
- `description` is one line, 120 characters at most. It is the hint text, so it carries the key fact.
- `paths` lists files or folders from the repo root. A folder ends with `/`. Matching is by prefix, with no globs. It may be empty. Name the narrowest folders or files the page is about; a whole top-level folder makes the page stale on nearly every branch.
- The body is 250 words at most. Split a longer page and link the parts.
- There is no `index.md` and no `log.md`. `acta wiki ls` builds the list from the frontmatter and git keeps the history. Both files would clash between parallel worktrees.

## Writing

Pages are written in the last step of `acta:build`, before review, so review sees the wiki diff. The agent that ran the plan writes them: the main agent, or omp under `dispatch`. A plan has no wiki task. All of it is one commit in the worktree, and none when nothing changed.

1. Run `acta wiki check <parent>..HEAD` in the worktree. It checks pages covering branch changes. It reports errors and stale warnings: a commit changed a file under its `paths` after the page was last checked. Errors block; warnings alone do not.
2. Read affected pages and relevant diffs; fix every line that is now wrong. `acta wiki ok <id>` is optional after reviewing an unchanged page.
3. Write each page the spec approved. Shape notes each yes in the spec: a term or decision it proposed as a page.
4. Add any other page only for a lesson a fresh agent would lose time without.
5. Run the check again until there are no errors and any incorrect facts found in review have been fixed, then commit what changed.

A page is written at most once per plan. An update changes lines. It does not rewrite the page. Nothing that git, plans or debt already hold gets a page, so no LANDED status and no review NOTEs.
