---
id: DBT-0128
hash: ekz2brg
parent: plans/2026-10-19-setup-skill-choices
---
# Review NOTEs: Setup skill choices plan

- [ ] (medium) setup skill sends a user to ! acta setup just to install a missing plugin, but the wizard has no install-only mode, so they answer every question again (internal/cli/setup_cmd.go)
- [ ] (medium) skill_setup_test only checks that "## Every path" exists; moving the override paragraph back under "## No AskUserQuestion" would still pass, so the round-1 BLOCKER has no guard
