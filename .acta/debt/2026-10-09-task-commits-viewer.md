---
id: DBT-0105
hash: gfq7x5t
parent: plans/2026-10-09-task-commits-viewer
---
# Review NOTEs: Task commits viewer

- [ ] (medium) internal/tui/model.go reloadCmd: commits.RefsFor runs trees.Others again on every reload although the board load just did the same scan, plus 2 git calls per ref; measured about +0.5 s per watcher reload on this repo (1.75 s vs 1.23 s). Build the refs from the trees the load already has.
- [ ] (low) internal/commits/commits.go Find: one full git log --grep per ref walks the shared history once per ref; a repo with many unmerged branches pays N walks on every reload. One git log --source over all refs avoids it.
- [ ] (low) internal/commits/commits.go: a user with log.showSignature=true gets gpg lines on stdout that break the record parse; pass --no-show-signature to log and show.
