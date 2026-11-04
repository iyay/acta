---
name: review
description: "acta: Use once, when every task of a plan is committed, on each fix round, and over a polish commit. By `## Review tiers`, runs the read-only reviewers its depth names (Spec and Standards axis) or a light inline review, over an explicit git range; sorts findings into BLOCKER or NOTE; stops at three rounds. CLEAN-round flow: `## After a CLEAN round`. Also covers receiving findings."
---

# Review

Review happens once, when every task of the plan is committed, then once per fix round. Never after each task. A CLEAN round runs the polish flow in `## After a CLEAN round`.

## Review depth

Pick the level in this order, and ask no question when one answers: the argument of `/acta:review quick|standard|deep` (this run only), then `review_depth` from `acta config show`, then `standard`. The levels, the floors and the review-notes line are in [depth.md](depth.md); the deep level adds [fact-check.md](fact-check.md). No setting drops a floor. `review_understanding` is a separate switch: no level turns it on, the review asks nothing about it, and `acta:land` adds what it asks for.

## Two reviewers, in parallel

The standard and deep levels (`## Review depth`) dispatch two read-only reviewer subagents at the same time, both from [code-reviewer.md](code-reviewer.md):

**Subagent models.** The Spec-axis reviewer runs on your own model. The Standards-axis reviewer runs on a cheaper, faster model tier than yours, for example sonnet in Claude Code. If the harness has no cheaper tier, use its default.

- **Spec axis:** is the change faithful to the approved spec and plan text? It judges against that text, not against a better design it can imagine.
- **Standards axis:** do the changed lines follow the repo's own standards, and did anything that was working break? It checks the callers of what changed.

Give each an explicit range: `BASE_SHA=$(git rev-parse <parent>)`, `HEAD_SHA=$(git rev-parse HEAD)`, and the plan path. Reviewers never edit, stage, commit or run a formatter; they may run the narrow tests that prove a finding, never the full suite. Ask them to test a property, not to hunt for any case you did not test.

## Finding bar: BLOCKER or NOTE

A **BLOCKER** is one of: wrong output for a real user today on a path the plan covers; data loss or corruption; a security hole (a user value reaching SQL, shell, a path, a template or HTTP unchecked, or an auth bypass); production cannot run the change. It must carry `file:line` and a concrete input that gives the wrong output. No reproducible scenario means it is not a BLOCKER.

Everything else is a **NOTE**: one line each. A NOTE is not a task, unless it carries the `[fix]` tag: see "Where findings go". Always a NOTE: comments, naming, wording, formatting, log text, test-helper design, guards for inputs the plan does not name, "could be more robust", duplicated code that works, a missing test for a path the plan does not cover.

## The three questions, and the deep lens

Outside the deep lens, the review answers three questions and stops:
1. Does it match the approved spec and plan?
2. Does it run: tests and type checks green, output shown?
3. Do the tests go red when the change is reverted?

The deep lens (trace every user value to its sink, check every caller, adversarial inputs, error swallowing, cross-tenant access) applies only to trust boundaries, auth, money, migrations and deletes.

Each reviewer reports: the model it ran on, BLOCKERs ranked (five at most), NOTEs as a flat list, and a last line of one word: CLEAN or BLOCKED.

## Budget: three rounds, then land or ask

Each NOTE is sorted into a bucket: `[fix]`, `[debt]` or `[note]`. Three rounds at most.

1. Round 1 reviews `<parent>..HEAD`. CLEAN, or NOTEs only: run `## After a CLEAN round`, which ends in `acta:land`.
2. BLOCKERs: one fix task holding all of them, added to the same plan from the same worktree with `acta plan add <plan id> fix-round --title <title> --verify <line>`, one BLOCKER per stdin line, landing as one commit. It prints the task id to tick. The `[fix]` NOTEs join that fix task.
3. Round 2 reviews the fix range (`<round-1 head>..HEAD`) plus the direct callers of every function the fix touched; fixed code on a trust boundary, auth, money, migration or delete path gets the deep lens again. CLEAN: run `## After a CLEAN round`. BLOCKER: one more fix task.
4. Round 3 reviews the second fix range only. CLEAN: run `## After a CLEAN round`. BLOCKER: stop and ask the user: land anyway, fix, or revert to the round-1 head. Land anyway: that round's `[fix]` NOTEs move to `[debt]`, so none is left with no home.

There is never a round 4. A round with no BLOCKER does not start; it runs `## After a CLEAN round`. Findings in test tooling are NOTEs. A reviewer cannot re-scope the plan. A fix that opens a new hole is reverted and both are noted, unless the original was a BLOCKER. Before every commit after the deliverable is green, ask: "Without this, does a real user see a wrong result today?" No: no commit, one NOTE; the `[fix]` NOTEs are handled in `## After a CLEAN round`.

## Where findings go

- A BLOCKER in the diff under review is unfinished work of that story, not a bug. It goes into this round's one fix task (`## Fix round <n>` in the same plan).
- A defect a reviewer finds in code already on the parent branch, that the diff did not bring in, is a bug. Record it with `acta:bug`, and fix it through its own plan with `parent: bugs/<file>`. It never widens this plan.
- A NOTE is one line, and it sorts into a bucket. A NOTE that fails the first test falls to the next one.
  - **`[fix]`:** the file is already in the diff, the fix adds no new logic, the path is not security, auth, money, migration or delete, and it fits in one small commit. A comment, a name, a wording change, a missing test, a one-line guard.
  - **`[debt]`:** leaving it costs something later that you can name ("when X changes, Y breaks", "a user hits it when Z", "the tests get slow"), and the fix needs new logic or a file outside the diff. A `[debt]` NOTE must start with `(medium) ` or `(high) `. A NOTE that would be `(low)` or has no priority tag is a `[note]`.
  - **`[note]`:** taste, nits, input the plan does not name, "could be more robust" with no real scenario behind it.

**Who sorts.** Reviewers stay read-only and start each NOTE with a suggested bucket tag. The orchestrator checks every tag against the code and moves it when the tag does not hold.

## After a CLEAN round

The polish commit counts as no round. This section is the whole CLEAN-round flow.

1. Sort every NOTE into `[fix]`, `[debt]` or `[note]` with the bucket rules in `## Where findings go`.
2. With `[fix]` NOTEs: first add the `## Polish` task in the worktree with `acta plan add <plan id> polish`, one `[fix]` NOTE per stdin line. It prints the task id to tick. Then one polish commit holding all of them. The orchestrator runs the narrow tests of the code the polish touched and the formatter check, with the output shown, never the full suite; the full suite waits for `acta:land`. Then the polish range gets the review its tier picks (`## Review tiers`). When the polish or the tests fail, revert that commit, and that moves those items to `[debt]`. On a revert also tick the `## Polish` task boxes, and say so in the review notes.
3. A BLOCKER in the polish review reverts the polish commit and moves those items to `[debt]`. It starts no fix round.
4. The polish-review NOTEs sort into `[debt]` or `[note]` only, so no second polish follows.
5. The `[debt]` NOTEs go to `acta debt new <plan id>` on the branch (one NOTE per stdin line). Before the call, drop the bucket tag from each line: the debt file stores no bucket. Keep the priority: the priority tag comes after the bucket tag, so it starts the line once the tag is gone. The debt file merges with the branch. With no `[debt]` NOTE, no debt file is written.
6. The `[note]` NOTEs are not written one by one. Write one summary line under `## Review notes` with `acta plan add <plan id> review-notes`, run in the worktree, the line on stdin; it commits the plan.
7. Then `acta:land`.

A polish sent to another agent goes out with `acta dispatch send --round polish`, the `[fix]` NOTE list on `--note-file -`; when `/acta:review` arrives with round polish, run the narrow tests of the code the polish touched and the formatter check, with the output shown, then the review its tier picks (`## Review tiers`) over the polish range, then step 3 onward. The `--round polish` send hands over the plan's `## Polish` task.

## Review tiers

This is the one place that says which diffs get a full review; `## Review depth` says how deep it goes. Tiers apply to round 1 and to a polish commit. A fix round always takes the two reviewers.

- **Light tier:** every changed line is a comment, docs or markdown no agent reads as instructions, log text, or a test name. Any number of files.
- **Full tier:** everything else: any logic line, config that changes behavior, text an agent reads as instructions (skills, prompts, hook text), any change on a security, auth, data migration, money or delete path. In doubt: full.

Light review: the orchestrator reads the full diff, checks every line is in the light tier, runs the narrow tests and the formatter with the output shown, and gives CLEAN or BLOCKED. It writes one line naming the tier and why with `acta plan add <plan id> review-notes`, run in the worktree. No reviewer runs. A full review dispatches the reviewers its depth names (`## Review depth`). Reading the diff yourself never replaces the reviewers its depth names, however small the diff.

A light review that finds any line outside the light tier switches to the full review.

**Receiving findings.** Read [receiving.md](receiving.md) before acting when any reviewer returns a finding, when the user or an outside reviewer gives review feedback, or when a light inline review finds something. With no feedback, do not read it.
