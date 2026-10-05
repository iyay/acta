---
id: SCR-0042
hash: k7ic4li
title: Generate the wiki for the first time in an existing project
status: raw
created: "2026-10-05 07:08:34"
schema: "1"
finished: "2026-10-05 22:21:32"
---
# Generate the wiki for the first time in an existing project

## Words

### 2026-10-05

User's words (2026-10-04 and 2026-10-05, chat in Indonesian, put into English):

- Existing projects (brownfield) need a way to generate the wiki for the first time.
- It must work for every acta user. It must not be built on one user's setup or example.
- A third-party tool to generate it is fine. It runs once, and the tool is removed after use.
- For example OKF, or another tool that is better and fits.

Depends on the project wiki: SPC-0071 / PLN-0081, `.acta/wiki/`, pages of type Decision, Gotcha, Runbook, Reference or Glossary, checked by `acta wiki check`.

Research (2026-10-05, read-only, no tool was run):

- OKF is Open Knowledge Format. It is an open Apache-2.0 spec from Google Cloud, now at v0.2 (https://github.com/GoogleCloudPlatform/open-knowledge-format). It is a format, not a plugin: Markdown files, one concept each, with YAML frontmatter. v0.2 adds `sources`, `generated`, `verified`, `status` and `stale_after`. Google's own generator reads BigQuery datasets, not code repos. acta's page fields are already close to OKF.
- OpenWiki (https://github.com/langchain-ai/openwiki, MIT) writes an OKF bundle from a codebase. Its direct mode is `npx -y openwiki --init` and needs its own model key (or Ollama). Its host mode uses the agent's session but needs a global install. It leaves `openwiki/`, marker blocks in AGENTS.md and CLAUDE.md, and `~/.openwiki/`.
- Repowise (https://github.com/repowise-dev/repowise, AGPL-3.0) mines decision records with evidence and the files they govern, which fits `paths`. Use its ideas only. Copy no code, because of the license.
- Agent OS `/discover-standards` (https://github.com/buildermethods/agent-os, MIT) is a prompt. It asks the user why for each code pattern, and the common mistake.
- Most generators write code structure: Serena onboarding, `/init`, DeepWiki and its clones, Google Code Wiki, Kiro steering, Operator Memory project-init. The wiki drops code structure, since the code index covers it. About 10 to 20% of their output would survive as pages, as an estimate. Hosted ones send the code out, and private repos cost money. BMAD's `document-project` is deprecated.

Proposed shape (not decided):

1. An opt-in "generate the wiki for the first time" step, offered only when the repo has code and `.acta/wiki/` is empty or missing.
2. Default generator `agent`: a built-in acta prompt that the host agent runs. It finds sources by convention: ADR folders, CHANGELOG, PR bodies through `gh`, commits that record a decision, `WHY:` and `HACK:` comments, README, and existing AGENTS.md, CLAUDE.md or .cursorrules. No install and no extra model key, so it works on every harness.
3. Third-party option: one OKF importer, so any tool that writes an OKF bundle can feed the seed (OpenWiki today). Repowise is a second option, for decisions.
4. A tool runs in a temporary clone under `$TMPDIR` with a temporary HOME, through `npx -y` or `uvx`. The real checkout is never touched, and nothing is installed globally. After the conversion, the clone and the temporary HOME are deleted. That delete needs the user's yes, since it is a recursive delete.
5. Convert and filter: only facts the code cannot tell become pages. Show a table first (source item, page type, path, `paths`, or "drop" with the reason). After the user's yes, write the pages in a worktree, pass `acta wiki check`, then review and land.

Open questions:
- Should seeded pages carry OKF's `generated` and `verified` flags until a person checks them?
- Is a hosted generator ever allowed, and only as opt-in?
- Where does it live: a section of `acta:migrate`, or its own step offered by `acta:setup`?
- What does a generator cost in quota or in model keys, and how is the user told before it runs?

## Context

## Log

### 2026-10-05

2026-10-05 user: do this before changing the dispatch brief. The brief still tells omp to read the Claude memory indexes (internal/cli/dispatch_brief.go:255) and has no wiki line; this repo has no .acta/wiki/ pages yet. Once the wiki is seeded, a Bounded change swaps the brief's MEMORY line for an `acta wiki ls` line. Seed sources here: the project memory folder holds many lasting facts (eval gotchas, land rules, commit format) that belong in wiki pages.

### 2026-10-05

2026-10-05 shaped in the same session as SCR-0041, on the user's explicit ruling ("di sini aja", take every recommendation, no confirmations). The pre-tool hook blocks `acta set ... brainstorming` for a second item, so the status was not set; this is noted on purpose, not worked around.
Rulings (recommendations, user pre-approved):
- Home: a "Seed the wiki" section in acta:migrate (migrate already brings existing knowledge into acta); acta:setup offers it in one line when the repo has code and no wiki pages.
- Generator: the host agent with a built-in prompt only. No third-party tool by default: no installs, no model keys, no temp clone to delete. OKF importer and Repowise ideas are out of scope for this plan.
- Hosted generators: never.
- Sources by convention: ADR folders, CHANGELOG, README, AGENTS.md / CLAUDE.md / .cursorrules, commits that record a decision, WHY:/HACK: comments, and the harness's own memory for this project when it has one (read-only).
- Filter: only facts the code cannot tell; show a table first (source, type, page path, paths, or drop with reason); after the yes, write pages in a worktree, pass acta wiki check, review, land.
- No generated/verified fields: the page format stays as it is; review checks the seed.
- Cost: say how many sources will be read before starting.
- This plan also seeds the acta repo itself from its project memory, and then swaps the dispatch brief's MEMORY line for a wiki line.

## Open questions
