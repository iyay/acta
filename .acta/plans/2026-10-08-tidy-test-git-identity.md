---
parent: specs/2026-10-08-tidy-test-git-identity
depth: minimal
id: PLN-0123
created: "2026-10-08 11:40:51"
hash: u2c588x
---
# Tidy tests set their own git identity

**Goal:** Every test in `internal/tidy` passes on a machine with no git user configured, as on a clean CI runner.

**Spec:** .acta/specs/2026-10-08-tidy-test-git-identity.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Only `internal/tidy/tidy_test.go` changes. No version bump.
- Identity is set with a local `git config` in the temp repo, never global.

## Waves

- Wave 1: Task 01

### Task 01: newRepo sets a local git identity

**Files:**
- Modify: `internal/tidy/tidy_test.go`

**verify:** No git call in `internal/tidy` tests depends on a user identity from outside the temp repo: with an empty HOME and `user.useConfigOnly=true`, the whole package passes. List every git call in the file that can create a commit and what identity it gets.

- [ ] Failing test: run `h=$(mktemp -d); HOME=$h XDG_CONFIG_HOME=$h GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_PARAMETERS="'user.useConfigOnly'='true'" scripts/test ./internal/tidy/ -count=1` and see `TestReplayClashFoldsIntoPrevious` fail with `Committer identity unknown`.
- [ ] Code: in `newRepo`, right after `git init`, run `git config user.name test` and `git config user.email test@example.com` in the temp repo, with one plain comment saying why (a CI runner has no git user, so a merge would stop).
- [ ] Run the same command passes; `scripts/test ./internal/tidy/` passes; `gofmt -l internal/tidy` empty, `go vet ./internal/tidy/` clean; commit `fix(tidy): tests set their own git identity`.
