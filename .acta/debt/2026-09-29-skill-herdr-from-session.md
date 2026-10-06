---
id: DBT-0028
hash: rabbzyg
parent: plans/2026-09-29-skill-herdr-from-session
---
# Review NOTEs: Skill Herdr Choice From Session Text Implementation Plan

- [ ] (low) No test ties the hook's herdr: line prefix to the skill's choice (b) trigger; renaming the prefix breaks choice (b) with every test green.
- [x] setup/SKILL.md:29 still has the agent check HERDR_ENV=1 before offering dispatch; same failure mode as BUG-0007. (stale)
- [ ] (low) brainstorm/SKILL.md:80 choice (b) is one long line while its neighbours wrap near 80 columns.
- [ ] (low) skill_brainstorm_test.go MustNot "HERDR_ENV" is case-sensitive; lowercase herdr_env passes.
- [ ] (low) Rule 8 reads worse than rules 1-7: two colons in one sentence, and the two ways come as a sentence with no verb after the id instruction.
- [ ] (low) Rule 8 "no Skill tool" is scoped to the filing only by punctuation; say "no Skill tool for the filing".
- [x] hook_test.go wanted strings are checked only with Herdr=false.
- [x] The recipient amended the plan's rule 8 target text (71309a8) instead of using it verbatim; additive only, but the brief said use it exactly. (stale)
- [ ] (low) Fix round 1 grew rule 8 to 12 negations to satisfy the grader outside herdr; the 60-line cap counts newlines, so byte growth goes unseen.
- [ ] (low) Scratch skill says add context right after new with acta scratch add; rule 8 bans that on the second-brainstorm path only.
- [ ] (medium) Brainstorm Step 0 already files the item before acta set; blockText and rule 8 saying "file it" again can make a duplicate scratch item.
