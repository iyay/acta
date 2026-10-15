---
id: DBT-0116
hash: s23q6e5
parent: plans/2026-10-15-optional-tui-install
---
# Review NOTEs: Optional TUI at install time Implementation Plan

- [-] The Variant to label mapping is written three times with three strings: help.go ("cli, no TUI"), doctor.go ("cli (no TUI)"), update_cmd.go ("acta-cli"); when a label changes, the others drift. One helper would keep them in step.
- [-] scripts/install_test.go run() starts install.sh with no deadline of its own, unlike ptyRun; a stuck curl only stops at the go test timeout.
- [-] The prompt's EOF / Ctrl-D path in install.sh has no test in TestInstallAsksOnTerminal.
