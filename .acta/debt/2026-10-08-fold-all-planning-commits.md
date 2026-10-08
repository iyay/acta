---
id: DBT-0101
hash: bqczlsr
parent: plans/2026-10-08-fold-all-planning-commits
---
# Review NOTEs: Every one-file planning commit folds

- [ ] (low) gitc trailerStart misses folded trailer lines that continue on an indented line, so a fold after such a trailer puts the new subject inside that paragraph and git stops reading the trailer block. Accept indented lines once the block starts with a real trailer.
- [ ] (low) planning-commit-folds eval never checks that the spec ends with one commit; a run that calls acta commit but leaves two commits passes.
- [x] (low) canFold does not check tags; HEAD can be folded while a tag points at it (since SPC-0110).
- [ ] (low) A chore(plan): tick fix round N commit that folds into an earlier plan-only chore commit becomes a body line, so board newestRound and tidy isReviewMarker stop seeing the round. Today a fix commit always comes first, so it is rare.
