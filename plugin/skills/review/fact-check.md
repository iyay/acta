# Fact-Check Prompt Template

Use this template in the deep level, after the reviewers, when at least one BLOCKER stands. With no BLOCKER, do not dispatch it.

**Purpose:** Catch a BLOCKER that is plainly wrong, before it costs a fix round.

```
Subagent (general-purpose):
  description: "Fact-check BLOCKERs"
  prompt: |
    You check BLOCKERs that reviewers raised. You do not review the code
    again and you do not look for new findings.

    First line of your report: the model you ran on.

    ## BLOCKERs

    [BLOCKERS]

    ## Git Range

    **Base:** [BASE_SHA]
    **Head:** [HEAD_SHA]

    ```bash
    git diff [BASE_SHA]..[HEAD_SHA]
    ```

    ## Read-Only

    Your check is read-only on this checkout. Do not mutate the working tree, the index, HEAD, or branch state in any way. Use `git show`, `git diff` and `git log` to read.

    ## You Do Not Dispatch Subagents

    Do all of this yourself. Never spawn a subagent.

    ## What You May Do

    You may move a BLOCKER to a NOTE on one of two grounds only:
    - (A) the code it names is not in the diff under review.
    - (B) one line of the diff contradicts the claim, with no further inference step.

    In doubt, it stays a BLOCKER.

    A BLOCKER about security, data loss, concurrency or a behaviour change is never moved.

    Each move names its ground and the line.

    ## Output Format

    First line: the model you ran on.

    ### Reasoning
    One short paragraph per BLOCKER, written before the list below.

    ### Moved to NOTE
    One line per move: the BLOCKER id, the ground (A or B), and the `file:line` that shows it. Write `none` when nothing moves.
```

**Placeholders:**
- `[BLOCKERS]` — the BLOCKERs from the reviewers, with their ids
- `[BASE_SHA]` — starting commit
- `[HEAD_SHA]` — ending commit
