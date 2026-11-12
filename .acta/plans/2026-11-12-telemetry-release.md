---
parent: specs/2026-11-12-usage-telemetry-design
depth: minimal
id: PLN-0230
created: "2026-11-12 06:31:58"
hash: tl164uk
after: [PLN-0228]
started: "2026-11-12 07:30:59"
finished: "2026-11-12 07:35:24"
---
# Telemetry endpoint in releases, setup line and README

**Goal:** Release builds carry the telemetry endpoint from a GitHub Actions secret, `acta setup` shows the telemetry state, and the README explains what is counted and how to turn it off.

**Spec:** `.acta/specs/2026-11-12-usage-telemetry-design.md` (SPC-0201)

**Tests:** fast `scripts/test ./scripts ./internal/cli`; full `scripts/test --full`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The endpoint URL never appears in the repo; only the secret name `ACTA_TELEMETRY_ENDPOINT` does.
- Uses `telemetry.Endpoint` and the store status from the core plan.

## Waves

- Wave 1: Task 1, Task 2

### Task 1: Endpoint from the release secret

**Files:** `.goreleaser.yaml`, `.github/workflows/release.yml`, `scripts/release_test.go`
**Seam:** the text of `.goreleaser.yaml` and `release.yml`, read by `scripts/release_test.go`
**verify:** Both goreleaser builds (`acta` and `acta-cli`) set `-X github.com/iyay/acta/internal/telemetry.Endpoint={{ .Env.ACTA_TELEMETRY_ENDPOINT }}`, and the release job passes `ACTA_TELEMETRY_ENDPOINT` from `secrets.ACTA_TELEMETRY_ENDPOINT` to goreleaser. No URL is written in either file. An empty secret gives a build with no endpoint.
- [x] Red: `release_test.go` fails because neither build sets `telemetry.Endpoint`.
- [x] Change: add the `-X` line to both builds and the env line to the release step.
- [x] Commit: build(release): bake the telemetry endpoint from a release secret

### Task 2: Setup line and README section

**Files:** `internal/cli/setup_cmd.go`, `internal/cli/setup_cmd_test.go`, `README.md`
**Seam:** `cli.Run([]string{"setup", ...})` output, and the README text
**verify:** Setup prints exactly one telemetry line naming the current state and `acta telemetry off`, in every setup mode, and asks no new question; the form and the keys it writes are unchanged. The README has a Telemetry section that lists what is counted, what is never sent, that only release builds send, and the three ways to turn it off (`acta telemetry off`, `DO_NOT_TRACK=1`, `ACTA_TELEMETRY=0`).
- [x] Red: `setup_cmd_test.go` fails because setup prints no telemetry line.
- [x] Change: print the line after the setup summary and add the README section.
- [x] Commit: docs(telemetry): show the telemetry state in setup and explain it in the README

## Polish

### Task 3: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] [fix] .goreleaser.yaml: a local goreleaser run with ACTA_TELEMETRY_ENDPOINT unset fails on the missing key; use {{ envOrDefault "ACTA_TELEMETRY_ENDPOINT" "" }} in both builds and update the release test to match.
- [x] [fix] .acta/wiki/release.md: put back the fact that both builds strip with -s -w, within the 250-word cap.
- [x] Commit: `polish: review notes for PLN-0230`

## Review notes

- Level deep (review_depth), deep lens on the release secret. BLOCKER 0, NOTE 15, pre-existing 1 (Endpoint scheme is not checked to be https), round 1 CLEAN on both axes; polish (goreleaser envOrDefault, -s -w fact back in the wiki) CLEAN. Open notes: setup line prints outside the rail, wrong source if the home folder cannot be found, off (environment) and off (setting) setup cases untested, one long README line, the baked URL is readable in the binary (accepted by the spec), the README lives in plugin/README.md since the repo has no root README.

## State

### Next

SPC-0201 parallel run
PLN-0228 landed
PLN-0229 running (subagent, worktree ../acta-telemetry-hooks)
PLN-0230 review round 1 (worktree ../acta-telemetry-release)
