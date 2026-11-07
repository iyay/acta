---
id: SPC-0191
created: "2026-11-07 17:05:16"
hash: enkpnc0
---
# Debug independent check

Bounded. The user approved the design on WhatsApp on 2026-11-07. Their goal is that a cause `acta:debug` reports can be trusted. The change ships only if an eval shows the current skill falls short where the new one does not.

## Problem

The Claude Code ultrareview docs (code.claude.com/docs/en/ultrareview) reproduce and verify every finding on its own before they report it. A finding that fails the check is dropped, and a stopped review reports nothing half done. Set against `acta:debug`, two gaps show:

- The agent that forms a belief about the cause is the same one that proves it. Nothing checks the proof without that belief, so a misread output, a probe run in the wrong tree, or a control that does not really remove the cause can pass.
- The Phase 3 report has no fixed shape, so a reader cannot tell at a glance what was proven and what was not.

## Goal

When doubt was high, a cause reaches the user only after a check that does not share the investigator's reasoning.

## Out of scope

- No cloud runs, billing, PR posts, diff size limits or exit codes.
- No check for a clear bug: one good probe stays enough.
- No parallel probes.
- How often agents load `acta:debug` on their own is a separate topic.
- The race case gap on Sonnet (0/2 in the SPC-0157 runs) gets its own spec.

## Design

New file `plugin/skills/debug/verify.md`, named from Phase 3 of `SKILL.md`:

1. **When.** The check runs only when doubt was high: more than one cause fit in Phase 2, or a fix already failed. Then a control is required, not "when you can". A cause with no possible control is reported as not proven.
2. **Inputs.** A fresh subagent gets only these: the source tree path and commit to copy, a repro command that runs on its own (probe scripts passed along with it), the one-sentence cause ("X at `file:line` does Y, so Z"), and the control as a patch or a command. It gets no notes and no reasoning.
3. **Check.** It copies the tree to its own temp folder, runs the repro, applies the control, runs the repro again, and tries to refute the cause. The cause holds only when the repro shows the symptom and the control removes it.
4. **Refuted.** A refuted cause goes back to Phase 2. After two refutations, or when the check cannot finish (the subagent errors or stops), stop and report only what was ruled out. No cause is sent.
5. **No subagent.** In a harness with no subagents, run steps 2 and 3 yourself from a fresh temp copy before reading your own notes again.
6. **Report.** The Phase 3 report has four lines: `Cause:` (with `file:line`), `Repro:` (a command the user can paste), `Evidence:` (probe and control output), `Unverified:` (what was not reachable). The "Name what you saw" ground rule stays for every other report.

In `SKILL.md`: Phase 3 step 2 says a control is required when `verify.md` applies, and step 5 points to `verify.md` for the check and the report. `SKILL.md` keeps its byte count (6546) and line count by tightening its existing text.

### Budget

`verify.md` gets its own byte cap in `internal/plugincheck/budget_test.go`. `MaxLines` in `internal/plugincheck/skill_debug_test.go` (350, the folder is at 350 now) rises by exactly the line count of `verify.md`.

### Eval first

New eval case `plugin/evals/debug-independent-check/`, in the layout of the other debug cases. Its scaffold plants a bug whose first suspect has a control that looks right but leaves the real cause in place, so the symptom still shows when the control is run as written. The grader passes a run that reports the real cause, or reports the first suspect as not proven. It fails a run that reports the first suspect as the cause.

Run the case three times on the current skill before writing `verify.md`. If all three pass, the current skill already holds: stop and tell the user, and close this spec as not needed. Otherwise, build, then the case must pass three of three on the new skill.

### Testing

- Red first in `skill_debug_test.go`: `verify.md` exists and names the trigger, the four inputs, the two-refutation stop, the no-subagent fallback and the four report lines. `SKILL.md` links `verify.md` from Phase 3. The existing `Must` list still passes.
- The budget test passes with the new caps.
- `scripts/test` green. At land, the six existing debug evals and the new case stay green.

### Scorecard

The user wants people to pick `acta:debug` because its results are right, so the gain must show in numbers. After the build, run all seven debug cases twice each on the old skill and on the new one, with the flags of `scripts/eval`. Put pass counts, tokens and cost per run in a table in the plan, next to the reference numbers SPC-0157 recorded (marked with their date, since they were not run again). The new skill must not pass fewer runs than the old one on any case. The order-dependent case is the one to watch: its failures left the paste-ready repro out of the report, which the `Repro:` line now asks for.
