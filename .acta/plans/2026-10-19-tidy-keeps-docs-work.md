---
parent: specs/2026-10-19-tidy-keeps-docs-work-design
depth: minimal
id: PLN-0178
created: "2026-10-19 21:01:24"
hash: cw04x8f
---

# Tidy keeps docs work and carries Task trailers

**Goal:** `acta tidy` keeps real docs work as commits and never loses a `Task:` trailer when it folds.

**Spec:** .acta/specs/2026-10-19-tidy-keeps-docs-work-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Production code changes stay in `internal/tidy/tidy.go`. The final tree must still equal the merge (the existing proof stays).

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Docs work before the review marker stays a commit

**Files:** `internal/tidy/tidy.go`, `internal/tidy/tidy_test.go`
**Seam:** `tidy.Run`
**verify:** A `docs` commit before the marker is kept exactly when it touches a file outside the planning root; `chore` and `wiki` always fold; the marker and everything after it fold as today. List each subject type and marker position checked; `TestReviewMarkerFoldsRest` passes unchanged.
- [ ] Red: a branch `docs(skills)`, `docs(setup)`, `polish:`, `docs(bug)` tidies to one commit today instead of the two docs commits.
- [ ] Change: split fold types so `docs` keeps when `touchesOutside` is true and the marker has not been seen.
- [ ] Commit: `fix(tidy): keep docs work that touches code`

### Task 2: Folded Task trailers move to the kept commit

**Files:** `internal/tidy/tidy.go`, `internal/tidy/tidy_test.go`, `.acta/wiki/land-tidy-history.md`
**Seam:** `tidy.Run`, read back through `internal/commits`
**verify:** No `Task:` trailer of any folded commit is lost on any fold path (fold into a kept commit, ride forward before the first kept commit, all-fold fallback); each lands once in the last paragraph of the kept message, and a mid-body `Task:` mention never counts as present. The all-fold subject comes from the newest commit touching files outside the planning root. List every fold path checked.
- [ ] Red: a folded commit's trailer is missing from the kept commit, and an all-fold branch takes a planning-only `docs(bug)` subject.
- [ ] Change: carry folded trailers into the kept message's last paragraph with dedupe there, pick the all-fold subject by `touchesOutside`, and rewrite the trailer line of the wiki page.
- [ ] Commit: `fix(tidy): carry Task trailers when commits fold`
