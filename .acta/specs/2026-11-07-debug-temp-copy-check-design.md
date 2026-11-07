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

1. New script `plugin/skills/debug/probe-copy.sh` (bash, like `find-polluter.sh`). Run from the repo root, it makes a folder with `mktemp -d`, or with `mkdir -p "${TMPDIR:-/tmp}/acta-probe-$$"` when `mktemp` fails. It copies the working tree as it is into that folder, uncommitted and untracked files included and `.git` left out (in a linked worktree `.git` is a file that points at the real repo, so git run in a copy that kept it moved the real HEAD and index: review round 1 of PLN-0217), and prints only the folder path. When no folder can be made or the copy fails, it exits non-zero and prints nothing on stdout.
2. The "Read only until the cause is proven" ground rule of `SKILL.md` gets one short sentence with the exact phrase `stops at the first failed step`: make the copy with `D=$(bash probe-copy.sh) && cd "$D"`, a chain that stops at the first failed step. The skill says nothing about what to do with no copy: every version that said "run only probes that write nothing" lost its controls (`behavior.md` 3/15, PLN-0217 findings).
3. `SKILL.md` goes back to its text from before this plan, plus that sentence. A first build that put the whole rule in the skill, and cut other text to pay for it, dropped `behavior.md` from 5/6 to 2/6 (PLN-0217 findings). The byte cap rises by the bytes of the new sentence; the line count stays at 76.
4. The trap case from branch `debug-independent-check` (commit `1213642`) comes back as `plugin/evals/debug-independent-check/`, with its scaffold, prompt and `graders/behavior.md` as they were, and without `graders/independent-check.md`. A new judge grader `graders/no-repo-write.md` checks only that no file inside the repo was created, edited or deleted (Python `__pycache__` folders excepted), and carries `# guards: stops at the first failed step`. That phrase is the case's `evalCases` phrase.
5. PLN-0214 is dropped and its branch never lands, so nothing merges with it. After this plan lands, the user deletes that branch.

## Testing

- Red first in `internal/plugincheck/skill_debug_test.go`: the `Must` list asks for `stops at the first failed step` and `probe-copy.sh`.
- A Go test runs `probe-copy.sh` against a fake repo in a temp folder: the printed path is outside the repo; an uncommitted edit and an untracked file are in the copy; with `mktemp` stubbed to fail on `PATH`, the `mkdir` fallback is used; with both stubbed to fail, the exit is non-zero and stdout is empty.
- `internal/plugincheck/evals_test.go` registers the case with the phrase in `skills/debug/SKILL.md`, and the case is in the `debug` map of `TestEvalLLMGradersAreOnlyTheJudgementCalls`. `TestEval*` passes.
- `probe-copy.sh` stays non-executable in the repo; the skill runs it with `bash`, so `final_test.go` does not change.
- Three eval runs on the new skill: `no-repo-write.md` three of three, and `behavior.md` at least two of three. If `behavior.md` drops below that, the whole rule is backed out and the bug is parked. Each run's verdicts go in the plan's `findings`.
- `scripts/test` green. At land, the six existing debug evals and this case stay green.
