---
id: SPC-0191
created: "2026-11-07 17:05:16"
hash: enkpnc0
---
# Debug independent check

Bounded. The user approved the design on WhatsApp on 2026-11-07, provided it makes `acta:debug` better.

## Problem

The Claude Code ultrareview docs (code.claude.com/docs/en/ultrareview) run many reviewers in parallel, then reproduce and verify every finding on its own before they report it. A finding that fails the check is dropped, and a stopped review reports nothing half done. Set against `acta:debug`, three gaps show:

- The agent that forms a belief about the cause is the same one that proves it. Nothing checks the proof without that belief.
- Phase 2 runs one probe at a time, even when two or three causes rank close and each probe is slow.
- The cause report has no fixed shape, so a reader cannot tell at a glance what was proven and what was not.

## Goal

A cause reaches the user only after a check that does not share the investigator's reasoning, when doubt was high.

## Out of scope

- No cloud runs, billing, PR posts, diff size limits or exit codes.
- No check for a clear bug: one good probe stays enough.
- No new plugin eval. The six debug evals already run at land.

## Design

New file `plugin/skills/debug/verify.md`, named from `SKILL.md`:

1. **When.** The check runs only when doubt was high: more than one cause fit in Phase 2, or a fix already failed.
2. **Independent check.** A fresh subagent gets only three things: the repro command, the one-sentence cause ("X at `file:line` does Y, so Z") and the control. It gets no notes and no reasoning. It reruns both in its own temp copy and tries to refute the cause. A cause it cannot reproduce, or whose control does not remove the symptom, is not proven: go back to Phase 2.
3. **No subagent.** In a harness with no subagents, rerun the repro and the control from a fresh temp copy before reading your own notes again.
4. **Parallel probes.** When two or three ranked causes each need a slow probe and the probes do not share state, run them at the same time, each in its own temp copy. This is optional.
5. **Report.** The Phase 3 report has four lines: `Cause:` (with `file:line`), `Repro:` (a command the user can paste), `Evidence:` (probe and control output), `Unverified:` (what was not reachable). A stopped or refuted check sends no cause, only what was ruled out.

In `plugin/skills/debug/SKILL.md`: Phase 3 gets one step that points to `verify.md`, Phase 2 step 3 points to it for parallel probes, and the four-line report replaces the "Name what you saw" ground rule. The added text is paid for by tightening existing text, so the file stays at or under its 6546-byte cap.

### Budget

`verify.md` gets its own byte cap in `internal/plugincheck/budget_test.go`. The debug folder line cap (277) rises only by what `verify.md` adds.

### Testing

- Red first in a plugincheck test: `verify.md` exists and names the trigger (more than one cause, or a failed fix), the three inputs, the no-subagent fallback, the parallel probe rule and the four report lines. `SKILL.md` links `verify.md` from Phase 3.
- The budget test passes with the new caps.
- `scripts/test` green. At land, the six debug evals stay green.
