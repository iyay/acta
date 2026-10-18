---
parent: specs/2026-10-19-state-resume-stop-design
depth: minimal
id: PLN-0164
created: "2026-10-19 06:56:46"
hash: r2nmet0
---
# state-resume stop plan

**Goal:** The state-resume eval case stops after the Next step, so a land never deletes the file its graders read.
**Spec:** `.acta/specs/2026-10-19-state-resume-stop-design.md`
**Tests:** `scripts/test` (fast), `scripts/test --full` (full, land only); task loop: `scripts/test ./internal/plugincheck/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Change only the prompt text of the case; graders and max_turns stay. Inline in a worktree, English in the repo, never push.

## Waves

- Wave 1: Task 1.

### Task 1: Stop after the Next step

**Files:** Modify `plugin/evals/state-resume/prompt.md`.
**Seam:** the eval case as `scripts/eval --case state-resume` runs it.
**verify:** No run of the case removes or lands the worktree the graders read; every run either changes `wt/src/invoice.py` as Next names or fails on that. List what each grader rejects. `scripts/eval --case state-resume --runs 3` passes 3 of 3.
- [ ] Red: Run the case on main and observe a failed run (1 of 3 failed on 2026-10-19, trace e-nMn6yY).
- [ ] Change: Add the stop line; pass the eval 3 of 3 and `scripts/test ./internal/plugincheck/`.
- [ ] Commit: `test(evals): state-resume stops after the Next step`.
