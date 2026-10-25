---
parent: specs/2026-10-26-optional-cold-read-design
depth: minimal
id: PLN-0194
created: "2026-10-26 05:56:36"
hash: b75vpyv
started: "2026-10-26 05:58:13"
finished: "2026-10-26 06:08:57"
---
# Optional cold read implementation plan

**Goal:** Offer a single fresh-context spec review in shape before final spec approval and planning.
**Spec:** `.acta/specs/2026-10-26-optional-cold-read-design.md`
**Tests:** Focused `scripts/test ./internal/plugincheck -run '<task pattern>'`; targeted `scripts/eval --case <case>`; full `scripts/test --full` and the existing land gates.

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Build inline in a worktree only after plan approval. Keep the existing spec approval, plan approval, implementation review, and landing gates. Never push.
- Cold read applies to Bounded and Architectural shape flows, never Spike. Ask once per spec in the current flow; another pass requires an explicit user request.
- Dispatch one read-only reviewer without inherited conversation, chat summaries, prior conclusions, or a plan. Findings are advice; only agreed edits are applied. Missing or failed delegation requires a user decision to retry or continue without cold read.
- Keep the existing shape folder cap of 180 Markdown lines and existing file byte caps. Register the new companion file with a byte cap of 2400. Shorten existing prose only without losing requirements; do not increase existing caps.
- No new command, setting, dependency, report directory, bootstrap work, or automatic review loop. Chat follows the configured language; repo content stays English.

## File map

- `plugin/skills/shape/SKILL.md`: offer placement and the two shape checklists.
- `plugin/skills/shape/cold-read.md`: dispatch, reviewer brief, triage, and failure handling, loaded only after acceptance.
- `internal/plugincheck/skill_shape_test.go` and `internal/plugincheck/budget_test.go`: skill contract and size checks.
- `plugin/evals/cold-read-{offer,skip,accept,unavailable}/`: one `case.yaml`, `prompt.md`, `scaffold.sh`, and `graders/behavior.md` per case.
- `internal/plugincheck/evals_test.go`: register the behavior cases using the existing eval registry and grader rules.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2, after Task 1.

### Task 1: Add the optional cold read to shape

**Seam:** `CheckSkill` and focused shape contract tests over the shipped skill files.
**verify:** Both applicable checklists offer cold read after spec self-review and before final approval; skipping preserves approval, acceptance alone loads the companion, dispatch excludes conversation history and plans, at most five findings include location/consequence/clarification, only agreed edits are committed, and failure never silently falls back or advances. Preserve Spike and all existing shape requirements. List the paths checked.
- [x] Red: Add focused `TestShapeColdRead` contract tests in `internal/plugincheck/skill_shape_test.go` and watch `scripts/test ./internal/plugincheck -run 'TestShapeColdRead'` fail because the offer and companion are absent.
- [x] Change: Update `plugin/skills/shape/SKILL.md`, create `plugin/skills/shape/cold-read.md`, and register its 2400-byte cap in `internal/plugincheck/budget_test.go`; run `scripts/test ./internal/plugincheck -run 'TestSkillShape|TestShape|TestBudgetFiles'` to prove the new contract and existing shape requirements still pass within the unchanged folder cap.
- [x] Commit: `feat(shape): offer optional cold read before planning`

### Task 2: Verify cold read decisions in greenfield evals

**Seam:** Existing behavior eval runner, native delegation traces, and `TestEvalCases` registry checks.
**verify:** Offer case stops for the user's choice; skip case reaches final spec approval without dispatch or companion loading; accept case dispatches exactly one fresh-context reviewer, checks the spec references, returns useful located findings, and waits before applying unapproved edits or planning; unavailable case reports incomplete review and waits for retry or skip. No case writes implementation code or a plan. Check dispatch inputs and available child trace for isolation; tool count alone is insufficient evidence. Report any trace visibility limit.
- [x] Red: Add the four registry entries in `internal/plugincheck/evals_test.go` and watch `scripts/test ./internal/plugincheck -run 'TestEvalCases'` fail on missing cases, then add the mapped case files and demonstrate that the offer/accept behavior fails against the pre-change skill in an isolated baseline checkout.
- [x] Change: Complete the four case fixtures and graders in the mapped eval folders using existing empty-directory and temporary-home guards, allow native delegation only in the accept case, use deterministic evidence wherever possible and register any necessary judgment grader explicitly, then run `scripts/test ./internal/plugincheck -run 'TestEval|TestScaffolds'` and `scripts/eval --case cold-read-offer`, `scripts/eval --case cold-read-skip`, `scripts/eval --case cold-read-accept`, and `scripts/eval --case cold-read-unavailable` against the candidate with native trace inspection for the accept case.
- [x] Commit: `test(shape): verify optional cold read decisions`

## Polish

### Task 3: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] [fix] Tag the four cold-read eval prompts claude-only because omp's LLM grader cannot inspect their required trace evidence or native Agent dispatch.
- [x] Commit: `polish: review notes for PLN-0194`

## Review notes

- Deep full-tier round 1 CLEAN on Spec and Standards axes; no BLOCKERs. Polish marked four trace-based cold-read cases claude-only and both polish reviewers returned CLEAN. Native Claude offer, skip, accept, and unavailable cases passed; native child trace confirms one reviewer read the spec and contract without parent context. Codex instructions were inspected but not executed.
- Integration validation: initial full native suite was 38/43; four unrelated failed cases passed focused baseline and candidate rechecks, and BUG-0054 was repaired separately in PLN-0195. Final combined tree passed scripts/test --full plus 4/4 Cold read and 6/6 build evals. A single full native-suite run was not wholly green; final evidence uses the full run and targeted rechecks. Native behavior coverage is Claude only.

## State

### Next

### Findings
