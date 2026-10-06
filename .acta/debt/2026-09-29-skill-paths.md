---
id: DBT-0017
hash: uc0dwq7
parent: plans/2026-09-29-skill-paths
---
# Review NOTEs: Skill Paths Implementation Plan

- [x] plugin/skills/brainstorm/SKILL.md flow graph: Bounded joins the shared "User reviews spec?" node, whose "changes requested" edge goes back to "Write design doc" (the Architectural node) instead of "Write short spec"; the step list is right.
- [x] Dispatch dropped its "main checkout git status --porcelain empty and HEAD = parent" check; plugin/skills/land/SKILL.md step 4 has no clean-main or HEAD check.
- [x] Spec says merge --no-ff must not appear anywhere in plugin/skills/dispatch/, but the new MustNot strings only match text; herdr-delivery.md has none today. (stale)
- [x] TestDispatchLandsThroughLand goes red on revert only through "Bugs found by recipient"; acta:land and herdr pane close were already in the old Landing section, so the MustNot list in TestSkillDispatch is what guards the old merge steps. (stale)
- [ ] (low) plugin/skills/plan/SKILL.md:99-100 still describes the no-spec Bounded form "**Spec:** none (Bounded, approved in chat on <date>)"; mostly dead now that Bounded writes a spec.
- [x] TestSkillBrainstorm MaxLines 395 against about 281 lines leaves slack that cannot catch growth.
