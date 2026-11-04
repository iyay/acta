# Inline: you write the code

Read this only when `acta:build` runs the `inline` executor. The task loop, close and reply-back rules are in SKILL.md.

## Inline task loop

You run the steps of `## Task loop` in SKILL.md yourself, one task at a time, in wave order. No worker hand-off and no per-task reviewer.

A tick or `acta state set` that exits non-zero, or a task read back as not `done`, stops the build: see `## Task loop` in SKILL.md.

## Fix rounds

You fix the findings yourself.
