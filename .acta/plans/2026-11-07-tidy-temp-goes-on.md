---
parent: debt/2026-11-07-worktree-debt-text
closes: [DBT-0161.01, DBT-0161.02]
depth: minimal
id: PLN-0222
created: "2026-11-07 21:12:14"
hash: ab1ptn3
started: "2026-11-07 21:13:00"
finished: "2026-11-07 21:36:29"
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
- [x] Red: the new land-goes-on regex fails on a hand-made log where the land stops at the refused remove.
- [x] Change: add the land-goes-on grader and tighten the judge to both choices, with who reruns the remove left open.
- [x] Commit: test(evals): pin offer and go on for a refused tidy temp remove

## Polish

### Task 2: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] [fix] the judge fails a reply where the agent offers to run a force itself; only a line that the user may force on their own is allowed.
- [x] Commit: `polish: review notes for PLN-0222`

## Review notes

- Review: deep, round 1 CLEAN on both axes; polish (judge fails an agent offer to force) light tier, grader wording only. Plan and spec text still say tidy.md does not change; the 21:32 ruling overrides that. Notes left: land-goes-on takes any refused remove and any merge --ff-only target.

## State

### Next

Parked 2026-11-07 21:50 by the user (deadline 21:55). Review CLEAN, polish committed, full suite green at 56bb9e8 (baseline written).
Eval gate (scripts/eval -j 8): red all 3 times on this branch: brainstorm-related-topic and debug-no-repro; debug-no-repro passed once on main as a control. Neither case touches tidy.md, so check whether -j 8 load or main moving explains it before calling it a skill gap.
Passed on a rerun, not yet recorded as flakes: brainstorm-mixed-context, brainstorm-resume-topic, debug-order-dependent, debug-independent-check, land-tidy-temp-refused, second-brainstorm-choices.
Next: rerun the two red cases alone without -j, record the flakes with acta flake record eval:<case> --plan PLN-0222, then acta tidy main tidy-temp-goes-on and land.

### Findings

Plan text still says tidy.md does not change; the 21:32 ruling changed that, and the spec Decision line is stale the same way.

### Open rulings

User ruled A on 2026-11-07 21:32 (WhatsApp): tidy step 3 gets one sentence to name both choices in the report even when it ends with one next action, so this plan now touches tidy.md and its land needs the full eval gate. The judge also lets the agent say the user may run a force on their own, since core rule 10 allows it.
