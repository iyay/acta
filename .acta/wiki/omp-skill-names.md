---
type: Reference
title: omp skills carry bare names
description: In omp, acta skills show under bare names like shape with no acta prefix, mixed with other plugins
paths: [plugin/omp/]
timestamp: 2026-10-05T15:47:58Z
---

Claude Code namespaces plugin skills (`acta:shape`); omp uses the SKILL.md frontmatter name as is (shape, slice), so acta skills mix with other plugins there. Ruling: keep the bare frontmatter names and prefix every skill description with `acta: ` instead. The user removes dead omp links in user scope; agents never touch them.
