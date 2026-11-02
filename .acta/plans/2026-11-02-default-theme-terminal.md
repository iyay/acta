---
parent: specs/2026-11-02-default-theme-terminal
depth: minimal
id: PLN-0208
created: "2026-11-02 11:31:01"
hash: f6j7dx7
---
# Default theme terminal implementation plan

**Goal:** Make `terminal` the default theme and make `acta config show` name the theme in use.
**Spec:** `.acta/specs/2026-11-02-default-theme-terminal.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only their owning packages.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
The theme name is the only default that changes; no saved value moves, and no new theme is added.
`config show` must keep working when a saved theme file is missing or broken: it prints the name and does not call `theme.Load` for it.

## Waves

Wave 1: Task 1, Task 2. They share no file.

### Task 1: default theme is the terminal

**Files:** Modify `internal/theme/theme.go`, `internal/theme/theme_test.go`, `internal/cli/doctor_test.go`, `internal/doctor/doctor_test.go`, `internal/tui/styles_test.go`.
**Seam:** `theme.Default` and `theme.Load("")`.
**verify:** An unset theme resolves by name to `terminal` on every path that used the old default. Cases: `Load("")` returns the theme named `terminal`, and in a clean HOME it is the built-in with an empty background; a user file named `terminal` wins over the built-in; `New` builds styles with the terminal theme, so the background is empty; doctor's no-theme line says `theme terminal loads`; a name that does not load falls back to the terminal theme, not tokyo-night.
- [ ] Red: `TestLoadDefault` (and the doctor, `New` and fallback expectations) fails because the default is still `tokyo-night`.
- [ ] Change: set `Default = "terminal"` in `internal/theme/theme.go`, update those expectations, then run `scripts/test ./internal/theme/ ./internal/doctor/ ./internal/cli/ ./internal/tui/ -run 'Theme|Default|Doctor'` to green.
- [ ] Commit: `feat(theme): default theme is the terminal`.

### Task 2: config show names the theme

**Files:** Modify `internal/cli/config_cmd.go`, `internal/cli/cli_test.go`.
**Seam:** `cli.Run` with `config show` and `config show --json`.
**verify:** Every `config show` prints exactly one `theme:` line and no `(default)` marker. Cases: a fresh home prints `theme: terminal`; `--json` reports `"theme": "terminal"`; after `acta config set --theme dracula` plain and JSON name dracula; `--clear-theme` returns the line to `terminal` instead of dropping it; a saved theme that no longer loads is still printed by name.
- [ ] Red: a test over a fresh home expects `theme: terminal` and fails because `show` prints no theme line.
- [ ] Change: fill an empty theme with `theme.Default` for both outputs, always print the line, update the clear-theme test, then run `scripts/test ./internal/cli/ -run 'TestVoice|TestConfig'` to green.
- [ ] Commit: `feat(config): show names the theme in use`.
