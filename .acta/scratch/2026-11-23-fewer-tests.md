---
id: SCR-0099
hash: x4kut01
title: Fewer, leaner tests
status: brainstorming
created: "2026-11-23 20:20:00"
schema: "1"
started: "2026-11-23 20:20:04"
finished: "2026-11-23 20:28:12"
---
# Fewer, leaner tests

## Words

### 2026-11-23

User (2026-11-23): "too many tests". Seen in every repo that uses acta. Picked C+D.

C. Test rules in tdd, debug and review: a red test is a new row in an existing table or test.each first; one behaviour, one test; no test that only reads source text.
D. acta:cleanup gets a test audit mode: report tests that pin source text and tests that add no coverage. Report only; the user approves each delete.

Evidence: acta 2046 Go test funcs, 67k test lines vs 29k code; clever-platform test files 73 to 144 in 4 days, 1604 cases, 45 test files read source text (13 read src/index.ts), deploy-guards.test.ts 1117 lines with 120 toContain.

## Context

## Log

### 2026-11-23

Round 1 answers (2026-11-23), all as recommended:
Q1 ban tests that read code text to check code behaviour; allowed when the file checked is shipped text (skill, template, config).
Q2 red test order: row in an existing table or test.each, else an assert in the nearest test of the same unit, else a new test func only for an unchecked behaviour. Applies to tdd, debug regression tests and review fix rounds.
Q3 review: a source-text pin test is a BLOCKER; a new test func where a table fits is a NOTE.
Q4 audit: always pin tests and near duplicates; optional coverage check when the repo has a coverage tool. Report only, user approves each delete.
Q5 first audit run: clever-platform, after its history rewrite.

### 2026-11-23

Spec 2 section 1 (red test order in tdd/slice/debug, ban on code-text pin tests, review BLOCKER exception, eval tdd-row-not-func) approved 2026-11-23.

### 2026-11-23

Spec 2 section 2 (test audit mode in cleanup/tests.md, report only, approved items become one debt item, eval cleanup-test-audit) approved 2026-11-23. Writing spec 2.

### 2026-11-23

Cold read of SPC-0218 (2026-11-23): user took every recommendation. 1 name every cap change, prefer replacing text over adding; 2 no new pin rule, enforce the existing "Behavior, not text" rule in writing-good-tests.md, drop the skill/config exemption (changes Q1); 3 BLOCKER exception in review/SKILL.md and code-reviewer.md, review-planted-bug prompt and a second grader; 4 new cleanup description and cap, audit mode goes straight to tests.md; 5 one slice case and one debug case. Asked the reviewer for its cut finding.

### 2026-11-23

SPC-0218 rewritten after cold read (2026-11-23): finding 6 (coverage scan) moved out of scope; no new text rule, existing Behavior-not-text rule enforced by a review BLOCKER; caps listed; slice and debug evals added.

## Open questions
