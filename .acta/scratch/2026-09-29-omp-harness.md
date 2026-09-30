---
id: SCR-0015
hash: mgfekh8
title: omp-harness
status: brainstorming
created: "2026-09-29"
started: "2026-09-30"
finished: "2026-09-30"
---
# omp harness: evals and brainstorm block

Split out of SCRATCH-6 on 2026-09-29. omp does not run Claude Code shell hooks, so the SCRATCH-6 session state, reminder and second-brainstorm block do not work in omp. The claude plugin eval suite also covers Claude Code only.

Idea: give plugin/omp/index.ts a handler for the omp tool-call event (check first that the event exists), and add an omp eval run for the same scenarios.

- Q1 scope: user picked b, one spec for A (omp extension hooks) + B (omp eval runner), two plans.

- Q2 cases: user picked a, omp runner reads the same plugin/evals/* cases; Claude-only cases get a skip marker (for example harness: claude in frontmatter).

- Q3 runner: user picked a, Go subcommand acta eval-omp in internal/, reuses plugincheck frontmatter parsing, scripts/eval --omp calls it.

- Q4 judge: user ruled no model pin for omp; case runs and llm grader judge both use omp's own default model (omp -p, no --model).

- Q5 sandbox: user picked a, reuse user omp auth, run with --no-extensions --no-skills plus --plugin-dir plugin and -e plugin/omp/index.ts in a temp dir with --no-session; first plan task checks acta skills still load under --no-skills and records it in plugin/omp/FACTS.md.

- Approach A: user picked 1, extension sends Claude-shaped JSON on stdin to acta hook pre-tool/post-tool/prompt with node:child_process spawnSync; no Go change.

- Section 1 approved (plan A, omp extension): only plugin/omp/index.ts + index.test.ts change. Run = spawnSync("acta", args, {input, cwd: sessionManager.getCwd(), timeout 10s}). payload() builds {session_id, tool_input:{command}}. createState has contextFor(sessionId) (prompt hook now gets stdin so the reminder works), onToolCall (exit 2 -> {block:true, reason: stderr}), onToolResult (post-tool, output dropped). tool_call and tool_result wired for toolName bash only. Any failure stays quiet. Tests with a fake Run; one real two-brainstorm check recorded in plugin/omp/FACTS.md.

- Section 2 approved (plan B, acta eval-omp): new package internal/evalomp (case.go, run.go, grade.go, report.go). Each case runs in a temp git dir with its scaffold, then omp -p --mode json --no-session --no-extensions --no-skills --plugin-dir plugin -e plugin/omp/index.ts, capped by timeout_seconds. Graders file_exists, regex, tool_used (Bash maps to bash), llm (omp -p default model, PASS/FAIL). max_turns and allowed_tools are ignored (omp has no flags); harness: claude skips a case (second-brainstorm-choices). First task probes the json event shape and --no-skills. Tests use a fake omp on PATH.
