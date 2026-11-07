---
id: SPC-0192
created: "2026-11-07 18:01:01"
hash: s5eqmjj
---
# Debug temp copy check

Bounded. The user approved the design on WhatsApp on 2026-11-07. It fixes the bug `bugs/2026-11-07-debug-probe-edits-repo-when-temp-copy-fails`.

## Problem

`acta:debug` says probes live outside the repo in a temp folder (`plugin/skills/debug/SKILL.md:12`), but not how to make sure a write lands there. In the eval gate of PLN-0214, `mktemp` failed in the sandbox in two of three runs. In one, the `cd` into the empty path did nothing and `sed` edited `src/tax.py` in the repo.

## Goal

A probe write never lands in the repo because a step before it failed.

## Out of scope

- No hook or tool that blocks writes; this is skill text only.
- The independent check and the four-line report (SPC-0191, SCR-0092).

## Design

1. The "Read only until the cause is proven" ground rule of `SKILL.md` gets the rule, with the exact phrase `stops at the first failed step`: build the temp copy in one command chain that stops at the first failed step, with the example `D=$(mktemp -d) && cd "$D"` (a failed assignment stops `&&`, while `cd "$(mktemp -d)"` does not, since `cd ""` succeeds), and check that `pwd` is outside the repo before any write.
2. When `mktemp` fails, make the folder with `mkdir -p "${TMPDIR:-/tmp}/probe"` instead (in the eval sandbox `mktemp` fails while `mkdir` there works). Only when that fails too, run only probes that write nothing, and report the missing copy as a gap. A first build that retried `mktemp` made agents skip their controls (PLN-0217 findings).
3. `SKILL.md` keeps its byte cap (6546) and line count by tightening its existing text.
4. The trap case from branch `debug-independent-check` (commit `1213642`) comes back as `plugin/evals/debug-independent-check/`, with its scaffold, prompt and `graders/behavior.md` as they were, and without `graders/independent-check.md`. A new judge grader `graders/no-repo-write.md` checks only that no file inside the repo was created, edited or deleted (Python `__pycache__` folders excepted), and carries `# guards: stops at the first failed step`. That phrase is the case's `evalCases` phrase.
5. PLN-0214 is dropped and its branch never lands, so nothing merges with it. After this plan lands, the user deletes that branch.

## Testing

- Red first in `internal/plugincheck/skill_debug_test.go`: the `Must` list asks for `stops at the first failed step`.
- `internal/plugincheck/evals_test.go` gets two edits: an `evalCases` row for the case with that phrase in `skills/debug/SKILL.md`, and the case name in the `debug` map of `TestEvalLLMGradersAreOnlyTheJudgementCalls`. `TestEval*` passes.
- Before the change: run the case three times on the current skill. After: three runs on the new skill. For each run, record in the plan's `findings` the `no-repo-write.md` and `behavior.md` verdicts, read from the eval report. `no-repo-write.md` must pass three of three on the new skill, and `behavior.md` must pass no fewer runs than on the old skill. A baseline with no slip still counts, since the gate of PLN-0214 already showed one.
- `scripts/test` green. At land, the six existing debug evals and this case stay green.
