---
id: SCR-0052
hash: mmzx0dw
title: Curl install script and release pipeline
status: raw
created: "2026-10-07 09:32:23"
schema: "1"
---
# Curl install script and release pipeline

## Words

### 2026-10-07

Split from SCR-0051 on 2026-10-07. A curl install script (`curl ... | sh`) that downloads a prebuilt acta binary for the user's platform, then execs `acta setup`. Needs a release pipeline first: the repo has no git remote, no GitHub releases and no goreleaser config yet. Open: where binaries are hosted, which platforms, checksum or signature check in the script.

## Context

## Log

## Open questions
