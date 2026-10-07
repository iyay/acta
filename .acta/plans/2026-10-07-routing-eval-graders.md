---
parent: debt/2026-10-07-routing-eval-set
depth: minimal
closes: [SPC-0098, DBT-0090.01, DBT-0090.02, DBT-0090.03]
id: PLN-0107
created: "2026-10-07 08:30:31"
hash: bdsgj85
started: "2026-10-07 08:32:45"
finished: "2026-10-07 08:44:10"
---
# Routing eval graders Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Routing cases tell the light route from the chat route by the workflow skill the agent loads, and Indonesian prompts run with chat_language: Indonesian.

**Spec:** .acta/specs/2026-10-07-routing-eval-graders.md

**Tests:** `scripts/test ./internal/plugincheck/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Grader types are only `tool_used` and `regex`, in the format of the existing routing graders (frontmatter with a `# guards: <route>` comment line). See `plugin/evals/FACTS.md`, section Graders.
- Prompt bodies do not change. Scaffolds change only in the chat_language value.
- Run tests with `scripts/test`, never bare `go test`. Never run `scripts/eval`; it spends the user's quota. Never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01, Task 02

Task 01 and Task 02 touch different files and can run together.

### Task 01: Skill grant, light and chat graders, scaffold language

**Files:** `internal/plugincheck/routing_evals_test.go`, every `plugin/evals/routing-*/prompt.md` (frontmatter `allowed_tools` line only), new `plugin/evals/routing-light-0{1..4}/graders/workflow-skill.md`, new `plugin/evals/routing-chat-0{1..4}/graders/no-skill-read.md`, `plugin/evals/routing-*/scaffold.sh` of the cases whose prompt is mainly Indonesian.

**verify:** For all 18 routing cases: `allowed_tools: [Skill]`; every light case has a `tool_used` Skill grader with `input_match: "acta:"` and `min: 1`; every chat case has both the Skill `acta:` and the Read `/skills/` graders with `min: 0` and `max: 0`; every scaffold writes `chat_language: Indonesian` when the test's language table marks the prompt Indonesian (mixed counts by its main language) and `English` otherwise. `TestRoutingEvalCases` fails when any one of these breaks. List each rule and the folders it covers.

- [x] Red: extend `TestRoutingEvalCases` with the four rules above, reusing its existing language table for the main language of each prompt; run `scripts/test ./internal/plugincheck/ -run TestRoutingEvalCases` and watch it fail on `allowed_tools: []`.
- [x] Green: set `allowed_tools: [Skill]` in all 18 prompts; add `workflow-skill.md` (`type: tool_used`, `tool: Skill`, `input_match: "acta:"`, `min: 1`) to each light case; add `no-skill-read.md` (`type: tool_used`, `tool: Read`, `input_match: "/skills/"`, `min: 0`, `max: 0`) to each chat case; in each Indonesian case's `scaffold.sh` change `chat_language: English` to `chat_language: Indonesian`.
- [x] Commit: `test(evals): routing cases grant Skill, light needs a workflow skill, Indonesian scaffolds (DBT-0090)`

### Task 02: Version 0.1.30

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.30, and `internal/plugincheck` passes. List each file and the version it holds.

- [x] Red: none needed; `scripts/test ./internal/plugincheck/` guards that the three agree.
- [x] Green: change 0.1.29 to 0.1.30 in all three files.
- [x] Commit: `chore: version 0.1.30`

## Review notes

- The scaffold language check uses strings.Contains, so a scaffold writing both chat_language values would pass; none does today.
- No test pins the prompt bodies; the review diffed them and only the allowed_tools line changed.
- Light cases now load a skill inside max_turns: 6; the turn budget is not measured yet.
- Commit b7357d9 says "re-check wiki" but changed no wiki page, and added a started: line to the DBT-0090 file.
