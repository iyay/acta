---
parent: scratch/2026-10-07-standalone-setup-installer
depth: minimal
closes: [SPC-0100]
id: PLN-0109
created: "2026-10-07 09:34:22"
hash: debv8d3
---
# acta setup wizard Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta setup` runs an interactive wizard in a plain terminal that writes the user config, installs the plugin into the harnesses it finds, and offers the acta block in the current repo.

**Spec:** .acta/specs/2026-10-07-acta-setup-wizard-design.md

**Tests:** `scripts/test ./internal/setup/ ./internal/cli/ ./internal/plugincheck/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- All decisions live in the pure function `setup.Plan`; `form.go` and the runner hold no rules. Config values go through `config.User.Validate` before any write.
- Tests never touch the real HOME: set `HOME`, `PM_VOICE_FILE` and `TMPDIR` to `t.TempDir()` paths. Tests never run the real `claude` or `omp`; harness commands go through a runner interface with a fake in tests.
- Plugin install commands, copied from `plugin/README.md`: Claude Code runs `claude plugin marketplace add <plugin dir>` then `claude plugin install acta@acta-local`; omp runs `omp plugin link <plugin dir>`. The plugin dir comes from `acta setup --plugin-dir <path>`. With no `--plugin-dir`, the wizard runs no install and prints these commands with `<plugin dir>` left for the user to fill.
- The wizard asks what the skill asks today: chat language, style (adhd or plain), tone, build executor, subagent models, questions, plan depth. Offer `dispatch` as an executor only when `HERDR_ENV=1`. Never ask about `coding_guide`.
- New dependency: `github.com/charmbracelet/huh` only.
- Run tests with `scripts/test`, never bare `go test`. Never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01, Task 03, Task 04
- Wave 2: Task 02

Task 02 builds on the `internal/setup` API from Task 01. Tasks 03 and 04 touch other files.

### Task 01: pure plan and acta block writer

**Files:** `internal/setup/plan.go`, `internal/setup/plan_test.go`, `internal/setup/block.go`, `internal/setup/block_test.go` (all new).

**verify:** For every combination of env (TTY or not, `claude` / `omp` found or not, plugin dir given or not, in a git repo or not, CLAUDE.md / AGENTS.md: both, one, none) `Plan` returns the action list the spec's Flow section names, and with no TTY it returns no write or install action. `WriteBlock` changes only the bytes between `<!-- acta:begin -->` and `<!-- acta:end -->` on a re-run, appends the block when the markers are missing, and creates a file holding only the block when the file is missing. The block text equals, byte for byte, the block in `git show main:plugin/skills/setup/SKILL.md` as of this plan (Task 03 removes it from the skill in the same wave). List every env combination and file state tested.

- [ ] **Test:** table-driven `TestPlan` over the env combinations above and `TestWriteBlock` over new file, file with no markers, file with markers and text around them; they fail because the package does not exist.
- [ ] **Code:** `type Env struct{ TTY bool; Harnesses []string; PluginDir, RepoRoot string; HasClaudeMD, HasAgentsMD bool; Current config.User }`, `type Answers struct{ User config.User; Install map[string]bool; BlockFile string }`, `type Action struct{ Kind string /* "config", "install", "print", "block" */; Harness, Path string; Argv [][]string; User config.User }`, `func Plan(a Answers, e Env) []Action`, plus `const Block` and `func WriteBlock(path string) error`.
- [ ] **Run:** `scripts/test ./internal/setup/` passes.

### Task 02: huh form, runner and the `acta setup` command

**Files:** `internal/setup/form.go`, `internal/setup/run.go`, `internal/setup/run_test.go` (new), `internal/cli/cli.go` (add `case "setup"` and the name to the unknown-command list), `internal/cli/setup_cmd.go`, `internal/cli/setup_cmd_test.go` (new), `go.mod`, `go.sum`.

**verify:** No path of `acta setup` writes a file or runs a harness command when stdin or stdout is not a TTY; it exits non-zero with a message naming `acta config set`. Every action from `Plan` is carried out by the runner: config through `config.SaveUser` after `Validate`, block through `WriteBlock`, installs through the runner interface, and a failed install prints its command and carries on. Form defaults come from the current config. List every action kind and the no-TTY paths tested.

- [ ] **Test:** `TestRunActions` with a fake runner (success, one failing install) and a temp HOME; `TestSetupNoTTY` calls `cli.Run([]string{"setup"}, ...)` with `stdinIsTTY=false` and checks the exit code, the message and that no config file appeared. Both fail first.
- [ ] **Code:** `type Runner interface{ Run(argv []string) error }`, `func Apply(actions []Action, r Runner, out io.Writer) error`, `func Ask(e Env) (Answers, error)` built on `huh` with one field per group, `cmdSetup(args, stdinIsTTY, stdout, stderr)` that builds `Env` (`exec.LookPath`, `HERDR_ENV`, git root, `--plugin-dir`), runs `doctor.Run` and prints its results, then `Ask`, `Plan`, `Apply` and a summary.
- [ ] **Run:** `scripts/test ./internal/setup/ ./internal/cli/` passes.

### Task 03: thin acta:setup skill

**Files:** `plugin/skills/setup/SKILL.md`, `internal/plugincheck/skill_setup_test.go`.

**verify:** The skill tells the agent to ask the user to run `! acta setup` in their terminal, and to fall back to `acta config set` flags when the user cannot. It no longer holds the block text or the per-question rules that now live in Go. The plugincheck rule for the skill matches the new text and still bans `superpowers:`. List every phrase dropped from and added to the `Must` list.

- [ ] **Test:** update `TestSkillSetup` `Must` to the new phrases (`acta setup`, `--plugin-dir`, `acta config set`, `acta doctor`) and lower `MaxLines`; it fails against the old skill.
- [ ] **Code:** rewrite `plugin/skills/setup/SKILL.md` to the thin version.
- [ ] **Run:** `scripts/test ./internal/plugincheck/` passes.

### Task 04: version bump

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.

**verify:** All three files hold the same version, one patch above 0.1.30, and `internal/plugincheck` accepts it.

- [ ] **Test:** the existing version check in `internal/plugincheck` is the test.
- [ ] **Code:** set the version to `0.1.31` in all three files.
- [ ] **Run:** `scripts/test ./internal/plugincheck/` passes.
