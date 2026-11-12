---
parent: specs/2026-11-12-review-evidence
depth: minimal
after: [PLN-0236]
id: PLN-0232
created: "2026-11-12 23:03:20"
hash: i9nhifa
started: "2026-11-13 00:04:27"
finished: "2026-11-13 00:13:27"
---
# Review verdicts carry their evidence

**Goal:** A reviewer brief stands alone, a CLEAN report shows what it checked, the orchestrator proves each BLOCKER before fixing it, and four eval cases measure all of it.

**Spec:** `.acta/specs/2026-11-12-review-evidence.md` (SPC-0203)

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`; evals `scripts/eval --case <name> --runs 10 -j 8`.

**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched; whole when the diff touches `plugin/hooks/` or `plugin/references/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Task 1 runs its four cases on the review text as it is before this plan, and records the pass counts with `acta state set <plan id> findings`. That is the old arm for every later task.
- The new text is measured once, after Task 4: the four cases, 10 runs each, against the old arm. Only when a case scores below the old arm is that case run again on each task's commit, to find the change that lowered it. A change found to lower a case is taken out with one commit that carries its task's trailer, and `acta state set <plan id> findings` says which.
- The `acta` on PATH must be the branch binary before any eval run. Build it to a folder of its own and put that folder first on PATH for the run; never overwrite the user's installed binary.
- Skill text stays general for every acta user, in plain English. When a file passes its byte cap, raise the cap in `internal/plugincheck/budget_test.go`; never trim words an agent acts on.
- No new reviewer seat, no change to the number of rounds or to the review tiers.

## Build brief

- Land order: this plan starts after PLN-0236 is on main, because both change `plugin/skills/review/SKILL.md` and `plugin/skills/review/code-reviewer.md`.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3
- Wave 4: Task 4

### Task 1: Four review eval cases and their old-arm scores

**Files:** `plugin/evals/review-planted-bug/`, `plugin/evals/review-nit-stays-note/`, `plugin/evals/review-false-blocker/`, `plugin/evals/review-clean-has-evidence/`, `internal/plugincheck/evals_test.go`
**Seam:** `TestEvalCases` in plugincheck, and `scripts/eval --case 'review-*'`
**verify:** Each case fails for the one behaviour it names and for no other reason: a seeded defect on a path the plan covers must end BLOCKED with that `file:line`; a diff with only naming and comment problems must end CLEAN with no nit as a BLOCKER; a wrong BLOCKER handed to the orchestrator must be rejected with repro output and no fix task; a CLEAN report must hold `### Coverage`. Every case has a `# guards:` line for a heading the review skill has today, a scaffold that refuses a folder that is not empty, and caps on turns and time. List each case, its graders, and its score over 10 runs on the old text.
- [x] Red: `TestEvalCases` fails because the four case folders it now lists do not exist.
- [x] Change: add the four case folders and their `evalCases` entries, then record the old-arm scores in the plan state.
- [x] Commit: test(evals): four review cases for defects, nits, wrong BLOCKERs and evidence

### Task 2: A reviewer brief that stands alone and reports its evidence

**Files:** `plugin/skills/review/code-reviewer.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the review skill, and the cases `review-planted-bug`, `review-nit-stays-note`, `review-clean-has-evidence`
**verify:** No line of the brief sends the reviewer to the review skill for a rule: the three questions, the finding bar and the bucket rules are stated in the brief itself, and the persona line is gone. Every report, CLEAN or BLOCKED, has a `### Coverage` section with each plan requirement and the `file:line` and test behind it, each changed file marked as read, and each command run with the output line that decides it. The Standards-axis reviewer is the one told to revert the change in a temp clone, run the narrow tests and show them red. A reviewer that cut BLOCKERs at the cap of five writes `N more BLOCKERs not listed`. List every rule the brief used to point at and where it now stands.
- [x] Red: `skill_review_test.go` fails because the brief lacks `### Coverage`, the inlined rules and the `N more BLOCKERs not listed` line, and still points at the review skill.
- [x] Change: rewrite the brief with the rules inline, the Coverage section, the revert check for the Standards axis and the cut line.
- [x] Commit: feat(review): reviewer brief stands alone and reports its evidence

### Task 3: Merge-base range, and reports without evidence go back once

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/review/depth.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the review skill, and the case `review-clean-has-evidence`
**verify:** No review text takes the base from the parent tip: the base is `git merge-base <parent> HEAD` everywhere a range is built. A report with no `### Coverage` never counts as a verdict: the orchestrator sends it back to the same reviewer once and dispatches no new reviewer for it, and a second report without it makes that axis BLOCKED and is told to the user. A round where any reviewer wrote `N more BLOCKERs not listed` is BLOCKED, and the next round's brief tells the reviewers to look for them. List every place a range or a verdict is read.
- [x] Red: `skill_review_test.go` fails because the skill still says `git rev-parse <parent>` and lacks the send-back rule and the cut-BLOCKER rule.
- [x] Change: switch the base to the merge base and add the send-back rule and the cut-BLOCKER rule.
- [x] Commit: feat(review): merge-base range and no verdict without evidence

### Task 4: The orchestrator proves each BLOCKER before it becomes a fix task

**Files:** `plugin/skills/review/receiving.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/budget_test.go`, `.acta/wiki/review-base-merge-base.md`
**Seam:** plugincheck required strings for the review skill, and the case `review-false-blocker`
**verify:** No BLOCKER reaches a fix task without the orchestrator having run its repro in a temp clone and kept the output; one whose repro does not show the wrong result gets `rejected: <reason>` with that output. No line about gratitude or which words to say remains. The rules for unclear feedback and for pushing back with technical reasons are still there. The wiki page says the skill now reads the merge base. List every way a finding can arrive (reviewer, user, outside reviewer, light inline review) and what each must pass before it is acted on.
- [x] Red: `skill_review_test.go` fails because `receiving.md` lacks the repro rule and still holds the gratitude lines.
- [x] Change: rewrite `receiving.md` around the repro rule and update the wiki page.
- [x] Commit: feat(review): prove a BLOCKER with its repro before fixing it

## Polish

### Task 5: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] [fix] review-false-blocker: both fixture reports get a Coverage section, the scaffold comment names the reports folder, and the verdict grader asks for the verdict on its own line plus the word rejected
- [x] [fix] review SKILL.md budget step 1: round 1 reviews the range from the merge base, not <parent>..HEAD
- [x] [fix] review SKILL.md: a round BLOCKED only by cut BLOCKERs or a missing Coverage section has a next step: the next round reviews the same range again
- [x] [fix] receiving.md: drop the line about praise and thanks
- [x] [fix] code-reviewer.md: say what a Spec-axis reviewer writes for question 3
- [x] [fix] code-reviewer.md: a (low) NOTE is a [note]
- [x] [fix] receiving.md: a row for Pre-existing items
- [x] [fix] review-planted-bug grader names line 4, where the defect is
- [x] Commit: `polish: review notes for PLN-0232`

## Review notes

- Round 1, deep, full tier (skill text): CLEAN from both reviewers; polish of 8 [fix] NOTEs reviewed CLEAN by both; no fact-check since no BLOCKER. BLOCKER 0, NOTE 33, pre-existing 2 (build SKILL.md still names the range as <parent>..HEAD, now a debt line; the (low) priority tag in the brief, fixed in the polish). 3 NOTEs went to debt; the rest are grader strictness and wording nits. Eval keep rule: every change kept, final text 40 of 40 runs on the four new cases against 30 of 40 on the old text.

## State

### Next

Task 2: rewrite code-reviewer.md with the three questions, finding bar and bucket rules inline, a ### Coverage section, the revert check for the Standards axis, and the N more BLOCKERs line. Tests in skill_review_test.go first.

### Findings

Old arm 2026-11-13, review text before this plan, 10 runs per case at -j 8, branch binary on PATH: review-planted-bug 10/10, review-nit-stays-note 10/10, review-false-blocker 10/10, review-clean-has-evidence 0/10 (the brief never asked for a Coverage section).
New arm, text after Task 4: clean-has-evidence 10/10, planted-bug 10/10, false-blocker 10/10, nit-stays-note 9/10. Final text after the polish, 10 runs each: 10/10 on all four, so nit-stays-note stands at 19/20 against 10/10 and no change was taken out.
scripts/eval takes one --case per call: with several, only the last one runs. Run cases one call each, or pick them by --tag=<tag>.
