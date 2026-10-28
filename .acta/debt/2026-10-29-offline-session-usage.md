---
id: DBT-0144
hash: wqe0wxb
parent: plans/2026-10-28-offline-session-usage
---
# Review NOTEs: Offline Codex session usage

- [ ] (medium) acta usage in a bare repository, or in a checkout reached only through GIT_DIR, exits 1 as "not inside a Git repository" and loses git's reason; the discriminator in internal/usage/usage.go repoRoot only looks for a .git entry above cwd.
