---
id: DBT-0094
hash: mr313rz
parent: plans/2026-10-07-acta-setup-wizard-look
---
# acta setup look review debt

- [ ] (low) A failed harness install shows only the command to run by hand, not why it failed; the install output is thrown away to keep the rail clean. Show the last stderr line as a dim rail line (internal/cli/setup_cmd.go install runner).
- [ ] (low) TestSetupRunnerNoStdin swaps the global os.Stdin; it breaks if internal/cli tests ever run with t.Parallel.
