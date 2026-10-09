---
type: Runbook
title: Cutting a release
description: How to cut a release with scripts/release, what the Release action then does, and the rules that keep versions honest
paths: [scripts/release, scripts/install.sh, scripts/install.ps1, .goreleaser.yaml, .github/workflows/, internal/update/]
timestamp: 2026-10-09T09:26:54Z
---

From a clean `main`:

1. Run `scripts/release patch` (or `minor`, or `X.Y.Z`). It checks `main`, a clean tree and a new tag. It writes the version into the three plugin version files. It runs `scripts/test --full`, `go vet ./...` and `gofmt -l .`; a failed gate restores the files. If all pass, it makes one commit `chore(release): vX.Y.Z` and one annotated tag.
2. It never pushes. Run the `git push origin main vX.Y.Z` it prints.
3. A `v*` tag starts the Release action (`.github/workflows/release.yml`).

It runs `go test ./...`, then goreleaser (`.goreleaser.yaml`), which builds static darwin and linux binaries (amd64, arm64). It publishes `acta_<os>_<arch>.tar.gz`, `checksums.txt` and `install.sh`. Windows is not released yet (SCR-0058). Names carry no version, so `releases/latest/download/<file>` links stay stable.

`acta update` downloads `acta_<os>_<arch>.tar.gz` and `checksums.txt` from the same release. Do not rename them.

`.goreleaser.yaml` sets ldflag `-X github.com/iyay/acta/internal/update.release=true`; `update.IsRelease()` reads it. `acta update` refuses a binary without it, so a source build is never replaced by a release.

`install.sh` checks SHA-256: it catches a broken download, not a tampered release.

Rules:

- Plans never bump the version. Only `scripts/release` does.
- `release.yml` refuses a tag that is not `v` plus the `plugin.json` version.
- A commit a tag points at never folds. `canFold` in `internal/gitc/gitc.go` says no, so a planning write cannot rewrite the release commit.
- `ci.yml` runs on every push: vet, gofmt, tests, the "Cross-build release targets" step (builds `./cmd/acta` for the four pairs, `CGO_ENABLED=0`), and a Windows job.
