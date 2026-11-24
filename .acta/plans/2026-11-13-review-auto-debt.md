---
parent: debt/2026-11-13-review-depth-auto
depth: minimal
closes: [DBT-0166.01, DBT-0166.02]
after: [PLN-0238]
id: PLN-0239
created: "2026-11-13 06:13:09"
hash: lsr8dl1
started: "2026-11-13 19:54:38"
finished: "2026-11-14 05:43:57"
---
# Review evals count reviewer roles

**Goal:** Review-depth evals distinguish reviewer identities and retries from helper calls and fix the polish case depth.

**Spec:** `.acta/specs/2026-11-13-workflow-debt.md` (SPC-0208)
**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched; whole when the diff touches `plugin/hooks/` or `plugin/references/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Build inline, serially, in a worktree under `~/.wwt/acta/`; never push.
- Keep source text generic across repos and harnesses; adjust byte caps only where needed, without cutting safety instructions.
- Validate changed graders with passing and failing samples before scoring a skill; use `evalomp.LoadCases` and `evalomp.Grade` in plugincheck for fixed graders, and measure judge rubrics on planted traces with the existing eval runner.
- For every skill-text task, run each named case once per call with `scripts/eval --case <name> --runs 10 -j 8` on old and new text under identical corrected graders and scaffolds; use an isolated copy for the old arm and the branch binary on PATH, never replace the installed binary.
- Record raw results with `acta state set`; isolate and exclude any regressing text change and keep its debt open, removing that item from `closes:` before land.
- Full suites run only at land; focused checks run after each change. Run scaffolds and mutation checks with fresh HOME and TMPDIR outside the repo.

## Build brief

- Land order: start only after PLN-0238 lands on main; these serial plans share plugincheck and eval files.

## Waves

- Wave 1: Task 1

### Task 1: Make reviewer-count evidence match the chosen depth

**Files:** `plugin/evals/review-auto-safe-one-reviewer/{prompt.md,scaffold.sh,graders/one-reviewer.md}`, `plugin/evals/polish-full-review/{prompt.md,scaffold.sh,graders/reviewer-dispatched.md,graders/standards-reviewer-dispatched.md}`, `internal/plugincheck/workflow_debt_evals_test.go`, `internal/plugincheck/evals_test.go`
**Seam:** existing review eval grader interface, with trace-focused judge rubrics where distinct reviewer identity requires judgment.
**verify:** The safe auto case explicitly records an inline build and accepts one combined reviewer, including a repeat request to that same reviewer for Coverage; a helper alone and two distinct reviewers fail. The standard polish case explicitly sets its depth and requires separate Spec and Standards reviewers; a combined brief containing both words cannot pass. Evaluate positive and negative traces for identities, resumed calls and helper roles, and retain literal dispatch checks only where they prove an actual role.
- [x] Red: add `TestWorkflowDebtReviewerRoles` samples and judge-trace controls, then run `scripts/test ./internal/plugincheck -run 'TestWorkflowDebtReviewerRoles|TestEvalCases|TestEvalLLM'` to expose the current raw-call maximum and incidental-word checks.
- [x] Change: declare the executor and standard polish depth in their scaffolds/prompts, grade distinct reviewer roles with existing grader types, update the narrow judge allowlist and samples, then run both named cases for 10 runs each and keep the corrected results separate from prior skill-effect claims.
- [x] Commit: test(review): count distinct reviewer roles and allow Coverage retries

## Review notes

- SPC-0212 (specs/2026-11-14-reviewer-role-evidence.md, o00a18d) supersedes Task 1's trace-focused LLM seam and judge-trace controls. Its written spec was approved by WhatsApp reaction 3A67321BCC55EF1303C8 on question 3EB08E4114DC08803C3245 at 2026-11-14 05:37. This revision awaits its own plan approval before implementation. Task 1 and Wave 1 remain the only task and wave; all original constraints and gates still apply.
- Files: retain every Task 1 path; additionally create plugin/evals/review-auto-safe-one-reviewer/graders/combined-role.md and plugin/evals/polish-full-review/graders/reviewer-count.md; restore the existing polish standards-reviewer-dispatched.md. No product skill, hook, CLI or dependency changes.
- Seam: evalomp.LoadCases and evalomp.Grade with actual Agent tool-input JSON and the loaded tool_used graders. Replace the superseded LLM-only wiring test and remove only the two newly added mixed-judge allowlist entries.
- verify: only fresh Agent prompts starting with the truthful reviewer-axis declaration count. The safe role graders jointly require exactly one BOTH reviewer; polish role graders jointly require exactly one SPEC and one STANDARDS and exactly two reviewers total. Helpers, incidental or quoted role text, and nonempty resume IDs never add a reviewer. Empty/null resume remains a fresh launch. Missing axes, single-axis safe reviewers, a combined-only polish reviewer, extra fresh launches, and an identity resumed as the other axis fail; correct roles with helpers or resumes pass. Enumerate these paths through the loaded graders.
- Red: replace TestWorkflowDebtReviewerRoles with positive and negative structured Agent-input samples against all role graders for each case, then watch scripts/test ./internal/plugincheck -run 'TestWorkflowDebtReviewerRoles|TestEvalCases|TestEvalLLM' fail on the current superseded graders.
- Change: keep the explicit inline-build safe fixture and standard-depth polish fixture; require truthful ACTA_REVIEW_AXIS=BOTH/SPEC/STANDARDS prefixes in actual reviewer prompts; implement fresh-launch axis/count checks using existing tool_used graders, with no judge or runner changes. Run the focused tests, then scripts/test ./internal/plugincheck. Run scripts/eval --case review-auto-safe-one-reviewer --runs 10 -j 8 --keep-temp and scripts/eval --case polish-full-review --runs 10 -j 8 --keep-temp as separate calls using the matching branch binary. Keep corrected results separate from earlier invalid judge evidence; claim no skill-effect improvement.
- Commit: test(review): count distinct reviewer roles and allow Coverage retries, with Task: PLN-lsr8dl1#1. Full suite remains at land, followed by the configured review/landing process; never push. No second task or plan split is needed for this single eval-evidence change.
- Review: auto selected standard (full tier: agent instructions); Round 1 CLEAN; BLOCKER 0, NOTE 0, pre-existing 0. Both axes completed Coverage over e04836e224479e7c7fe1929a087fe4b9c6c4be03..870fc31b20410343b43ba1c9db8f56b5c5650b4f. Standards independently passed focused checks0.039s and proved retained real grader samples RED with original graders restored in a private clone with fresh HOME/TMPDIR. Spec focused0.034s PASS. No polish task. Worktree list contains only main and this build; no unexpected commits. New wiki lesson inspected separately; original wiki restored byte-for-byte. Native corrected safe10/10 and polish10/10 first-run, both RUN_EXIT0, no flakes.

## State

### Next

Both final round1 reviewers CLEAN0blockers0notes0preexisting with complete Coverage overe04836e..870fc31; no polish. Next acta:land full suite with complete output/true rc and PASS baseline, then gates selection (only evals/plugincheck changed, named case20runs already green; plan slow gate not triggered), tidy and final ancestor/plan checks, cleanup. Enduring missing-evidence lesson recorded in eval-reviewer-launch-evidence wiki; original raw evidence retained separately.
After land close DBT0166.01/.02, report7/14 original items/3of5files, then continue approved PLN0240 followed0241 serially. No push/attachments.

### Findings

Task1 new method RED: after correcting nil-judge test setup, real loaded grader checks rejected superseded LLM/missing fixed graders and valid samples. GREEN focused0.051s, plugincheck14.418s; gofmt/bash-n/diffcheck clean. Sixteen structured call samples jointly cover the five actual fixed role/count graders, helpers, quoted markers, combined/separate axes, extra/missing reviewers, resumes before/after prompt, empty/null resume and SendMessage.
Native corrected method safe10/10 first-run, 139s tmp.lQHRvgRYkV; polish10/10 first-run,142s tmp.YaeJ2XjGRS. Both RUN_EXIT=0. Logs/binary /Users/iyay/.wwt/acta/reviewer-roles-fixed-20261114/{logs,bin}. Agent tool checks saw exactly1BOTH safe and exactly1SPEC+1STANDARDS/exactly2total polish. These are corrected grader/fixture results, no skill-effect improvement claim. No skill/hooks/runtime changes.
Previous rejected method and raw counterexamples remain retained: short16/16, safe10/10 tmp.IOtofzfdjv, polish7/10 tmp.tPqYcUNze1, launch evidence absent from lossy focus even on PASS; unsupported hook/tool_calls-focus probes separate. Do not count those as reliable role proof or ordinary flakes.

### Open rulings

SPC0212 written spec approved by reaction3A67321BCC55EF1303C8 on3EB08E4114DC08803C3245 at05:37; revised PLN0239 approved by reaction3A1278C6D6C2EB5A7512 on3EB0C8DAB2E9BE10F4542E at2026-11-14 05:38. Resume Task1 inline with no pending design/plan approval. Original serial-plan approvals persist. No push/attachments.
