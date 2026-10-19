---
type: Gotcha
title: Evals run from a tmp worktree go falsely red
description: A checkout under /private/tmp made probe-round fail 6 of 6 on code that passes from a folder next to the repo
paths: [plugin/evals/, scripts/eval]
---

On 2026-10-19 `scripts/eval --case probe-round` failed 6 of 6 runs from a detached worktree under the session scratchpad (`/private/tmp/...`). The same tree as main, also checked out there, failed 2 of 2. From the main checkout it passed 5 of 5, and the branch under test passed 2 of 2 from `../<repo>-<slug>`. The failing runs never read `plugin/skills/shape/probe.md`. The cause inside the runner is not known.

Run land gates and eval comparisons from the main checkout or a worktree next to the repo, never from a temp folder. When only a temp-folder run is red, repeat it from a sibling worktree before blaming the diff.
