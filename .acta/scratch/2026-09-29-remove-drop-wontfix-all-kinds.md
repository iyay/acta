---
id: SCR-0018
hash: cwogc6x
title: Remove, drop and wont-fix actions for every kind
status: raw
created: "2026-09-29"
---
kita juga belom ada opsi buat remove/delete, drop dam wont-fix  ya buat semua. scratch - tasks


---
Agent notes (2026-09-29, found by reading):
- Statuses today (internal/board/board.go:84-88): story `draft, approved, in-progress, done, dropped`; bug `open, fixing, fixed, wontfix`; debt `open, done, wontfix`; scratch `raw, brainstorming, dropped`. Task lines have a wontfix mark `-` (internal/board/parse.go:23).
- So `dropped` and `wontfix` exist in the data for some kinds only, and they are set by hand with `acta set <kind>/<stem> status ...`. There is no delete or remove command, and the TUI has no key for any of these.
- Missing per kind: dropped for bug and debt, wontfix for story and scratch, and remove/delete everywhere.
- Open questions: does remove delete the file or only hide it (the scratch skill says dropped items stay in the repo for history)? Is it a CLI command, a TUI key, or both? What happens to the children when a parent is dropped?
