---
parent: specs/2026-10-07-acta-setup-question-text
depth: minimal
closes: [SPC-0102]
id: PLN-0111
created: "2026-10-07 12:59:29"
hash: w9waw9o
started: "2026-10-07 12:59:53"
finished: "2026-10-07 13:01:22"
---
# acta setup question text Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta setup` shows short titles, only useful descriptions, a note on every option, and never speaks as "I".

**Spec:** .acta/specs/2026-10-07-acta-setup-question-text.md

**Tests:** `scripts/test ./internal/setup/ ./internal/cli/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Text exactly as the spec table says. Only wizard text changes; values written, order, layout, colors and logic stay.
- Tests pin HOME, PM_VOICE_FILE and TMPDIR to `t.TempDir()`. Run tests with `scripts/test`, never bare `go test`. Never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01, Task 02

### Task 01: question text

**Files:** `internal/setup/form.go`, `internal/setup/form_test.go`, plus any setup or cli test that pins the old titles.

**verify:** Every title, description (or no description line) and option note matches the spec table on every env (with and without `HERDR_ENV=1`, with and without harnesses); no text the wizard prints matches the whole words I, me or my; the answered line shows the bare option value, not its note. List every question and env checked.

- [x] **Test:** a table test of all 8 questions against the spec, a no-"I" scan of every title, description and option label, and an answered-line check; it fails on today's text.
- [x] **Code:** replace the titles, descriptions and option labels in `form.go`; drop the description line when the spec says none.
- [x] **Commit:** `feat(setup): short titles and useful help text`.

### Task 02: version bump

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.

**verify:** All three files hold 0.1.33 and `internal/plugincheck` accepts it.

- [x] **Test:** the existing version check in `internal/plugincheck`.
- [x] **Code:** set 0.1.33 in all three files.
- [x] **Commit:** `chore(plugin): bump version`.

## Review notes

- `internal/setup/export_test.go` is the repo's first export_test file; it exposes `AnsweredLines` to the external test package only.
- `pinHome` in `form_text_test.go` repeats the HOME/PM_VOICE_FILE/TMPDIR setup other setup tests do inline; one shared helper would be a later cleanup.
- `TestFormAnsweredLine` swaps the global lipgloss color profile; safe while no setup test runs with `t.Parallel`.
- No eval run: this branch touches no file under plugin/skills or plugin/hooks.
