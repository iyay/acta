---
id: SCR-0052
hash: mmzx0dw
title: Curl install script and release pipeline
status: brainstorming
created: "2026-10-07 09:32:23"
schema: "1"
started: "2026-10-07 13:24:34"
---
# Curl install script and release pipeline

## Words

### 2026-10-07

Split from SCR-0051 on 2026-10-07. A curl install script (`curl ... | sh`) that downloads a prebuilt acta binary for the user's platform, then execs `acta setup`. Needs a release pipeline first: the repo has no git remote, no GitHub releases and no goreleaser config yet. Open: where binaries are hosted, which platforms, checksum or signature check in the script.

## Context

Ruling 2026-10-07: plugin files ship inside the binary (go:embed, extracted to ~/.acta/plugin by acta setup). The curl script only needs to fetch the binary and exec `acta setup`.

## Log

### 2026-10-07

Q1 hosting: GitHub Releases on a public repo (goreleaser + GitHub Actions). User answer 2026-10-07.

### 2026-10-07

Q2 platforms: macOS + Linux (amd64, arm64) plus native Windows (amd64, arm64) with an install.ps1. User answer 2026-10-07.

### 2026-10-07

Q3 Windows hook scope: spike first. Check whether the bash hooks run under Claude Code and omp on Windows, then decide between binary + install.ps1 only and cross-platform hooks. User answer 2026-10-07.

### 2026-10-07

Q4 Windows test: free path. CI runs plugin/hooks scripts through Git Bash on windows-latest and checks output; docs research covers how Claude Code and omp call hooks on Windows. No paid API key test for now. User answer 2026-10-07.

### 2026-10-07

Q5 verification: SHA-256 only. goreleaser checksums.txt, install script checks the downloaded archive hash. No signatures. User answer 2026-10-07.

### 2026-10-07

Spike result (docs research, 2026-10-07): Claude Code on Windows runs string hook commands through Git Bash, else PowerShell (bash scripts fail). Extensionless shebang scripts run under Git Bash. CLAUDE_PLUGIN_ROOT path mangling is a known bug (anthropics/claude-code #18527, #21878). Without Git Bash the Bash tool is absent, so Bash matchers in hooks.json never fire. Fix path: exec-form hooks calling acta directly (loses default-rules.md fallback) and matchers Bash|PowerShell. omp: plugin/omp/index.ts spawns acta directly with no shell, so a Windows acta.exe on PATH should be enough; unverified on Windows.

### 2026-10-07

Q6 release trigger: manual tag, CI builds on tag push and refuses a tag that does not match plugin.json. User also wants a release script that does the version bump and the rest. User answer 2026-10-07.

### 2026-10-07

Q7 version policy change: plans stop bumping the version; only scripts/release bumps it (then tags). Changes the CLAUDE.md Version rule and anything that enforces a per-plan bump. User answer 2026-10-07.

### 2026-10-07

Q8 Windows hooks: keep bash hooks, Git Bash required. install.ps1 checks for Git Bash and warns when missing. hooks.json matchers widen to Bash|PowerShell. CLAUDE_PLUGIN_ROOT path bug filed as debt. User answer 2026-10-07.

### 2026-10-07

Q9 install location: per-user, no sudo. ~/.local/bin/acta on macOS/Linux, %LOCALAPPDATA%\acta\bin\acta.exe on Windows, ACTA_INSTALL_DIR overrides. Script prints the PATH line when missing and never edits rc files or user PATH. User answer 2026-10-07.

### 2026-10-07

Approach: goreleaser + GitHub Actions on tag push v*. Separate windows-latest CI runs plugin/hooks through Git Bash on every push. Depends on PLN-0113 landing first (embedded plugin + acta setup). User answer 2026-10-07.

### 2026-10-07

Repo creation on hold 2026-10-07: user wants to rewrite git history first (all authors to rtriharyana@gmail.com, maybe more) before gh repo create iyay/acta --public. Not part of this spec; a prerequisite before the first release. Quick regex scan of all history found no secret patterns; /Users/iyay appears 19 times. Worktree setup-tidy (PLN-0113) is open, so a rewrite should wait until it lands.

### 2026-10-08

Design section 1 approved 2026-10-08 (release pipeline): .goreleaser.yaml builds cmd/acta for darwin/linux/windows x amd64/arm64 with CGO off; tar.gz (zip on Windows); archive names carry no version so releases/latest/download URLs work; checksums.txt SHA-256; no ldflags since the version comes from the embedded plugin.json. release.yml on tag v*: tag must equal v + plugin.json version, then go test, then goreleaser; install.sh and install.ps1 upload as release assets. ci.yml on every push: ubuntu vet + test, windows-latest runs plugin/hooks scripts through Git Bash and checks output.

### 2026-10-08

Design section 2 approved 2026-10-08 (scripts/release + version rule): POSIX sh scripts/release <patch|minor|x.y.z> refuses off main, dirty tree or existing tag; writes the version to plugin.json, marketplace.json and package.json; runs scripts/test --full, go vet and gofmt -l, and reverts the three files on red; commits chore(release): vX.Y.Z and makes annotated tag vX.Y.Z; never pushes, prints the push command. CLAUDE.md Version section becomes: plans never bump; only scripts/release bumps and tags. plugincheck keeps the three-files-agree check. Tests run in a temp clone like scripts/test_test.go. DBT-0101.03 folds in: canFold refuses to fold a HEAD that a tag points at.

### 2026-10-08

Design section 3 approved 2026-10-08 (install scripts): install.sh (POSIX sh, curl | sh from releases/latest/download) detects OS and arch with uname, ACTA_VERSION pins, checks SHA-256 against checksums.txt with sha256sum or shasum -a 256 and installs nothing on mismatch, extracts to ACTA_INSTALL_DIR or ~/.local/bin, prints the PATH line and never edits rc files, then runs acta setup </dev/tty (setup needs a TTY stdin) or prints the command when no /dev/tty. install.ps1 (irm | iex) does the same with Get-FileHash into %LOCALAPPDATA%\acta\bin and warns when Git Bash is missing. Tests: install.sh from Go against a local fake HTTP server via ACTA_DOWNLOAD_URL (hash ok, hash bad, unknown platform, dir not on PATH); install.ps1 parse-checked on windows-latest, run against a real release after the first one.

### 2026-10-08

Design section 4 approved 2026-10-08 (Windows hooks, docs, wiki): hooks.json matchers widen to Bash|PowerShell|Read|Edit|Write|MultiEdit (pre) and Bash|PowerShell (post); internal/hook treats PowerShell like Bash when reading tool_input.command. CLAUDE_PLUGIN_ROOT Windows path bug (anthropics/claude-code #18527, #21878) filed as debt, not fixed. plugin/README.md Install switches to curl | sh and irm | iex, go install becomes the dev path, marketplace steps go (acta setup does them), Windows note says Git Bash is required. New wiki Runbook release.md written in build. Public repo creation and first push stay with the user, before the first release.

## Open questions
