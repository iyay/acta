---
id: DBT-0086
hash: vrgqbl8
parent: plans/2026-10-06-omp-eval-runner-safety
---
# omp eval runner safety review notes

- [ ] (low) internal/evalomp/run.go: a scaffold that times out has only its top bash killed, not its process group; a background child of the scaffold can outlive the run once WaitDelay ends.
