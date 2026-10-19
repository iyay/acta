---
parent: specs/2026-10-19-setup-single-enter-design
depth: minimal
id: PLN-0167
created: "2026-10-19 07:22:08"
hash: fzlwre8
started: "2026-10-19 07:22:35"
finished: "2026-10-19 07:23:25"
---
**Goal:** Setup questions need one Enter each when stdin is redirected from `/dev/tty`, as under `curl | sh`.
**Spec:** `.acta/specs/2026-10-19-setup-single-enter-design.md`
**Tests:** `scripts/test ./internal/setup/` (fast); `scripts/test --full` (full, land only); `gofmt -l internal/setup`; `go vet ./internal/setup/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Keep one form per question and the collapsed answer lines. When `/dev/tty` cannot be opened, behave as today.
- Build inline in a worktree from main. No release or push.

## Waves

- Wave 1: Task 1.

### Task 1: Give every setup form the named tty

**Files:** Modify `internal/setup/form.go`; test in `internal/setup/form_test.go`.
**Seam:** A small function in `internal/setup` that returns the input for the forms (an open file named `/dev/tty`, or nil), with the open call swappable in tests.
**verify:** Every form `Ask` runs, override screens included, gets the same input; that input is a file named `/dev/tty` whenever one can be opened, and nil (today's default) when the open fails. List each form-run call site checked.
- [x] Red: test the seam for both the open-works and open-fails paths; watch `scripts/test ./internal/setup/ -run <name>` fail.
- [x] Change: open the tty once in `Ask`, close it when done, pass it with `WithInput` to the question forms and to `runGroup`; tests, gofmt and vet pass.
- [x] Commit: `fix(setup): read answers from the named tty so one Enter is enough`.

## Polish

### Task 2: Review polish

**verify:** NOTE below applied, and nothing else changes.
- [x] Test that every form Ask runs, override screens included, gets the one opened tty (a partial revert of the wiring must go red).
- [x] Commit: `polish: review notes PLN-0167`

## Review notes

- Deep review, full tier (input handling for setup forms): Spec and Standards reviewers over d201611..e6ef15c both CLEAN. The one [fix] NOTE (no test for the tty wiring) went into the polish commit, and its review was CLEAN. [note]s: stdin from a terminal other than the controlling one now reads the controlling tty; the char-device check repeats the cli gate; Ask and the test both close the fake tty. Manual check through curl | sh in herdr still to do after landing.
