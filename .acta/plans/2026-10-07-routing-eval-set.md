---
parent: scratch/2026-10-07-routing-eval-set
depth: minimal
closes: [SPC-0097]
id: PLN-0106
created: "2026-10-07 06:42:37"
hash: n94p6ul
started: "2026-10-07 06:47:30"
finished: "2026-10-07 07:12:04"
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

- [x] Red: `internal/plugincheck/routing_evals_test.go` with `TestRoutingEvalCases`: a table of the 16 folder names; it fails when a folder is missing, when an extra `routing-*` folder exists, or when any rule in verify breaks (frontmatter keys, scaffold present, grader present, no `type: llm`, banned words absent from the prompt body). Run `scripts/test ./internal/plugincheck/ -run TestRoutingEvalCases` and watch it fail on the missing folders.
- [x] Green: write the 16 cases. `prompt.md` frontmatter: `name` (the folder name), `description` (one line: the prompt and the route it should get), `tags: [routing]`, `runs: 3`, `max_turns: 6`, `timeout_seconds: 240`, `allowed_tools: []`, `context` is not needed because the case folder holds `scaffold.sh`. Each prompt ends with `The acta CLI for this repo is ./bin/acta. It is on PATH as well.` Graders per route: scratch = `tool_used` Bash `input_match: "acta scratch new"` min 1, plus `tool_used` Write `input_match: ".acta/specs/"` min 0 max 0; arch = `tool_used` Bash `input_match: "status brainstorming"` min 1; light = `tool_used` Bash `input_match: "acta scratch new"` min 0 max 0, plus `tool_used` Bash `input_match: "status brainstorming"` min 0 max 0; chat = `tool_used` Write min 0 max 0, `tool_used` Edit min 0 max 0, `tool_used` Bash `input_match: "acta scratch"` min 0 max 0; second = `regex` on `last_message` with `pattern: "claude --bg"`, plus `tool_used` Write `input_match: ".acta/specs/"` min 0 max 0. Each grader file starts with a `# guards: <route>` comment line inside the frontmatter, as the existing graders do. Prompt ideas: scratch = a side thought about something worth doing someday while the user says they are busy with other work; arch = a request that changes how many modules fit together (auth rewrite, move storage to another database, plugin system); light = one wrong value or word in one named file; chat = a plain question about git, Go or how a tool works; second = the user says an open design is already running in this session, then asks for another big change.
- [x] Commit: `test(evals): routing eval set, 16 cases with action graders (SPC-0097)`

### Task 02: Version 0.1.29

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.29, and `internal/plugincheck` passes. List each file and the version it holds.

- [x] Red: none needed; `scripts/test ./internal/plugincheck/` guards that the three agree.
- [x] Green: change 0.1.28 to 0.1.29 in all three files.
- [x] Commit: `chore: version 0.1.29`

## Fix round 1

### Task 03: Fix round 1

**Files:** `plugin/evals/routing-light-01/` to `-04/` (`prompt.md`, `graders/`), `plugin/evals/routing-chat-01/` to `-04/graders/`, `plugin/evals/routing-arch-01/` to `-04/graders/`, the prompts of whichever mixed-language cases drop their mix, `internal/plugincheck/routing_evals_test.go`, this plan file (Task 01 text only).

**verify:** No routing case can pass when the agent takes a route other than its own: for every route, list the wrong routes and the grader that fails each one. The light prompts neither list the route options nor tell the agent to stop or to leave files alone. Exactly 2 or 3 prompts in the set mix English and Indonesian, and the test fails on 1 or on 4.

- [ ] BLOCKER (light graders cannot tell routes apart): rewrite the four light prompts as a plain request to fix the problem. Drop "full process with competing approaches and a sectioned write-up, or is the short route enough", drop "Answer in a few lines and stop" and "Do not change any file", and their Indonesian versions. Put the wrong line itself in the prompt (the file is not in the scaffold). With the question gone, an agent that takes the Architectural path files a scratch item and sets `status brainstorming`, which the existing `max: 0` graders catch. Add a banned phrase check to `TestRoutingEvalCases` so a light prompt holding "stop" or "do not change" (case-insensitive) fails.
- [ ] [fix] Chat graders: add `tool_used` on `Skill` with `input_match: "acta:"`, `min: 0`, `max: 0`, so loading a workflow skill fails a chat case, as spec item 2 says.
- [ ] [fix] Arch graders: add `tool_used` on `Bash` with `input_match: "acta scratch new"`, `min: 1`, next to the existing `status brainstorming` grader.
- [ ] [fix] Mixed-language count: keep exactly 3 mixed prompts (light-02, light-04 and second-02 after the rewrite keep their mix; arch-02 and arch-04 become plain Indonesian). Change the test so it fails below 2 and above 3.
- [ ] [fix] Task 01 text in this plan: "16 new" becomes "18 new" (the spec counts 4+4+4+4+2), and "mode 0755" becomes "mode 0644, like the existing scaffolds".
- [ ] Commit: `fix(evals): light routing cases ask for the fix, stricter chat and arch graders (PLN-0106 fix round 1)`
