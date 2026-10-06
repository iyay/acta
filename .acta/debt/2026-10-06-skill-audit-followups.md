---
id: DBT-0085
hash: dtnl0df
parent: plans/2026-10-06-skill-audit-followups
---
# Review NOTEs: skill audit follow-ups Implementation Plan

- [ ] (low) debug/SKILL.md:112 and root-cause-tracing.md:68 are clipped fragments ('CI-only probes: ask user to run; never push.', 'If stuck, use a throwaway clone outside the repo:'); full sentences need byte-cap room, and the line must keep 'outside the repo'.
- [ ] (low) No Must pin holds 'outside the repo' in the debug folder, so a rewording that drops it passes the tests.
- [ ] (low) implementer-prompt.md: no pin stops the old '**ask questions**' line coming back; add MustNot 'ask questions'.
- [ ] (low) MustNot pins match exact old strings only; a slight rewording of removed text passes.
- [ ] (low) build/dispatch.md:29 still gives a fixed English line for the report's last line to the user.
- [ ] (medium) debug/root-cause-tracing.md:97 runs ./find-polluter.sh without saying where; run in the real checkout it writes files there, against SKILL.md:48.
