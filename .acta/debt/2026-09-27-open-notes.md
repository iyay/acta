---
id: DBT-0001
hash: ohwd3gg
parent: plans/2026-09-26-open-notes
---
# Review NOTEs: Close open review NOTEs Implementation Plan

- [x] Another local user can pre-create /tmp/pmb-<uid>: safeDir fails safe but the sticky bit means only root can remove it, so ticks stay blocked; error text is the same for symlink, bad mode and wrong owner.
- [x] Tests that call lock/Tick without useLockBase still leave lock files in the real /tmp/pmb-<uid> (harmless, never cleaned). (stale)
- [ ] (low) TestLockRejectsUnsafeFolder's "plain file" case passes without needing the IsDir check on its own; the owner-check branch still has no test (needs a second user).
- [ ] (low) The dropped TestSkillBuild Must line for "references/house-rules.md", "implementer-prompt.md", "## Waves" was never restored.
