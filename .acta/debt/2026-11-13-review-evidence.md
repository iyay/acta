---
id: DBT-0165
hash: pitbluq
parent: plans/2026-11-12-review-evidence
started: "2026-11-13 18:44:11"
---
# Review NOTEs: Review verdicts carry their evidence

- [x] (medium) Review SKILL.md: the next step for a round BLOCKED with nothing to fix has three gaps. In round 3 it points at a round 4 that two other lines forbid, and should say ask the user. "A round with no BLOCKER does not start" can be read to send a round with only a cut line to the CLEAN flow. And a round with a real BLOCKER plus a cut line, or plus an axis with no Coverage, takes the fix task, so round 2 reviews only the fix range and never sees the cut ones.
- [x] (medium) plugin/skills/build/SKILL.md hands review the range `<parent>..HEAD` while review now reads the base as the merge base; the two texts name the range in two ways, and the wiki page review-base-merge-base covers both folders.
- [x] (medium) The review eval graders check a stand-in for two rules: review-clean-has-evidence checks that the brief asks for Coverage, not that a report without it goes back; review-false-blocker is settled by reading one line, so nothing checks that a repro ran. Its verdict pattern also fails a right reply that writes REJECTED in capitals or puts CLEAN in bold.
