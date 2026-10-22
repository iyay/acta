---
parent: specs/2026-10-22-wiki-advisory-freshness-design
depth: minimal
id: PLN-0187
created: "2026-10-22 23:06:09"
hash: frbmu2s
started: "2026-10-22 23:08:42"
finished: "2026-10-22 23:18:25"
---
# Advisory wiki freshness

**Goal:** Make stale notes advisory while keeping wiki errors blocking and agent review meaningful.
**Spec:** .acta/specs/2026-10-22-wiki-advisory-freshness-design.md
**Tests:** fast `scripts/test ./internal/wiki ./internal/cli ./internal/plugincheck`; full `scripts/test --full`; required landing eval gate `scripts/eval` with the branch-built acta on PATH.

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Executor is inline; use an isolated worktree, failing tests before implementation, deep review at close, and tidy landing; never push.
- Keep ancestry, range selection, no-wiki behavior, `.checked`, optional `wiki ok`, output ordering, and one-line messages; add no setting, dependency, or history query.
- Before the first wiki edit, capture no-range check output and record full-page/code review evidence in this plan; page edits alone do not prove review.
- Do not edit or confirm the baseline-stale `plugin-loads-from-repo` and `plugin-ships-in-binary` just to clear warnings; release noise stays in SCR-0077.
- The sibling paths plan owns only its three page files; this plan must not change them or other pages' paths.

## File map and waves

- Task 1 owns wiki diagnostic severity, CLI status/output, help copy, and their package tests.
- Task 2 owns build/land/migrate instructions, the freshness decision page, existing plugin checks, and the existing wiki-close eval.
- Wave 1: Task 1. Wave 2: Task 2, consuming advisory checks from Task 1. Execute sequentially inline.

### Task 1: Distinguish stale warnings from check errors

**Files:** Modify `internal/wiki/check.go`, `internal/wiki/check_test.go`, `internal/cli/wiki.go`, `internal/cli/wiki_test.go`, and `internal/cli/help.go`; reuse existing help tests in `internal/cli/help_test.go` without adding copy-only tests.
**Seam:** `wiki.Check(repo, pages, rng)` produces `Problem` values; `cmdWiki` consumes them and supplies the CLI exit status.
**verify:** Only actual staleness carries warning severity; field/path errors, bad ranges with pages, and history failures remain errors, including when mixed on one page; existing ancestry, dirty-page, optional confirmation, range, and fixed git-call-count properties remain unchanged. At the CLI, load errors always fail, stale-only output succeeds with and without a range, mixed diagnostics fail while printing both, and no-wiki behavior stays unchanged; list each path checked. The stale message retains commit/page context, starts its message with `warning: stale:`, uses the spec's review instruction, and never tells the user to run `wiki ok`.
- [x] Red: Add severity assertions and stale-only/mixed CLI expectations, including history-read failure and load errors, then run `scripts/test ./internal/wiki ./internal/cli -run 'TestCheck|TestWiki'` and observe failure because all current diagnostics are fatal and stale copy mandates confirmation.
- [x] Change: Add a zero-default `Warning bool` to `wiki.Problem`, set it only where staleness is detected rather than parsing message text, preserve hard errors and existing formatting order, make `wikiCheck` fail only on errors while printing every diagnostic, replace stale instructions, update help summary/comments, and pass `scripts/test ./internal/wiki ./internal/cli -run 'TestCheck|TestWiki|TestHelp'`.
- [x] Commit: `feat(wiki): report stale pages as advisory warnings` with this task's trailer.

### Task 2: Align workflow instructions and prove advisory close behavior

**Files:** Modify `plugin/skills/build/SKILL.md`, `plugin/skills/build/wiki.md`, `plugin/skills/land/SKILL.md`, `plugin/skills/migrate/SKILL.md`, `.acta/wiki/wiki-freshness-from-git.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/skill_wikiseed_test.go`, `internal/plugincheck/evals_test.go`, and `plugin/evals/wiki-close/{scaffold.sh,prompt.md,graders/page-fixed.md}`; add `plugin/evals/wiki-close/graders/check-succeeded.md` inside that existing case.
**Seam:** Existing plugin skill checks and the wiki-close scaffold/graders; `scripts/eval` supplies the end-to-end agent behavior gate at land.
**verify:** Build, land, wiki-writing, migrate seeding, help from Task 1, and the freshness page all agree that errors block, warnings require relevant review, incorrect facts must be corrected, and unchanged notes need no edit or confirmation; list these six surfaces. Remove mandatory `wiki ok`/empty-output rules from checks without weakening unrelated gates. The extended existing eval includes one incorrect page and one still-correct stale page, requires correction plus a real successful check, and accepts either recording or skipping confirmation of the correct page. Evidence capture must preserve the real CLI output/status and overwrite the last-check record on failure too; replay success and failure so a stale passing record cannot hide a later error.
- [x] Red: Revise existing skill checks for advisory wording and add `TestWikiCloseCheckEvidence` in `internal/plugincheck/evals_test.go` exercising the existing scaffold in a temporary HOME/cwd, then run `scripts/test ./internal/plugincheck -run 'TestBuildCloseWritesWikiFirst|TestWikiRecheckHasNoTimestamp|TestBuildWikiFileHoldsTheRules|TestLandWikiGate|TestSkillWikiSeed|TestWikiCloseCheckEvidence'` and observe old workflow rules and missing real-check evidence fail.
- [x] Change: Update the four skill files and reviewed freshness page, retain `wiki ok` only as optional review recording, extend the current scaffold with a valid still-correct stale page and last-check output/exit capture, add a deterministic successful-check file grader accepting warning output and optional confirmation, update the existing page grader/prompt as needed, and pass `scripts/test ./internal/plugincheck` plus `bash -n plugin/evals/wiki-close/scaffold.sh` without adding an eval case or LLM grader.
- [x] Commit: `docs(wiki): make unchanged-note confirmation optional` with this task's trailer, then close through the existing deep review and required full test/eval gates.

## Close

Run the gates with the branch-built CLI so the installed old binary cannot enforce the former fatal-stale behavior. The required `scripts/eval` runs at land; development checks do not replace it. Report remaining warnings as warnings, and keep the page-review evidence with this plan.

## State

### Next

Land using the reviewed implementation and passing full suite. The user explicitly waived the known build-close-no-suite eval failure after seeing it also fail on main.

### Findings

The existing build-close-no-suite failure was moved to the plan's debt file before landing. No known failure was relabeled as a passing gate.

### Open rulings

User ruling: do not investigate or fix the existing eval failure; land this plan with that specific failure waived. The full eval remains recorded as exit 1 (32/33 passed).

## Verification evidence

Task 1 RED: TestWikiCheckRange showed stale-only exit 1 and mandatory wiki ok text. GREEN: focused wiki/CLI/help tests passed; go vet passed for both packages. Typed Warning is set only at actual staleness; CLI adds the warning label, preserving the raw stale prefix used by doctor test helpers.

Before the first page edit, branch-built no-range check exited 0 with six stale warnings: the three sibling path notes, the two release-stale notes from ccaffcf, and wiki-freshness-from-git from Task 1. No release-stale page was edited or confirmed. Read the entire freshness page and compared its ancestry, .checked, dirty-page and merge rules with wiki/check.go, wiki/checked.go and gitc/wikihistory.go. Existing rules remain; added advisory status, optional review ownership, persistent warnings and review before page edits.

## Review notes

Deep, full tier: Spec and Standards reviewers both CLEAN, no BLOCKERs; no fact-check pass needed. One note records that focused tests do not replace the pending full suite and required agent eval.

## Landing gate evidence

- Full suite passed, including notui CLI: /Users/iyay/Nayakatara/acta/.git/acta-baseline-advisory-20261022T162126Z-43470.log (23 ok package results, no failures).
- Scoped vet, full go vet, formatter and shell syntax checks passed.
- Focused wiki-close eval passed all three graders, score 1.00: /var/folders/32/w_l_8__17m37y41yzzct4llh0000gn/T/tmp.EikY2g8O69/report.html.
- Full eval log: /tmp/acta-wiki-advisory-eval.bNOHlE. Existing build-close-no-suite failed score 0.67; the same targeted case failed on main with a main-built CLI, report /var/folders/32/w_l_8__17m37y41yzzct4llh0000gn/T/tmp.22NBjR22jg/report.html. Retained branch replay /private/tmp/e-WntzR7/out/trace.jsonl shows a real scripts/test call, so this is not a phantom marker. Gate remains blocking until resolved.

Final full eval result: exit 1, 32/33 cases passed. Only build-close-no-suite failed (0.67); wiki-close and wiki-hint passed (1.00). Report: /var/folders/32/w_l_8__17m37y41yzzct4llh0000gn/T/tmp.7LF2EF7Kiv/report.html. No landing was attempted for this plan. The independent path plan landed at faaac5c with full gates reused because only planning-root files differed; its worktree and temporary branch were removed with user approval.
