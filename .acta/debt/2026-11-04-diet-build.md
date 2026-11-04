---
id: DBT-0154
hash: avzxcx9
parent: plans/2026-11-04-diet-build
---
# Review NOTEs: Diet: build skill implementation plan

- [ ] (medium) The build pins in TestSkillBuild check the whole skill folder, so a later edit that moves a shared body rule into a mode file such as dispatch.md stays green while subagent and inline runs lose it. Pin shared rules against SKILL.md alone.
