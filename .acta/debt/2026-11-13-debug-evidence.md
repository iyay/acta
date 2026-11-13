---
id: DBT-0167
hash: czu566q
parent: plans/2026-11-12-debug-evidence
started: "2026-11-14 05:50:46"
---
# Review NOTEs: Debug proof is required, and each text change is measured

- [x] (medium) Debug SKILL.md control line copies with `cp -R "$OLDPWD/<path>" .`, and $OLDPWD is the folder the shell call started in. When calls start in the main checkout and the bug lives in a worktree, the copy takes the main checkout code with no error, so the control can run on the wrong tree. The line needs the root of the tree under debug named, and a measure.
- [x] (medium) Some writes in debug Phases 1 to 3 ride on no folder check: a repo test that writes caches or build output run as the feedback loop, the polluter hunt of find-polluter.sh, a profiler that writes into the current folder, and any write through a file tool that takes a path and never passes the shell. acta flake record and acta:bug also commit a planning file while the ground rule says no file in the repo changes. The rule needs to say which of these are meant and which are allowed.
- [x] (medium) /tmp is the default probe folder when TMPDIR is unset; on a machine with several users a clone of the repo there can be read by others, and a fixed folder name can hold files from an earlier session.
