---
name: review
description: "acta: Use once, when every task of a plan is committed, and on each fix round after that, and over a polish commit. Dispatches two read-only reviewers in parallel (Spec axis and Standards axis) over an explicit git range, sorts every finding into BLOCKER or NOTE, and stops after three rounds. The CLEAN-round flow lives in `## After a CLEAN round`. Also covers how to receive findings."
---

# Review

Review happens once, when every task of the plan is committed, then once per fix round. Never after each task. A CLEAN round runs the polish flow in `## After a CLEAN round`.

## Two reviewers, in parallel

Dispatch two read-only reviewer subagents at the same time, both from [code-reviewer.md](code-reviewer.md), both on your own model (never lower):

**Subagent models.** Only when `acta config show` lists `subagent_models: split` and you run in Claude Code: subagents that write code use `model: "sonnet"`; all other subagents (mapping, explore, planning help, debug investigation, spikes) use `model: "opus"`; reviewers use your own model alias. Otherwise name no model and follow the user's own config.

- **Spec axis:** is the change faithful to the approved spec and plan text? It judges against that text, not against a better design it can imagine.
- **Standards axis:** do the changed lines follow the repo's own standards, and did anything that was working break? It checks the callers of what changed.

Give each an explicit range: `BASE_SHA=$(git rev-parse <parent>)`, `HEAD_SHA=$(git rev-parse HEAD)`, and the plan path. Without a range a reviewer looks at the wrong span. Reviewers never edit, stage, commit or run a formatter; they may run the narrow tests that prove a finding, never the full suite. Ask them to test a property, not to hunt for any case you did not test.

## Finding bar: BLOCKER or NOTE

A **BLOCKER** is one of: wrong output for a real user today on a path the plan covers; data loss or corruption; a security hole (a user value reaching SQL, shell, a path, a template or HTTP unchecked, or an auth bypass); production cannot run the change. It must carry `file:line` and a concrete input that gives the wrong output. No reproducible scenario means it is not a BLOCKER.

Everything else is a **NOTE**: one line each. A NOTE is not a task, unless it carries the `[fix]` tag: see "Where findings go". Always a NOTE: comments, naming, wording, formatting, log text, test-helper design, guards for inputs the plan does not name, "could be more robust", duplicated code that works, a missing test for a path the plan does not cover.

## The three questions, and the deep lens

Outside the deep lens, the review answers three questions and stops:
1. Does it match the approved spec and plan?
2. Does it run: tests and type checks green, output shown?
3. Do the tests go red when the change is reverted?

The deep lens (trace every user value to its sink, check every caller, adversarial inputs, silent error swallowing, cross-tenant access) applies only to trust boundaries, auth, money, migrations and deletes.

Each reviewer reports: the model it ran on, BLOCKERs ranked (five at most), NOTEs as a flat list, and a last line of one word: CLEAN or BLOCKED.

## Budget: three rounds, then land or ask

NOTEs are not one pile. Each one is sorted into a bucket: `[fix]`, `[debt]` or `[note]`. The budget does not change: three rounds at most, and a round with no BLOCKER does not start.

1. Round 1 reviews `<parent>..HEAD`. CLEAN, or NOTEs only: run `## After a CLEAN round`, which ends in `acta:land`.
2. BLOCKERs: one fix task holding all of them, appended to the same plan as a `## Fix round <n>` section, in the same worktree, landing as one commit. The `[fix]` NOTEs join that fix task.
3. Round 2 reviews the fix range (`<round-1 head>..HEAD`) plus the direct callers of every function the fix touched; fixed code on a trust boundary, auth, money, migration or delete path gets the deep lens again. CLEAN: run `## After a CLEAN round`. BLOCKER: one more fix task.
4. Round 3 reviews the second fix range only. CLEAN: run `## After a CLEAN round`. BLOCKER: stop and ask the user: land anyway, fix, or revert to the round-1 head. Land anyway: that round's `[fix]` NOTEs move to `[debt]`, so none is left with no home.

There is never a round 4. A round with no BLOCKER does not start; it runs `## After a CLEAN round`. Findings in test tooling are NOTEs. A reviewer cannot re-scope the plan. A fix that opens a new hole is reverted and both are noted, unless the original was a BLOCKER. Before every commit after the deliverable is green, ask: "Without this, does a real user see a wrong result today?" No: no commit, one NOTE; the `[fix]` NOTEs are handled in `## After a CLEAN round`.

## Where findings go

- A BLOCKER in the diff under review is unfinished work of that story, not a bug. It goes into the one fix task of this round (`## Fix round <n>` in the same plan), where it already shows on the board under the story.
- A defect a reviewer finds in code already on the parent branch, that the diff did not bring in, is a bug. Record it with `acta:bug`, and fix it through its own plan with `parent: bugs/<file>`. It never widens this plan.
- A NOTE is one line, and it sorts into a bucket. A NOTE that fails the first test falls to the next one.
  - **`[fix]`:** the file is already in the diff, the fix adds no new logic, the path is not security, auth, money, migration or delete, and it fits in one small commit. A comment, a name, a wording change, a missing test, a one-line guard.
  - **`[debt]`:** leaving it costs something later that you can name ("when X changes, Y breaks", "a user hits it when Z", "the tests get slow"), and the fix needs new logic or a file outside the diff.
  - **`[note]`:** taste, nits, input the plan does not name, "could be more robust" with no real scenario behind it.

**Who sorts.** Reviewers stay read-only. Each NOTE they print starts with its suggested bucket tag. The orchestrator decides: it checks every tag against the code, the same way it checks any finding, and moves it when the tag does not hold.

**The flow.**

- A CLEAN round: no new round, no new fix task. Run `## After a CLEAN round`.

## After a CLEAN round

The polish commit counts as no round. Every commit gets the two reviewers. This section is the whole CLEAN-round flow.

1. Sort every NOTE into `[fix]`, `[debt]` or `[note]` with the bucket rules in `## Where findings go`.
2. With `[fix]` NOTEs: one polish commit holding all of them. The orchestrator runs the full test suite with the output shown, then the two reviewers review the polish range. When the polish or the tests fail, revert that commit, and that moves those items to `[debt]`. The polish commit uses no round.
3. A BLOCKER in the polish review reverts the polish commit and moves those items to `[debt]`. It starts no fix round.
4. The polish-review NOTEs sort into `[debt]` or `[note]` only, so no second polish follows.
5. The `[debt]` NOTEs go to `acta debt new <plan id>` on the branch (NOTEs on stdin, one per line). Before the call, drop the bucket tag from each line: it routes the NOTE inside the review, and the debt file stores no bucket. Keep the priority: the priority tag comes after the bucket tag, so it starts the line once the tag is gone. A NOTE may start with `(high) `, `(medium) ` or `(low) ` when it matters more or less than the rest; with no tag it is unset. The debt file merges with the branch. With no `[debt]` NOTE, no debt file is written.
6. The `[note]` NOTEs go into a `## Review notes` section in the plan file, one line each, and that section is committed with the plan.
7. Then `acta:land`.

A polish sent to another agent goes out with `acta dispatch send --round polish`, the `[fix]` NOTE list on `--note-file -`; when `/acta:review` arrives with round polish, run the full test suite with the output shown, the two reviewers over the polish range, then step 3 onward.

## Small changes

Small means one file, and only text or config with no code logic. Two or more files, or any code change, take the two reviewers. A small change may take an inline self-review instead: read the full diff, answer the three questions, state a verdict. Security, auth, data migration and money always take the two reviewers. Work another agent wrote (the `dispatch` executor) always takes the two reviewers: you did not watch it being written. A polish commit is never a small change: it goes through `## After a CLEAN round`.

## Receiving findings

Verify each finding against the code before acting. Push back with reasoning when a reviewer is wrong; a wrong finding routed as a task wastes a round.

### The Response Pattern

```
WHEN receiving code review feedback:

1. READ: Complete feedback without reacting
2. UNDERSTAND: Restate requirement in own words (or ask)
3. VERIFY: Check against codebase reality
4. EVALUATE: Technically sound for THIS codebase?
5. RESPOND: Technical acknowledgment or reasoned pushback
6. IMPLEMENT: One item at a time, test each
```

### Forbidden Responses

**NEVER:**
- "You're absolutely right!" (explicit instruction-file violation)
- "Great point!" / "Excellent feedback!" (performative)
- "Let me implement that now" (before verification)

**INSTEAD:**
- Restate the technical requirement
- Ask clarifying questions
- Push back with technical reasoning if wrong
- Just start working (actions > words)

### Handling Unclear Feedback

```
IF any item is unclear:
  STOP - do not implement anything yet
  ASK for clarification on unclear items

WHY: Items may be related. Partial understanding = wrong implementation.
```

**Example:**
```
your human partner: "Fix 1-6"
You understand 1,2,3,6. Unclear on 4,5.

❌ WRONG: Implement 1,2,3,6 now, ask about 4,5 later
✅ RIGHT: "I understand items 1,2,3,6. Need clarification on 4 and 5 before proceeding."
```

### When To Push Back

Push back when:
- Suggestion breaks existing functionality
- Reviewer lacks full context
- Violates YAGNI (unused feature)
- Technically incorrect for this stack
- Legacy/compatibility reasons exist
- Conflicts with your human partner's architectural decisions

**How to push back:**
- Use technical reasoning, not defensiveness
- Ask specific questions
- Reference working tests/code
- Involve your human partner if architectural

**If you're uncomfortable pushing back out loud:** Name that tension, then tell your partner about the issue you've seen. They'll appreciate your honesty.

### Acknowledging Correct Feedback

When feedback IS correct:
```
✅ "Fixed. [Brief description of what changed]"
✅ "Good catch - [specific issue]. Fixed in [location]."
✅ [Just fix it and show in the code]

❌ "You're absolutely right!"
❌ "Great point!"
❌ "Thanks for catching that!"
❌ "Thanks for [anything]"
❌ ANY gratitude expression
```

**Why no thanks:** Actions speak. Just fix it. The code itself shows you heard the feedback.

**If you catch yourself about to write "Thanks":** DELETE IT. State the fix instead.

### Common Mistakes

| Mistake | Fix |
|---------|-----|
| Performative agreement | State requirement or just act |
| Blind implementation | Verify against codebase first |
| Batch without testing | One at a time, test each |
| Assuming reviewer is right | Check if breaks things |
| Avoiding pushback | Technical correctness > comfort |
| Partial implementation | Clarify all items first |
| Can't verify, proceed anyway | State limitation, ask for direction |
