---
id: DBT-0032
hash: o7qusw1
parent: plans/2026-09-30-config-file-name
started: "2026-09-30"
---
# Review NOTEs: Config File Name Implementation Plan

- [x] plugin/skills/build/SKILL.md:3 and :16 still call subagent the default executor; the body at :22 reads the config first, so only the wording is stale.
- [x] internal/cli/voice.go:20: after a failed rename, acta voice show prints config.yaml (exists: true) while the values came from voice.yaml; the broken-YAML error at :85 names config.yaml in that case too.
- [x] Old and new acta binaries side by side: an old binary still reads and writes ~/.acta/voice.yaml, and the new one ignores it once config.yaml exists; rebuild the PATH binary right after landing. (stale)
- [ ] (low) internal/voice/voice.go:91: os.Rename replaces config.yaml if a non-acta writer creates it between Load and the rename; os.Link plus os.Remove would never overwrite.
- [ ] (low) internal/voice/voice.go:94 shadows v, exists, err in an inner scope; correct but harder to read.
- [ ] (low) plugin/skills/setup/SKILL.md:8 says "An old voice.yaml in the same folder" instead of the plan text, because the plan's own MustNot banned ~/.acta/voice.yaml.
- [ ] (low) plugin/skills/setup/SKILL.md:41: asking to create a CLAUDE.md does not tell the user that /init also writes codebase docs into it.
- [x] Task 1 verify grep also matches voice.go comments about the old file and the MustNot entry in skill_setup_test.go; the verify line did not expect them. (stale)
- [x] TestResolveTwoReadersRacingBothSeeTheSetting goes beyond the plan and races goroutines in one process, not two hook processes. (stale)
