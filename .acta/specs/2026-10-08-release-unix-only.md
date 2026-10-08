---
parent: specs/2026-10-08-curl-install-release-pipeline-design
created: "2026-10-08 12:12:02"
id: SPC-0115
hash: xx9en35
---
# First release ships macOS and Linux only, and CI cross-builds every release target

Status: Bounded, approved by the user in chat on 2026-10-08 (option 1). Changes SPC-0113.

Why: the v0.1.44 release run (37730081152) failed in goreleaser. `GOOS=windows` does not compile: `internal/write/runone.go`, `internal/write/tick.go`, `internal/hook/wikihint.go` and `internal/evalomp/run.go` use Unix-only syscalls (`Flock`, `Stat_t`, `Setpgid`, `Kill`). CI never compiled for another OS, so the tag went out first. darwin and linux on amd64 and arm64 build fine. The Windows port is SCR-0058.

Design:
- `.goreleaser.yaml` builds darwin and linux only, on amd64 and arm64. The windows zip override and `scripts/install.ps1` in `release.extra_files` go. `install.ps1` stays in the repo for SCR-0058.
- `ci.yml` job `test` gets a step that cross-builds `./cmd/acta` with `CGO_ENABLED=0` for every goos and goarch pair `.goreleaser.yaml` lists, so a target that does not compile fails CI before anyone tags.
- `internal/plugincheck/release_test.go` follows: goos is `darwin linux`, no format override, extra files are only `scripts/install.sh`, and a new check that the ci cross-build step names exactly the goreleaser goos and goarch pairs. The windows-hooks CI job and its checks stay; they guard the hook scripts for later.
- `plugin/README.md`: the Windows install line goes, with one line saying Windows is not supported yet.
- `.acta/wiki/release.md` drops windows from the build list and the assets, and names the cross-build step.
- Next release is v0.1.45 through `scripts/release patch`. Tag v0.1.44 stays on origin with no release.
