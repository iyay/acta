---
id: DBT-0027
hash: b7mmh0z
parent: plans/2026-09-29-rule8-names-choices
---
# Review NOTEs: Rule 8 Names the Choices Implementation Plan

- [ ] hook_test.go herdr-on check matches "herdr tab" from the skill index, so it passes even without the herdr line; the CLI test covers it.
- [ ] hook_test.go "acta:brainstorm" check always passes via the skill index and rule 7; only the claude --bg and new-session checks test rule 8.
- [ ] hook_test.go slices out[strings.Index(out,"Core rules:"):] and panics instead of failing if the header changes.
- [ ] cli/hook_test.go t.Setenv("PM_VOICE_FILE","") does nothing; the test reads the real voice file.
- [ ] hook.go herdrExtra comment says a session outside herdr never hears the word, but the skill index line always says "herdr tab".
- [ ] Plan body (Steps 1 and 3, File Map) was not updated after the herdr design moved into the hook; ticks are done but text is stale.
- [ ] scripts/eval does not clear HERDR_ENV; a run from a herdr tab may get the herdr line and fail the grader.
- [ ] Eval pass rate was 10/12 before BUG-7 is fixed; three in a row is not proof.
- [ ] The acta binary on PATH was built from the branch before merge; live sessions ran unmerged hook text.
- [ ] rule 8 says acta scratch new even when the second brainstorm is an existing scratch item; can file a duplicate.
- [ ] session.go blockText says "file this one as a scratch item" although the blocked item already exists.
- [ ] session.go texts say "rule 8"; the user's CLAUDE.md has its own rule 8; "acta rule 8" is clearer.
- [ ] The phrase "Answer the user first, from this rule" was dropped from rule 8 in fix round 1.
- [ ] plugin/evals/second-brainstorm-choices/prompt.md description still says "the three choices, minus the herdr tab".
