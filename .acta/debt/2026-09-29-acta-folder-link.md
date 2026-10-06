---
id: DBT-0013
hash: tbpcmkr
parent: plans/2026-09-28-acta-folder-link
---
# Review NOTEs: Keep EnsureGitignore Inside the Repo Implementation Plan

- [x] internal/hook/gitignore_test.go "folder in another repo" links to <other>/.pm, not the other repo top, so the start-from-parent step (gitignore.go:30-32) is untested; removing it keeps all tests green.
- [ ] (medium) internal/hook/gitignore.go:29-31 an explicit --root, ACTA_ROOT or absolute root: that is a link to the repo top now gets no .agents.json line (doctor --fix exits 3 when the link's parent is a git repo); doctor.inRepoPath says inside, the hook says outside.
- [ ] (low) internal/hook/gitignore.go containment check repeats doctor.inRepoPath and realPath; the two disagree on a root linked to the repo top.
- [ ] (low) A worktree whose .acta links into the main checkout's .acta no longer gets the line from session-start or tick; matches doctor, but the plan did not list it.
- [ ] (medium) .acta linked into a nested repo or submodule inside the repo still gets the line written into the nested repo.
- [ ] (medium) acta tick with .acta linked into another repo still writes .agents.json there through write.RecordAgent, now without an ignore line.
- [ ] (low) On macOS a .acta link to the same repo spelled in a different letter case is refused silently.
- [ ] (low) internal/hook/gitignore.go:30 a root with a trailing slash would make Lstat follow the link and skip the parent step; no caller passes one today.
- [ ] (low) The implementation renamed the test helper gitTop to repoTop and rewrote WorksThroughALinkedParentFolder; the plan text still shows the old versions.
- [ ] (low) internal/hook/gitignore.go:48 packs the whole Rel test into one condition; doctor splits it with a why comment.
