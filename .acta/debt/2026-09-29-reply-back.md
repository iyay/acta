---
id: DEBT-16
hash: hr7s
parent: plans/2026-09-29-reply-back
---
# Review NOTEs: Dispatch Reply-Back Implementation Plan

- [x] plugin/skills/dispatch/SKILL.md:400 still says the fix brief REPLY-BACK goes over <fixed-from>..<new-head>; leftover of the hand-filled reply-back, and TestNoHandFilledReplyBack misses it.
- [x] dispatch init defaults the round to the branch name; a branch like feat/x, an uppercase name or a detached HEAD fails roundPattern and exits 1, and the error does not say to pass --round.
- [ ] checkRecord keeps the plan inside the root by text only (Clean + Rel), no symlink resolution; harmless today since the plan must name a board item.
- [ ] dispatch init warns and still writes the record when EnsureGitignore fails, so the record could be committed by accident.
- [ ] TestDispatchNewThenGoal fix-round wording loosened from "fix round: /goal only" to "then /goal only".
- [ ] cmdReplyBack in internal/cli/dispatch.go is about 80 lines; the open-task scan and message building could be split out.
- [ ] checkRecord joins the plan with cfg.RepoRoot but checks containment against cfg.Root; correct, but a one-line why comment would help.
- [ ] dispatch_test.go storedRecord repeats dispatchRecord; kept for the red-first run, could go now.
- [ ] No reviewer ran a revert mutation; red-on-revert was judged by reading the tests.
