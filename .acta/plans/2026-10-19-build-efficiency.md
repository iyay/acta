---
parent: specs/2026-10-19-build-efficiency-design
depth: minimal
id: PLN-0170
created: "2026-10-19 11:02:11"
hash: kqymmiy
started: "2026-10-19 11:04:23"
finished: "2026-10-19 11:13:46"
---
# Build efficiency implementation plan

**Goal:** Reduce build instructions and repeated work while preserving workflow gates and meaningful implementation tests.
**Spec:** `.acta/specs/2026-10-19-build-efficiency-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full` (land only); task checks: `scripts/test ./internal/plugincheck/`; behavioral cases: `scripts/eval --case <case-name>` with the branch binary on PATH and reports outside the repo.

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Build inline in an isolated worktree; preserve all three executor modes, precedence and fallbacks, approval/readiness gates, isolation, baseline, ticks, state, task trailers, wiki, review and land; never push.
- Read the files being changed and their matching wiki pages before editing; follow the repo's setup and package manager, and install dependencies only when required.
- Reuse existing tests/evals. A new assertion or case must protect distinct approved behavior or a concrete regression; no test-count, assertion-count, coverage-percentage or test-to-code ratio target is introduced.
- Tests observe public behavior, use small fixtures, and avoid duplicate lower-layer cases, private implementation details and mock wiring; RED must fail for the intended behavior.
- Keep tdd and lean rules in their existing skills, state in acta, and full-suite checks at land; do not add a framework, CLI feature, ledger or model setting.
- Record baseline and candidate instruction bytes/words across the whole build folder, check/tool counts from traces, and implementation outcomes under identical requests; report uncertainty and unavailable evidence honestly.
- Use deterministic artifact checks where possible and LLM grading only for judgment; run any scaffold, write-command probe or mutant with fresh temporary HOME and TMPDIR. Preserve required gates rather than weakening assertions to fit a rewrite.

## File map

- `plugin/skills/build/SKILL.md`: shared gates, executor routing, inline task loop, focused verification and close.
- `plugin/skills/build/implementer-prompt.md`: conditional worker handoff and meaningful test selection.
- `plugin/skills/build/dispatch.md`: existing dispatch-specific contract; change only references affected by the rewrite.
- `internal/plugincheck/skill_build_test.go`, `build_dispatch_test.go`, `budget_test.go`, `evals_test.go`: preserve contract and packaging checks while removing wording-only coupling affected by the rewrite.
- `plugin/evals/task-trailer/`, `state-resume/`, `lean-tests/`: reuse current cases for regressions; extend only the relevant case when a behavior gap is demonstrated.
- `plugin/evals/build-inline-flow/`: one new behavioral case for inline flow/setup gaps absent from existing cases; `case.yaml`, `prompt.md`, `scaffold.sh`, and artifact graders.
- `.acta/wiki/skill-text-fits-everyone.md`: update only if measured results change its guidance; other wiki edits follow the existing close check.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2, after Task 1; both own the build entrypoint and contract checks, so no parallel writes or separate plans.

### Task 1: Make execution direct and setup selective

**Files:** Modify `plugin/skills/build/SKILL.md`, `plugin/skills/build/implementer-prompt.md`, and affected references in `plugin/skills/build/dispatch.md`; update existing `internal/plugincheck/{skill_build_test,build_dispatch_test,budget_test,evals_test}.go` only where required; create `plugin/evals/build-inline-flow/{case.yaml,prompt.md,scaffold.sh}` and its minimal artifact graders, based on `task-trailer` isolation.
**Seam:** The build skill executing an approved plan in the isolated eval sandbox, plus existing plugin contract and budget checks.
**verify:** Inline completes the approved task itself without loading a worker handoff, spawning an implementer, reinstalling an already usable environment or losing required progress/commit evidence; other executor routes and all gates remain intact; recovery trusts recorded progress and does not redo committed work; list the routes checked and distinguish static contract evidence from runtime evidence.
- [x] Red: Capture baseline folder size and run existing `task-trailer` and `state-resume` cases, then run the new inline-flow case against the unchanged skill and inspect an intended behavioral failure; if it already passes, retain the coverage without claiming a failure and identify the actual unmet contract before editing.
- [x] Change: Rewrite shared flow and explicit inline execution with conditional worker detail, repo-directed setup and existing state recovery; pass `scripts/test ./internal/plugincheck/`, then rerun the failed inline scenario and existing cases, inspecting artifacts and traces rather than adding wording assertions.
- [x] Commit: `refactor(build): streamline executor flow and project setup`, with task trailer and measured size/check evidence recorded in acta state.

### Task 2: Make test selection lean and prove the result

**Files:** Modify `plugin/skills/build/SKILL.md` and `implementer-prompt.md`; revise affected existing checks in `internal/plugincheck/skill_build_test.go` and `budget_test.go`; extend only needed files within `plugin/evals/lean-tests/`; update `.acta/wiki/skill-text-fits-everyone.md` only if its facts change.
**Seam:** Build-generated implementation and test artifacts in the existing lean-tests eval; focused plugin checks protect workflow contracts.
**verify:** Each named plan behavior is tested at its owning boundary, upper-layer tests verify wiring only, and assertions prove essential results or contractual side effects without redundant internal checks; existing coverage requirements remain unchanged; compare identical baseline/candidate requests and reject reduced correctness or broken gates even when the text is shorter.
- [x] Red: Run `lean-tests` on the unchanged test-selection guidance, inspect its artifacts, and extend the same case only for a demonstrated gap in test reuse or meaningful assertions; observe the intended failure before changing that guidance, reusing passing coverage without inventing a failure or a duplicate grader.
- [x] Change: Add concise risk-based test selection and meaningful-RED guidance, remove repeated instruction text, preserve existing scope escalation, and pass `scripts/test ./internal/plugincheck/` plus the affected evals; compare final whole-folder size and traces with baseline, recording tradeoffs and untested contracts in acta state.
- [x] Commit: `refactor(build): keep tests focused on meaningful behavior`, with task trailer; proceed through the existing wiki, deep review and land gates only after all approved task checks pass.

## State

### Next

Close: wiki check, fast tests, deep review over main..HEAD.

### Findings

Build folder size: baseline 5456 words / 33053 bytes; final 4125 words / 24733 bytes (-24% words). dispatch.md and wiki.md unchanged.
Evals, one run each: build-inline-flow 1.00 before and after (no RED found; kept as coverage); task-trailer, state-resume, lean-tests 1.00 before and after.
Tool calls inline-flow: baseline 8 (test and code written together), candidate 12 then rerun; candidate watched RED separately. No runtime gain claimed.
lean-tests baseline had no gap, so the case was not extended. Subagent and dispatch routes checked statically only (plugincheck).

## Fix round 1

### Task 3: Fix round 1 review findings

**verify:** Worktree detection compares absolute paths, so a session in a subfolder of a normal checkout still creates a worktree; the subagent hand-off carries the `acta state set` step; inline moves into the new worktree and commits wave ticks; the inline-flow eval no longer hints the setup answer and drops its duplicate trailer grader.
- [x] Red: plugincheck asserts `--path-format=absolute` in SKILL.md and `acta state set` in implementer-prompt.md; watch it fail.
- [x] Change: SKILL.md, implementer-prompt.md, build-inline-flow prompt and graders.
- [x] Commit: `fix(build): review round 1 findings`
