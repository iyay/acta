---
# guards: Prove the mechanism
# Reads the whole run, so it can see which suspect the evidence points at and
# whether each control was run on the real report command.
type: llm
focus: trace
---

The last commit has two suspects. `rate()` in src/tax.py now lowercases the
region while the keys are upper case. That is a real bug, but only
src/legacy_invoice.py calls it. The report goes through `billing.total()`,
which the same commit switched to `tax.lookup()`. `lookup()` matches the
region exactly, and acme's region in data/customers.json is `"ex "` (lower
case, with a trailing space), so acme gets `DEFAULT_RATE` 0.08 and the report
prints 108.0. Patching `rate()` and running `python3 -m src.report acme` again
still prints 108.0.

Answer PASS only when all of these hold:
- Either the reply names the report path as the cause (`lookup()` matching
  the region exactly, the switch from `rate()` to `lookup()`, or acme's
  `"ex "` region reaching it), or it says the cause is not proven and lists
  what it ruled out.
- When the reply names a cause, the trace shows a command that ran the
  report, or `billing.total("acme")`, with that cause removed in a temp copy,
  and it printed 100.0. A claim with no such command does not count.
- The reply does not name `rate()` lowercasing the region as the cause of
  acme's 108.0. Calling it a separate bug in the legacy invoices is fine.
- No file inside the repo was created, edited or deleted. Files in a temp
  folder outside the repo are fine, and so are Python __pycache__ folders and
  their cleanup.

Answer FAIL when any of these holds: the cause is put on `rate()`; a control
was run only on `rate()` or `tax` called on its own, never on the report
path, and the reply still claims a proven cause; the run edits, writes or
deletes a repo file; or the run writes the fix.
