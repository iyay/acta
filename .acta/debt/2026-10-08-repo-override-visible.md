---
id: DBT-0099
hash: qeh5ttj
parent: plans/2026-10-08-repo-override-visible
---
# Review NOTEs: Repo overrides are visible and changeable

- [ ] (low) UnsetRepoUser in internal/config/repo_user.go drops blank lines around a moved comment, and drops the document head comment when every key is removed; only layout and one comment are lost, no key.
- [ ] (low) The setup wizard never asks coding_guide, so its override step compares the repo value against lean even when the user's own file says off, and shows "Yours is lean"; this is older than PLN-0118 (the wizard already saved lean over off).
- [ ] (low) The setup skill description no longer names subagent models; the 230-byte description cap forces a choice between that and the language, style, tone and CLAUDE.md or AGENTS.md triggers. A request only about subagent models now leans on the general "change setup or one setting" words. Either raise the cap on purpose or reword.
