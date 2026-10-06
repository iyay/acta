---
id: DBT-0089
hash: dkkdwcg
parent: plans/2026-10-06-debt-write-safety
---
# debt write safety review notes

- [ ] (medium) internal/write/priority.go setLinePriority still finds its debt line by number only; a hand edit above the item can retag another item's line. Same fix as TickLine: check the line by text.
- [ ] (low) internal/write/mark.go MarkItem releases the lock after TickLine and then commits; a NewDebt append landing in that gap rides along in the tick commit.
