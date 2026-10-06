---
parent: scratch/2026-10-07-routing-eval-set
depth: minimal
closes: [SPC-0097]
id: PLN-0106
created: "2026-10-07 06:42:37"
hash: n94p6ul
---
# Routing eval set Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** 16 new eval cases measure how often the skill rules pick the right route from a prompt that does not name it.

**Spec:** .acta/specs/2026-10-07-routing-eval-set.md

**Tests:** `scripts/test ./internal/plugincheck/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Grader types are only `tool_used` and `regex` (see `plugin/evals/FACTS.md`, section Graders). No `llm` grader in any routing case.
- A prompt never names its route: no "note this", "later", "idea", "follow the acta workflow", "design", "brainstorm", "light path", "Architectural", "Bounded", "scratch". It reads like a user talking normally. The cases do not tell the agent to stop early either, except the chat cases may ask only a question.
- Run tests with `scripts/test`, never bare `go test`. Do not run `scripts/eval` in a task; it spends the user's quota.
- Never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01, Task 02

Task 01 and Task 02 touch different files and can run together.

### Task 01: 16 routing cases and their layout test

**Files:** `internal/plugincheck/routing_evals_test.go` (new), `plugin/evals/routing-scratch-01/` to `-04/`, `plugin/evals/routing-arch-01/` to `-04/`, `plugin/evals/routing-light-01/` to `-04/`, `plugin/evals/routing-chat-01/` to `-04/`, `plugin/evals/routing-second-01/` to `-02/`. Each case folder holds `prompt.md`, `scaffold.sh` (a byte-for-byte copy of `plugin/evals/note-to-scratch/scaffold.sh`, mode 0755) and `graders/*.md`.

**verify:** Every `plugin/evals/routing-*` folder is one of the 16 names, and every one of the 16 has: frontmatter with `tags: [routing]`, `runs: 3`, `max_turns:` and `timeout_seconds:`; a scaffold; at least one grader; no grader with `type: llm`; and a prompt body that contains none of the banned route words from Global Constraints (case-insensitive). The language split holds: for scratch, arch, light and chat, two English and two Indonesian prompts; second has one of each; at least two prompts in the set mix English and Indonesian. List each check and the folders it ran over.

- [ ] Red: `internal/plugincheck/routing_evals_test.go` with `TestRoutingEvalCases`: a table of the 16 folder names; it fails when a folder is missing, when an extra `routing-*` folder exists, or when any rule in verify breaks (frontmatter keys, scaffold present, grader present, no `type: llm`, banned words absent from the prompt body). Run `scripts/test ./internal/plugincheck/ -run TestRoutingEvalCases` and watch it fail on the missing folders.
- [ ] Green: write the 16 cases. `prompt.md` frontmatter: `name` (the folder name), `description` (one line: the prompt and the route it should get), `tags: [routing]`, `runs: 3`, `max_turns: 6`, `timeout_seconds: 240`, `allowed_tools: []`, `context` is not needed because the case folder holds `scaffold.sh`. Each prompt ends with `The acta CLI for this repo is ./bin/acta. It is on PATH as well.` Graders per route: scratch = `tool_used` Bash `input_match: "acta scratch new"` min 1, plus `tool_used` Write `input_match: ".acta/specs/"` min 0 max 0; arch = `tool_used` Bash `input_match: "status brainstorming"` min 1; light = `tool_used` Bash `input_match: "acta scratch new"` min 0 max 0, plus `tool_used` Bash `input_match: "status brainstorming"` min 0 max 0; chat = `tool_used` Write min 0 max 0, `tool_used` Edit min 0 max 0, `tool_used` Bash `input_match: "acta scratch"` min 0 max 0; second = `regex` on `last_message` with `pattern: "claude --bg"`, plus `tool_used` Write `input_match: ".acta/specs/"` min 0 max 0. Each grader file starts with a `# guards: <route>` comment line inside the frontmatter, as the existing graders do. Prompt ideas: scratch = a side thought about something worth doing someday while the user says they are busy with other work; arch = a request that changes how many modules fit together (auth rewrite, move storage to another database, plugin system); light = one wrong value or word in one named file; chat = a plain question about git, Go or how a tool works; second = the user says an open design is already running in this session, then asks for another big change.
- [ ] Commit: `test(evals): routing eval set, 16 cases with action graders (SPC-0097)`

### Task 02: Version 0.1.29

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.29, and `internal/plugincheck` passes. List each file and the version it holds.

- [ ] Red: none needed; `scripts/test ./internal/plugincheck/` guards that the three agree.
- [ ] Green: change 0.1.28 to 0.1.29 in all three files.
- [ ] Commit: `chore: version 0.1.29`
