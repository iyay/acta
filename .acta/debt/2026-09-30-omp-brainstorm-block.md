---
id: DBT-0047
hash: d0ycxdh
parent: plans/2026-09-30-omp-brainstorm-block
---
# omp brainstorm block review notes

- [ ] plugin/omp/FACTS.md: two blank lines before the "Brainstorm block in omp" heading; other sections use one.
- [ ] plugin/omp/FACTS.md: trailing "Temp dir left in place per plan ... T5" line is build bookkeeping, not a fact.
- [ ] Spec Plan A says index.ts and index.test.ts only, yet its point 7 and the plan name FACTS.md; spec wording mismatch only.
- [ ] No test proves omp sets isError on a blocked bash call; tool_result skip relies on the fake event field.
- [ ] before_agent_start still spawns acta hook prompt with an empty session id; ParseEvent rejects it, one wasted spawn.
- [ ] plugin/omp/index.test.ts has no newline at end of file.
- [ ] plugin/omp/index.ts: spawnSync blocks omp's event loop up to 10s per hook call (pre-tool, post-tool, prompt).
- [ ] plugin/omp/index.ts: a bash call with its own input.cwd in another repo still runs hooks in sessionManager.getCwd(); same limit as the Claude shell hooks.
- [ ] plugin/omp/index.ts: export type Run is one long line; rest of file wraps shorter.
- [ ] plugin/omp/index.ts: realRun sets no maxBuffer; default 1MB is far above hook output.
