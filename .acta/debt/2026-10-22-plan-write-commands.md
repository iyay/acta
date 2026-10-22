---
id: DBT-0140
hash: n38c3v0
parent: plans/2026-10-22-plan-write-commands
---
# Review NOTEs: Plan writes through acta commands

- [ ] (medium) The session-start runningRule in internal/hook/hook.go tells the agent to run acta state set <plan> next but not from the worktree; the new guard refuses that call from the main checkout, so a main-checkout session following the hook gets an error and loses a turn.
