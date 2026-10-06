---
id: DBT-0025
hash: kslwc2z
parent: plans/2026-09-29-short-id-format
---
# Review NOTEs: Short Id Format Implementation Plan

- [ ] (low) Old example ids left in comments and usage strings: board.go:55-56 (PLAN-12 / PLAN-k3f2), write/ops.go:34 (SCRATCH-1), cli.go:102,446 (<SCRATCH-n>), hook/session.go:145 (SCRATCH-n).
- [ ] (low) freeHash only rejects exact matches, so a new 7-char hash can start with a 4-char hash in an unmigrated file and make that 4-char prefix lookup ambiguous.
- [ ] (low) readCloses does not read block-style YAML closes lists, so those entries are not migrated (they still resolve through Canon).
- [ ] (low) findHash scans every item on each Get miss; fine at today's board size.
- [ ] (low) findHash treats any input with 4+ characters after a dash as a hash prefix, so a mistyped path does a harmless prefix search.
- [ ] (low) land/SKILL.md:65 says acta id rewrites old ids after a merge, but the landing step runs acta id --fix-duplicates, which does not migrate; plain acta id must also run.
- [ ] (low) migratedCloses also rewrites closes entries that are already valid (path id, hash) into the target short id; spec only asked for old ids.
- [ ] (low) readCloses keeps a trailing YAML comment inside the last closes entry, which can make that entry name nothing.
- [ ] (low) longHash keeps the old 4 characters, so until every branch is migrated the old and new hash both match by prefix.
