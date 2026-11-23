---
id: DBT-0172
hash: qu79qxd
parent: plans/2026-11-23-land-bookkeeping-in-tidy
---
# Review NOTEs: Land bookkeeping rides the land

- [ ] (medium) internal/tidy/tidy.go:195: a land stopped by a red gate leaves refs/acta/tidy/<branch>, so the retry counts as a rerun and plan_landed is never counted for it; key the rerun check by plan, or count once the parent moves.
- [ ] (medium) internal/tidy/tidy.go:194: the rerun ref is keyed by branch, so a stale ref from a stopped land on a reused branch name hides the next plan count.
