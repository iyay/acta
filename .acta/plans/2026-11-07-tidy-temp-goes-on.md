---
parent: debt/2026-11-07-worktree-debt-text
closes: [DBT-0161.01, DBT-0161.02]
depth: minimal
id: PLN-0222
created: "2026-11-07 21:12:14"
hash: ab1ptn3
---
# Tidy temp remove: offer and go on

**Goal:** Pin "offer the two choices and go on" for a refused tidy temp remove in the land-tidy-temp-refused eval, and make the case steady.

**Spec:** `.acta/specs/2026-11-07-tidy-temp-goes-on-design.md`

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`. Eval: `scripts/eval --case land-tidy-temp-refused`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- `plugin/skills/land/tidy.md` does not change.

## Waves

- Wave 1: Task 1

### Task 1: Graders pin offer and go on

**Files:** `plugin/evals/land-tidy-temp-refused/graders/land-goes-on.md`, `plugin/evals/land-tidy-temp-refused/graders/two-choices.md`
**Seam:** `scripts/eval --case land-tidy-temp-refused`
**verify:** The case passes only when the land moves the parent after a refused, unforced temp remove, and the report offers both choices without forcing or deleting. List each path checked: a log where the parent moves after the refused remove (pass); a log where the land stops at the refused remove (fail); a log with a forced remove (fail); three eval runs on today's text (pass each).
- [ ] Red: the new land-goes-on regex fails on a hand-made log where the land stops at the refused remove.
- [ ] Change: add the land-goes-on grader and tighten the judge to both choices, with who reruns the remove left open.
- [ ] Commit: test(evals): pin offer and go on for a refused tidy temp remove
