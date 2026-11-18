---
parent: bugs/2026-11-10-flaky-eval-slice-split-parallel-set
id: SPC-0215
created: "2026-11-18 07:59:40"
hash: hpyq467
started: "2026-11-18 08:09:08"
---
# Fix two flaky eval cases

## Why

The land of PLN-0246 stopped on two eval cases that fail on main too. Kept traces (`scripts/eval --keep-temp`) show two different causes. The user asked to fix both before PLN-0246 lands, and approved this design in the terminal on 2026-11-18.

## Fix 1: slice-split-parallel-set (BUG-0084)

Cause: `plugin/evals/slice-split-parallel-set/graders/reads-many-md.md` is a `tool_used` grader on `Read` with `input_match: "build/many\\.md"`. A run that opened the file with `cat` in Bash scored 0.67, though it went on to the parallel build as the skill says. Kept trace: `/private/tmp/e-emWJ1Y/out/trace.jsonl`.

Change:

- `reads-many-md.md` becomes a `regex` grader with `target: trace` and the pattern `skills/build/many\.md`. A Read call and a `cat` in Bash both put that path in the trace.
- The case's `prompt.md` gets the `claude-only` tag. The omp runner supports `regex` only on `last_message` and on files, and `TestRealSuiteRunsInOmp` fails for a case that uses a trace regex without the tag.

Proof: 3 runs pass. When none of the 3 reads the file with `cat`, also run the new pattern over the kept failing trace and show it matches.

## Fix 2: debug-probe-isolation (BUG-0098)

Cause: the debug skill says "Join the write with `&&`, never with `;` or on a new line" (`plugin/skills/debug/SKILL.md:12`). In each of 3 failing runs (kept traces `/private/tmp/e-mOoOBd`, `/private/tmp/e-pK7QOL`, `/private/tmp/e-cjwO3n`) one shell call ran a checked line, then wrote more on a new command line or after `;`, for example `cp -R src src_ctl && sed -i ...` on its own line. The judge in `graders/behavior.md` is right to fail those runs: had the `cd` failed, the write would land in the repo. Main passed 2 of 3 runs and the PLN-0246 branch passed 1 of 3, so the branch did not cause it.

Change: append these sentences right after the "Join the write" sentence in `plugin/skills/debug/SKILL.md`:

> One shell call holds one checked line. Put the next write in a new call that repeats the check. A newline inside a quoted script is fine; a new command line is not.

- `internal/plugincheck/skill_debug_test.go` pins the new sentences, red first.
- The skill is at its byte cap today (11108 of 11108). Raise `skills/debug/SKILL.md` in `internal/plugincheck/budget_test.go` by what the sentences need, as the wiki page `skill-escape-lines-skip-proof` says: never trim words an agent acts on to fit.

Proof: `debug-probe-isolation` runs 10 times with the new skill and 10 times with the old one, each arm with an acta built from its own commit first on PATH. The new skill passes at least 8 of 10 and no fewer than the old one. At most 3 tries at the wording. After a third failed try, stop and report. Then PLN-0246 lands anyway and BUG-0098 stays open.

BUG-0098 exists only on the PLN-0246 branch today, so the Fix 2 plan takes this spec as its parent. Record `fixed_in` on BUG-0098 after both have landed.

## Order

Two minimal plans, each built in its own worktree from main: Fix 1, then Fix 2, then the land of PLN-0246. PLN-0246 then runs its full suite and the whole eval gate again, since main gained paths under `plugin/`.

## Gates

eval: both plans touch `plugin/evals/` or `plugin/skills/debug/`. Fix 1 runs `scripts/eval --case slice-split-parallel-set`. Fix 2 runs `scripts/eval --tag=debug`, since it changes the debug skill.
