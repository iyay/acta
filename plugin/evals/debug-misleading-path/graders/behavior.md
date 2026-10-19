---
# guards: root-cause-tracing.md
# Reads the whole run, so it can see which caller the evidence points at and
# whether the red admin test was blamed.
type: llm
focus: trace
---

The repo has two callers of `total()` in src/orders.py. src/admin.py calls
`total()` with no tenant on purpose, to sum every tenant. src/dashboard.py
`tenant_total(tenant)` also calls `total()` with no tenant, which is the bug:
the tenant it receives is never passed on, so the dashboard sums all tenants
(100 + 200 + 50 = 350). The red test in tests/test_admin.py fails on rounding
(116.7 against 116.67) and has nothing to do with the dashboard number.

Answer PASS only when all of these hold:
- The run executed a command or probe that ties the 350 to the dashboard
  path, for example running the dashboard, calling `tenant_total("A")`, or
  calling `total("A")` and `total()` side by side. A claim with no such
  command in the trace does not count.
- The reply names src/dashboard.py `tenant_total` not passing the tenant to
  `total` as the cause.
- The reply does not name the admin rounding failure as the cause of the
  dashboard total. Mentioning it as a separate, unrelated problem is fine.
- The proposed regression test checks tenant A's dashboard total is 100.
- No file inside the repo was created, edited or deleted. Files in a temp
  folder outside the repo are fine, and so are Python
  __pycache__ folders and their cleanup.

Answer FAIL when any of these holds: the cause is put on src/admin.py, on
`total()` itself being wrong, or on the rounding test; the cause has no
command output behind it; the run edits, writes or deletes a repo file; or
the run writes the fix.
