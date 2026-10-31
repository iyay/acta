---
id: SCR-0088
hash: bmc0ny9
title: acta id commits a new spec, then acta commit fails with an empty error
status: raw
created: "2026-10-31 20:20:07"
schema: "1"
---
# acta id commits a new spec, then acta commit fails with an empty error

## Words

### 2026-10-31

Seen twice on 2026-10-31 while shaping SPC-0185 and SPC-0186: after writing a new spec, `acta id` committed the spec file itself (subjects "chore: migrate ids to 3-letter prefix" at f7d8fc8 and "chore: new 2026-10-31-block-destructive-commands-design" at 9cd3d85, not chore(spec)). The shape skill's next step, `acta commit <spec> -m "..."`, then printed `written, not committed: commit failed: git commit --only ... exit status 1:` with no reason, because nothing was left to commit. Either acta id should not commit a file it only stamped, or acta commit should say "nothing to commit" and exit 0.

## Context

## Log

## Open questions
