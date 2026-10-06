---
id: DBT-0062
hash: fviavuo
parent: plans/2026-10-01-four-old-bugs
---
# Review NOTEs: Four Old Bugs Implementation Plan

- [ ] (medium) The debt item refusal in SetValue (internal/write/ops.go) talks only about titles, but it also refuses status, type and ref; the TUI status popup on a debt line shows that title message after the user picks a value.
- [ ] (low) The TUI still opens the status and type popups for a debt item and refuses only after a pick (internal/tui/model.go status popup); it could refuse on the key press like tasks do.
- [ ] (low) No regression test for status on a debt item, the BUG-0025 path the debt item guard now closes.
- [ ] (low) No test covers a fresh tab whose item on show changes or disappears on reload, the case the reloadMsg fix exists for; the item-gone subtest moves the cursor first.
- [ ] (low) The empty, blank and line-break title refusal tests would also pass with no title field at all, because an unknown field is bad input too; check the error text.
- [ ] (low) No title tests for a `# ` line inside a code fence.
- [ ] (low) A title with spaces at the start or end is written as typed, and the board trims it on read.
- [ ] (low) The doctor `rm <path>` hint and the `omp plugin link <target>` hint next to it are not shell-quoted, so a path with a space breaks if pasted.
- [ ] (low) The doctor comment says "omp cannot unlink itself"; the real reason is that omp has no unlink action.
- [ ] (low) The SetValue doc comment still says "sets the status or type of one story or bug".
- [ ] (low) setTitle on a frontmatter that ends with --- and no newline and no heading puts the new heading on the fence line.
- [ ] (low) setTitle drops a bare \r on a last heading line with no newline after it.
- [ ] (low) scripts/test GIT_CONFIG_COUNT replaces any GIT_CONFIG_* the caller already set.
- [ ] (low) The spec puts the BUG-0010 fix in moveTo; the code puts it in the reloadMsg case, which review ruled better, so the spec text no longer matches.
