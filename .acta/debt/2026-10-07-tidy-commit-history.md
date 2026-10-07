---
id: DBT-0096
hash: glecsw6
parent: plans/2026-10-07-tidy-commit-history
---
# Review NOTEs: Tidy commit history Implementation Plan

- [x] (medium) The land skill lost three rules in fix round 1 to fit its byte cap: "Thinking 'just this once'", "Tired and wanting work over", and the row "I'm tired | Tired is not an excuse". Put them back once the cap is settled.
- [ ] (medium) remapText rewrites any 7 to 40 char hex word that is a unique prefix of an old branch hash, even when the word names another object; the proof uses the same function, so it cannot catch this.
- [ ] (low) A replay clash maps the clashed commit's old hash to the earlier new commit, but its change only shows up in the last commit.
- [ ] (low) After reset --keep, folded parent chore commits leave history; planning files that name their hashes now point at gone commits, since remap covers only branch hashes.
- [ ] (low) With no @{upstream}, the refs/acta/last-land fold point can sit behind commits the user pushed by hand, so "pushed commits are never touched" holds only with an upstream.
- [ ] (low) An unrelated base prints the raw git merge-base error instead of saying the two share no history.
- [ ] (low) A branch given as @{-1} passes check-ref-format --branch; the chain is built before update-ref fails. Check refs/acta/tidy/<short> up front instead.
- [ ] (low) git rev-parse --short=7 prints more than 7 chars when 7 is ambiguous, while acta tidy prints exactly 7, so land stops when it need not. Compare by prefix or print the full sha.
- [x] (low) Land skill full-mode step 8 says fixed_in takes the last task commit, tidy step 4 says the tidy tip; make the two lines say the same thing.
- [x] (low) The land skill is at 9094 of 9100 bytes; the next edit must cut text or raise the cap.
- [ ] (medium) tidy.md names step numbers from two files ("step 1", "step 4", "Steps 9 and 10" mean SKILL.md; "step 2", "the commit from 4" mean tidy.md); say "SKILL.md step N" so an agent reading only tidy.md does not guess.
- [ ] (low) TestLandTidyPath does not pin --short=7, "Pushed commits stay", chore(...), "red stops the land" or the last-land update-ref step.
- [ ] (low) In tidy mode the landing report's wiki diff <merge-sha>^1..<merge-sha> sees only the last replayed commit; diff from the old parent tip instead.
