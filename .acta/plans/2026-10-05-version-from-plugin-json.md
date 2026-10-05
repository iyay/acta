---
parent: specs/2026-10-05-version-from-plugin-json
depth: minimal
id: PLN-0083
created: "2026-10-05 08:37:30"
hash: leucc00
started: "2026-10-05 11:55:05"
---
# Version from plugin.json Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The TUI and `acta doctor` show the version from `plugin/.claude-plugin/plugin.json`, not the Go pseudo-version.

**Spec:** `.acta/specs/2026-10-05-version-from-plugin-json.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- A bad or empty `version` field gives `dev`, never a panic.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: embed the plugin version

**Files:** Create `plugin/version.go`, `plugin/version_test.go`; modify `internal/cli/cli.go`, `internal/cli/doctor.go`, `internal/cli/doctor_test.go`.
**verify:** No version text acta shows (TUI header, doctor) ever comes from the Go module pseudo-version; each reads the plugin.json version, and every bad input (empty field, bad JSON, missing field) gives `dev`. List every place that shows a version and every bad input checked.
- [x] Failing test: `plugin.Version()` equals the `version` in `.claude-plugin/plugin.json`, and a parse helper given bad JSON or an empty field returns `dev`; fails because the package does not exist.
- [x] Code: `plugin/version.go` embeds `.claude-plugin/plugin.json` with `go:embed` and parses `version`; `cli.go` passes `"v" + plugin.Version()` (or `dev`) to `WithVersion`; `buildVersion` in `doctor.go` uses `plugin.Version()` and keeps the commit in brackets when the build stamped one.
- [x] Commit: `version: TUI and doctor show the plugin.json version`.

### Task 2: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [ ] Failing test: none new; run `scripts/test ./internal/plugincheck ./plugin` before and after the bump to watch it stay green.
- [ ] Code: add 1 to the patch in all three files.
- [ ] Commit: `plugin: bump patch version`.
