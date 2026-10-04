---
parent: scratch/2026-10-04-acta-wiki
id: SPC-0071
created: "2026-10-04 18:59:24"
hash: qemt9le
started: "2026-10-04 19:27:20"
---
# Project wiki under .acta/wiki

Status: design approved by the user in chat on 2026-10-04, section by section. Architectural: it adds a planning folder, a CLI command, a hook path and skill steps. Rulings and answers live in the SCR-0040 Log. Live work state was split out to SCR-0041.

Why: agents keep project knowledge in private agent memory. In this repo, 44 of 79 memory index lines are LANDED status that git already holds. Gotchas stay where omp cannot see them. A grep hook adds unrelated hits to every prompt. None of the user's repos has the CONTEXT.md, docs/adr or .okf/ that the rules name. The wiki keeps that knowledge in the repo. It shows a page only when the agent touches the files the page covers, and it checks that pages stay true. It must stay cheap in tokens, both in what gets loaded and in what agents write. The code index stays with codebase-memory-mcp.

## 1. Page format

- Pages live in `.acta/wiki/` under the acta root, the same way specs do. One concept per file. Subfolders are fine. A page id is its path minus `.md`.
- The board reads only the folders it knows, so it skips the wiki. Pages have no acta id and no status.
- Frontmatter:

```yaml
---
type: Gotcha
title: x/ansi Wrap overflows
description: Wrap can return lines wider than the limit after " -"; use Hardwrap(Wordwrap(...))
paths: [internal/tui/]
timestamp: 2026-10-04T14:05:00+07:00
---
```

- `type` is one of five: `Decision` (also holds what ADRs and conventions held), `Gotcha`, `Runbook` (also playbooks), `Reference` (also architecture), `Glossary` (one page, `glossary.md`, holds what CONTEXT.md held).
- `description` is one line, 120 characters at most. It is the hint text, so it carries the key fact.
- `paths` lists files or folders from the repo root. A folder ends with `/`. Matching is by prefix, with no globs. It may be empty.
- `timestamp` is ISO 8601 with a time. Bump it on every change and on every re-check.
- The body is 250 words at most. Split a longer page and link the parts.
- There is no `index.md` and no `log.md`. `acta wiki ls` builds the list from frontmatter, and git keeps the history. Both files would clash between parallel worktrees.
- A repo with no `.acta/wiki/` has the feature off and pays 0 tokens.
- The name "okf" is never used. The fields come from OKF, but the format text is acta's own.

## 2. Reading

1. Session start. Only when `.acta/wiki/` holds pages, `acta hook session-start` adds 45 words at most (about 60 tokens). The size does not grow with the page count. The text says how many pages exist. It tells the agent to read a page that a `wiki:` line names before it changes those files. It says project knowledge goes to the wiki, not to agent memory.
2. Hints. The PreToolUse matcher in `plugin/hooks/hooks.json` widens from `Bash` to `Bash|Read|Edit|Write|MultiEdit`. `acta hook pre-tool` takes the file path, or each word of a Bash command with a leading `./` and a trailing `/...` cut off. It makes each one relative to the root of the checkout the agent works in. Then it matches them by prefix against every page's `paths`. A match prints one line through `additionalContext`: `wiki: .acta/wiki/tui-wrap.md: <description>`. A hint always exits 0. Exit 2 stays for the blocks that exist today.
3. Once per context. Pages already shown are kept in `.acta/state/` (gitignored), keyed by `session_id` plus `agent_id`. Claude Code sends `agent_id` only when a hook fires inside a subagent, so the main thread and each subagent count as their own context. The list for a session resets when session start fires with source `clear` or `compact`, since the old hints left the context.
4. omp. The extension's `tool_call` handler can return `additionalContext` (checked in omp's `shared-events.d.ts` on 2026-10-04). omp's file tools are `read`, `edit` and `write`, with the path in `event.input.path` or `event.input.paths`. So omp gets the same hints through `acta hook pre-tool`, and no fallback is needed.
5. Shape. The Shared language section of `acta:shape` changes. Before its questions, shape reads `glossary.md` and `acta wiki ls --type Decision`. A new term or decision is proposed as a page and waits for a yes, as today.

No text is added to every prompt.

## 3. Writing and commands

1. Pages are written in the last step of `acta:build`, before review, so review sees the wiki diff. The agent that ran the plan does it: the main agent, or omp under dispatch. It runs `acta wiki check <parent>..HEAD`. It re-checks each page the branch touched: it fixes lines that are now wrong and bumps `timestamp`. It adds a page only for a lesson a fresh agent would lose time without. All of it is one commit in the worktree. Plans get no wiki task.
2. Commands. All of them are read-only and never commit.
   - `acta wiki ls [--type T]`: one line per page, with path and description.
   - `acta wiki match <file>...`: the pages whose paths cover those files. The hook uses the same code.
   - `acta wiki check [<range>]`: reports missing or bad fields, an unknown type, a description over the limit, a body over 250 words, paths that do not exist, and stale pages. A page is stale when a commit that touched a file under its paths is newer than its `timestamp`. A range touches a page when a file changed in that range falls under the page's paths. With a range, only touched pages are checked. Any problem exits 1.
3. `acta:land` runs `acta wiki check <parent>..HEAD` with its other gates. A problem blocks the merge. Most often the fix is a `timestamp` bump after a fix round touched the paths. The landing report lists the pages added or changed.
4. The format and the writing rules live in one file, `plugin/skills/build/wiki.md`. Build's last step and shape point to it. It is read only when a page gets written.
5. A page is written at most once per plan. An update changes lines. It does not rewrite the page. Nothing that git, plans or debt already hold gets a page, so no LANDED status and no review NOTEs.

Limit: a session with no build (questions, investigation) writes no page, since the main checkout is read-only for agents. Its findings go to a scratch or bug item first.

## 4. Success, all tested

1. With no `.acta/wiki/`, the session start text is byte-identical to today.
2. With a wiki, session start adds 45 words at most, the same at 1 page and at 100.
3. A hint is one line per page per context. A second touch prints nothing. After compaction it prints again.
4. In a temp git repo, the land gate fails on a stale touched page and passes after a `timestamp` bump.
5. Two eval cases. Hint: a Gotcha page covers `src/parser.go` and the task changes the parser. The grader checks that the gotcha's fix is used. Write: the branch changes a covered file. The grader checks that the page was re-checked before review.

## 5. Out of scope

Semantic search. Background jobs. A user-level wiki, which stays in `~/.claude/memory`. A TUI pane. The code index. Live work state (SCR-0041). Editing the live context the way the CLM paper does. Moving the 79 agent-memory notes, a chore after land. The user's own CLAUDE.md and `memory-turn` hook, which the user changes after land.

## 6. Risks and open checks

1. Resolved at slice time: omp's `tool_call` result has `additionalContext` (see 2.4).
2. Resolved at slice time: Claude Code hook input carries `agent_id` inside a subagent (Claude Code hooks docs, common input fields), so the key is `session_id` plus `agent_id` (see 2.3).
3. Claude Code auto-memory stays on. The session start line points knowledge to the wiki. Turning auto-memory off is the user's call.
4. An agent-written page can carry a prompt injection into every later session. Pages are written before review, so review reads them, and the landing report lists every changed page.
5. The hook now runs on every Read and Edit. The plan measures its time per call. Target: under 30 ms.

Version: the plan ends with the patch bump in the three plugin files, as every plan does.
