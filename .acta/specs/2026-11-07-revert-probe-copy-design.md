---
id: SPC-0195
created: "2026-11-07 21:06:05"
hash: c2xc2ha
started: "2026-11-07 21:07:29"
finished: "2026-11-07 21:19:48"
---
# Revert the probe copy rule

Bounded. The user chose this on WhatsApp on 2026-11-07 (option B1) after a 10 against 10 eval.

## Problem

PLN-0217 added one sentence to the debug skill's ground rule, `D=$(bash probe-copy.sh) && cd "$D"`, and the `probe-copy.sh` helper; PLN-0220 hardened the helper. On `debug-independent-check`, the skill text from before that work beat the new text on both graders: `behavior.md` 6/10 against 4/10, `no-repo-write.md` 9/10 against 8/10. Agents almost never call the helper.

## Goal

The debug skill reads as it did before PLN-0217, and no unused helper stays behind.

## Design

1. `plugin/skills/debug/SKILL.md` goes back to its text at `20a15a7` (before PLN-0217).
2. Delete `plugin/skills/debug/probe-copy.sh` and `internal/plugincheck/probe_copy_test.go`.
3. `internal/plugincheck/skill_debug_test.go` drops `stops at the first failed step` and `probe-copy.sh` from `Must`. `internal/plugincheck/budget_test.go` puts the `SKILL.md` cap back to 6546. `internal/plugincheck/final_test.go` no longer allows `probe-copy.sh` as an executable.
4. The eval case `plugin/evals/debug-independent-check/` stays, for later measurement. Its `evalCases` phrase and the `no-repo-write.md` guards line go back to `Prove the mechanism`.
5. The wiki page `skill-escape-lines-skip-proof` gets one line with the 10 against 10 result and the revert.
6. The bug `2026-11-07-debug-probe-edits-repo-when-temp-copy-fails` is open again: its fix is undone.

## Testing

- `scripts/test ./internal/plugincheck/ ./internal/setup/ ./plugin/` green.
- `git diff 20a15a7 -- plugin/skills/debug/` shows no change.
- At land: full suite, and the seven debug eval cases.
