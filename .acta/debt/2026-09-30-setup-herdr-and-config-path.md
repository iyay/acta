---
id: DBT-0052
hash: waan7t3
parent: plans/2026-09-30-setup-herdr-and-config-path
---
# Review NOTEs: Setup Herdr Check and Config Show Path Implementation Plan

- [ ] No test covers a broken config.yaml in ResolveUserFile; by reading it returns config.yaml, exists true, and the parse error.
- [ ] If ~/.acta cannot be searched, the rename fails and config show prints file: ~/.acta/voice.yaml (exists: false, old file; ...), an odd line with unchanged values.
- [ ] When UserPath fails (no HOME), ResolveUserFile returns an empty path; config show prints the error and never uses it.
- [ ] TestSkillSetup still requires the bare word herdr; the MustNot on "or `herdr` on PATH" is what blocks the old wording.
- [ ] internal/config/user.go rename-failed branch shadows v, exists and err with :=; correct, reads oddly.
