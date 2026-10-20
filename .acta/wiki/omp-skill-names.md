---
type: Reference
title: omp skills carry bare names
description: In omp, acta skills show under bare names like shape with no acta prefix, mixed with other plugins
paths: [plugin/omp/]
---

Claude Code namespaces plugin skills (`acta:shape`); omp uses the SKILL.md frontmatter name as is (shape, slice), so acta skills mix with other plugins there. Ruling: keep the bare frontmatter names and prefix every skill description with `acta: ` instead.

The omp extension sends its own line with the session rules (`SKILL_NOTE` in `plugin/omp/index.ts`): read a skill as `skill://` plus its bare name, so `acta:build` is `skill://build` and never `skill://acta-build`. It goes out with the first context, after a session switch or a compaction, and also when the binary is missing and the default rules stand in; later prompts carry only the usual reminder. `plugin/omp/index.test.ts` pins that text and checks that every installed skill keeps a bare frontmatter name, since the URI is exactly that name.

The user removes dead omp links in user scope; agents never touch them.
