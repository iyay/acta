---
parent: specs/2026-10-20-build-without-baseline-run-design
depth: minimal
id: PLN-0179
created: "2026-10-20 05:31:04"
hash: blphkrk
started: "2026-10-20 05:33:12"
finished: "2026-10-20 05:42:39"
---
# Start builds without a baseline test run

**Goal:** Start task implementation after worktree setup and save baseline evidence from full suites already run at landing.
**Spec:** `.acta/specs/2026-10-20-build-without-baseline-run-design.md`
**Tests:** `scripts/test` (fast, close only), `scripts/test --full` (full, land only); task loop: `scripts/test ./internal/plugincheck/ -run '<task test selection>'`; behavior gate: `scripts/eval` at landing, with focused cases during implementation.

## Global Constraints

- Follow acta:lean: understand the task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- No baseline test before the first task, even when a saved result is absent, stale, unreadable, or from another test context; keep task TDD, focused checks, close checks, review, and full landing gates.
- The baseline is historical evidence only; preserve existing merge and tidy gate-reuse rules and never relabel reused results as a fresh run.
- Keep output from the original run and read it for summaries; never rerun tests only to recover output.
- Store the local baseline and logs in the absolute Git common directory, shared across worktrees and outside tracked files; use no new command, dependency, runner integration, or cache engine.
- Inline executor in a worktree; write repository content in English and never push.

## File map and waves

- Build startup: `plugin/skills/build/SKILL.md`, its plugin checks, and build evals.
- Landing evidence: `plugin/skills/land/SKILL.md`, new `plugin/skills/land/baseline.md`, its plugin checks, landing evals, and the landing runbook.
- Wave 1: Task 1. Wave 2: Task 2; both register evals in `internal/plugincheck/evals_test.go`, so execute sequentially in this single plan.

### Task 1: Start the first task without baseline tests

**Files:** Modify `plugin/skills/build/SKILL.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/evals_test.go`, and `plugin/evals/build-inline-flow/` (scaffold, prompt, graders); create `plugin/evals/build-stale-baseline/` (case, scaffold, prompt, graders); update `.acta/wiki/skill-text-fits-everyone.md` with the startup rule.
**Seam:** The build skill's setup-to-task transition, `CheckSkill`, and sandboxed build eval command traces.
**verify:** Every approved build starts its first task after setup without baseline tests or a baseline-approval pause, regardless of record availability or validity; enumerate fresh/resumed, inline/delegated, absent/stale/unreadable/context-mismatch paths, keep task regression tests and close checks, and prove absent/stale behavior in eval traces without instructing the agent to skip tests in the prompt.
- [x] Red: Add startup-rule regression checks and eval coverage for absent and stale saved results, with no setup test calls before the first task starts; observe `scripts/test ./internal/plugincheck/ -run 'Test(SkillBuild|Build.*Baseline|Eval)'` fail against the current build skill.
- [x] Change: Replace the clean-baseline run with direct task startup and optional reading of saved evidence, keep task/close checks, finish the two build evals using existing deterministic graders, update their registrations and the wiki, and pass the focused checks and build evals.
- [x] Commit: `perf(build): start tasks without baseline test runs`.

### Task 2: Save passing full-suite evidence and reuse its output

**Files:** Modify `plugin/skills/land/SKILL.md`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/evals_test.go`, and `.acta/wiki/land-gates-then-merge.md`; create `plugin/skills/land/baseline.md`, `plugin/evals/land-baseline-pass/`, and `plugin/evals/land-baseline-fail/` (each with case, scaffold, prompt, deterministic graders).
**Seam:** The common full-suite landing precondition before ordinary/tidy landing, its baseline record/log artifacts, `CheckSkill`, and sandboxed landing eval command traces.
**verify:** Every actual full-suite landing run keeps its complete output and true command exit status; only PASS atomically replaces `acta-baseline.md` with tested commit/tree, command, time, relevant runtime/dependency context, exit status, summary, and an absolute retained-log path; failed/interrupted runs and write errors preserve the previous PASS and its log, retain failure output, and never permit a false PASS or merge; linked worktrees share evidence, paths with spaces work, logs are unique per run, a summary never reruns the suite, and gate reuse preserves the original tested identity without publishing a fresh result.
- [x] Red: Require the landing evidence policy and its shared placement before ordinary/tidy paths, register passing/failing evals with counted test commands and a pre-existing PASS for the failure case, and observe `scripts/test ./internal/plugincheck/ -run 'Test(SkillLand|Land.*Baseline|Eval)'` fail against the current landing skill.
- [x] Change: Add concise shared instructions and `baseline.md` for capturing original output/status without pipeline masking, recording PASS by atomic rename with an immutable log, and preserving prior evidence on failure; finish artifact/trace graders and the runbook, preserve all existing gates/reuse rules, and pass focused checks and landing evals.
- [x] Commit: `feat(land): save passing suite baselines and logs`.

## Fix round 1

### Task 3: Fix review findings

**verify:** the baseline policy covers every full-suite run, the eval graders fail for the right reason, and nothing else changes.

- [x] Scope the baseline policy to step 1, step-6 and tidy-tip runs, naming the tested checkout.
- [x] Grade the failure log path from the report, any merge form, the full old record and all record fields.
- [x] Use seven-character hashes in the stale eval and drop the unused requirements file.
- [x] Keep the tidy-history fold fallback and drop the double run-one wrap.
- [x] Commit: `fix(land): close the review findings`

## Fix round 2

### Task 4: Fix round 2 findings

**verify:** the eval matchers see the JSON-encoded command, the registry phrase has a live carrier, and nothing else changes.

- [x] Anchor the negative suite matchers on the JSON command value and keep the positive matcher plain.
- [x] Point the pass registry at a phrase a surviving grader carries.
- [x] Require an absolute failure log path and allow the run-one wrapper in the record fields.
- [x] Record the input_match regex semantics in FACTS.md.
- [x] Commit: `fix(evals): read the JSON tool input`

## Polish

### Task 5: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] Anchor the absolute failure log path so a relative path with a slash fails.
- [x] Let the negative suite matcher cross an escaped quote or newline in the JSON command.
- [x] Commit: `polish: review notes for PLN-0179`

## Close

Run the existing close checks and one deep review over the complete branch, then the normal full suite, behavior evals, and wiki gate before landing; no extra test run belongs before Task 1. The new skill behavior applies after this change lands, so the executor must still follow the currently loaded build skill for this implementation.

## State

### Next

Close: wiki check, fast tests, deep review, then land.

### Findings

## Review notes

- Round 1, deep, two reviewers: BLOCKED. The baseline policy covered step 1 only, while the step-6 and tidy reruns also run the full suite.
- Fix round 1: that blocker, plus the eval matcher, hash, record and tidy-page notes.
- Round 2, two reviewers: BLOCKED. The suite matchers could not see JSON-encoded Bash input, and the deleted pass grader carried the eval registry phrase.
- Fix round 2: both blockers, plus the absolute log path, the run-one record form and the FACTS semantics.
- Round 3, two reviewers: CLEAN.
- Polish, light tier, markdown graders only: CLEAN.
- Notes kept: the merge matcher misses `-m` before the branch, which is outside the tested flow; the red log is proven by the report path, since the omp file lister skips `.git`; the step-1 `acta run-one -- <full command>` wording is pre-existing and now caveated.
