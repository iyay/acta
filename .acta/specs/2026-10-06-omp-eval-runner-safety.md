---
parent: debt/2026-09-30-omp-eval-runner
closes: [DBT-0048.01, DBT-0048.02, DBT-0048.03, DBT-0048.07, DBT-0048.08, DBT-0048.09]
id: SPC-0092
created: "2026-10-06 08:10:31"
hash: lro201m
---
Status: Bounded, approved by the user in chat on 2026-10-06.
Why: the omp eval runner (`acta eval-omp`, `internal/evalomp`) has one test that can wipe the user's real `~/.acta/config.yaml` when the code under test regresses, a `--case` gate that shows green when it ran nothing, and child processes that can hang a run or outlive it.

# omp eval runner: safe tests, honest --case, no stuck runs

## Changes

1. **HOME is never the real one in tests.** `TestMain` in `internal/evalomp/main_test.go` sets `HOME` to a fresh temp folder before `m.Run()`. `TestRunCaseScaffold` also gives the test its own fake `HOME` and checks that the scaffold did not write `.acta/config.yaml` there. If `RunCase` stops passing the throwaway home to the scaffold, this test fails. (DBT-0048.01)
2. **`--case` fails loudly.** `RunAll` checks the glob once, before the loop. A bad glob (for example `[abc`) is an error naming the glob. A glob that matches no case is an error: `no case matches "<glob>"`. In both cases no case runs and `acta eval-omp` exits 1. An empty glob still means every case. (DBT-0048.02, DBT-0048.03)
3. **The scaffold has a time limit.** The scaffold `bash` runs under `exec.CommandContext` with the case's own `TimeoutSeconds`. Going over it is an error that says the scaffold timed out. (DBT-0048.07)
4. **A timed-out omp takes its children with it.** omp starts in its own process group (`Setpgid`). On timeout the whole group is killed, not only omp. `WaitDelay` stays. (DBT-0048.08)
5. **The judge reports why it failed.** `OmpJudge` sets `WaitDelay`, and when omp fails, the error carries its stderr. (DBT-0048.09)

## Out of scope

DBT-0048.04, .05, .06, .10, .11 and .13 stay open as debt.

## Testing

Every change starts with a failing test that uses the fake omp already in `run_test.go`. Process-group and timeout tests use a fake that starts a sleeping child and checks it is gone after the run. Gate: `scripts/test --full`.

## Close

Last task bumps the patch version to 0.1.24 in `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`.
