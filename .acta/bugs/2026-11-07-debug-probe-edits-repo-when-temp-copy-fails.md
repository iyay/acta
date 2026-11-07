---
id: BUG-0068
hash: ma4vk63
started: "2026-11-07 18:05:01"
---
# Debug probe edits a repo file when its temp copy was never made

## Symptom
While debugging, an agent meant to edit a temp copy but edited a tracked file in the repo instead. In the debug-independent-check eval (run 1 of 3, 2026-11-07), `mktemp` failed in the sandbox, the following `cd` into the empty path did nothing, and `sed` changed `src/tax.py` in the repo. The agent put the file back and said so, but Phases 1 to 3 must change no repo file. Run 3 hit the same `mktemp` failure and ran the report in the repo root, which left `src/__pycache__/`.

## Root cause
`plugin/skills/debug/SKILL.md:12` says probes live outside the repo in a temp folder, but never says to check that the copy exists and that the shell is inside it before a probe writes. A command chain like `D=$(mktemp -d); cp ... $D; cd $D; sed -i ...` keeps going after a failed step, so the write lands in the repo.

## Repro
From the acta-debug-independent-check worktree: `scripts/eval --case debug-independent-check --runs 3 --keep-temp`. Run 1's trace shows the failed `mktemp`, then `sed` on the repo file. The case lives on branch `debug-independent-check` (not landed).

## Found in
Branch `debug-independent-check`, eval gate of PLN-0214 (SPC-0191), read from the eval trace.
