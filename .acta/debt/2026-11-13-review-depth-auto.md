---
id: DBT-0166
hash: srt6bxz
parent: plans/2026-11-12-review-depth-auto
started: "2026-11-13 19:54:38"
---
# Review NOTEs: Review depth auto: the review a diff needs, picked by rule and reason

- [ ] (medium) review-auto-safe-one-reviewer counts Agent calls, not reviewers: a right run whose reviewer report lacks the Coverage section and is sent back makes two calls and fails the max of 1, and one call to any helper subagent passes. The prompt also does not say which executor built the plan, so a strict reader in doubt about dispatch work may send two reviewers.
- [ ] (medium) polish-full-review sets no review_depth, so under the new default it gets one reviewer where it used to get two; its graders match the words Spec and Standards in a brief, which a combined brief holds only by its Axis line. When that line changes the case goes red for no fault of the run.
