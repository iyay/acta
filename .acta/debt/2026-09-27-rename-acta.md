---
id: DBT-0007
hash: gqalwue
parent: plans/2026-09-27-rename-acta
---
# Review NOTEs: Rename pm and pmb to acta Implementation Plan

- [ ] (low) Plugin hooks (session-start, prompt-reminder) still look only for "acta"; a user whose PATH has only "pmb" still silently gets default rules and no reminder, no fallback added.
- [x] config.Default() sets Root to .pm; Load always overwrites, and the comment disagreed.
- [x] migrate-root still moves ignored files under .pm along with the folder (fine, still untested).
- [x] Plan header still names .pm/specs/... for the spec path.
- [ ] (medium) migrate undo still swallows git errors and always says "move undone".
- [ ] (low) migrateRoot is still ~85 lines; rootValue/withoutRootLine still duplicate the root-line check; CRLF is still untested; post-move failures still exit 3 not 1.
- [ ] (low) checkMovable is still 62 lines; the .acta.yaml existence check now uses os.Lstat, but the earlier .acta folder check (actaDir) still uses os.Stat, so a dangling symlink named .acta still slips past.
- [ ] (low) migrate_states_test.go is still 605 lines.
- [ ] (low) Root-only .pm.yaml still leaves an empty .acta.yaml committed.
