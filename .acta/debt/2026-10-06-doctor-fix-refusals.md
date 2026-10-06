---
id: DBT-0087
hash: w6uqhnv
parent: plans/2026-10-06-doctor-fix-refusals
---
# doctor --fix refusals review notes

- [ ] (medium) internal/doctor/doctor.go refusal: the symlink walk assumes ActaRoot is spelled under RepoRoot; a root spelled through another alias (root: /tmp/r/link while git reports /private/tmp/r) skips the walk, so the no-symlink rule is not enforced there. Run the walk on realPath(RepoRoot) and the cleaned root.
- [ ] (low) internal/doctor/doctor.go checkRepo picks the fix line by substring match on the refusal reason, which includes user paths; a reworded reason or a path holding '.gitignore' or 'not a regular file' gets the wrong fix line. Return a reason kind from refusal instead.
