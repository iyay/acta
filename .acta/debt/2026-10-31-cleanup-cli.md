---
id: DBT-0146
hash: v1wt831
parent: plans/2026-10-31-cleanup-cli
---
# Review NOTEs: Cleanup CLI and data implementation plan

- [ ] (medium) cleanup apply and --fix write planning files with bare os.WriteFile: no file lock and no tmp+rename, so a TUI tick on the same debt file during an apply is lost, and a crash mid-write can leave a torn file. Route them through the internal/write helpers that take the lock.
- [ ] (medium) cleanup apply has no rollback: when a later os.WriteFile fails, the files already written stay changed and uncommitted, and the next apply or --fix refuses them as dirty.
- [ ] (medium) A closed debt line keeps its reason only inside the line text, so board.Item.Reason, acta show and the TUI REASON row stay empty for debt items and --json readers cannot see why a line closed.
- [ ] (medium) Plain acta cleanup turns every cleanup.Plan error into stdout text with exit 0, not only the dirty-file one; an unreadable file or a git failure reads as success to a script. Return a sentinel for the dirty case and keep other errors on exitOther.
