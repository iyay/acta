---
id: SPC-0191
created: "2026-11-07 17:05:16"
hash: enkpnc0
status: dropped
finished: "2026-11-07 17:56:08"
reason: 'Eval gate on the current skill: 3 of 3 runs found the real cause with a control on the report path; the independent check adds nothing here. See PLN-0214 findings on branch debug-independent-check.'
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

1. **When.** The check runs only when doubt was high: more than one cause fit in Phase 2, or a fix already failed (one the user reports, or one tried in this session). Then a control is required, not "when you can". A cause with no possible control is reported as not proven.
2. **Inputs.** A fresh subagent gets only these: the path of the working tree to copy as it is, uncommitted and untracked files included (the commit is only a label for the report), a repro command that runs on its own (probe scripts passed along with it), the one-sentence cause ("X at `file:line` does Y, so Z"), and the control as a patch or a command. It gets no notes and no reasoning.
3. **Check.** It copies the tree to its own temp folder, runs the repro, applies the control, and runs the repro again. It returns exactly one verdict, plus the repro output before and after the control:
   - `holds`: the repro shows the symptom and the control removes it.
   - `refuted`: the repro shows no symptom, or the symptom stays after the control.
   - `could not finish`: the copy, the repro or the control did not run.
4. **Refuted.** A `refuted` cause goes back to Phase 2. After two `refuted` verdicts in one debug session (over all causes), or one `could not finish`, stop and report `Cause: not proven`.
5. **No subagent.** In a harness with no subagents, run steps 2 and 3 yourself from a fresh temp copy before reading your own notes again.

In `SKILL.md`:

6. **Report.** Phase 3 step 5 gives every Phase 3 report four lines, whether or not the check ran: `Cause:` (with `file:line`, or `not proven` followed by what was ruled out), `Repro:` (a command the user can paste), `Evidence:` (probe and control output), `Unverified:` (what was not reachable). The "Name what you saw" ground rule stays for every other report.
7. Phase 3 step 2 says a control is required when `verify.md` applies, and a new step points to `verify.md`. `SKILL.md` keeps its byte count (6546) and line count by tightening its existing text.

### Budget

`verify.md` gets its own byte cap in `internal/plugincheck/budget_test.go`. `MaxLines` in `internal/plugincheck/skill_debug_test.go` (350, the folder is at 350 now) rises by exactly the line count of `verify.md`.

### Eval first

New eval case `plugin/evals/debug-independent-check/`, in the layout of the other debug cases, except that `allowed_tools` adds `Agent` and `max_turns` and `timeout_seconds` leave room for the check run. Its scaffold leaves two causes that both fit the symptom, so the trigger fires. The first suspect has a control that looks right but leaves the real cause in place, so the symptom still shows when the control is run as written. The graders pass a run that dispatches a subagent for the check and reports the real cause, or reports `Cause: not proven`. They fail a run that reports the first suspect as the cause, or one that never dispatches a subagent.

Run the case three times on the current skill before writing `verify.md`. If all three pass, the current skill already holds: stop and tell the user, and close this spec as not needed. Otherwise, build, then the case must pass three of three on the new skill.

### Testing

- Red first in `skill_debug_test.go`: `verify.md` exists and names the trigger, the inputs, the three verdicts, the two-refutation stop and the no-subagent fallback. `SKILL.md` links `verify.md` from Phase 3 and holds the four report lines. The existing `Must` list still passes.
- The budget test passes with the new caps.
- `scripts/test` green. At land, the six existing debug evals and the new case stay green.

### Scorecard

The user wants people to pick `acta:debug` because its results are right, so the gain must show in numbers. After the build, run all seven debug cases twice each on the old skill and on the new one, with the flags of `scripts/eval`. The old arm is main's `plugin/skills/debug/` swapped into a temp copy of the branch's `plugin/`, so it has the new case too. Put pass counts, tokens and cost per run in a table in the plan, next to the reference numbers SPC-0157 recorded (marked with their date, since they were not run again).

When the new skill passes fewer runs than the old one on a case, run that case two more times on both arms. Still fewer over the four runs: it is a blocker, and the work goes back to build. The order-dependent case is the one to watch: its failures left the paste-ready repro out of the report, which the `Repro:` line now asks for.
