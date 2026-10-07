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

## Open questions
