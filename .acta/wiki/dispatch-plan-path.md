---
type: Reference
title: dispatch init needs the md path
description: dispatch init --plan takes .acta/plans/<stem>.md, while tick takes plans/<stem>#task-N; the forms differ
paths: [internal/cli/]
timestamp: 2026-10-07T06:14:14Z
---

`acta dispatch init --plan plans/<stem>` fails with `plan %q is outside %s`. Pass `--plan .acta/plans/<stem>.md` instead. `acta tick` uses the `plans/<stem>#task-N` form, so the two commands differ on purpose. Put the exact command, id included, in every implementer hand-off: short ids fail with "unknown id".
