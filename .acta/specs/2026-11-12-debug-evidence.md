---
parent: scratch/2026-11-07-debug-four-line-report
id: SPC-0204
created: "2026-11-12 23:01:55"
hash: k2tsm6x
started: "2026-11-13 00:59:44"
finished: "2026-11-13 01:51:46"
---
# Debug proof is required, and each text change is measured

Bounded. Worked out with the user in the terminal on 2026-11-12. The user approved the design in chat. The review skill has its own spec, SPC-0203.

## Problem

The debug skill asks for evidence, but its strongest checks are optional and two of its rules cannot both be followed.

- Phase 2 names `git bisect`, and the ground rules say Phases 1 to 3 change no file in the repo. Bisect checks out other commits in the working tree. This is BUG-0074.
- The skill says probes live in a temp folder and never says how to make one or how to check the shell is inside it. An agent whose temp folder was never made wrote to a tracked file. This is BUG-0068.
- The control in Phase 3 runs only "when you can". No step asks what result would show the stated cause is wrong.
- A Phase 3 report has no fixed shape, so the command that reproduces the bug is often missing from the final reply. This is scratch item SCR-0092.
- Phase 2 has no step for what changed lately, and Phase 3 has no step that looks for the same defect in other callers.
- The skill's description covers a hang, and no recipe names the evidence for one. The rule "Three failed fixes" sits in Phase 3, which writes no fix.

Earlier text changes to this skill lost against the old text on the eval cases: the probe-copy line (reverted in PLN-0221) and the independent check (SPC-0191, closed). Six of the seven `debug-*` eval cases are recorded as flaky. The wiki page `skill-escape-lines-skip-proof.md` holds what those runs showed.

## Changes

1. Baseline first. Before any text changes, run each of the seven `debug-*` cases 10 times on today's skill text and record the pass counts in the plan with `acta state set <plan id> findings`.
2. Bisect in a temp clone. Phase 2 says bisect runs in a clone, and shows the command: `git clone . "$TMPDIR/<name>"`, change into it, then `git bisect`. The repo's own working tree is never bisected.
3. A probe folder that is checked. The ground rules show one command pattern an agent can copy: make the folder with `mkdir -p "$TMPDIR/<name>"`, change into it, and check `pwd` prints that folder before any command that writes. The text has no fallback line that allows less work when a step fails.
4. Required proof in Phase 3.
   - The control is required. A report whose control could not run says why under Unverified.
   - Before stating the cause, write the result that would show the cause is wrong, then run that probe.
   - The report has four fixed lines: Cause (with `file:line`, or "not proven" plus what was ruled out), Repro (a command the user can paste), Evidence, Unverified.
5. Three smaller additions.
   - Phase 2 starts with what changed lately: `git log` on the files involved, dependency versions, environment.
   - Phase 3, once the cause is proven, searches for the same pattern in other callers and lists what it found.
   - A hang recipe: take a stack dump of the stuck process and name who waits on whom. "Three failed fixes" moves to Phase 4.

## Rule for keeping a text change

Each numbered change is judged on the `debug-*` cases, 10 runs per arm, old text against new text. The cases are measured once for the whole plan first; a single change is measured on its own only when the whole scores below the old text on a case. A change that scores below the old text on any of those cases is left out, and the plan's `## Review notes` say which. Change 2 removes a contradiction and adds no new behaviour to measure, so it is kept when the cases that use bisect do not get worse.

An outcome where only some of the changes pass is a correct result of this spec.

## Files

- `plugin/skills/debug/SKILL.md`, `root-cause-tracing.md`
- One new case folder, `plugin/evals/debug-hang`
- `internal/plugincheck/skill_debug_test.go` for the required strings, and `internal/plugincheck/budget_test.go` for the byte caps. `SKILL.md` is at 6.4K of a 6546 byte cap, so the cap goes up. No words an agent acts on are trimmed to fit.
- Wiki pages that the change makes wrong, found by `acta wiki check` at the close of the build.

## Testing

- The baseline of change 1, then the new text at 10 runs per case. That is about 150 runs when no case gets worse, about 20 minutes at `scripts/eval -j 8`, and it costs quota.
- `internal/plugincheck` keeps the required strings and the size caps.

## Closes

BUG-0074 when change 2 lands. BUG-0068 when change 3 lands. SCR-0092 is the parent.

## Out of scope

The four phases as a structure, the race recipe, the review skill, and DBT-0160 (a `GIT_DIR` leak in a script that was reverted).
