---
created: "2026-10-08 11:40:51"
id: SPC-0114
hash: dcq43dd
---
# Tidy tests set their own git identity

Status: Bounded, approved by the user in chat on 2026-10-08.

Why: the first GitHub CI run (run 37726346387, commit 6b8217d) failed in `TestReplayClashFoldsIntoPrevious`. `internal/tidy/tidy_test.go:183` runs `git merge` with only date env vars, and the runner has no git user, so git stops with `Committer identity unknown`. It passes on a dev machine only because git can guess a name there. Repro: `HOME` set to an empty dir plus `GIT_CONFIG_PARAMETERS="'user.useConfigOnly'='true'"`, then `scripts/test ./internal/tidy/ -count=1`.

Design:
- `newRepo` in `internal/tidy/tidy_test.go` sets `user.name` and `user.email` with a local `git config` in the temp repo right after `git init`, so every git call in these tests has an identity, not only the merge.
- The `commit` helper keeps its own author and committer env, so the author of each test commit does not change.
- Only `internal/tidy/tidy_test.go` changes. No version bump: plans never bump the version.
- Test: the repro command fails before the change and passes after it, and every other test in `internal/tidy` stays green.
