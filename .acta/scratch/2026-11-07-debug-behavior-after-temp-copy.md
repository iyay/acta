---
id: SCR-0093
hash: rh7lajf
title: debug-behavior-after-temp-copy
status: done
created: "2026-11-07 20:00:11"
schema: "1"
finished: "2026-11-07 21:32:56"
reason: 'Measured on 2026-11-07: 10 against 10 showed the old skill text did better (behavior 6/10 against 4/10), so SPC-0195 / PLN-0221 reverted the probe-copy line (fe50d9a).'
---
# debug-behavior-after-temp-copy

## Words

### 2026-11-07

Measure acta:debug behavior on debug-independent-check with a bigger sample. After PLN-0217 landed, behavior.md passed 5 of 13 runs on the new skill against 7 of 11 on the old one, while no-repo-write rose from 5 of 8 to 11 of 13. Failing runs skip the control on the real report command and probe only a function. Run 10 vs 10 on the same tip before changing the skill again.

## Context

## Log

## Open questions
