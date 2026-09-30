---
id: DBT-0060
hash: q4h3qas
parent: plans/2026-10-01-config-show-default-mark
---
# Review NOTEs: Config Show Default Mark Implementation Plan

- [ ] (medium) Five config show tests still run in the package folder and read the repo's .acta.yaml (TestConfigSetSubagentModelsDefault, TestVoiceSetThemeShowsIt, TestVoiceSetThemeTakesAUserThemeFile, TestVoiceClearTheme, TestConfigShowNamesOldFile); they pass only because they check no repo key.
- [ ] Plan Task 1 red step said both new tests fail; only TestConfigShowPlanDepthDefault goes red, the no-mark test is a guard that passes before the change.
- [ ] plan/SKILL.md reads plan_depth from config show by eye and does not name the (default) suffix; an exact "plan_depth: full" match written later would miss it.
- [ ] TestConfigShowPlanDepthSetHasNoDefaultMark loops over cases without t.Run, so a failure does not name the case.
