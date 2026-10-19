---
id: DBT-0134
hash: wef8lth
parent: plans/2026-10-19-drop-settings-go
---
# Review NOTEs: Drop repo language and subagent models: Go side

- [ ] (medium) CLAUDE.md:37 still lists repo language and subagent models as user settings; an agent reading it looks for settings that no longer exist.
- [ ] (medium) Three CLI test cases (cli_test.go --subagent-models all, config_repo_test.go --repo --subagent-models split and --repo-language with --unset) now pass only because the flag is gone, so they no longer test the personal-flag refusal or set-and-unset rule their comments name.
