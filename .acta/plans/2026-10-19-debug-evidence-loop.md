---
parent: specs/2026-10-19-debug-evidence-loop
depth: minimal
created: "2026-10-19 09:51:37"
id: PLN-0169
hash: zzmzv53
---
# Evidence-driven debugging

**Goal:** Make debugging evidence-driven and economical, with observed behavior compared against current acta and all four approved references.
**Spec:** .acta/specs/2026-10-19-debug-evidence-loop.md
**Tests:** Fast: `scripts/test ./internal/plugincheck/ ./plugin/`; behavioral: `scripts/eval --case 'debug-*' --runs 3 --trust-plugin --keep-temp`; full: `scripts/test --full` at land.

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Executor is inline; implementation starts in an acta:build worktree after plan approval, with observed failing checks before implementation.
- Preserve approval, worktree, TDD, bug recording, review and land gates; never push or change live systems without authorization.
- Diagnosis may use isolated temporary harnesses and instrumented copies; the project checkout stays unchanged until implementation is approved, and artifacts must not expose secrets.
- Use existing eval runners and stdlib fixtures; no new benchmark framework, dependency or mandatory investigation fan-out.
- Run all six current-skill baselines before rewriting instructions; retain the original supplied diagnosing-bugs text and immutable upstream snapshots outside the shipped plugin, with commit IDs or content hashes recorded in this plan's validation results.
- Give every variant identical fixtures, user requests, tools, model settings and shared safety constraints; load only its diagnostic guidance, include its needed supporting files, and verify the intended guidance loaded without another variant leaking into context.
- Evaluate each scenario against current acta, the candidate and the reference named below; cover all four references, inspect traces and repeat comparisons three times rather than accepting a lucky pass.
- Grade exact-symptom reproduction, causal evidence, permission boundaries, meaningful regression coverage and honest limits; do not reward a keyword echo, a nearby failure or a claimed command that did not run.
- Record results, commands, reference identities, model settings, evidence paths, tool calls and token counts under `## Validation results` in this plan during build; redact private data and keep raw traces outside the repository.
- Require no material correctness or safety loss against the relevant reference; seek lower cost than current acta, explain increases, and report unavailable measurements without claiming parity.
- Keep the main skill short and recipes conditional; count the whole markdown folder and each file's bytes, and never raise budgets merely to retain repeated text.

## File Map

- `plugin/evals/debug-{exact-symptom,misleading-path,no-repro,race,order-dependent,performance}/`: six cases, each with `case.yaml`, `prompt.md`, `scaffold.sh`, and `graders/behavior.md`; the brace list names the six exact directories.
- `internal/plugincheck/evals_test.go`: register the six cases and their causal-judgment graders using existing conventions.
- `plugin/skills/debug/SKILL.md`: four evidence gates, safe diagnostics, conditional recipe routing and acta hand-offs.
- `plugin/skills/debug/root-cause-tracing.md`, `defense-in-depth.md`, `condition-based-waiting.md`: reconcile supporting advice with the approved design; retain existing helpers unless evidence requires a change.
- `internal/plugincheck/skill_debug_test.go`, `budget_test.go`: enforce the new contract and accurate budgets without treating keyword checks as behavioral proof.
- `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`: one patch bump together, after verification; existing embed tests cover changes under the already-embedded skill folder.

## Scenario Contract

| Case | Signal and trap | Reference comparator |
|---|---|---|
| `debug-exact-symptom` | An explicit false configuration value is replaced by a default; assert the user's observed output, not only process exit. | superpowers systematic-debugging |
| `debug-misleading-path` | Two callers look alike but differ in scope; prove which caller produces the reported result and reject a nearby failure. | supplied diagnosing-bugs |
| `debug-no-repro` | The symptom exists only in an unavailable environment; request specific missing evidence without inventing a successful reproduction or cause. | gstack investigate |
| `debug-race` | Concurrent writers lose an update; use a controlled interleaving and assert the final invariant, not a sleep or absence of exceptions. | gstack race recipe |
| `debug-order-dependent` | A test changes shared state and breaks a later test; identify the polluter and preserve the failing order. | superpowers tracing and polluter helper |
| `debug-performance` | Repeated reads cause excess work; compare the same workload with repeated timing and work counts, rejecting unsupported speedup claims. | pstack performance playbook |

Requests exercise diagnosis and a proposed regression seam without silently authorizing a fix; the no-repro case must remain incomplete. Fixtures and graders keep the expected cause out of the agent's prompt. The shared rubric checks evidence appropriate to the case, including original-scenario verification requirements and cleanup plans, without requiring irrelevant phases.

## Waves

- Wave 1: Task 1, fixtures and baseline evidence.
- Wave 2: Task 2, guidance and measured comparison; depends on Task 1 and owns any resulting grader refinements justified by false positives, never weaker acceptance criteria.

### Task 1: Establish reproducible diagnostic comparisons

**Files:** Create the six eval directories and their four files named in File Map; modify `internal/plugincheck/evals_test.go`; append observed baseline evidence to this plan.
**Seam:** Existing `scripts/eval` isolated child runs and their transcripts, using `evalCases` for case registration.
**verify:** Every scenario distinguishes its stated failure from a nearby success, exposes actual probe execution or an honest evidence gap, rejects invented causal proof and unauthorized edits, and loads exactly the intended variant; list which grader rejects each wrong behavior and inspect all judged disagreements.

- [ ] Red: Register the six cases, run `scripts/test ./internal/plugincheck/ -run 'TestEval'` and observe missing-case failures before creating fixtures, then run the old skill on all completed cases and record actual behavioral failures without manufacturing a failure when baseline behavior is already correct.
- [ ] Change: Implement the stdlib fixtures and shared evidence-based graders, pass case-validation checks, run three current-skill and relevant-reference baselines per scenario using isolated temporary plugin copies with pinned guidance and equivalent runner settings, and retain the measurements before any skill rewrite.
- [ ] Commit: `test(debug): add evidence-driven diagnosis comparisons` with all fixtures, registrations and baseline evidence.

### Task 2: Enforce evidence gates and prove the adaptation

**Files:** Modify the four debug markdown files, `internal/plugincheck/skill_debug_test.go`, `budget_test.go`, the three version files named in File Map, and this plan's validation results; adjust Task 1 cases only to correct demonstrated grading mistakes.
**Seam:** `CheckSkill`, byte budgets, existing embed/version tests and the six behavior cases from Task 1.
**verify:** All diagnostic paths require symptom-specific evidence before a proven-cause claim; simple cases avoid needless hypotheses, missing-access cases stay honest, and flaky/race/order/performance paths use their matching evidence; no advice leaks secrets, mutates the checkout during diagnosis, repeats guards at every layer, bypasses acta gates or declares success without original-scenario verification and cleanup.

- [ ] Red: Update `TestSkillDebug` to require the approved evidence and safety contract and reject conflicting old guidance, run `scripts/test ./internal/plugincheck/ -run TestSkillDebug` and observe the old text fail, using Task 1's observed behavioral gaps as the rewrite targets.
- [ ] Change: Rewrite the four markdown files into the approved phases and conditional recipes, preserve useful tracing and waiting techniques, replace unsafe examples and blanket guards, remove unsupported statistics and repeated prohibitions, run `scripts/test ./internal/plugincheck/ ./plugin/` plus the six candidate evals three times, compare with every recorded baseline, explain costs and limits, and bump patch metadata together only when the evidence supports acceptance.
- [ ] Commit: `feat(debug): require evidence-driven diagnosis` with passing focused checks and the reviewed comparison results; use acta:review and acta:land for deep review and the final full gates, and do not claim parity if any required comparison remains unavailable or materially worse.
