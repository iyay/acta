---
parent: specs/2026-10-07-acta-setup-install-rows
depth: minimal
closes: [SPC-0103]
id: PLN-0112
created: "2026-10-07 13:07:53"
hash: onfa7nh
started: "2026-10-07 13:08:12"
finished: "2026-10-07 13:09:11"
---
# acta setup install rows Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** the plugin install rows in `acta setup` line up, with a gap between each tool name and its Yes/No marks.

**Spec:** .acta/specs/2026-10-07-acta-setup-install-rows.md

**Tests:** `scripts/test ./internal/setup/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Only the install screen's look and its title/description change. Tests pin HOME, PM_VOICE_FILE and TMPDIR to `t.TempDir()`. Run tests with `scripts/test`, never bare `go test`. Never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01, Task 02

### Task 01: aligned install rows

**Files:** `internal/setup/form.go`, `internal/setup/form_test.go` or `internal/setup/form_text_test.go`.

**verify:** For any set of tool names (one, two or three, different lengths), every rendered install row has its first Yes/No mark in the same column, at least one space after the longest name; the title is `Plugin install` with description `One row per tool found.`; the answered line format is unchanged. List every name set rendered.

- [x] **Test:** render the harness group for name sets like [claude], [claude, omp], [a, claude, omp] and check the mark column and the gap; check the title and description; it fails on today's `claude● Yes`.
- [x] **Code:** pad each confirm title to the longest tool name plus two spaces (or put the gap in the theme), and change the title and description.
- [x] **Commit:** `fix(setup): line up the plugin install rows`.

### Task 02: version bump

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.

**verify:** All three files hold 0.1.34 and `internal/plugincheck` accepts it.

- [x] **Test:** the existing version check in `internal/plugincheck`.
- [x] **Code:** set 0.1.34 in all three files.
- [x] **Commit:** `chore(plugin): bump version`.

### Task 03: shorter closing summary

**Files:** `internal/setup/look.go`, `internal/setup/look_test.go`, plus any setup or cli test that pins the old summary lines.

**verify:** On every run path the closing summary prints only `config: <path>` (or `config: (unknown)`) on the rail, then `└  Run acta in a repo to browse specs, plans and bugs.`; no `installed:`, `block:` or `next:` line appears in any form, with installs or none, with block files or none. List every path checked.

- [x] **Test:** summary tests for installs and none, block files and none, unknown config path; they fail on today's `installed:`/`block:`/`next:` lines.
- [x] **Code:** cut the three lines from `SummaryBox` and put the TUI hint on the closing `└` line; drop parameters that become unused.
- [x] **Commit:** `fix(setup): shorter closing summary with the TUI hint`.

## Review notes

- Install row padding counts bytes with `len`, not screen width; fine while every tool name is ASCII.
- `TestApplySetupHardError` in `internal/cli/setup_cmd_test.go` keeps the old `applySetup` name but now tests `setup.Apply`; `internal/setup/run_test.go` would be its natural home.
- Dropping `failTracker` and `TestApplySetupSkipsFailedInstall` lost no live path: the summary no longer lists installs, and `run_test.go` still pins "a failed install prints its command and carries on".
- No eval run: this branch touches no file under plugin/skills or plugin/hooks.
