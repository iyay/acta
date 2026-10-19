---
created: "2026-10-19 09:51:37"
id: SPC-0157
hash: yrl9k97
started: "2026-10-19 09:55:08"
finished: "2026-10-19 10:17:12"
---
# Evidence-driven debugging

Bounded. The user approved the design and written spec in chat on 2026-10-19, provided the skill reaches the quality of the four references.

## Goal

Make `acta:debug` accurate, thorough and efficient. Keep root-cause discipline, make evidence gates concrete, and scale investigation effort to uncertainty. Quality parity is an acceptance target to test, not a claim based on reading the references.

## Design

Use four phases with observable exit criteria:

1. **Symptom and feedback loop.** Record expected and actual behavior, environment, and one executed command with its result that catches the exact symptom. Read code as needed to build the loop. Prefer existing tests, then a small replay or harness. Missing reproduction means collect evidence and report the gap; do not claim a proven cause. Redact secrets in commands, output, artifacts and web queries.
2. **Narrow and distinguish.** Remove irrelevant inputs and steps while preserving failure. Compare working and failing paths. For ambiguous cases, rank competing hypotheses with predictions; test one variable at a time using the probe that best separates them. Clear cases need no fixed hypothesis count or exhaustive minimization.
3. **Prove the mechanism.** Connect trigger, state transition and symptom with source locations and probe results. Distinguish observed facts, hypotheses and unverified coverage. Test a shared premise when repeated attempts fail under it. Three failed fixes stop for reassessment; failed hypotheses alone do not prove architectural failure. Record confirmed bugs through `acta:bug`.
4. **Fix, verify and clean up.** Preserve acta approval, plan, worktree, TDD, review and land gates. Test the real bug pattern at a suitable seam, observe failure before the fix, then success. Re-run the original scenario and relevant regression checks. Remove temporary probes and report remaining limits. Missing test coverage is a gap, never evidence of a fix.

Keep the project checkout unchanged during diagnosis. Temporary harnesses and instrumented copies live outside it; explain their limits compared with the real environment. Do not change live systems or production instrumentation without authorization. Do not push.

Use short conditional recipes for flaky failures, races, order-dependent tests and performance. Record failure counts and conditions for flakes; control interleavings and assert invariants for races; find state polluters for ordering failures; compare repeated measurements under matched workloads for performance. Reruns until green, larger timeouts and skipped assertions do not prove a fix.

Rewrite the main skill and conflicting supporting guidance in plain English. Replace blanket validation at every layer with validation at trust boundaries and necessary invariants. Remove unsafe environment-output examples, unsupported statistics and repeated prose. Reuse existing tooling; add no general framework, dependency or mandatory agent fan-out.

## Scope and verification

Change `plugin/skills/debug/`, focused plugin checks and behavioral evals, plus required embed, budget or release metadata. Keep line and byte budgets honest; splitting files must not hide total growth. Leave unrelated workflows unchanged.

Before rewriting instructions, run behavioral baselines. Compare the candidate with current acta and the relevant reference on each shared scenario: a clear bug, a misleading nearby failure, missing reproduction access, a flaky or race failure, an order-dependent failure, and a performance regression. Use isolated contexts, equivalent tools and fixtures, pinned reference versions, and a shared rubric. Repeat noisy comparisons at least three times and inspect disagreements; do not weaken graders to pass.

Accept only when the candidate catches the exact symptom, avoids unsupported causal claims and unauthorized edits, preserves meaningful regression checks, and shows no material correctness or safety loss against the relevant reference. Record commands, outcomes, tool calls and token usage. Seek lower cost than current acta without quality loss; explain any cost increase. Report unavailable comparisons explicitly and do not claim parity without evidence. Run required repository checks before landing.

## References

- [pstack bug fix](https://github.com/cursor/plugins/blob/main/pstack/skills/poteto-mode/playbooks/bug-fix.md): runtime evidence and hypothesis elimination; its runtime and performance playbooks supply focused measurement techniques.
- `diagnosing-bugs`, supplied by the user: exact-symptom feedback loop, minimization, falsifiable predictions, regression seam and cleanup.
- [superpowers systematic debugging](https://github.com/obra/superpowers/blob/main/skills/systematic-debugging/SKILL.md): backward tracing, working-path comparison and single-variable probes.
- [gstack investigate](https://github.com/garrytan/gstack/blob/main/investigate/SKILL.md): conditional diagnostic recipes, scope discipline and fresh verification.
