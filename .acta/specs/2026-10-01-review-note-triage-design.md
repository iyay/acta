---
id: SPC-0061
created: "2026-10-01 15:09:40"
hash: lxzz03x
started: "2026-10-01 15:14:18"
finished: "2026-10-01 15:16:31"
---
Status: approved by the user on 2026-10-01 (Bounded).

# Sort review NOTEs into fix, debt and note

## Problem

Today every NOTE from every review round goes into `acta debt new`. The agent never chooses. A NOTE is also never a task, so even a one-line fix waits as debt. The debt triage on 2026-09-29 found 160 items: 115 were only tidy-ups and 12 were stale.

## Design

The agent sorts each NOTE into one of three buckets. A NOTE that fails a bucket's test falls to the next one.

1. **fix**: all of these hold: the file is already in the diff; the fix adds no new logic; the path is not security, auth, money, migration or delete; it fits in one small commit. Examples: a comment, a name, wording, a missing test, a one-line guard.
2. **debt**: leaving it has a real cost later that you can name ("when X changes, Y breaks", "a user hits it when Z", "the tests get slow"), and the fix needs new logic or a file outside the diff.
3. **note**: taste, nits, inputs the plan does not name, "could be more robust" with no real scenario.

**Who sorts.** Reviewers stay read-only. They tag each NOTE with a suggested bucket: `[fix]`, `[debt]` or `[note]`. The orchestrator decides. It checks each tag against the code, the same way it checks any finding.

**Flow.**
- A round with BLOCKERs: the `[fix]` NOTEs join that round's one fix task, so the next round reviews them.
- A CLEAN round: all `[fix]` NOTEs land as one polish commit. No new round. The agent reviews it itself: it reads the full diff and runs the full test suite with the output shown. When the polish fails or the tests go red, the agent reverts the commit and moves those items to `[debt]`.
- Before `acta:land`: the `[debt]` NOTEs go to `acta debt new` as today. The `[note]` NOTEs go into a `## Review notes` section in the plan file. With no `[debt]` NOTEs, no debt file is written.
- The three-round budget does not change. A round with no BLOCKER still does not start.

## Files

- `plugin/skills/review/SKILL.md`: "Where findings go" and "Budget". "NOTE is never a task" gets an exception for `[fix]`.
- `plugin/skills/review/code-reviewer.md`: the NOTE output format carries the bucket tag.
- `plugin/references/house-rules.md` line 8 and `plugin/skills/build/dispatch.md` line 262: "files them with acta debt new" becomes "sorts them".
- `internal/plugincheck/skill_review_test.go`, and the dispatch and house-rules checks if they pin the old text: new required phrases for the three buckets.

## Testing

- Red first: the plugincheck tests require the new phrases and fail on the current skill text.
- `scripts/test --full` green.
- Optional eval case: a review whose NOTEs do not all land in debt. It costs quota, so ask the user before running it.
