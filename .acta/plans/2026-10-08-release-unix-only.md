---
parent: specs/2026-10-08-release-unix-only
depth: minimal
id: PLN-0124
created: "2026-10-08 12:12:02"
hash: hy35zmo
started: "2026-10-08 12:55:12"
finished: "2026-10-08 12:57:09"
---
# First release ships macOS and Linux only, and CI cross-builds every release target

**Goal:** goreleaser builds only targets that compile, and CI fails on any release target that does not compile before a tag is cut.

**Spec:** .acta/specs/2026-10-08-release-unix-only.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Release targets: goos `darwin`, `linux`; goarch `amd64`, `arm64`. Asset names stay `acta_<os>_<arch>.tar.gz`, `checksums.txt`, `install.sh`.
- `scripts/install.ps1` and the `windows-hooks` CI job stay in the repo.
- No version bump; no Go source change outside `internal/plugincheck`.

## Waves

- Wave 1: Task 01
- Wave 2: Task 02

### Task 01: Release config and CI cross-build

**Files:**
- Modify: `.goreleaser.yaml`, `.github/workflows/ci.yml`
- Test: `internal/plugincheck/release_test.go`

**verify:** Every goos and goarch pair `.goreleaser.yaml` builds is cross-built by a `ci.yml` step on every push, and no other pair; nothing in the release config builds or ships a Windows asset. List each pair and where the test checks it.

- [x] Failing test: in `release_test.go` expect goos `darwin linux`, no `format_overrides`, extra files only `scripts/install.sh`; add a check that the `test` job in `ci.yml` has a step whose run line cross-builds `./cmd/acta` with `CGO_ENABLED=0` and names exactly the goreleaser goos and goarch values. Run `scripts/test ./internal/plugincheck/ -run TestRelease` and see it fail.
- [x] Code: drop windows, the zip override and the install.ps1 extra file from `.goreleaser.yaml`; add to the `test` job in `ci.yml` a step that loops over `darwin linux` and `amd64 arm64` running `GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -o /dev/null ./cmd/acta`. Plain comments saying why.
- [x] Run `scripts/test ./internal/plugincheck/` passes, gofmt and vet clean, commit `fix(release): ship macOS and Linux only and cross-build every target in CI`.

### Task 02: README and wiki

**Files:**
- Modify: `plugin/README.md`, `.acta/wiki/release.md`

**verify:** No text a user or agent reads promises a Windows binary or `install.ps1` download today. List every file searched and every line changed.

- [x] Check first: `grep -rn -i "install.ps1\|windows" plugin/README.md .acta/wiki/release.md` lists the lines to change; `scripts/test ./internal/plugincheck/` passes before and after.
- [x] `plugin/README.md`: drop the Windows install block; add one line that Windows is not supported yet. `.acta/wiki/release.md`: build list is darwin and linux, assets are the tar.gz files, `checksums.txt` and `install.sh`, and the ci line names the cross-build step; bump `timestamp`; body under 250 words; `acta wiki check main..HEAD` prints nothing.
- [x] Commit `docs: release covers macOS and Linux only for now`.

## Polish

### Task 03: Review polish

**verify:** Each NOTE below applied, and nothing else changes.

- [x] `.github/workflows/ci.yml`: the header comment no longer says install.ps1 is run for real on Windows; it says the Windows job runs the hook scripts through Git Bash and parse-checks install.ps1 so it stays ready for the Windows port.
- [x] Commit: `polish: review notes PLN-0124`

## Review notes

- Round 1: full tier, two reviewers, both CLEAN. Polish: light tier (one comment in ci.yml), read inline by the orchestrator, CLEAN.
- The ci cross-build check matches goos and goarch words, not the loop shape; a step that skipped a pair but named it in a comment would pass.
- The wiki release.md edit reworded some lines that have nothing to do with Windows; the meaning is the same.
- release.yml runs the full `go test ./...`, not -short; not yet proven on a clean runner (see DBT-0102).
