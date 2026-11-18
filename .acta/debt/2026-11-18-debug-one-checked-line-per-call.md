---
id: DBT-0171
hash: z2hlvyk
parent: plans/2026-11-18-debug-one-checked-line-per-call
---
# Review NOTEs: Debug skill: one checked line per shell call

- [ ] (medium) Rerun the old-skill arm of debug-probe-isolation under the final graders (bug-record exception, stdout-only control-ran); the 16/20 against 4/10 comparison used different graders for the old arm.
- [ ] (medium) The debug skill bans ; outside quotes in a write call, while the flaky and race recipes ask for many runs; a shell for loop needs ; so agents may break the rule or skip repeats. Say how to repeat a write safely, then measure debug-race and debug-order-dependent.
