---
parent: specs/2026-10-07-acta-setup-tidy-install
depth: minimal
closes: [SPC-0104]
id: PLN-0113
created: "2026-10-07 13:19:43"
hash: yk36z7h
started: "2026-10-07 13:20:04"
finished: "2026-10-07 13:25:36"
---
# acta setup tidy install Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta setup` only asks about tools it can really install, keeps every line on the rail, shows the block once, and starts empty on a first run.

**Spec:** .acta/specs/2026-10-07-acta-setup-tidy-install.md

**Tests:** `scripts/test ./internal/setup/ ./internal/cli/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Values written, question order, layout, colors, `WriteBlock` and the no-TTY guard stay as they are. `config.ResolveUser` keeps its legacy fallback for every other caller.
- Reuse doctor's own check for "plugin installed" rather than a second copy of that logic.
- Tests pin HOME, PM_VOICE_FILE and TMPDIR to `t.TempDir()` and never run real claude or omp. Run tests with `scripts/test`, never bare `go test`. Never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01
- Wave 2: Task 02, Task 03
- Wave 3: Task 04

Task 01 changes `Env` and the install actions that Tasks 02 and 03 build on.

### Task 01: honest install step

**Files:** `internal/setup/plan.go`, `internal/setup/plan_test.go`, `internal/setup/form.go`, `internal/setup/form_test.go`, `internal/setup/run.go`, `internal/setup/run_test.go`, `internal/setup/look.go`, `internal/setup/look_test.go`, `internal/cli/setup_cmd.go`, `internal/cli/setup_cmd_test.go`, and the doctor file that holds the plugin check if it needs an exported helper.

**verify:** No tool with the plugin already installed ever gets a Yes/No row, and it shows `✓ <tool>  already installed`. With no `--plugin-dir`, no tool gets a Yes/No row; each tool that needs the plugin shows `▲ <tool>  run by hand:` with every command on its own rail line. With `--plugin-dir`, only not-installed tools get a row. No printed line of any install output starts without the rail. The block shows once as `✓ acta block → <path>` per file, and the config line reads `✓ config saved`. List every mix of installed, not installed and plugin dir tested.

- [x] **Test:** cases for installed, not installed with and without plugin dir, and a mix; a scan that every output line starts with a rail mark; block shown once; `✓ config saved`. They fail on today's output.
- [x] **Code:** add which tools already have the plugin to `Env` (from doctor's check), skip those and skip all rows without a plugin dir in the form and `Plan`, print the run-by-hand lines on the rail, drop the block preview line, and change the saved line.
- [x] **Commit:** `fix(setup): ask only about installs the wizard can do`.

### Task 02: empty defaults on a first run

**Files:** `internal/cli/setup_cmd.go`, `internal/cli/setup_cmd_test.go`, `internal/setup/form.go`, `internal/setup/form_test.go`, plus a small reader in `internal/config/user.go` only if no existing function reads just the config file.

**verify:** With no `~/.acta/config.yaml` (and no `PM_VOICE_FILE`), the wizard starts with empty text inputs and selects on their first option, even when `~/.acta/voice.yaml` or `~/.pm/voice.yaml` exist; with a config file it starts from that file's values. An empty chat language or repo language is refused with `▲ required` and the question stays open; an empty tone is fine. `config.ResolveUser` still falls back for other callers. List every file state tested.

- [x] **Test:** temp HOME with only legacy files, with a config file, with nothing; the required check on both language fields; a test that `ResolveUser` still reads the legacy file.
- [x] **Code:** read defaults with `config.LoadUser(config.UserPath())` (or the existing equivalent), leave inputs empty with a dim placeholder when there is no file, and add a required validator to the two language inputs.
- [x] **Commit:** `fix(setup): start empty on a first run`.

### Task 03: version bump

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.

**verify:** All three files hold 0.1.35 and `internal/plugincheck` accepts it.

- [x] **Test:** the existing version check in `internal/plugincheck`.
- [x] **Code:** set 0.1.35 in all three files.
- [x] **Commit:** `chore(plugin): bump version`.

### Task 04: plugin files inside the binary

**Files:** `plugin/embed.go` (new), `plugin/embed_test.go` (new), `internal/setup/extract.go` (new), `internal/setup/extract_test.go` (new), `internal/cli/setup_cmd.go`, `internal/cli/setup_cmd_test.go`.

**verify:** The embedded tree holds every file Claude Code and omp need to load the plugin (`.claude-plugin/`, `skills/`, `hooks/`, `output-styles/`, `references/`, `omp/`, `package.json`, `README.md`, `NOTICE`) and none from `evals/`, `evals-routing/` or any `*_test.go`; a test walks the embed and compares it with the files on disk under `plugin/`, so a new plugin file that is not embedded fails the test. Extraction writes the tree to `<HOME>/.acta/plugin/` byte for byte, keeps hook scripts executable, replaces an older extraction fully (no stale file left), and never writes outside that folder. With no `--plugin-dir`, every install command uses the extracted folder; with `--plugin-dir`, the given folder wins and nothing is extracted. If extraction fails, the install step shows the run-by-hand lines and the rest of setup still runs. List every case tested.

- [x] **Test:** embed contents against disk; extraction into a temp HOME (fresh, over an old tree with a stale file, hook mode bits); install argv with and without `--plugin-dir`; extraction failure path. They fail because none of this exists.
- [x] **Code:** add `//go:embed` with explicit `all:` paths in `plugin/embed.go`, an extractor in `internal/setup/extract.go` that writes to a temp folder in `~/.acta` and renames it into place, and wire `cmdSetup` to use it as the default plugin dir.
- [x] **Commit:** `feat(setup): ship the plugin inside the binary`.

## Fix round 1

### Task 05: refuse a relative home before extracting

**Files:** `internal/setup/extract.go`, `internal/setup/extract_test.go`.

**verify:** No run of the extractor ever creates, renames or deletes anything unless home is an absolute path: an empty home, `.`, `relative/dir` and `./x` are all refused with an error before any file call, and a test proves nothing appears under the current directory. When the swap fails and putting the old tree back also fails, the error says where the old tree now sits instead of being dropped. List every home value and failure path tested.

- [ ] **Test:** extractor calls with each relative home inside a temp working dir, checking the error and that no `.acta` appears; a failed swap plus failed restore that checks the returned error names the moved-aside path. They fail on today's code.
- [ ] **Code:** refuse `!filepath.IsAbs(home)` next to the empty check, and return the restore error with the aside path.
- [ ] **Commit:** `fix(setup): never unpack the plugin under a relative home`.
