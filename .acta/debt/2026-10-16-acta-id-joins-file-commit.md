---
id: DBT-0124
hash: dhht245
parent: plans/2026-10-16-acta-id-joins-file-commit
---
# Review NOTEs: acta id joins the file's birth commit Implementation Plan

- [ ] (medium) canFold checks the author by name only and treats a commit as unpushed when no remote-tracking ref holds it; a push with no tracking ref can still be amended.
- [ ] (medium) a fold that deletes a file HEAD added keeps that name in a built subject.
- [ ] (medium) newWhat assumes chore(scratch) is the longest subject prefix; a longer kind pushes the subject past 72 characters.
