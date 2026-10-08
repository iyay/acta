---
type: Runbook
title: Cutting a release
description: How to cut a release with scripts/release, what the Release action then does, and the rules that keep versions honest
paths: [scripts/release, scripts/install.sh, scripts/install.ps1, .goreleaser.yaml, .github/workflows/]
timestamp: 2026-10-08T03:48:37Z
---

Steps, from a clean `main`:

1. Run `scripts/release patch` (or `minor`, or `X.Y.Z`). It checks for `main`, a clean tree and a new tag. It writes the new version into `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`. It runs `scripts/test --full`, `go vet ./...` and `gofmt -l .`. If any gate fails, it puts the three files back. If all pass, it makes one commit, `chore(release): vX.Y.Z`, and one annotated tag `vX.Y.Z`.
2. The script never pushes. Run the `git push origin main vX.Y.Z` it prints.
3. Watch the Release action. A `v*` tag starts it.

What the Release action does (`.github/workflows/release.yml`): it checks that the tag is `v` plus the version in `plugin.json`, runs `go test ./...`, then goreleaser (`.goreleaser.yaml`), which builds static binaries for darwin, linux and windows on amd64 and arm64. It publishes `acta_<os>_<arch>.tar.gz` (a zip on windows), `checksums.txt`, `install.sh` and `install.ps1`. Names carry no version, so `releases/latest/download/<file>` links never change.

`install.sh` downloads over HTTPS only. SHA-256 catches a broken download, not a tampered release, since `checksums.txt` ships beside it.

Rules:

- Plans never bump the version. Only `scripts/release` does.
- `release.yml` refuses a tag that is not `v` plus the `plugin.json` version. The check runs before anything is published.
- A commit a tag points at never folds. `canFold` in `internal/gitc/gitc.go` says no, so a later planning write cannot rewrite the release commit.
- `ci.yml` runs on every push: vet, gofmt, tests, and a Windows job that runs each hook without `acta` and parses `install.ps1`.
