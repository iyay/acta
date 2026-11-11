---
parent: specs/2026-11-12-usage-telemetry-design
depth: minimal
id: PLN-0230
created: "2026-11-12 06:31:58"
hash: tl164uk
after: [PLN-0228]
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
- [ ] Red: `release_test.go` fails because neither build sets `telemetry.Endpoint`.
- [ ] Change: add the `-X` line to both builds and the env line to the release step.
- [ ] Commit: build(release): bake the telemetry endpoint from a release secret

### Task 2: Setup line and README section

**Files:** `internal/cli/setup_cmd.go`, `internal/cli/setup_cmd_test.go`, `README.md`
**Seam:** `cli.Run([]string{"setup", ...})` output, and the README text
**verify:** Setup prints exactly one telemetry line naming the current state and `acta telemetry off`, in every setup mode, and asks no new question; the form and the keys it writes are unchanged. The README has a Telemetry section that lists what is counted, what is never sent, that only release builds send, and the three ways to turn it off (`acta telemetry off`, `DO_NOT_TRACK=1`, `ACTA_TELEMETRY=0`).
- [ ] Red: `setup_cmd_test.go` fails because setup prints no telemetry line.
- [ ] Change: print the line after the setup summary and add the README section.
- [ ] Commit: docs(telemetry): show the telemetry state in setup and explain it in the README
