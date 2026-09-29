---
id: SCR-0025
hash: ptoej3m
title: Bug fix plans visible in bug detail and Activities
status: raw
created: "2026-09-29"
---
User: "ini plan-33-nya mana? kok gak ada?" then "oooh.. tetep harus keliatan di Detail Activity kalo itu Bug" then picked 3 (both) (2026-09-29)

Agent lines:
- Seen with PLAN-33 (parent: bugs/2026-09-29-second-brainstorm-choices-not-offered): the plan shows only nested under BUG-6, not in the Plans tab.
- Want 1: when a bug is selected, its detail pane lists its fix plan(s) and their tasks with progress, the way a spec shows its plans.
- Want 2: the Activities tab (internal/tui/sidebar.go activityRows) keeps showing a bug-fix plan's tasks even though the plan's parent is a bug.
- Open question for brainstorm: Activities only lists tasks that are in progress (inProgress in internal/tui/model.go), so a finished task leaves it anyway; decide whether that part is expected.
