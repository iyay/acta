---
parent: scratch/2026-11-23-fewer-tests
id: SPC-0218
created: "2026-11-23 20:28:12"
hash: ldgt96k
---
# Fewer, leaner tests

## Why

Repos that use acta grow tests faster than code. Counts from 2026-11-23:
- acta: 2046 Go test funcs, 67k test lines against 29k code lines. Last 30 days added 9.7k test lines and 5k code lines.
- walkie: 85k test lines against 47k code lines.
- clever-platform: test files went from 73 to 144 in 4 days, 1604 test cases. 45 test files read source text with `readFileSync` or `Bun.file`, 13 of them `src/index.ts`. `tests/deploy-guards.test.ts` is 1117 lines with 120 `toContain` checks, one test per edge case.

What in acta pushes this way:
- `acta:tdd` says "Write one minimal test" for every change, and agents read that as a new test func each time.
- `acta:debug` phase 4 turns every repro into a regression test, and `acta:review` fix rounds start with a red test, so each bug and each finding adds a test.
- `writing-good-tests.md` already forbids text tests ("Behavior, not text", lines 46-50, and the warning sign "The test greps source text"). Nothing enforces it: the review finding bar makes every test finding a NOTE, and the tdd RED step never points to the rule.

The user asked for this on 2026-11-23 and picked C (test rules) and D (test audit).

## Design

Rule for every skill edit below: replace text where the new rule takes over an old one, and add only what is new. Caps go up only by what the plan measures it needs (see Caps).

### Red test order

`plugin/skills/tdd/SKILL.md`, RED step: the sentence "Write one minimal test: one behavior, a clear name, real code (no mocks unless unavoidable)." becomes this order:
1. A new row in an existing table test, `test.each` or subtest list that checks the same unit.
2. Else, a new assert in the nearest existing test of the same unit.
3. Else, a new test func, only for a behaviour no test checks yet.

The row, assert or func still checks one behaviour with real code, and it must be seen failing before the code changes. The same step points to "Behavior, not text" in `writing-good-tests.md`.

`plugin/skills/slice/SKILL.md`: a task names the existing test and the row or assert it adds. A task may name a new test func only when steps 1 and 2 do not fit, with a one-line reason in the task.

`plugin/skills/debug/SKILL.md` phase 4 step 1: the regression test follows the same order.

### Text tests

No new rule. "Behavior, not text" in `writing-good-tests.md` stays as written and is the one rule. It covers code, scripts, skills and config: a test that asserts a file contains an exact line or phrase proves nothing. Skill files are proven by evals, as the rule says.

This also binds the acta repo: a plan that adds a `Must` phrase to a skill test in `internal/plugincheck` adds a text test. Structure checks there (line and byte caps, frontmatter, eval layout) assert no text, so they are not text tests.

### Review

`plugin/skills/review/SKILL.md` (finding bar, line 29) and `plugin/skills/review/code-reviewer.md` (Finding Bar, line 76) both get the same exception, next to the BLOCKER definition:
- A test the diff adds that asserts a code, script, skill or config file contains an exact line or phrase is a BLOCKER. Its evidence is the test's `file:line` and the path it reads, in place of the concrete input the other BLOCKERs need.
- It is the one test finding that is a BLOCKER. "Findings in test tooling are NOTEs" and the "Always a NOTE" list say so.

A new test func where a table, `test.each` or nearby test of the same unit would fit is a NOTE.

### Test audit in cleanup

`acta:cleanup` gets a second mode in a new file `plugin/skills/cleanup/tests.md`.

The cleanup description gets the trigger. New text, within a raised cap: `acta: Use when the user asks to clean up, tidy or check stale planning items (scratch, bugs, debt lines, plans or specs that are done, dropped or no longer valid), or to audit tests for text tests and duplicates. Proposes each change with evidence and writes only what the user approves.`

SKILL.md gets one line before step 1: when the user asks to audit tests, skip steps 1 to 7 and follow `tests.md`. The rule "Cleanup never deletes a file" stays: removing a test is a code change, so it goes through a plan.

`tests.md`:
1. Scan, read-only: text tests, as "Behavior, not text" defines them, and near duplicates, meaning tests in one file or one unit whose names or asserts are almost the same. A big repo gets one read-only subagent per test folder.
2. Report: at most 10 items per message. Each item has `file:line`, the kind, the evidence and a proposal: delete, fold into a table, or rewrite as a behaviour test.
3. The items the user approves go into one debt item with `acta debt new`, one line per test. Nothing is deleted here. The work then goes through `acta:slice` and `acta:build` like any debt.

### Caps

`internal/plugincheck/budget_test.go` and the skill tests get these changes, each by the bytes or lines the edit needs after replacing old text:
- byte caps (`fileCaps`): `skills/tdd/SKILL.md`, `skills/slice/SKILL.md`, `skills/debug/SKILL.md`, `skills/review/SKILL.md`, `skills/review/code-reviewer.md`, `skills/cleanup/SKILL.md`, and a new entry for `skills/cleanup/tests.md`;
- description cap (`descriptionCaps`): `skills/cleanup/SKILL.md`, for the text above;
- folder line caps (`MaxLines`): cleanup (60), debug (361), slice (266), review (482) where the edit needs it.

`writing-good-tests.md` does not change, so its cap stays.

## Out of scope

- A coverage scan in the audit. It needs one coverage run per test or per-test profiles, a different tool in each language, and hours on a big repo. It comes back as its own spec if the text and duplicate scan proves too weak.
- Deleting tests in any repo as part of this spec. The first audit runs on clever-platform after its history rewrite, as its own job.
- Removing the `Must` phrase tests acta already has in `internal/plugincheck`. The audit reports them like any text test.
- Planning commits, which are spec `2026-11-23-fewer-planning-commits-design.md`.
- A Go command for the audit.

## Verification

Skill text is proven with evals:
- New case `tdd-row-not-func` (tags `tdd`, `build`): a small repo with a table test and a bug, and a plan task to fix it. Pass when the fix adds a row to that table, the row is seen failing first, and no new test func is added.
- New case `slice-row-not-func` (tag `slice`): a spec for a change to a unit that has a table test. Pass when the plan's task names that test and a new row, and names no new test func.
- New case `debug-regression-row` (tag `debug`): a bug in a unit with a table test. Pass when the proposed regression test is a row in that table.
- Case `review-planted-bug` (tag `review`): the scaffold adds a test file that reads `src/discount.py` and asserts it contains a line. The prompt drops "The repo has no tests". A second grader passes only when that test's `file:line` is listed as a BLOCKER. The existing grader for `src/discount.py:4` stays.
- New case `cleanup-test-audit` (tag `cleanup`): a scaffold with one text test, one near duplicate pair and one sound test, and the prompt "audit the tests". Pass when the report names the first two, leaves the sound one out, runs no `acta cleanup`, deletes no file and asks before any `acta debt new`.

`internal/plugincheck` checks caps, frontmatter, and that each new eval case carries its skill tag: `scripts/test ./internal/plugincheck`. The diff touches `plugin/skills/`, so the eval gate runs `scripts/eval --tag=tdd`, `--tag=slice`, `--tag=debug`, `--tag=review` and `--tag=cleanup` at land.
