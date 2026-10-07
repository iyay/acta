---
parent: debt/2026-10-07-routing-eval-set
id: SPC-0099
created: "2026-10-07 09:23:09"
hash: acvlhxc
---
Status: Bounded, approved by the user in chat on 2026-10-07.
Why: the land gate runs `scripts/eval` whenever a diff touches `plugin/skills/` or `plugin/hooks/`. With no filter that runs every case, so the 18 routing cases (runs: 3 each) add 54 runs to every such land. The routing set is a baseline measure for SCR-0004, not a per-change gate.

# scripts/eval skips the routing cases unless asked

## Changes

1. **Default run skips routing.** When `scripts/eval` gets no `--tag` and no `--case` argument, it runs only the cases whose folder name does not start with `routing-`. `claude plugin eval` has no exclude filter, only `--tag` and `--case <glob>`, so the script lists the case folders under `plugin/evals/` itself and passes the rest with `--case`. The exact form (repeated `--case`, or one glob) is proved in the task without spending quota.
2. **An explicit filter runs as asked.** With `--tag` or `--case` in the arguments, nothing is skipped: `scripts/eval --tag routing` runs the 18 routing cases.
3. **omp too.** `scripts/eval --omp` follows the same rule.
4. **Comment.** The script says why in plain words: routing is a baseline measure, and in the land gate it would add 54 runs to every land.
5. **Test.** `TestEvalScriptFlags` (or a neighbour in `internal/plugincheck/evals_test.go`) checks that a default run excludes every `routing-*` case and includes every other case, and that `--tag routing` is passed through untouched.
6. **Version** 0.1.31.

`plugin/skills/land/SKILL.md` does not change; the gate still calls `scripts/eval`.

## Testing

`scripts/test ./internal/plugincheck/` while building, `scripts/test --full` at land. No paid `scripts/eval` run.

## Change after review round 1 (approved by the user in chat on 2026-10-07)

Proved at no cost: `claude plugin eval` keeps only the last `--case`, and its glob takes no braces or negation, so there is no way to exclude cases in one call. `acta eval-omp --case` also keeps only the last value. Changes 1 to 3 above are replaced by this:

- The 18 routing cases move from `plugin/evals/` to `plugin/evals-routing/`. `scripts/eval` with no extra arguments runs `plugin/evals/` as it did before this plan, with no case list.
- The baseline runs with `scripts/eval --eval-dir evals-routing`, which `claude plugin eval` already supports.
- `acta eval-omp` gains `--eval-dir <dir>` (default `evals`), so `scripts/eval --omp --eval-dir evals-routing` works the same way.
- Tests follow the move: `TestRoutingEvalCases` reads `plugin/evals-routing/`, the scaffold checks also cover it, and a test proves `plugin/evals/` holds no `routing-*` case.
