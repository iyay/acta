---
id: DBT-0024
hash: nqjxbpr
parent: plans/2026-09-29-body-schema-scratch
---
# Review NOTEs: Body Schema Plan 1: Base Rule, Date Meta and Scratch Implementation Plan

- [ ] (low) ids.go: the first id overwrites an existing created value, so a scratch item made on one day and given its id on a later day gets the later date.
- [ ] (low) tick writes started on any task tick, not only on tick --start (a harmless widening of the spec).
- [x] markTaskDates dereferences plan with no nil check after the board reload.
- [ ] (low) CheckBody counts a ## line inside a code fence as a section heading.
- [ ] (low) CheckBody does not check that the # title matches the frontmatter title, which the spec base rule asks for.
- [ ] (low) doctor schemaProblems prints the raw kind name, not the words the spec error text uses (spec, plan); ids.go maps them, doctor does not.
- [x] The scratch SKILL.md edit joined three paragraphs into single long lines.
- [ ] (low) Commit b48610d loosened the full-file equality check for tick --start in cmd/acta/tick_test.go to a box count.
- [ ] (medium) A tick writes started/finished into the spec file but tick never commits, and the orchestrator commits only the plan file, so the spec can stay dirty and the next write on it skips its commit.
- [ ] (medium) closeScratchParent writes the parent scratch file without a dirtyBefore check, so pending edits there ride the id commit.
- [ ] (low) NewScratch writes schema and created as quoted strings, unlike hand-written files.
- [ ] (low) An untick never clears finished, so a plan closed, reopened by untick, and closed again keeps the old day; the status path gives the new day.
- [ ] (low) An explicit acta set status to a closed status, or fixed_in, on an item already closed moves finished to today (allowed by the plan).
