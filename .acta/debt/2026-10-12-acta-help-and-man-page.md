---
id: DBT-0112
hash: nxo6r9c
parent: plans/2026-10-12-acta-help-and-man-page
---
# acta help and man page review notes

- [ ] (medium) internal/cli/setup_cmd.go: the setupManStep(stdout) call in cmdSetup has no test; deleting it keeps every test green, because cmdSetup needs a terminal and the tests only call setupManStep on its own. Guarding it needs a way into the TTY-only setup flow.
- [-] (low) internal/cli/help.go usageLine panics at package init when a name is missing from helpTable; a Go panic exits 2, the code plugin/hooks/pre-tool reads as block, so such a binary would block every tool call. Tests catch it today.
- [-] (low) CLAUDE.md still lists voice among the internal/cli commands, but acta voice was removed and now gives the unknown-command error.
