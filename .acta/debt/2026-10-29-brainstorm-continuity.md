---
id: DBT-0145
hash: c3zelrb
parent: plans/2026-10-29-brainstorm-continuity
---
# Review NOTEs: Brainstorm continuity implementation plan

- [ ] (medium) plugin/hooks/hooks.json PostToolUse and plugin/omp/index.ts onToolResult still spawn the now no-op acta hook post-tool on every Bash call; drop the plugin wiring and the plugin_test.go:86 assertion, keep the CLI subcommand for old installs.
- [ ] (medium) Text outside PLN-0200 still describes the removed brainstorm block: plugin/omp/index.ts:4 and :211 comments, plugin/omp/index.test.ts:202 sample block message, plugin/README.md:47 "the brainstorm guard"; a later agent reads them as true.
- [ ] (medium) The new evals brainstorm-mixed-context and brainstorm-resume-topic each passed 2 of 3 runs on the branch binary (land run judges FAIL FAIL FAIL and FAIL PASS FAIL, rerun green); tighten the prompt or grader so a land gate does not go red on judge noise.
