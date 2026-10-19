---
name: debug
description: "acta: Use for any bug, error, failing test, stack trace, wrong output, slowness or hang - and whenever the user says something is broken - before reading code for a fix or editing anything. Four gated phases - root cause, pattern, hypothesis, then a fix with a regression test."
---

# Debugging

Find the cause with evidence before changing code. A cause is proven only when a command you ran shows it. Spend effort in line with doubt: a clear bug needs one good probe, an unclear one needs ranked guesses and probes that tell them apart.

## Ground rules

- **Read only until the cause is proven.** Phases 1 to 3 change no file in the repo. Probes, scripts and instrumented copies live outside the repo, in a temp folder.
- **Redact secrets** in commands, output, saved files and web searches. Print whether a secret is set, never its value.
- **No live systems.** Do not change production, shared services or their logging without the user's yes, and never push. When only the user can reach a system, give them the exact commands to run.
- **Name what you saw.** Every report splits three things: observed (a command ran and printed it), hypothesis (not yet tested), and unverified (not reachable from here).
- **A confirmed bug gets a file.** Once the cause is proven (`file:line` plus a repro), or a bug is found while reading and reported to the user, record it with `acta:bug`.
- **Subagent models.** Name no model for any subagent; the harness default picks it.

## Phase 1: Symptom and feedback loop

Exit: one command you ran catches the exact symptom the user reported.

1. Write down expected and actual behaviour, and where it happens (versions, config, data, environment).
2. Build the loop: an existing test or the user's own command, else a small replay script. A command that writes files runs only in a temp copy.
3. Check the output itself, not only the exit code. A pass on a nearby failure (another test, error or caller) is not the symptom.
4. Symptom fits a recipe below (flaky, race, order, slow)? Pick it now; its evidence gates Phase 3. Shared state under load comes up short or varies: race, so read `race.md` now.
5. Cannot reproduce, or the symptom lives where you cannot reach? Stop short of a cause. Say the cause is not proven, list what you ruled out locally, and ask for specific evidence: exact values, logs for a time window, a row count, a deploy diff. "Check the logs" is not specific.

## Phase 2: Narrow and distinguish

Exit: one explanation fits every observation, and the others are ruled out by a probe.

1. Cut inputs and steps while the failure still shows. Stop once the cause is clear; full minimization is not required.
2. Compare a working path with the failing one: a similar caller, an older commit (`git bisect`), other data. List every difference.
3. When more than one cause fits, rank them and write what each predicts. Run the probe whose result splits them best. Change one thing at a time.
4. For a bug deep in a call stack, trace the bad value back to where it starts: see `root-cause-tracing.md`.

## Phase 3: Prove the mechanism

Exit: trigger, state change and symptom are linked with `file:line` and probe output, plus the evidence the matching recipe names, from a command you ran.

1. State the cause: "X at `file:line` does Y, so Z." Back each link with output you saw.
2. Use a control when you can: the same probe with the suspected cause removed (in the temp copy) makes the symptom go away.
3. When several tries fail under one shared belief, test that belief directly.
4. Three failed fixes: stop and tell the user. Question the design before a fourth.
5. Report the cause, the evidence, and what stays unverified. Propose the regression test the matching recipe below asks for. Do not write the fix in this skill.

## Phase 4: Fix, verify and clean up

A fix is code, so it goes through `acta:shape`, a plan, and its own worktree through `acta:build`.

1. The repro becomes the regression test (`acta:tdd`). Test the real bug pattern at the layer that owns it. Watch it fail, then pass.
2. Re-run the user's original command or scenario, plus nearby checks.
3. Remove temp probes and instrumentation. Report what remains unverified.
4. Use `acta:land` before claiming it is fixed. Missing test coverage is a gap, never proof of a fix.

A rerun that turns green, a longer timeout, or a skipped assertion does not prove a fix.

## Recipes

Each names the evidence that counts.

- **Race.** Run the probe in `race.md`, which must force the interleaving: each worker waits after its read until the other has read. A barrier at start does not. A lock control proves the cause but is not a test. The proposed regression test holds the workers the same way and asserts the final invariant (for example count equals updates). More threads, sleeps, repeats or "no exception" only make it likely: weak evidence, weak tests.
- **Flaky test.** Run it many times under the same conditions and record the failure count and what changed between runs (load, order, seed, time). For waits, see `condition-based-waiting.md`.
- **Order-dependent test.** Find the polluter by running pairs or bisecting the suite (`find-polluter.sh` finds tests that leave files behind). Put the smallest command that keeps the failing order in the report as a line the user can paste, for example the polluter then the victim, and name the shared state it leaves changed.
- **Slow path.** Measure the same workload more than once, at two sizes if you can, and count the work (calls, file reads, queries) or profile it. Name the cost that grows. A speedup you did not measure before and after on the same workload is unmeasured; say so. Prefer a test that counts work over a wall-clock limit.
- **Checks after the fix.** Add checks at trust boundaries and where an invariant must hold, not at every layer: see `defense-in-depth.md`.

## Signals you are off track

- About to edit code before Phase 3 exits.
- "It's probably X" with no probe behind it.
- Several changes at once, or one more fix after two failed.
- The user asks "is that actually happening?" or "stop guessing": go back to Phase 1.
