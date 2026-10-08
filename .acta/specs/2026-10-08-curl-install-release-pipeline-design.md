---
parent: scratch/2026-10-07-curl-install-release-pipeline
created: "2026-10-08 10:21:02"
id: SPC-0113
hash: kaafb5s
---
# Curl install script and release pipeline

Status: Architectural, approved by the user in chat section by section on 2026-10-08. Answers and rulings are in the Log of SCR-0052.

Why: today a user needs Go and a checkout to install acta (`go install`, then `/plugin marketplace add`). Since PLN-0113 the plugin ships inside the binary and `acta setup` unpacks and installs it. So one prebuilt binary is enough. This spec adds the release pipeline that builds that binary and the scripts that fetch it.

Depends on: PLN-0113 (embedded plugin and `acta setup`), landed. A public GitHub repo `iyay/acta` with `main` pushed. The user creates it; it is not part of this spec, and it must exist before the first release.

## 1. Release pipeline

- `.goreleaser.yaml` builds `./cmd/acta` for darwin, linux and windows, each on amd64 and arm64, with `CGO_ENABLED=0`.
- Archives are `.tar.gz`, and `.zip` on Windows. Archive names carry no version, for example `acta_darwin_arm64.tar.gz`, so `releases/latest/download/<name>` always works.
- goreleaser writes `checksums.txt` with SHA-256 sums.
- No ldflags. The version comes from the embedded `plugin.json` through `plugin.Version()`.
- `.github/workflows/release.yml` runs on a pushed tag `v*`. Steps in order: fail when the tag is not `v` plus the version in `plugin/.claude-plugin/plugin.json`; run the tests; run `goreleaser release`. `install.sh` and `install.ps1` upload as release assets, so each script always matches its binaries.
- `.github/workflows/ci.yml` runs on every push. An ubuntu job runs `go vet`, `gofmt -l` and the tests. A `windows-latest` job runs each `plugin/hooks/*` script through Git Bash with sample hook input and checks its output.

## 2. `scripts/release` and the version rule

- `scripts/release <patch|minor|x.y.z>` is POSIX `sh`, like `scripts/test`.
- It refuses to run when the branch is not `main`, when the worktree has changes, or when the tag already exists.
- It writes the new version to `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`.
- It runs `scripts/test --full`, `go vet ./...` and `gofmt -l .`. On any failure it restores the three files and stops, leaving no commit and no tag.
- On green it commits `chore(release): vX.Y.Z` and makes an annotated tag `vX.Y.Z`.
- It never pushes. It prints `git push origin main vX.Y.Z` for the user to run.
- The Version section of `CLAUDE.md` changes to: plans never bump the version; only `scripts/release` bumps it, then tags. `internal/plugincheck` keeps checking that the three files agree and that the version is `x.y.z`.
- This plan follows the new rule, so it has no version bump task.
- DBT-0101.03 folds in: `gitc` `canFold` refuses to fold when a tag points at HEAD, so an `acta commit` right after a release never rewrites the tagged commit. The plan lists DBT-0101.03 in `closes:`.

## 3. Install scripts

`install.sh`, for macOS and Linux, is POSIX `sh`. Use: `curl -fsSL https://github.com/iyay/acta/releases/latest/download/install.sh | sh`.

- It finds OS and arch with `uname` and fails with a clear message on a platform with no build.
- `ACTA_VERSION=vX.Y.Z` pins a version; the default is latest. `ACTA_DOWNLOAD_URL` overrides the base URL, for tests.
- It downloads the archive and `checksums.txt` into a temp dir and checks the SHA-256 with `sha256sum`, else `shasum -a 256`. On a mismatch it stops and installs nothing.
- It extracts `acta` to `ACTA_INSTALL_DIR`, else `~/.local/bin`, creating the dir when needed. No sudo.
- When that dir is not on PATH, it prints the line to add. It never edits rc files.
- Last it runs `acta setup </dev/tty`, because `acta setup` needs a terminal on stdin (`internal/cli/setup_cmd.go`) and under `curl | sh` stdin is the script. When `/dev/tty` cannot be opened, it prints `run: acta setup` instead.

`install.ps1`, for Windows. Use: `irm https://github.com/iyay/acta/releases/latest/download/install.ps1 | iex`.

- Same flow: arch from the environment, `ACTA_VERSION`, `ACTA_DOWNLOAD_URL`, hash check with `Get-FileHash`, install to `ACTA_INSTALL_DIR`, else `%LOCALAPPDATA%\acta\bin`.
- It prints the PATH step when missing and never edits the user PATH.
- It looks for Git Bash and warns when it is missing, because the hooks need it.
- Last it runs `acta setup`.

Tests:
- `install.sh` runs from a Go test against a local fake HTTP server through `ACTA_DOWNLOAD_URL`. Cases: hash matches and the binary lands; hash is wrong and nothing lands; unknown platform fails; install dir not on PATH prints the PATH line.
- `install.ps1` gets a parse check in the `windows-latest` CI job. After the first release that job also runs it against the real release.

## 4. Windows hooks, docs and wiki

- `plugin/hooks/hooks.json`: the PreToolUse matcher becomes `Bash|PowerShell|Read|Edit|Write|MultiEdit`, the PostToolUse matcher becomes `Bash|PowerShell`.
- `internal/hook` treats a `PowerShell` tool call like a `Bash` call when it reads `tool_input.command`, so checks such as the bare `go test` block work on both.
- The Windows `CLAUDE_PLUGIN_ROOT` path bug (anthropics/claude-code #18527, #21878) is filed as a debt item, not fixed here.
- `plugin/README.md` Install section: the curl and irm one-liners come first; `go install` stays as the path for developers; the `/plugin marketplace add` steps go, since `acta setup` does them; a Windows note says Git Bash is required.
- Wiki: a new Runbook `.acta/wiki/release.md`, written in build. It covers how to cut a release (`scripts/release`, push the tag, watch Actions), the rule that plans never bump the version, and the tag check in `release.yml`.

## Out of scope

- Creating the public repo and the first push.
- Signatures (cosign, GPG). SHA-256 only, per Q5.
- Cross-platform hooks without Git Bash.
- Homebrew, scoop or other package managers.
