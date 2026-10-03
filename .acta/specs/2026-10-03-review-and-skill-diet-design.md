---
parent: scratch/2026-10-03-review-diet
closes: [SCR-0039]
id: SPC-0069
created: "2026-10-03 20:58:15"
hash: vt4gf9q
started: "2026-10-03 21:05:18"
---
# Review diet, dispatch command and skill diet

Status: design approved by the user in chat on 2026-10-03, section by section. Architectural: it changes how review runs, adds a CLI command the build skill depends on, and rewrites the skills every cycle loads. SCR-0039 (skill diet) was merged into this brainstorm by the user.

Steps 4 and 5 of the 2026-10-03 roadmap. Goal: acta costs fewer tokens than superpowers, mattpocock/skills and gstack for the same guarantees. Measured on PLN-0077: subagents used about 1.07M tokens, review took 54%, and a polish of four tiny fixes took about 308k.

Every piece of text taken from somewhere else is rewritten, never copied. Ideas from anthropics/claude-code are "All rights reserved", so only ideas are used there, no text.

## Plans

Three plans, run one after another. All three touch `internal/plugincheck/budget_test.go`, so they cannot run side by side.

1. `acta dispatch send` and `acta dispatch close`, and one merged dispatch file.
2. Review diet.
3. Diet for slice, build `SKILL.md`, tdd, debug, land, `implementer-prompt.md`, the 12 skill descriptions, and the two PLN-0077 debt items (DBT-0069).

Dispatch goes first: it is the biggest cost, and `review/SKILL.md` names the dispatch command, so that line is written once. Each plan ends by adding 1 to the patch version (0.1.5 after all three).

## 1. `acta dispatch send` and `acta dispatch close`

Today the orchestrator types about ten herdr steps by hand and writes the brief itself. The two dispatch files cost about 10.8k tokens per build. Dispatch is the user's default executor.

**Command.** `acta dispatch send --plan <path> [--round <slug>] [--note-file <path>|-]`, run in the worktree, with `HERDR_ENV=1`. New file `internal/cli/dispatch_send.go`. It reuses the record code of `acta dispatch init`.

In order:

1. Read the plan: each task with its `verify:` line and its wave. With `--round`, the tasks are the ones in the last `## Fix round <n>` section. For `--round polish`, the `[fix]` NOTEs come from the note. A task with no `verify:` line stops the command with an error, before anything is sent.
2. Write the dispatch record, the same as `dispatch init` (base = HEAD now).
3. Write the brief to `.claude/dispatch/<slug>-brief.md`, where the slug is the branch name. The brief holds: plan and spec paths, the tasks with verify lines and waves, worktree, branch, parent, base, the absolute path of `references/house-rules.md`, the absolute paths of both `MEMORY.md` files, the plan's gate command, the note text, and the REPLY-BACK line.
4. Find the tab: `herdr agent get <slug>`. Found: reuse it. Not found: create a tab with `--workspace` (the orchestrator's) and `--no-focus`, run `omp` with no argument in the worktree, and rename the agent to the slug. A fresh omp process is a new session, so `/new` is never sent.
5. Wait until omp is ready (its empty input box is on screen), 60 seconds at most. A held goal (`⏸ Goal`, or the "Resume the current goal first" warning) is cleared with `/goal drop` and two enters.
6. Send `/goal` with a one-line summary, `ultrathink orchestrate`, the brief path and the REPLY-BACK line. Wait for `🎯 Goal` in the status bar; try once more if it does not show. `omp "/goal ..."` as a launch argument does not start goal mode (omp v18.4.4), which is why the goal goes through `herdr agent prompt`.
7. Wait about 20 seconds, read the pane once, and check that every task id shows. Print one line: `checkpoint ok`, `checkpoint drift: missing <ids>`, or `checkpoint unconfirmed` when no todo list is on screen yet. The pane text is printed only on drift. The old wave check is dropped: the brief now carries the plan's waves.
8. Print about six lines: slug, pane, base, brief path, checkpoint, and the exact watcher command to run in the background.

Exit codes: 0 for ok or unconfirmed, 1 for bad input with nothing sent (1 already means bad input in this CLI), 2 when delivery failed, with herdr's reason, 4 for drift.

**`acta dispatch close`** finds the pane from the slug and runs `herdr pane close`.

**Skill file.** `build/dispatch.md` and `build/herdr-delivery.md` become one rewritten `build/dispatch.md`, about 1.5-2k tokens. It says: run `send`, read its lines, end the turn and never wait, the idle watcher and its four steps, verify from git, fix rounds with `send --round`, `close` at land, then harvest omp memory. Dropped: the magic keyword section (now in code), the advisor section (one line stays), "Spread the work" (already in `house-rules.md`), the property section (`acta:slice` owns it), the stopping rule (`acta:review` owns it), the brief format and the operation map.

**Tests.** Go tests with the existing `fakeHerdr` in `internal/cli/dispatch_test.go`, extended so each pane read can return scripted text. Cases: new tab and reused tab, omp never ready (timeout), held goal, goal mark never shows, checkpoint ok, drift and unconfirmed, a task with no verify line. No eval: the eval sandbox has no herdr.

## 2. Review diet

**`review/code-reviewer.md`**, rewritten to about 0.5k tokens (from about 1.0k):

- Axis, range and plan path.
- Read the diff first. Open another file only to confirm a finding. Check callers only when a signature or a behaviour changed.
- Do not report: issues already on the parent branch, what a linter or type check catches, nitpicks, issues silenced in code on purpose (a comment or an ignore says so), wishes for more coverage the plan does not ask for.
- The BLOCKER bar is the review skill's. At most five NOTEs. Each NOTE names a written rule or a concrete failure.
- Never spawn a subagent. Output: the model, BLOCKERs, NOTEs with their bucket tags, and CLEAN or BLOCKED as the last line.
- Dropped: the generic checklist, "acknowledge strengths", the Answers section (the orchestrator answers the three questions).
- **Polish block:** the reviewer gets only the polish range and the `[fix]` list, no spec or plan. The Spec axis checks each item was done as written. The Standards axis checks nothing broke on the changed lines. Two reviewers stay.
- **Validator block:** the validator gets every BLOCKER from both reviewers, merges duplicates, and runs each one's concrete input again. Output per BLOCKER: `valid` with the proof, or `dropped` with the reason.

**`review/SKILL.md`**, about 1.6k tokens (from about 2.9k):

- A BLOCKED round runs one validator before the fix round, on the orchestrator's model. The validator is not a round. When every BLOCKER is dropped, the round is CLEAN and `## After a CLEAN round` runs. A dropped BLOCKER that still points at something real becomes a NOTE.
- "Receiving findings" keeps its heading and becomes about five lines: check each finding against the code; push back with technical reasons when the reviewer is wrong; ask about unclear items before doing any of them; no praise and no thanks, state the fix; one item at a time, test each.
- A polish sent to another agent goes out with `acta dispatch send --round polish`.
- The buckets, the three-round cap, the small-change rule and two reviewers on a polish all stay. Only the wording gets tighter.

## 3. Skill diet

Rewrite targets, in tokens (bytes / 4):

| File | Now | Target |
|---|---|---|
| `slice/SKILL.md` | 3.0k | 1.8k |
| `build/SKILL.md` | 3.8k | 2.0k |
| `tdd/SKILL.md` | 2.5k | 1.5k |
| `debug/SKILL.md` | 2.6k | 1.6k |
| `land/SKILL.md` | 1.4k | 0.9k |
| `build/implementer-prompt.md` | 1.9k | 1.0k |

One cycle (shape, slice, build, tdd, review, land) goes from about 18k to about 12.2k. Shape stays at 4.4k here; SCR-0037 owns it. With shape at about 2.5k later, a cycle is about 10.3k.

Rules for every rewrite:

- Every rule a plugincheck `Must` list pins still holds. Its wording may change only when the skill text and the test phrase change in the same commit. No grader is loosened.
- Each file's cap in `budget_test.go` drops to its new size in the same commit, so it cannot grow back.
- Word counts before and after go in the plan's Review notes.

**Descriptions.** Each of the 12 skill descriptions is about 160 bytes (40 tokens) at most: when to use the skill and its trigger words. The rest moves into the skill body. `descriptionCaps` drops to match. No new trigger evals.

**Debt DBT-0069.** Drop `SkillRule.MaxLines`, so a skill has one size cap, in bytes. Add `*.md text eol=lf` to `.gitattributes`, so byte caps hold on CRLF checkouts.

Out of scope: the shape skill, on-demand reference files (`tdd/writing-good-tests.md`, the three debug references), and a benchmark across plugins.

## Proof of same quality

Everything below runs in this repo only. Nothing new ships to user repos except the leaner skill text, the new review rules and the new command.

- At every land: `scripts/test --full`, every old eval case green, plugincheck green, word counts before and after.
- New eval cases in plan 2: a BLOCKED reviewer report makes the agent run the validator before any fix round; a finding in code already on the parent branch goes to `acta:bug`, not to a fix round.
- After plan 2 lands, one replay on the PLN-0077 range (`64b934f..c8f5e24`): old review against new review, total tokens and the findings side by side. The numbers go in plan 2's Review notes.
