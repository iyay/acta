---
id: DBT-0156
hash: v790bmg
parent: plans/2026-11-07-stray-worktrees-command
started: "2026-11-07 19:57:16"
---
# Review NOTEs: Stray worktrees: the acta worktrees command

- [ ] (medium) acta doctor counts dirty strays, but its fix acta worktrees --clean never removes dirty ones, so the warning stays after the fix with no hint why.
- [ ] (medium) acta doctor runs git status in every idle worktree on each run, acta setup included, so a repo with many big worktrees makes doctor and setup slow.
- [ ] (medium) lastActivity returns the zero time when the admin folder exists but holds no HEAD, index or logs/HEAD, so an unknown age reads as idle.
- [ ] (medium) a missing branch worktree whose branch ref was deleted by hand is removed by --clean, and its last commit loses its only holder.
- [ ] (medium) acta state <plan> run inside a build worktree printed "no running plan" for the plan being built, right after acta state set wrote its State section, so an agent cannot read its own state back there.
