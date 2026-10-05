---
parent: specs/2026-10-05-handoff-language-eval-read
depth: minimal
id: PLN-0088
created: "2026-10-05 14:25:27"
hash: zeb4nps
started: "2026-10-05 14:36:45"
---
# Hand-off language and eval Read Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Agents write subagent prompts in the repo language, and evals may use the Read tool.

**Spec:** `.acta/specs/2026-10-05-handoff-language-eval-read.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- New voice line, with the repo language filled in: `- Write prompts and hand-offs to subagents or other agents in <repo language>.`
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3

### Task 1: voice line for hand-offs

**Files:** Modify `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`.
**verify:** Every session start text that names a repo language (set voice, voice read error fallback, no voice yet, the default-rules.md fallback) tells the agent which language to use for prompts to subagents and other agents, and it is always the repo language, never the chat language. List each path checked.
- [x] Failing test: session start with chat Indonesian and repo English must hold `- Write prompts and hand-offs to subagents or other agents in English.`; fails because the line does not exist.
- [x] Code: add the line after the repo language line in SessionStart; regenerate default-rules.md the way hook_test.go's update flag does.
- [x] Commit: `hook: prompts to subagents use the repo language`.

### Task 2: evals may use Read

**Files:** Modify `scripts/eval`, `plugin/evals/FACTS.md`.
**verify:** The probe-round case can read sibling skill files with the Read tool, so its result no longer depends on which tool the agent picks; three runs of `scripts/eval --case probe-round` are all green.
- [x] Failing test: `scripts/eval --case probe-round` with the branch binary on PATH; a run where the agent uses Read on probe.md shows the denial in the trace.
- [x] Code: `--allow-tools Bash Read` in scripts/eval; one dated FACTS.md line saying why.
- [x] Commit: `eval: grant Read so cases can open sibling skill files`.

### Task 3: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [ ] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [ ] Code: add 1 to the patch in all three files.
- [ ] Commit: `plugin: bump patch version`.
