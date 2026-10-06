---
id: DBT-0084
hash: pcnorso
parent: plans/2026-10-06-skill-prompt-audit-fixes
---
# Review NOTEs: skill prompt audit fixes Implementation Plan

- [ ] debug/root-cause-tracing.md:66-84 and defense-in-depth.md:72-82 still tell the agent to add instrumentation to the real code, against debug SKILL.md 'Phases 1 to 3 change no file'.
- [ ] debug/SKILL.md:80,93-110: the CI/codesign example can often run only in real CI, where a clone outside the repo cannot; say whether a pushed probe branch counts as outside the checkout.
