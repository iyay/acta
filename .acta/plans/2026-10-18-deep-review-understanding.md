---
parent: specs/2026-10-18-deep-review-understanding-design
depth: minimal
id: PLN-0158
created: "2026-10-18 19:24:42"
hash: jwjcbtm
started: "2026-10-18 19:31:53"
finished: "2026-10-18 19:52:37"
---
# Optional Review Understanding Plan

**Goal:** An independent, default-off understanding switch adds a short explanation and one optional behavior question after successful landing at any review depth.

**Spec:** `.acta/specs/2026-10-18-deep-review-understanding-design.md`

**Tests:** `scripts/test` (fast), `scripts/test --full` (full, at land only); task loops name their packages below; behavior: `scripts/eval --case 'review-understanding-*'`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Build is not authorized by this plan-writing turn; stop after saving and committing the plan, even though its depth is minimal.
- When build is requested, use an isolated worktree and the inline executor; begin with observed failing checks, review once at the close, and never push.
- Keep review, fact-check, fix, polish, gates and auto-land intact; add only the `review_understanding` setting, with no checkpoint state, provider, extra reviewer or question engine.
- `review_understanding` defaults to false and is independent of `review_depth`; read its effective value from config or session settings, never infer it from a review level, lens or one-run depth override.
- Read the applicable wiki before edits: `plugin-loads-from-repo`, `plugin-ships-in-binary`, `plugincheck-caps-per-folder`, `skill-text-fits-everyone`, `eval-sandbox-greenfield` and `test-locked-runs`.
- Write repo content in English; keep skill text plain and usable across users and harnesses, and adjust only the text budgets that need it.
- At land, run `scripts/test --full` and the required `scripts/eval` gate, plus the existing wiki and state checks; do not substitute the selected eval cases for the landing gate.
- For config/hook/setup edits, also read `codex-harness`, `block-write-keeps-symlinks` and `rebuild-binary-after-merge`; use a branch binary on the eval PATH so fixtures see the new config, and put rebuild/reload instructions in the landing report.

## Waves

- Wave 1: Task 1 (inline).
- Wave 2: Task 2 (inline, depends on Task 1).

### Task 1: Add the independent default-off config switch

**Files:** Modify `internal/config/user.go`, `internal/config/repo_user.go`, `internal/config/user_test.go`, `internal/config/repo_user_test.go`, `internal/cli/config_cmd.go`, `internal/cli/config_repo_test.go`, `internal/cli/help.go`, `internal/cli/help_test.go`, `internal/hook/hook.go`, `internal/hook/hook_test.go`, `internal/setup/form.go`, `internal/setup/form_test.go` and, only if needed for the session text, `internal/plugincheck/budget_test.go`.

**Seam:** `config.LoadUser`, `SaveUserFile`, `MergeRepo`, `SaveRepoUser`, `UnsetRepoUser`, `Overrides`; `acta config set/show`; `hook.SessionStart`; existing setup answer assembly. Represent absence separately from false (for example, a `*bool` in `config.User`); preserve the current setup questions and retain the configured value when setup saves its answers.

**verify:** Config alone owns switch semantics: missing means false, explicit booleans round-trip without losing false, an explicit repo value overrides either user value, unset restores inheritance, and choosing any review depth leaves the switch unchanged. Reject invalid values and types with `review_understanding` in the error and without writing either config; preserve unrelated YAML keys/comments. CLI wiring owns `--review-understanding true|false` for user and repo, typed JSON plus existing source/default marks in show, and help for set/unset. Session wiring owns the effective boolean in settings for every harness; setup wiring owns retaining the saved value without a new question. Cover the cases at their owning seams, with only one success and one propagated error at higher wiring layers; list the paths checked.

- [x] Red: Add focused `Test.*ReviewUnderstanding` tests at the listed seams and observe `scripts/test ./internal/config/ ./internal/cli/ ./internal/hook/ ./internal/setup/ -run 'Test.*ReviewUnderstanding'` fail because the switch and its config surfaces do not exist.
- [x] Change: Add the boolean with default false and repo precedence, validate it before writes, wire set/show/help and session settings, preserve it during setup, then pass the focused tests and `scripts/test ./internal/config/ ./internal/cli/ ./internal/hook/ ./internal/setup/ ./internal/plugincheck/`.
- [x] Commit: `feat(config): add an independent review understanding switch`.

### Task 2: Add the optional question to enabled landing reports

**Files:** Modify `plugin/skills/review/SKILL.md`, `plugin/skills/review/depth.md`, `plugin/skills/land/SKILL.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/skill_land_test.go`, and, only if needed, `internal/plugincheck/budget_test.go`; create the following eval files:

- `plugin/evals/review-understanding-report/case.yaml`, `plugin/evals/review-understanding-report/prompt.md`, `plugin/evals/review-understanding-report/scaffold.sh`, `plugin/evals/review-understanding-report/graders/behavior.md`.
- `plugin/evals/review-understanding-disabled/case.yaml`, `plugin/evals/review-understanding-disabled/prompt.md`, `plugin/evals/review-understanding-disabled/scaffold.sh`, `plugin/evals/review-understanding-disabled/graders/behavior.md`.
- `plugin/evals/review-understanding-followup/case.yaml`, `plugin/evals/review-understanding-followup/prompt.md`, `plugin/evals/review-understanding-followup/scaffold.sh`, `plugin/evals/review-understanding-followup/graders/behavior.md`.

**Seam:** Existing `acta:review` close and `acta:land` report, exercised by plugin contract checks and greenfield eval prompts and traces; follow the existing `review-depth-floor` and `polish-light-review` fixture and grader formats without touching the user's home or Git config.

**verify:** At every review depth, an enabled successful landing keeps its required evidence and report fields, explains the main behavior and its meaningful risk or boundary and test limits using inspected final landed code, and ends with exactly one optional outcome-and-reason question; no answer or question precedes or gates the merge. Contract checks own both switch values across quick, standard and deep, including a deep lens, plus blocked/failed landings, no pass/skip status or extra reviewer, and skip, silence or a new request never blocking or restarting completed landing. The report eval owns an enabled quick review and a final landed revision that differs from the earlier reviewed snapshot, including light polish; its trace proves landing completes before the question and its explanation matches final code. The disabled eval owns deep review with explicit false and no added explanation, question or offer. The follow-up eval owns enabled standard review and a short, differently worded prediction with a missed concept: explain the actual cause against source, resolve ambiguity without grading wording, and offer at most one optional further scenario. Config is wiring only here: one enabled value and one disabled value from Task 1, without repeating merge or validation cases. List the paths checked; inspect eval traces as well as grader results.

- [x] Red: Add focused contract checks and the three bounded greenfield cases, run `scripts/test ./internal/plugincheck/ -run 'Test(ReviewUnderstanding|LandUnderstanding)'` and `scripts/eval --case 'review-understanding-*'` before changing skills, and observe the missing opt-in report contract fail.
- [x] Change: Read the independent switch, put the enabled explanation, final optional question and follow-up rules in the existing skills, fit only affected budgets, then make the narrow checks, all plugin checks and selected evals green while retaining all existing landing requirements.
- [x] Commit: `feat(review): add an opt-in understanding question after land`.

## State

### Next

None. Both tasks landed as c1a8277 and 0fa83b1.

## Review notes

- Round 1, deep, full tier, two reviewers (Spec, Standards), no BLOCKER so no fact-check: CLEAN. NOTEs, all [note]: session line only when the key is set (missing reads as false in the skills); added wiki page eval-merge-needs-libexec beyond the spec's "no new page"; evals_test.go registry edit outside the Task 2 file list; CLI stricter than yaml on True/TRUE; setup override shows an unset user value as "false"; byte caps raised to exact size.
