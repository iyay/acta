---
id: DBT-0135
hash: r6qg3jr
parent: plans/2026-10-19-tidy-keeps-docs-work
---
# Review NOTEs: Tidy keeps docs work and carries Task trailers

- [ ] (medium) tidy carryTrailers starts a new paragraph when the kept message ends in other git trailers (Co-Authored-By) but no Task line, so git stops reading those as trailers; append to a git-trailer last paragraph instead (internal/tidy/tidy.go carryTrailers).
