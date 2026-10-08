---
type: Runbook
title: Debug a red eval case skill-last
description: Check the case shape, then main, then PATH binary, before calling a red case a skill gap
paths: [plugin/evals/, plugin/evals-routing/]
timestamp: 2026-10-08T02:25:38Z
---

1. Read the case: empty `allowed_tools` means the child cannot load the skill it tests; a prompt that says "answer and stop" gets no tool calls; too-low `max_turns` looks like failure.
2. Run it on main: a case red there is not caused by the diff.
3. Run with the branch binary first on PATH: plugin hooks call the installed acta, not the branch code.
4. Treat one LLM-judged red run as noise until repeated.
5. Never loosen a grader to go green. Never fix a symlink under the home folder that the sandbox cannot read without asking; it blocks Bash cases.
