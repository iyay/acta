---
created: "2026-10-08 08:56:02"
id: SPC-0111
hash: e26s8qp
started: "2026-10-08 08:58:19"
finished: "2026-10-08 09:01:47"
---
# Review depth follows what the diff changes

Status: Bounded, approved by the user in chat on 2026-10-08. It loosens the 2026-10-01 ruling that every polish commit gets the two reviewers.

Why: the review skill sends a change to the two reviewers unless it is one file of text or config, and it never treats a polish commit as small. So a polish that only rewords comments in three files still costs two full reviews. `256bda8 Polish detail markdown comments` changed 0 logic lines, 18 comment lines and 111 markdown lines, and still went through both reviewers. The user wants the main agent to pick the depth from what the diff changes.

Design:
- `## Small changes` in `plugin/skills/review/SKILL.md` becomes `## Review tiers`. It is the one place that says how deep a review goes; other lines point to it.
- Light tier: every changed line is a comment, docs or markdown no agent reads as instructions, log text, or a test name. Any number of files.
- Full tier: everything else. That includes any logic line, config that changes behavior, text an agent reads as instructions (skills, prompts, hook text), and any change on a security, auth, data migration, money or delete path. In doubt: full.
- Light review: the orchestrator reads the full diff, checks every line is in the light tier, runs the narrow tests and the formatter with the output shown, and gives CLEAN or BLOCKED. It writes one line in the plan's `## Review notes` naming the tier and why. The rule that work another agent wrote always takes the two reviewers does not apply to the light tier.
- The tiers apply to round 1 and to a polish commit. A fix round always takes the two reviewers.
- `## After a CLEAN round`: a polish commit gets the review its tier picks, in place of always the two reviewers. Everything else in that section stays: the polish uses no round, a BLOCKER reverts it and moves its items to `[debt]`, and its NOTEs sort into `[debt]` or `[note]` only.
- Skill text is plain and fits every repo: no repo paths, no user names.
- `internal/plugincheck/skill_review_test.go` swaps the required strings that name the old rule for the new ones, and refuses the old sentences ("Small means one file", "Every polish commit gets the two reviewers").
- `plugin/skills/build/dispatch.md` changes only if a line there still says the polish always gets the two reviewers.
- Two new eval cases, each listed in `evalCases`, because one run can only face one polish: `polish-light-review`, where a polish that only changes comments is reviewed inline with no reviewer subagent, and `polish-full-review`, where a polish with one logic line still gets the two reviewers.
- The last task adds 1 to the patch version in the three plugin files.
