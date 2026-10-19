---
type: Runbook
title: Debug a red eval case skill-last
description: Check the case shape, then main, then PATH binary, before calling a red case a skill gap
paths: [plugin/evals/, plugin/evals-routing/]
---

1. Read the case: empty `allowed_tools` means the child cannot load the skill it tests; a prompt that says "answer and stop" gets no tool calls; too-low `max_turns` looks like failure.
2. Run it on main: a case red there is not caused by the diff.
3. Run with the branch binary first on PATH: plugin hooks call the installed acta, not the branch code.
4. Treat one LLM-judged red run as noise until repeated.
5. Check the scaffold plan does not name the fix. A Change step that says "make slug drop the spaces" sends every run to slug, so the old skill fails too and the red proves nothing.
6. Read the kept trace (`--keep-temp`, `out/trace.jsonl`): list the tool calls and check whether the run read the file the skill points to. A judge stores votes, not reasons. In debug-race every red run skipped `race.md`; text-only rewrites of the skill did not change that.
7. Never loosen a grader to go green. Never fix a symlink under the home folder that the sandbox cannot read without asking; it blocks Bash cases.
