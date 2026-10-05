---
id: DBT-0081
hash: l4z74u8
parent: plans/2026-10-05-wiki-first-time-seed
---
# Review NOTEs: Wiki first-time seed Implementation Plan

- [ ] (medium) Wiki pages with folder-wide paths (internal/tui/, internal/cli/, plugin/omp/) fire pre-tool hints on unrelated files; one Bash call drew 4-5 hints, about 100+ tokens, paid again on every call. Narrow the paths to files, or cap hints per call.
- [ ] (low) Some pages keep a dated story (land-gates-then-merge, rebuild-binary-after-merge, orphan-test-binaries); the wiki rule says no history, so the dates could go.
- [ ] (low) eval answers-appended scored 0.50 three times on 2026-10-05 and 1.00 on every rerun; cause not checked.
