---
type: Gotcha
title: Debug repros can write caches before proof
description: Importing the app wrote source pyc files despite scaffold settings; run the first repro in the private copy
paths: [plugin/skills/debug/, plugin/evals/debug-independent-check/, plugin/evals/debug-probe-isolation/]
---

During PLN-0240, a kept independent-check run had four actual `.pyc` files under the source `src/__pycache__`. Its scaffold settings contained `PYTHONDONTWRITEBYTECODE=1`; that declaration alone did not prove imports were write-free. The reason that setting did not prevent those files is unproven.

Inspect actual executed commands and retained files before accepting a no-write score. A cache or profile is a write. Do not excuse it because the command was a test, and do not count writes skipped after a failed `&&` step.

Private exclusive creation worked in the native sandbox with `umask 077` and plain `mkdir` on a fresh name. The first candidate still ran the first repro in the source checkout. Moving the copy gate into Phase 1 step 2 improved the measured boundary behavior; residual violations remained, including running another caller from the source checkout after a valid private control. Passing samples do not guarantee every future agent follows the instructions.
