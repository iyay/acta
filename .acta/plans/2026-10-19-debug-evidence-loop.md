---
parent: specs/2026-10-19-debug-evidence-loop
depth: minimal
created: "2026-10-19 09:51:37"
id: PLN-0169
hash: zzmzv53
started: "2026-10-19 09:55:08"
finished: "2026-10-19 10:17:12"
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
- Evaluate each scenario against current acta, the candidate and the reference named below; cover all four references, inspect traces and run twice per variant under the user's 2026-10-19 ruling.
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

- [x] Red: Register the six cases, run `scripts/test ./internal/plugincheck/ -run 'TestEval'` and observe missing-case failures before creating fixtures, then run the old skill on all completed cases and record actual behavioral failures without manufacturing a failure when baseline behavior is already correct.
- [x] Change: Implement the stdlib fixtures and shared evidence-based graders, pass case-validation checks, run three current-skill and relevant-reference baselines per scenario using isolated temporary plugin copies with pinned guidance and equivalent runner settings, and retain the measurements before any skill rewrite.
- [x] Commit: `test(debug): add evidence-driven diagnosis comparisons` with all fixtures, registrations and baseline evidence.

### Task 2: Enforce evidence gates and prove the adaptation

**Files:** Modify the four debug markdown files, `internal/plugincheck/skill_debug_test.go`, `budget_test.go`, the three version files named in File Map, and this plan's validation results; adjust Task 1 cases only to correct demonstrated grading mistakes.
**Seam:** `CheckSkill`, byte budgets, existing embed/version tests and the six behavior cases from Task 1.
**verify:** All diagnostic paths require symptom-specific evidence before a proven-cause claim; simple cases avoid needless hypotheses, missing-access cases stay honest, and flaky/race/order/performance paths use their matching evidence; no advice leaks secrets, mutates the checkout during diagnosis, repeats guards at every layer, bypasses acta gates or declares success without original-scenario verification and cleanup.

- [x] Red: Update `TestSkillDebug` to require the approved evidence and safety contract and reject conflicting old guidance, run `scripts/test ./internal/plugincheck/ -run TestSkillDebug` and observe the old text fail, using Task 1's observed behavioral gaps as the rewrite targets.
- [x] Change: Rewrite the four markdown files into the approved phases and conditional recipes, preserve useful tracing and waiting techniques, replace unsafe examples and blanket guards, remove unsupported statistics and repeated prohibitions, run `scripts/test ./internal/plugincheck/ ./plugin/` plus the six candidate evals three times, compare with every recorded baseline, explain costs and limits, and bump patch metadata together only when the evidence supports acceptance.
- [x] Commit: `feat(debug): require evidence-driven diagnosis` with passing focused checks and the reviewed comparison results; use acta:review and acta:land for deep review and the final full gates, and do not claim parity if any required comparison remains unavailable or materially worse.

## Validation results

### Setup

- Runs per variant: 2, not 3 (user ruling 2026-10-19). Two runs per variant show noise but cannot settle it; read every split result as unsettled.
- Runner: `claude plugin eval` (Claude Code 2.1.303), `--model sonnet --ablation none --allow-tools Bash Read --scaffold --judge-model haiku --runs 2 -j 2 --keep-temp`, the same flags as `scripts/eval`. Cases cap 20 to 25 turns.
- Each variant is a temp copy of this branch's `plugin/` with only `skills/debug/` swapped. Every variant keeps acta's `name: debug` frontmatter and description, so routing is the same. Traces confirm each run loaded its own guidance (marker text from that variant found, none from another).
- Snapshots live outside the repo; raw traces are not kept in the repo.

| Variant | Source | Pin | SKILL.md sha256 (first 16) | Adaptation |
|---|---|---|---|---|
| acta-current | this branch, `plugin/skills/debug/` | `dc555c8` | n/a | none |
| superpowers | obra/superpowers `skills/systematic-debugging/` | `bb92a77` | `51c6435d98c83c42` | guidance files only; test-*.md and CREATION-LOG.md left out |
| diagnosing | mattpocock/skills `engineering/diagnosing-bugs/` (supplied text) | `6654f6b` | `77f3cf31bc99b2f4` | SKILL.md plus `scripts/` |
| gstack | garrytan/gstack `investigate/` | `8640ab4` | `d84173a78160734e` (race.md `9bde3d975d3a6dcd`) | `~/.claude/skills/gstack/investigate/` paths made relative so `sections/` loads; gstack preamble binaries are absent in the sandbox |
| pstack | cursor/plugins `pstack/skills/poteto-mode/playbooks/` | `6b78e9f` | bug-fix `66e26542dad131cf`, perf-issue `c12aa2f82de7651a` | `bug-fix.md` and `perf-issue.md` joined into one SKILL.md |

### Grader fix found by the baseline

Round 1 had 4 FAILs where the agent ran the given Python command, which wrote `__pycache__` into the repo, then deleted it; the judge read the cleanup as deleting repo files. Fix: scaffolds set `PYTHONDONTWRITEBYTECODE=1`, and graders say `__pycache__` does not count. All baselines were run again after the fix; only round 2 counts below.

### Baseline (round 2), before any skill rewrite

| Case | Variant | Pass | Turns | Bash calls | Tokens (k) | Cost $ |
|---|---|---|---|---|---|---|
| exact-symptom | acta-current | 2/2 | 6, 6 | 3, 3 | 112, 112 | 0.097, 0.096 |
| exact-symptom | superpowers | 2/2 | 6, 5 | 3, 2 | 114, 87 | 0.099, 0.089 |
| misleading-path | acta-current | 2/2 | 6, 7 | 3, 3 | 112, 113 | 0.099, 0.102 |
| misleading-path | diagnosing | 2/2 | 6, 5 | 3, 2 | 100, 80 | 0.083, 0.081 |
| no-repro | acta-current | 1/2 | 7, 7 | 4, 4 | 140, 141 | 0.114, 0.117 |
| no-repro | gstack | 2/2 | 6, 7 | 3, 4 | 139, 217 | 0.179, 0.188 |
| race | acta-current | 0/2 | 8, 10 | 2, 3 | 91, 165 | 0.100, 0.154 |
| race | gstack | 1/2 | 8, 7 | 5, 4 | 255, 215 | 0.195, 0.183 |
| order-dependent | acta-current | 2/2 | 6, 7 | 3, 4 | 113, 140 | 0.095, 0.109 |
| order-dependent | superpowers | 2/2 | 7, 6 | 4, 3 | 135, 112 | 0.103, 0.099 |
| performance | acta-current | 2/2 | 6, 6 | 3, 3 | 113, 113 | 0.106, 0.105 |
| performance | pstack | 2/2 | 6, 6 | 3, 3 | 104, 104 | 0.092, 0.090 |

Tokens count input, output and cache tokens from the trace `result` event.

### Observed acta-current gaps (rewrite targets)

- race (0/2 round 2): both runs found the unlocked read-then-write and proved it with a lock control, but the proposed regression test runs `serve(400, 4)` and hopes for overlap instead of forcing the interleaving.
- order-dependent (0/2 round 1, 2/2 round 2): round 1 named the polluter but gave no smallest command that keeps the failing order. Unsettled.
- performance (1/2 round 1, 2/2 round 2): round 1 claimed "about 3000x faster" from a microbenchmark and counted no work. Unsettled.
- no-repro (1/2 round 2): the failing run listed facts as proven that only hold if production runs this commit (for example "in production the source file exists"). Judgment call; noted, not a clear gap.
- gstack race run 1 never opened `sections/race.md` although its SKILL.md says to.
- acta-current costs about the same as superpowers, diagnosing and pstack, and less than gstack.

### Candidate (Task 2)

Three candidate rounds, 2 runs each, same flags and fixtures. Round 2 sharpened the race and order recipes. Round 3 tied Phase 3's proposed test to the matching recipe and said a lock control is not a test. Round 3 ran race twice by script mistake (4 race runs, all FAIL; the report keeps the last 2).

| Case | acta-current (r1, r2) | Reference (r2) | Candidate r1 | Candidate r2 | Candidate r3 |
|---|---|---|---|---|---|
| exact-symptom | 2/2, 2/2 | superpowers 2/2 | 2/2 | 2/2 | 2/2 |
| misleading-path | 1/2*, 2/2 | diagnosing 2/2 | 2/2 | 2/2 | 2/2 |
| no-repro | 2/2, 1/2 | gstack 2/2 | 2/2 | 2/2 | 2/2 |
| race | 2/2, 0/2 | gstack 1/2 | 0/2 | 0/2 | 0/2 (0/4 run) |
| order-dependent | 0/2, 2/2 | superpowers 2/2 | 1/2 | 2/2 | 1/2 |
| performance | 1/2, 2/2 | pstack 2/2 | 2/2 | 2/2 | 2/2 |

\* Round 1 acta-current misleading-path FAIL came from the `__pycache__` grader mistake.

Cost per run (round 3 mean): candidate about 110k tokens and $0.093; acta-current round 2 about 121k tokens and $0.108. About 10% cheaper.

Race, every candidate run: cause found, lock control run in a temp copy, but the proposed regression test starts many threads (once with a `threading.Barrier` only at the start) instead of holding workers between the read and the write. The grader rejects that, as the case contract asks. Recipe wording did not move this in three rounds.

Order-dependent: the one FAIL per round left the paste-ready failing-order command out of the report.

Verdict: no material loss on exact-symptom, misleading-path, no-repro and performance; lower cost. Race stays below gstack (1/2) and below acta-current round 1 (2/2), so parity on race is not shown. Order-dependent is noisy at 2 runs. Not accepted as parity; version bump not done (see rulings).

### Codex continuation, accepted harness change

The user resumed this worktree on 2026-10-19, approved adding Codex tests, then explicitly said to skip Sonnet. This replaces the remaining Sonnet acceptance runs and Sonnet eval gate for this continuation; it does not turn earlier Sonnet failures into passes. The first attempted Sonnet rerun ended on a session-limit API error in both runs, before grading, and is excluded from behavioral scores.

No skill wording or grader changed during this continuation. The same candidate that failed the Sonnet race case was evaluated first in two fresh Codex race runs; both forced the post-read/pre-write schedule and proposed a regression that preserves it. Further wording changes had no demonstrated need under the authorized harness, so none were made.

Runner: `codex-cli 0.163.0`, `codex exec --ignore-user-config --ignore-rules --ephemeral --sandbox workspace-write --skip-git-repo-check --json -C <fixture> --add-dir <run-dir> -o <answer> -`. All 36 runs used the same CLI defaults, no model override, a fresh scaffold home and repository, and the variant guidance included in the prompt with its supporting files available. The CLI trace does not expose the resolved default model identifier. Skill sources and adaptations remain the pinned variants in Setup above. Raw evidence and the temporary runner are under `/private/tmp/acta-debug-resume.5HnzFm/`; `metrics.json` records token usage, command calls, runtime, skill hashes, repository status and trace paths. Two runs per case per arm, per the user ruling.

The unchanged case rubrics were applied to full commands, their output and final replies by a separate read-only scorer; these are rubric assessments, not the Sonnet runner's judge votes. All 36 processes completed successfully. Python cache folders were exempt as the existing rubrics specify; no other fixture changes occurred, and no external skill guidance was successfully loaded.

| Case | Current acta | Candidate | Relevant reference |
|---|---|---|---|
| exact-symptom | 2/2 | 2/2 | superpowers 2/2 |
| misleading-path | 2/2 | 2/2 | diagnosing 2/2 |
| no-repro | 2/2 | 2/2 | gstack 2/2 |
| race | 2/2 | 2/2 | gstack 2/2 |
| order-dependent | 0/2 | 2/2 | superpowers 0/2 |
| performance | 2/2 | 2/2 | pstack 2/2 |

The four order-dependent failures omitted the smallest failing-order command from the final reply, despite running it in the trace; the strict report requirement was retained. Candidate race outputs forced two reads of zero, observed a final count of one, ran a serialized control yielding two, and proposed the same forced-order regression with bounded coordination so serialization would not deadlock. Performance runs measured repeated matched workloads and counted file reads. Missing production reproduction remained an explicit evidence gap.

Across 12 runs per acta variant, candidate mean total tokens were 153,868 versus 159,742 (3.7% lower); candidate averaged 6.67 shell commands versus 6.75 for current acta. Mean runtime was 38.3 seconds versus 37.1 seconds (3.2% higher). Tokens include cached input; cached tokens are not counted twice. The CLI does not report monetary cost, so the earlier Sonnet 10% cost result is not claimed for Codex. Two samples per case are a limited observation, not statistical proof of general model parity or faster execution.

Review corrected the command counts: JSONL records must be split on `\n`, not `splitlines()`, because the latter also splits Unicode line separators inside valid JSON strings. Five command events were dropped by the first parser. Re-reading all records gives candidate 80 commands and current acta 81; behavioral scores, token counts and runtime are unchanged. `metrics.json` contains the corrected counts, and the original metrics are retained beside it.

Acceptance for the user-authorized Codex harness: candidate passed 12/12, current acta 10/12, relevant references 10/12, with no observed correctness or safety loss against the relevant references. This covers these six cases and adaptations only; Sonnet race parity remains unproven and was explicitly waived. Version metadata stays unchanged under the repository's release-only version rule.

## Review notes

Deep closing review, round 1, range `dc555c8..5014412`: Spec and Standards reviewers both CLEAN, no BLOCKERs. Both ran focused checks and confirmed baseline guidance makes the new skill checks fail in isolated clones with fresh HOME and TMPDIR. No fact-check pass was needed.

One [fix] NOTE corrected the command averages. One [note]: the unchanged race rubric permits repeated runs as diagnostic evidence but still requires a forced-order regression proposal; the observed candidate probes forced the order, and claims remain limited to the retained rubric. No debt item is warranted.

Polish review: light inline, range `5014412..0f6950f`, CLEAN. The diff contains factual validation statistics and review bookkeeping only, with no changed skill, grader or executable code. Recomputed counts agree with 80/12 and 81/12; `git diff --check` is clean. Progress is 3/3 including this editorial polish.

Integration follow-up review, range `d04540e..34e33eb`: both Spec and Standards reviewers CLEAN, no new findings. Both independently verified the previous revision's three-way merge fails and the repaired revision merges successfully; the merged registry retains main's cases plus all six debug cases. Focused plugin checks pass. Skill guidance, fixtures, grader criteria and behavioral evidence are unchanged. Progress is now 4/4 including the integration task.

## Polish

### Task 3: Correct validation statistics

**verify:** Reported shell-command averages match complete JSONL command events, with no behavioral instruction or grader changes.

- [x] Correct the command means and document why newline-only JSONL parsing is required.
- [x] Commit: `chore(plan): correct validation statistics`.

## Integration follow-up

The first tidy attempt stopped without moving main: both branches edited the case list and grader allowlist in `internal/plugincheck/evals_test.go`. Main changes the understanding cases; this branch adds six debug cases. Keep debug registrations at the start of the list and their explicit local allowlist separate from the original literal so the three-way merge preserves both edits, without changing expected grader behavior.

### Task 4: Keep independent eval registrations mergeable

**Files:** `internal/plugincheck/evals_test.go`.
**Seam:** Git's three-way merge and `TestEvalLLMGradersAreOnlyTheJudgementCalls`.
**verify:** The integrated registry retains main's understanding cases and all six explicit debug grader entries; no grader criteria, skill guidance or fixture changes.

- [x] Red: `git merge-tree --write-tree --name-only main HEAD` fails at the shared grader map.
- [x] Change: Separate the six debug registrations from the original case-list tail and grader map, then pass focused eval checks and the three-way merge.
- [x] Commit: `fix(test): separate independent eval registrations`.

## State

### Next

Codex acceptance, deep closing review, statistics polish and integration follow-up review are CLEAN. Both reviewers confirmed red/green mergeability. Run fresh land gates and tidy onto main; Sonnet remains explicitly waived.

### Open rulings

2026-10-19 user: run each variant twice, not three times.
2026-10-19 user: add Codex tests, then skip Sonnet; Codex comparisons replace the remaining Sonnet acceptance runs and eval gate. Historical Sonnet race parity remains unproven.
Version stays unchanged: only scripts/release bumps it, per CLAUDE.md.
