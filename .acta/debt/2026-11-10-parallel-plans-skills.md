---
id: DBT-0163
hash: qhzbbau
parent: plans/2026-11-10-parallel-plans-skills
---
# Review NOTEs: Parallel plans: slice ask and build many.md

- [ ] (medium) build/implementer-prompt.md line 3 says the template is only for the subagent executor, but many.md briefs from it too, so an agent reading both gets two rules.
- [ ] (medium) In a parallel run each subagent writes its own acta state set next after every task and can overwrite the orchestrator's parallel run marker; a session that dies then resumes that plan alone and leaves its siblings.
- [ ] (medium) many.md step 7 puts the review round in the state text, while build SKILL.md says state never holds the review round; it also says from that plan's worktree, which a waiting or landed plan does not have.
