---
id: DEBT-24
hash: nqjx
parent: plans/2026-09-29-body-schema-scratch
---
# Review NOTEs: Body Schema Plan 1: Base Rule, Date Meta and Scratch Implementation Plan

- [ ] ids.go: the first id overwrites an existing created value, so a scratch item made on one day and given its id on a later day gets the later date.
- [ ] tick writes started on any task tick, not only on tick --start (a harmless widening of the spec).
- [ ] markTaskDates dereferences plan with no nil check after the board reload.
- [ ] CheckBody counts a ## line inside a code fence as a section heading.
- [ ] CheckBody does not check that the # title matches the frontmatter title, which the spec base rule asks for.
- [ ] doctor schemaProblems prints the raw kind name, not the words the spec error text uses (spec, plan); ids.go maps them, doctor does not.
- [ ] The scratch SKILL.md edit joined three paragraphs into single long lines.
- [ ] Commit 47810b5 loosened the full-file equality check for tick --start in cmd/acta/tick_test.go to a box count.
- [ ] A tick writes started/finished into the spec file but tick never commits, and the orchestrator commits only the plan file, so the spec can stay dirty and the next write on it skips its commit.
- [ ] closeScratchParent writes the parent scratch file without a dirtyBefore check, so pending edits there ride the id commit.
- [ ] NewScratch writes schema and created as quoted strings, unlike hand-written files.
- [ ] An untick never clears finished, so a plan closed, reopened by untick, and closed again keeps the old day; the status path gives the new day.
- [ ] An explicit acta set status to a closed status, or fixed_in, on an item already closed moves finished to today (allowed by the plan).
