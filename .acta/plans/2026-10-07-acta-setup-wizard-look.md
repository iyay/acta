---
parent: specs/2026-10-07-acta-setup-wizard-look
depth: minimal
closes: [SPC-0101]
id: PLN-0110
created: "2026-10-07 11:28:32"
hash: zbv85ob
---
# acta setup look Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta setup` asks one question per screen with a title, step count and one-line help, in a minimal grey look with one calm blue accent, and prints short doctor, install, block and summary lines.

**Spec:** .acta/specs/2026-10-07-acta-setup-wizard-look.md

**Tests:** `scripts/test ./internal/setup/ ./internal/cli/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- No behaviour change: `Plan`, `Apply`'s actions, `WriteBlock`, the no-TTY guard and the config values written stay exactly as they are. Only what the user sees changes.
- Look: almost all grey; question text in the terminal's normal fg; descriptions and answered lines in dim grey; one accent, a calm blue, only for the focus marker, the selected option and `✓`; bold title with no color; no filled color blocks (yes/no reads `● Yes  ○ No`, focused option reads `› adhd`); errors in the terminal's red. It does not follow the acta theme.
- Colors are `lipgloss.AdaptiveColor` (one value for dark, one for light). Every dim grey keeps a WCAG contrast ratio of at least 1.6 against both `#1a1b26` (dark) and `#ffffff` (light). See wiki `ghostty-minimum-contrast`.
- Only installed dependencies: `huh`, `lipgloss`. Tests pin HOME, PM_VOICE_FILE and TMPDIR to `t.TempDir()`, and never run real claude or omp. Run tests with `scripts/test`, never bare `go test`. Never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01, Task 04
- Wave 2: Task 02, Task 03

Tasks 02 and 03 use the theme and print helpers from Task 01 and touch different files.

### Task 01: setup theme and print helpers

**Files:** `internal/setup/look.go`, `internal/setup/look_test.go` (both new).

**verify:** Every color the setup screens use comes from one place in `look.go`, and no filled background color is set on any style. Every dim grey has contrast 1.6 or more against both `#1a1b26` and `#ffffff`. The helpers print exactly: one line `✓ install checks ok (N)` when every doctor result is ok, else only the results that are not ok; one `acta block → <path>` line per block file; one line per harness install (`✓ claude` or `✗ omp: <command>`); and a short closing box naming the config path, installs, block files and the next step. List every helper and every input state tested.

- [ ] **Test:** `TestDimContrast`, `TestNoBackgrounds`, `TestDoctorSummary` (all ok, one failing), `TestBlockLines`, `TestInstallLines`, `TestSummaryBox`; they fail because `look.go` does not exist.
- [ ] **Code:** add the adaptive palette, `func Theme() *huh.Theme` built on `huh.ThemeBase()` with the palette, and `DoctorSummary([]doctor.Result) string`, `BlockLines(paths []string) string`, `InstallLine(harness string, argv []string, err error) string`, `SummaryBox(...) string`.
- [ ] **Commit:** `feat(setup): minimal mono theme and short print helpers`.

### Task 02: one question per screen

**Files:** `internal/setup/form.go`, `internal/setup/form_test.go` (new or existing).

**verify:** The form has one group per question, in the fixed order (language, style, tone, repo language, build executor, subagent models, plan depth, questions), then one harness group listing every harness found with a yes/no each. Every question group shows `acta setup` as a bold title, its step count `k/n` with the right k and n, and a one-line plain description. The form uses `Theme()` from Task 01. Defaults still come from `FormDefaults`, and `dispatch` is still offered only when `HERDR_ENV=1`. List every group and its title, count and description.

- [ ] **Test:** a test that builds the form for an env with and without harnesses and checks the group count, order, step counts and descriptions; it fails because today the form is one group.
- [ ] **Code:** split `Ask` into one `huh.NewGroup` per question with a description, use `huh.NewForm(...).WithTheme(Theme())`, and keep the answers mapping as it is.
- [ ] **Commit:** `feat(setup): one question per screen with step count`.

### Task 03: short output in `acta setup`

**Files:** `internal/cli/setup_cmd.go`, `internal/cli/setup_cmd_test.go`, `internal/setup/run.go`, `internal/setup/run_test.go`.

**verify:** On every run path `acta setup` prints no full doctor list when all checks pass, never prints the acta block text, prints one line per harness install, and ends with the summary box. A failing check, a failing install and a refused block write each still show their own line with the reason. The no-TTY path prints exactly what it prints today. List every output path tested.

- [ ] **Test:** update `setup_cmd_test.go` and `run_test.go` to expect the short lines from Task 01 (doctor summary, block lines, install lines, summary box); they fail against today's long output.
- [ ] **Code:** replace the doctor print, `showBlock` and the closing line in `setup_cmd.go` with the Task 01 helpers, and make `Apply` print install results with `InstallLine`.
- [ ] **Commit:** `feat(cli): short doctor, install and block output for setup`.

### Task 04: version bump

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.

**verify:** All three files hold the same version, one patch above the current one on main, and `internal/plugincheck` accepts it.

- [ ] **Test:** the existing version check in `internal/plugincheck` is the test.
- [ ] **Code:** add 1 to the patch in all three files.
- [ ] **Commit:** `chore(plugin): bump version`.
