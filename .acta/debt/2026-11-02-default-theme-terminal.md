---
id: DBT-0152
hash: l9yvpym
parent: plans/2026-11-02-default-theme-terminal
---
# Review NOTEs: Default theme terminal implementation plan

- [ ] (medium) internal/tui/view_test.go:1589,1597 and internal/tui/model_test.go:1512 pin the accent as tokyo-night 256-color 111, but the models now default to the terminal theme, whose accent is ANSI slot 12, so the check can never trip; pin those models with .WithTheme("tokyo-night", true) like the sibling tests.
- [ ] (medium) plugin/skills/setup/SKILL.md:15: a fresh run asks every setting `config show` lists, so the new always-printed `theme:` line makes setup ask about the theme, which the spec puts out of scope and the skill defines no question for; one bad theme answer also aborts the batch `config set`. Exempt theme from that first-run clause.
