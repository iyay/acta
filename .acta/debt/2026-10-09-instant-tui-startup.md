---
id: DBT-0104
hash: tr53gz4
parent: plans/2026-10-09-instant-tui-startup
---
# Instant TUI startup review notes

- [ ] (medium) No test covers the runTUI wiring in internal/cli/cli.go: switching the first board back to trees.Load, or WithLoad back to plain trees.Load, keeps every test green. The second makes the AUTHOR line vanish from the TUI with nothing failing. Closing it needs a small helper pulled out of runTUI that returns the first board and the loader, plus a test in internal/cli.
- [ ] (low) FillAuthors passes every absolute path of a checkout to one git log call (internal/board/closed.go). On this repo that is about 450 paths and 28 KB of arguments. When the Windows port lands (SCR-0058), its 32 KB command line cap is close, and a failed call gives every item in the checkout the current user.name instead of its real author. Relative paths or git's --pathspec-from-file in gitc.Authors would avoid it.
- [ ] (low) Nothing tests the path where the watcher cannot start inside tui.Start (internal/tui/watch.go): the send of watchFailedMsg, the stop that does nothing, and the full load that must still follow. If that path broke, a machine out of file watch slots would open the TUI with only the main tree and no word about manual mode. Testing it needs a way to hand Start a watcher that fails, which is new logic.
