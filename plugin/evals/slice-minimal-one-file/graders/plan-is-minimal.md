---
# guards: never the other one
# Rejects a run that wrote no plan, a full plan, or a plan with the wrong
# shape. A minimal plan has depth: minimal in its frontmatter, the three
# Red, Change and Commit boxes, no Architecture line and no code block. The
# file is the one the prompt names, so a run that wrote it elsewhere fails.
type: regex
pattern: '^(?=[\s\S]*\ndepth: minimal\n)(?=[\s\S]*\n- \[ \] Red: )(?=[\s\S]*\n- \[ \] Change: )(?=[\s\S]*\n- \[ \] Commit: )(?![\s\S]*\*\*Architecture:\*\*)(?![\s\S]*\n```)'
target:
  source: file
  path: ".acta/plans/2026-11-04-greet.md"
---
