---
id: DBT-0114
hash: enkufdu
parent: plans/2026-10-14-tidy-chore-commits
---
# Review NOTEs: Tidy chore commits Implementation Plan

- [ ] (low) internal/doctor/wiki.go linkedWikiFile checks only the last path part; a linked .acta/wiki folder lets fixWiki and wiki ok write .checked through the link.
- [ ] (low) internal/wiki/check.go runs one git log -p over .checked history per page, so wiki check costs pages times history; read it once per run.
- [ ] (low) internal/doctor/wiki.go lines() keeps CR of CRLF lines, so a CRLF .gitattributes never matches the union line and --fix appends it on every run.
