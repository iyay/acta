---
parent: specs/2026-10-08-curl-install-release-pipeline-design
closes: [SCR-0052, DBT-0101.03]
depth: minimal
id: PLN-0122
created: "2026-10-08 10:23:15"
hash: w5k2o2v
started: "2026-10-08 10:26:28"
---
# Curl install script and release pipeline

**Goal:** A tag push builds acta for six platforms on GitHub Releases, `scripts/release` is the only thing that bumps the version and tags, and `install.sh` / `install.ps1` fetch a hash-checked binary and run `acta setup`.

**Spec:** .acta/specs/2026-10-08-curl-install-release-pipeline-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- No plan bumps the version, this one included. Only `scripts/release` writes it, to `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`.
- Nothing in this plan pushes or creates a remote. `scripts/release` prints the push command and never runs it.
- Shell scripts are POSIX `sh` (no bashisms), like `scripts/test`.
- Repo URL is exactly `https://github.com/iyay/acta`. Release asset names carry no version: `acta_<os>_<arch>.tar.gz`, `acta_windows_<arch>.zip`, `checksums.txt`, `install.sh`, `install.ps1`. `<os>` is `darwin`, `linux` or `windows`; `<arch>` is `amd64` or `arm64`.
- Download URLs: latest is `<base>/latest/download/<file>`, pinned is `<base>/download/<ACTA_VERSION>/<file>`, where `<base>` is `ACTA_DOWNLOAD_URL`, default `https://github.com/iyay/acta/releases`.
- Install dir: `ACTA_INSTALL_DIR`, else `~/.local/bin` (sh) or `%LOCALAPPDATA%\acta\bin` (ps1). Scripts never edit rc files or the user PATH.
- Tests use temp dirs and temp repos only; never run `scripts/release` or acta write commands in this repo.
- Comments in plain short English that say why.

## Waves

- Wave 1: Task 01, Task 02, Task 03, Task 04
- Wave 2: Task 05
- Wave 3: Task 06

### Task 01: Never fold a tagged HEAD

**Files:**
- Modify: `internal/gitc/gitc.go`
- Test: `internal/gitc/gitc_test.go`

**verify:** `canFold` returns false whenever any tag (lightweight or annotated) points at HEAD, and every older guard still holds; a tagged commit's hash never changes through `CommitOrFold`. List each path checked and its result.

- [x] Failing test: in `TestCommitOrFold`, a one-file `chore(spec): x` HEAD with a lightweight tag on it, then a second edit of that file, gives a new commit and the tag still points at the old hash; same with an annotated tag. Run `scripts/test ./internal/gitc/ -run TestCommitOrFold` and see it fail.
- [x] Code: in `canFold`, before the branch checks, run `git tag --points-at HEAD`; any output or error means false. One plain comment: a tag marks a release, so its commit must never be rewritten.
- [x] Run `scripts/test ./internal/gitc/` passes, `go vet ./internal/gitc/` and `gofmt -l internal/gitc` clean, commit `fix(gitc): never fold a commit a tag points at`.

### Task 02: Hooks fire on the PowerShell tool

**Files:**
- Modify: `plugin/hooks/hooks.json`
- Test: `internal/plugincheck/plugin_test.go`, `internal/hook/gotest_test.go`

**verify:** Every hook check that reads `tool_input.command` for a `Bash` call does the same for a `PowerShell` call, and the matchers name `PowerShell` wherever they name `Bash`. List each hook check and its result for both tool names.

- [x] Failing test: in `plugin_test.go` change the expected PreToolUse matcher to `Bash|PowerShell|Read|Edit|Write|MultiEdit` and PostToolUse to `Bash|PowerShell`. In `gotest_test.go` add a case to `TestGoTestBlock` that sends `tool_name: "PowerShell"` with `tool_input.command: "go test ./..."` in a repo with `scripts/test` and expects the block. Run `scripts/test ./internal/plugincheck/ ./internal/hook/ -run 'TestGoTestBlock|TestHooksJSON'` and see the plugincheck one fail.
- [x] Code: change the two matchers in `hooks.json`. `internal/hook` already reads `ToolInput.Command` without looking at `ToolName`, so no Go change unless the new hook case fails; then make it treat `PowerShell` like `Bash`.
- [x] Run the same command passes, commit `feat(hooks): fire Bash hooks on the PowerShell tool too`.

### Task 03: scripts/release bumps, gates and tags

**Files:**
- Create: `scripts/release`, `scripts/release_test.go`

**verify:** `scripts/release` makes exactly one commit `chore(release): vX.Y.Z` and one annotated tag `vX.Y.Z` only when the branch is `main`, the tree is clean, the tag is new and every gate is green; on every other path it leaves no commit, no tag and the three version files unchanged. It never runs `git push`. List each path checked and its result.

- [x] Failing test: `scripts/release_test.go` (package `scripts`, reuse the fake-binary pattern of `runScript` in `test_test.go`). Each case builds a temp git repo on `main` holding a copy of `scripts/release`, a fake `scripts/test` that exits 0 or 1, and the three version files at `0.1.43` with the same field shape as the real ones; a fake `go` and `gofmt` on PATH exit 0 with no output. Cases: `patch` gives `0.1.44` in all three files, one commit, tag `v0.1.44` annotated, output names `git push origin main v0.1.44`; `minor` gives `0.2.0`; `1.2.3` gives `1.2.3`; bad argument fails; on another branch fails; dirty tree fails; existing tag fails; red `scripts/test` fails with files restored, no new commit, no tag. Run `scripts/test ./scripts/ -run TestRelease` and see it fail.
- [x] Code: `scripts/release`, POSIX `sh`, `set -eu`, `cd` to the repo root. Read the current version from `plugin/.claude-plugin/plugin.json`; compute the new one; check branch, `git status --porcelain`, `git rev-parse -q --verify refs/tags/vX.Y.Z`; rewrite the `"version": "..."` line in the three files with `sed` into a temp file and `mv`; run `scripts/test --full`, `go vet ./...`, and fail when `gofmt -l .` prints anything; on failure `git checkout --` the three files and exit 1; then `git commit -m "chore(release): vX.Y.Z" -- <three files>`, `git tag -a vX.Y.Z -m vX.Y.Z`, print the push line. Mark it executable.
- [x] Run `scripts/test ./scripts/` passes, commit `feat(scripts): release script bumps the version, runs the gates and tags`.

### Task 04: install.sh fetches a hash-checked binary

**Files:**
- Create: `scripts/install.sh`, `scripts/install_test.go`

**verify:** `install.sh` puts `acta` in the install dir only when the archive's SHA-256 matches its line in `checksums.txt`; on a mismatch, a missing line, a failed download or an unknown OS or arch it exits non-zero and the install dir holds no `acta`. It never edits rc files. List each path checked and its result.

- [x] Failing test: `scripts/install_test.go` builds a `.tar.gz` holding a fake `acta` shell script that writes its args to a file, serves it plus `checksums.txt` from `httptest.Server` under both `/latest/download/` and `/download/v9.9.9/`, and runs `sh scripts/install.sh` with `ACTA_DOWNLOAD_URL`, `ACTA_INSTALL_DIR` set to temp dirs and a fake `uname` on PATH. Cases: good hash installs and the fake acta was called with `setup` or the output says `run: acta setup`; `ACTA_VERSION=v9.9.9` hits the pinned path; bad hash fails and nothing lands; fake `uname` saying `Plan9` fails with a message naming the platform; install dir not on PATH prints an `export PATH=` line. Run `scripts/test ./scripts/ -run TestInstall` and see it fail.
- [x] Code: `scripts/install.sh`, POSIX `sh`, `set -eu`. Map `uname -s` (`Darwin`, `Linux`) and `uname -m` (`x86_64`/`amd64`, `arm64`/`aarch64`) to the names in Global Constraints; download with `curl -fsSL` into `mktemp -d` (cleaned by `trap`); pick `sha256sum` else `shasum -a 256`; compare to the `checksums.txt` line for the archive; `tar -xzf` and install with `mkdir -p` plus `cp` and `chmod 755`; print the PATH line when the dir is not in `$PATH`; when `( : </dev/tty ) 2>/dev/null` works run `acta setup </dev/tty`, else print `run: acta setup`.
- [x] Run `scripts/test ./scripts/` passes, commit `feat(scripts): install.sh downloads, checks and installs acta`.

### Task 05: goreleaser, workflows and install.ps1

**Files:**
- Create: `.goreleaser.yaml`, `.github/workflows/release.yml`, `.github/workflows/ci.yml`, `scripts/install.ps1`, `internal/plugincheck/release_test.go`

**verify:** The release config builds all six os/arch pairs with CGO off, names every archive without a version, ships `checksums.txt`, `install.sh` and `install.ps1` as assets, and the release workflow cannot publish a tag that differs from `v` plus the `plugin.json` version. List each property and where the test checks it.

- [ ] Failing test: `internal/plugincheck/release_test.go` parses `.goreleaser.yaml` with `gopkg.in/yaml.v3` and checks: main `./cmd/acta`, env `CGO_ENABLED=0`, goos `darwin linux windows`, goarch `amd64 arm64`, archive `name_template` is `acta_{{ .Os }}_{{ .Arch }}` with no `Version`, a windows `zip` format override, checksum `name_template: checksums.txt` with `algorithm: sha256`, `release.extra_files` lists `scripts/install.sh` and `scripts/install.ps1` and both files exist. It reads `release.yml` and checks the trigger is tags `v*` and a step compares `GITHUB_REF_NAME` with the version read from `plugin/.claude-plugin/plugin.json` before the goreleaser step. It reads `ci.yml` and checks a `windows-latest` job with `shell: bash` runs `plugin/hooks/` scripts and a step parses `scripts/install.ps1`. Run `scripts/test ./internal/plugincheck/ -run TestRelease` and see it fail.
- [ ] Code: write `.goreleaser.yaml` (version 2) to match. `release.yml`: on push tags `v*`; checkout, setup-go from `go.mod`, a tag check step with `jq -r .version plugin/.claude-plugin/plugin.json`, `go test ./...`, `goreleaser/goreleaser-action` with `args: release --clean` and `GITHUB_TOKEN`. `ci.yml`: on push; job `test` on `ubuntu-latest` runs `go vet ./...`, fails on `gofmt -l .` output, `go test ./...`; job `windows-hooks` on `windows-latest` with `shell: bash` pipes sample `{"tool_name":"Bash","tool_input":{"command":"ls"}}` JSON into each `plugin/hooks/*` script and checks exit 0, and a `pwsh` step parses `scripts/install.ps1` with `[System.Management.Automation.Language.Parser]::ParseFile` and fails on errors. `scripts/install.ps1`: arch from `$env:PROCESSOR_ARCHITECTURE` (`AMD64`, `ARM64`), same URL and env rules as Global Constraints, `Invoke-WebRequest` into a temp dir, `Get-FileHash -Algorithm SHA256` against `checksums.txt`, `Expand-Archive`, copy `acta.exe` to the install dir, print the PATH step when missing, warn when neither `git.exe` nor Git Bash `bash.exe` is found, then run `acta setup`.
- [ ] Run `scripts/test ./internal/plugincheck/` passes, commit `feat(release): goreleaser config, release and CI workflows, install.ps1`.

### Task 06: Version rule, README, debt and wiki

**Files:**
- Modify: `CLAUDE.md`, `plugin/README.md`
- Create: `.acta/wiki/release.md`, one debt file through `acta debt new`

**verify:** No text in the repo tells a plan to bump the version, the README's first install path is the one-liner for each OS, and the wiki page matches what `scripts/release`, `release.yml` and `ci.yml` do. List every file searched for the old bump rule and every claim in the wiki page with the file that backs it.

- [ ] Check first: `grep -rn -i "adds 1 to the patch" CLAUDE.md plugin internal` shows the old rule only in `CLAUDE.md`; `scripts/test ./internal/plugincheck/` passes before and after (README is checked there).
- [ ] `CLAUDE.md` Version section becomes: "Until the first release the version stays on 0.1.x. Plans never bump the version. Only `scripts/release` bumps it in `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`, then commits and tags. `internal/plugincheck` fails when the three files disagree or the version is not `x.y.z`."
- [ ] `plugin/README.md` Install: step 1 is `curl -fsSL https://github.com/iyay/acta/releases/latest/download/install.sh | sh` for macOS and Linux and `irm https://github.com/iyay/acta/releases/latest/download/install.ps1 | iex` for Windows, saying it installs `acta` and runs `acta setup`, which installs the plugin for Claude Code and omp; a Windows note that Git Bash is required for the hooks; `go install github.com/iyay/acta/cmd/acta@latest` then `acta setup` as the path for developers. Drop the `/plugin marketplace add` and `omp plugin link` steps. Keep the rest of the file as is.
- [ ] Debt: `acta debt new PLN-0122 --title "Windows CLAUDE_PLUGIN_ROOT path bug"` with a body naming anthropics/claude-code #18527 and #21878: on Windows the hook command path can come out mangled, so hooks may not run until that is fixed upstream or hooks call acta directly.
- [ ] Wiki: `.acta/wiki/release.md`, type Runbook, `paths: [scripts/release, scripts/install.sh, scripts/install.ps1, .goreleaser.yaml, .github/workflows/]`. Steps: on clean `main` run `scripts/release patch`, then `git push origin main vX.Y.Z`, then watch the Release action. Rules: plans never bump the version; `release.yml` refuses a tag that is not `v` plus the `plugin.json` version; a tagged commit never folds. Run `acta wiki check main..HEAD` clean.
- [ ] Commit `docs: release runbook, one-line install, plans stop bumping the version`.
