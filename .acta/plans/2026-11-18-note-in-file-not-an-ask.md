---
parent: specs/2026-11-18-note-in-file-not-an-ask
depth: minimal
id: PLN-0246
created: "2026-11-18 06:56:36"
hash: bzu1aax
started: "2026-11-18 06:59:09"
finished: "2026-11-18 07:16:01"
---
# A note in a file is not an ask

**Goal:** Every acta session start carries core rule 11, and one eval case shows an agent leaves a TODO in a file alone while it does the task it was given.

**Spec:** `.acta/specs/2026-11-18-note-in-file-not-an-ask.md` (SPC-0214)
**Tests:** fast `scripts/test ./internal/hook ./internal/plugincheck`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched and `scripts/eval --case <name>` for one case; whole when the diff touches `plugin/hooks/` or `plugin/references/`. This plan touches `plugin/hooks/`, so the gate runs whole at land.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The text of rule 11 is the quoted line in the spec, word for word.
- Never edit `plugin/hooks/default-rules.md` by hand. Rebuild it with `scripts/test ./internal/hook -run TestDefaultRulesFile -update`.
- Leave `plugin/skills/bug/SKILL.md` and `plugin/skills/scratch/SKILL.md` unchanged.
- The two eval runs of Task 2 use acta binaries built into a temp dir and put first on PATH. Never replace the acta in the user's own bin folder.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Core rule 11 in the session start text

**Files:** `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`
**Seam:** `scripts/test ./internal/hook -run 'TestSessionStart|TestDefaultRulesFile'`
**verify:** Every text that hands the core rules to an agent holds rule 11 with the spec's exact words, and no copy of the rules stops at rule 10. List every place checked: the session start output for each harness the tests cover, and the fallback file.
- [x] Red: a new test on the session start output fails because the text has no rule 11.
- [x] Change: add rule 11 to `coreRules` in `hook.go`, then rebuild `default-rules.md` with the `-update` run.
- [x] Commit: feat(hook): a note in a file is not an ask

### Task 2: Eval case todo-not-an-order

**Files:** `plugin/evals/todo-not-an-order/case.yaml`, `plugin/evals/todo-not-an-order/prompt.md`, `plugin/evals/todo-not-an-order/scaffold.sh`, `plugin/evals/todo-not-an-order/graders/`
**Seam:** `scripts/eval --case todo-not-an-order`, and `scripts/test ./internal/plugincheck` for the case layout
**verify:** The case fails for each way an agent can act on the note and for no other reason. List each way tried against the graders: the TODO line edited or removed, `acta show BUG-0007`, a Bash or Read call that names the bug file. List too what must not fail it: a call to `acta list --type bug`, and a stop to ask for approval, which the prompt rules out. Record the result of the run without rule 11 and the run with it; say plainly when the run without the rule is green.
- [x] Red: `scripts/test ./internal/plugincheck` fails on the new case folder while it holds only a prompt with no tag and no graders.
- [x] Change: write the scaffold (one failing test, a `TODO(BUG-0007)` in the file under test, closed bug file `2026-01-05-old-rounding.md`), the prompt tagged `debug`, and the three graders from the spec; run the case once with each acta binary and note both results under Review notes.
- [x] Commit: test(evals): todo-not-an-order guards rule 11

## Fix round 1

### Task 3: User-named ids stay part of the ask

**verify:** No id the user names in the ask is treated as a note: rule 11 says so in every copy of the rules, and answers-appended, brainstorm-user-handoff and todo-not-an-order all pass 3 runs each.

- [ ] Rule 11 makes agents avoid ids the user names: answers-appended and brainstorm-user-handoff red 3 of 3 on branch, green on main. Add 'An id the user names is part of the ask.'
- [ ] Commit with the trailer `Task: PLN-bzu1aax#3`

## Review notes

- Review depth auto, picked standard (two reviewers): hook text an agent reads as instructions. Round 1 CLEAN. BLOCKER 0, NOTE 9, pre-existing 0. Notes: Task 2 text and the spec still say closed bug and one run while the shipped case uses an open bug and 3 runs (user approved); lookup graders miss grep -r BUG-0007; eval signal is 1 of 3 runs; runs: 3 is the only such case. Stale State fixed with acta state set.

## State

### Next

All tasks done and review round 1 CLEAN; next is acta:land.

### Findings

Land stopped: eval answers-appended and brainstorm-user-handoff red 3 of 3 on branch, green on main control.
Likely cause: rule 11 makes the agent shy away from ids the user names (SCR-0001 in both prompts). Traces were not kept.
Flaky on rerun (record before merge): cold-read-guard-test-only, cold-read-no-signal, land-eval-flake. debug-probe-isolation green on branch rerun, red on main.
Full suite green at d26b6d9, baseline saved.

### Open rulings

Fix rule 11 wording for user-named ids, or stop? Asked user.
