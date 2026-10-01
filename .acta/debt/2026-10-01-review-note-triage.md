---
id: DBT-0066
hash: yt8k8ri
parent: plans/2026-10-01-review-note-triage
---
# Review NOTEs: Review NOTE Triage Implementation Plan

- [ ] (medium) plugincheck has no cross-file check: a new polish exception written outside the review skill's `## After a CLEAN round` stays green.
- [ ] `plugin/skills/build/SKILL.md` Close still says "Only the fix round is sent differently"; a polish is now sent to the same tab too.
- [ ] The bucket rules in `## Where findings go` (`plugin/skills/review/SKILL.md`) are no longer pinned by any plugincheck test, so they can drift silently.
- [ ] `plugin/skills/build/dispatch.md` step 10 pointer to `## After a CLEAN round` is not pinned; the test only forbids a land there.
- [ ] A polish reply-back runs the full suite then jumps to step 3; a red suite on that path has no revert instruction (it sits in step 2). The `acta:land` gate still stops the merge.
