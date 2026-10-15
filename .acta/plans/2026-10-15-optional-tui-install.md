---
parent: specs/2026-10-15-optional-tui-install-design
depth: minimal
id: PLN-0143
created: "2026-10-15 09:35:38"
hash: hp5mv0h
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

- [ ] **Failing test:** `variant_test.go` checks `Variant` per build and the deps list; the `notui` package does not build yet, so `scripts/test ./internal/cli` fails.
- [ ] **Change:** move `runTUI` and its imports to `tui_full.go` (`!notui`), add `tui_none.go` (`notui`) with `Variant` and the help line, tag `tuitrace.go`, and add the tagged run to `scripts/test --full`.
- [ ] **Commit:** `feat(cli): notui build tag leaves the TUI out`

### Task 02: release ships the acta-cli build

**Files:** Modify `.goreleaser.yaml`, `internal/plugincheck/release_test.go`, `.github/workflows/ci.yml`, `.acta/wiki/release.md`

**verify:** The release config can never drop either build or mix their flags. List checked: exactly two builds, both stripped with the release flag; only `acta-cli` has `tags: [notui]`; archives `acta_{{ .Os }}_{{ .Arch }}` and `acta-cli_{{ .Os }}_{{ .Arch }}`, each holding `acta`; CI builds with `-tags notui` for the four os/arch pairs and runs `go test -tags notui ./internal/cli`.

- [ ] **Failing test:** extend `TestReleaseGoreleaserBuilds`, `TestReleaseGoreleaserStrips` and `TestReleaseGoreleaserAssets` for two builds; they fail on the one-build config.
- [ ] **Change:** add the `acta-cli` build and archive, the CI steps, and one sentence in the wiki page.
- [ ] **Commit:** `build: release an acta-cli asset with no TUI`

### Task 03: install.sh picks the build

**Files:** Modify `scripts/install.sh`, `scripts/install_test.go`

**verify:** The installer never fetches a build the user did not pick, and never fetches anything on a bad `ACTA_TUI`. List checked: `ACTA_TUI=1`, `ACTA_TUI=0`, `ACTA_TUI=2` (fails, no request made), unset with no tty (full), and through a pty: Enter (full), `n` (small), a bad answer then `n` (asks again, small). Each case names the asset fetched.

- [ ] **Failing test:** the cases above in `install_test.go`; the script ignores `ACTA_TUI`, so they fail.
- [ ] **Change:** read `ACTA_TUI` or ask on `/dev/tty` before the download, then set the asset prefix.
- [ ] **Commit:** `feat(install): ACTA_TUI or a prompt picks the build`

### Task 04: acta update keeps the build

**Files:** Modify `internal/update/fetch.go`, `internal/update/fetch_test.go`, `internal/cli/update_cmd.go`

**verify:** `update` never swaps one build for the other and never installs bytes that failed a check. List checked at the `Fetch` layer: prefix `acta-cli` fetches and checks `acta-cli_*`; a release with no `acta-cli_*` checksum line fails with a clear error and no bytes. `update_cmd.go` is wiring only: it passes the prefix for `cli.Variant`, one success path.

- [ ] **Failing test:** `Fetch` cases above; `Fetch` has no prefix argument yet, so the test does not build.
- [ ] **Change:** add the prefix argument and pass it from `update_cmd.go`.
- [ ] **Commit:** `feat(update): fetch the asset that matches the build`

### Task 05: doctor and --version name the build

**Files:** Modify `internal/cli/doctor.go`, `internal/cli/cli.go` (version line only), `internal/cli/doctor_test.go`

**verify:** The build is named the same way everywhere a user can see it. List checked: `acta doctor` prints `build: full` or `build: cli (no TUI)` as info, in both builds, and its exit code does not change; `acta --version` names the build in both.

- [ ] **Failing test:** doctor and version tests look for the build line and fail.
- [ ] **Change:** add the doctor info line and the build word in the version output.
- [ ] **Commit:** `feat(doctor): show which build is installed`
