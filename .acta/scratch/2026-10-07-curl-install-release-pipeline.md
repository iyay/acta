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

## Open questions
