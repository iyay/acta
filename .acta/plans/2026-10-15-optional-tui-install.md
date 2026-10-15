---
parent: specs/2026-10-15-optional-tui-install-design
depth: minimal
id: PLN-0143
created: "2026-10-15 09:35:38"
hash: hp5mv0h
started: "2026-10-15 09:37:25"
finished: "2026-10-15 09:45:27"
---
# Optional TUI at install time Implementation Plan

**Goal:** A `notui` build of acta ships as `acta-cli_<os>_<arch>.tar.gz`, `install.sh` lets the user pick it with `ACTA_TUI` or a prompt, and `acta update` keeps each user on the build they have.

**Spec:** `.acta/specs/2026-10-15-optional-tui-install-design.md`

**Tests:** fast `scripts/test ./internal/cli ./internal/update ./internal/plugincheck ./scripts`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Build tag name `notui`. `cli.Variant` is exactly `"full"` or `"cli"`. Asset prefixes are exactly `acta` and `acta-cli`; the file inside both archives is `acta`.
- Env `ACTA_TUI`: `1` full, `0` small, any other value fails before download with `ACTA_TUI must be 0 or 1`. Prompt text, verbatim: `Install the TUI board too? It adds about 7M. [Y/n]`. Unset with no tty gives the full build.
- Doctor line, verbatim: `build: full` or `build: cli (no TUI)`; info only, never warn or fail.
- `scripts/install.ps1` is not touched. No version bump.
- Tests never run `rm -rf` or `rm -f`; temp dirs come from `t.TempDir()`.

## Waves

- Wave 1: Task 01, Task 02, Task 03
- Wave 2: Task 04, Task 05

### Task 01: notui build tag splits the TUI out of the CLI

**Files:** Modify `internal/cli/cli.go`, `internal/cli/tuitrace.go`, `scripts/test`, `scripts/test_test.go`; Create `internal/cli/tui_full.go`, `internal/cli/tui_none.go`, `internal/cli/variant_test.go`

**verify:** A `notui` build never links the TUI, and both builds still run every command. List checked: `go list -deps -tags notui ./cmd/acta` holds neither `internal/tui` nor `github.com/charmbracelet/glamour`; `cli.Variant` is `full` without the tag and `cli` with it; bare `acta` in the `notui` build prints help plus the `ACTA_TUI=1` line and exits 0; `scripts/test --full` also runs `go test -tags notui ./internal/cli`.

- [x] **Failing test:** `variant_test.go` checks `Variant` per build and the deps list; the `notui` package does not build yet, so `scripts/test ./internal/cli` fails.
- [x] **Change:** move `runTUI` and its imports to `tui_full.go` (`!notui`), add `tui_none.go` (`notui`) with `Variant` and the help line, tag `tuitrace.go`, and add the tagged run to `scripts/test --full`.
- [x] **Commit:** `feat(cli): notui build tag leaves the TUI out`

### Task 02: release ships the acta-cli build

**Files:** Modify `.goreleaser.yaml`, `internal/plugincheck/release_test.go`, `.github/workflows/ci.yml`, `.acta/wiki/release.md`

**verify:** The release config can never drop either build or mix their flags. List checked: exactly two builds, both stripped with the release flag; only `acta-cli` has `tags: [notui]`; archives `acta_{{ .Os }}_{{ .Arch }}` and `acta-cli_{{ .Os }}_{{ .Arch }}`, each holding `acta`; CI builds with `-tags notui` for the four os/arch pairs and runs `go test -tags notui ./internal/cli`.

- [x] **Failing test:** extend `TestReleaseGoreleaserBuilds`, `TestReleaseGoreleaserStrips` and `TestReleaseGoreleaserAssets` for two builds; they fail on the one-build config.
- [x] **Change:** add the `acta-cli` build and archive, the CI steps, and one sentence in the wiki page.
- [x] **Commit:** `build: release an acta-cli asset with no TUI`

### Task 03: install.sh picks the build

**Files:** Modify `scripts/install.sh`, `scripts/install_test.go`

**verify:** The installer never fetches a build the user did not pick, and never fetches anything on a bad `ACTA_TUI`. List checked: `ACTA_TUI=1`, `ACTA_TUI=0`, `ACTA_TUI=2` (fails, no request made), unset with no tty (full), and through a pty: Enter (full), `n` (small), a bad answer then `n` (asks again, small). Each case names the asset fetched.

- [x] **Failing test:** the cases above in `install_test.go`; the script ignores `ACTA_TUI`, so they fail.
- [x] **Change:** read `ACTA_TUI` or ask on `/dev/tty` before the download, then set the asset prefix.
- [x] **Commit:** `feat(install): ACTA_TUI or a prompt picks the build`

### Task 04: acta update keeps the build

**Files:** Modify `internal/update/fetch.go`, `internal/update/fetch_test.go`, `internal/cli/update_cmd.go`

**verify:** `update` never swaps one build for the other and never installs bytes that failed a check. List checked at the `Fetch` layer: prefix `acta-cli` fetches and checks `acta-cli_*`; a release with no `acta-cli_*` checksum line fails with a clear error and no bytes. `update_cmd.go` is wiring only: it passes the prefix for `cli.Variant`, one success path.

- [x] **Failing test:** `Fetch` cases above; `Fetch` has no prefix argument yet, so the test does not build.
- [x] **Change:** add the prefix argument and pass it from `update_cmd.go`.
- [x] **Commit:** `feat(update): fetch the asset that matches the build`

### Task 05: doctor and --version name the build

**Files:** Modify `internal/cli/doctor.go`, `internal/cli/cli.go` (version line only), `internal/cli/doctor_test.go`

**verify:** The build is named the same way everywhere a user can see it. List checked: `acta doctor` prints `build: full` or `build: cli (no TUI)` as info, in both builds, and its exit code does not change; `acta --version` names the build in both.

- [x] **Failing test:** doctor and version tests look for the build line and fail.
- [x] **Change:** add the doctor info line and the build word in the version output.
- [x] **Commit:** `feat(doctor): show which build is installed`

## Fix round 1

### Task 06: Review round 1 fixes

**Files:** Modify `scripts/install_test.go`, `scripts/install.sh`, `internal/doctor/doctor.go`, `internal/plugincheck/release_test.go`, `.acta/wiki/release.md`

**verify:** The installer tests pin the platform on every host, so they pass or fail the same on darwin/arm64, Intel Mac and the ubuntu runner; no installer test can hang past its own deadline; any `ACTA_TUI` value but `0` and `1`, the word `unset` included, fails before a download; every archive rule is tar.gz.

- [x] **Failing test:** `bothRig` serves Linux x86_64 only, so the pty test fails on this Mac while it runs the real `uname`; `ACTA_TUI=unset` joins the bad values; the format check runs on every archive.
- [x] **Change:** pass the python3 dir through `pathExtra`, give the pty driver a real deadline with `select` and skip when python3 is missing, read `ACTA_TUI` with `${ACTA_TUI+set}`, move the `checkBinary` comment back, name `ACTA_TUI` in the wiki.
- [x] **Commit:** `fix: round 1 review findings for PLN-0143`

## Review notes

- Round 1 (deep, two reviewers): BLOCKED on one finding, the pty test running the real uname (passed only on darwin/arm64). Fixed in Task 06.
- Round 2 (deep, two reviewers): CLEAN. Range also held the merge of main (update-notify); the cli.go conflict moved `.WithUpdate(newerRelease())` into `tui_full.go`.
- install.sh: typing `n` then Ctrl-D twice with no Enter gives the full build, a safe default; tagged [fix] by a reviewer, kept as a note because changing it needs another full review.
- After Ctrl-D at the prompt no newline is printed, so "Downloading ..." lands on the prompt line.
- The notui build still links bubbletea and lipgloss through internal/setup (huh); the spec bans only internal/tui and glamour.
- On a Mac without the command line tools, LookPath finds the /usr/bin/python3 stub, so the pty test fails instead of skipping.
- `scripts/test --full ./...` passes `./...` to the notui call too, so that run tests the whole repo twice.
- release.yml runs only `go test ./...`, not the notui tests; ci.yml covers them.
- Setsid in install_test.go does not build on Windows; matters only for the Windows port (SCR-0058).
- Bare `acta` in the notui build never shows the update notice; every other command does.
- Task 05 touched internal/doctor/doctor.go and internal/cli/help.go, which the plan did not list, because the checks and versionLine live there.
