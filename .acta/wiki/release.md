---
type: Runbook
title: Cutting a release
description: How to cut a release with scripts/release, what the Release action then does, and the rules that keep versions honest
paths: [scripts/release, scripts/install.sh, scripts/install.ps1, .goreleaser.yaml, .github/workflows/, internal/update/]
---

From a clean `main`:

1. Run `scripts/release patch` (or `minor`, or `X.Y.Z`). It checks `main`, a clean tree and a new tag, writes the version into the three plugin version files, and runs `scripts/test --full`, `go vet ./...` and `gofmt -l .` (a failed gate restores them). Then it makes one commit `chore(release): vX.Y.Z` and one annotated tag.
2. It never pushes. Run the `git push origin main vX.Y.Z` it prints.
3. A `v*` tag starts the Release action (`.github/workflows/release.yml`).

It runs `go test ./...`, then goreleaser (`.goreleaser.yaml`), which builds static darwin and linux binaries (amd64, arm64). It publishes `acta_<os>_<arch>.tar.gz`, `acta-cli_<os>_<arch>.tar.gz` (no TUI, `ACTA_TUI=0`), `checksums.txt` and `install.sh`. Windows is not released yet (SCR-0058). Names carry no version, so `releases/latest/download/<file>` links stay stable.

`acta update` fetches its own build's archive and `checksums.txt`. Do not rename them.

`.goreleaser.yaml` strips with `-s -w` and sets ldflag `-X github.com/iyay/acta/internal/update.release=true`; `update.IsRelease()` reads it. `acta update` refuses a binary without it, so it never replaces a source build.

Both builds set `telemetry.Endpoint` from the `ACTA_TELEMETRY_ENDPOINT` secret through `release.yml` (SPC-0201); empty means no telemetry.

`install.sh` checks SHA-256: it catches a broken download, not a tampered release.

Rules:

- Plans never bump the version. Only `scripts/release` does.
- `release.yml` refuses a tag that is not `v` plus the `plugin.json` version.
- A tagged commit never folds (`canFold` in `internal/gitc/gitc.go`), so a planning write cannot rewrite it.
- `ci.yml` runs on every push: vet, gofmt, tests, the "Cross-build release targets" step (both builds, four pairs), and a Windows job.
