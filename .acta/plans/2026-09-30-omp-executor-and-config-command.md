---
created: "2026-09-30"
parent: bugs/2026-09-30-omp-build-runs-dispatch-executor
id: PLN-0051
hash: nln9dod
started: "2026-09-30"
finished: "2026-09-30"
---
# omp Executor Rule and `acta config` Command Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** An omp session never runs the dispatch executor, and the `acta voice` command is renamed to `acta config` with no alias.

**Architecture:** The omp rule is skill text only: `acta:build` maps `dispatch` to `subagent` on omp, and `acta:dispatch` refuses inside omp. The rename changes the CLI case, the command file, usage and error strings, the hook reminder prefix, and every plugin text. Package `internal/voice` keeps its name. A plugincheck guard stops `acta voice` from coming back.

**Tech Stack:** Go, the repo's own `internal/plugincheck` skill tests.

**Spec:** `.acta/specs/2026-09-30-omp-executor-and-config-command-design.md`

**Tests:** fast `scripts/test <touched packages>`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- `acta voice` is removed with no alias. `acta voice ...` exits 1 through the existing `unknown command` branch in `internal/cli/cli.go`.
- Package `internal/voice`, type `voice.Voice`, `PM_VOICE_FILE`, `voice.Resolve`, `ErrBad` and the "voice" word in hook struct fields stay as they are. Only the command name and the text that names the command change.
- Planning history under `.acta/` is not rewritten.
- Comments are plain English a 10-year-old reads back without stopping. They say why, not what.
- Run `gofmt -l .` and `go vet` on the touched packages before each commit, inside the task commit.

## File Map

| File | Task | Change |
|---|---|---|
| `plugin/skills/build/SKILL.md` | 1 | omp rule in the executor table and the `acta voice show` paragraph |
| `plugin/skills/dispatch/SKILL.md` | 1 | Step -3 guard: refuse inside omp |
| `internal/plugincheck/skill_build_test.go` | 1 | Must strings for the omp rule |
| `internal/plugincheck/skill_dispatch_test.go` | 1 | Must strings for the guard |
| `internal/cli/voice.go` | 2 | `git mv` to `internal/cli/config_cmd.go`; `cmdConfig`, `configUsage`, `acta config` strings |
| `internal/cli/cli.go` | 2 | `case "config"` in place of `case "voice"` |
| `internal/cli/cli_test.go` | 2 | `"voice"` args become `"config"`; comment at line 62 |
| `cmd/acta/voice_test.go` | 2 | `git mv` to `cmd/acta/config_test.go`; args become `"config"`; `acta voice` exits 1 |
| `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go` | 2 | fix text `acta config set --theme` |
| `cmd/acta/hook_test.go` | 2, then 3 | Task 2: the `voice set` call becomes `config set`; Task 3: the expected prompt line |
| `internal/hook/hook.go`, `internal/hook/hook_test.go` | 3 | skill line and prompt prefix `acta config:` |
| `internal/cli/hook_session_test.go`, `cmd/acta/hook_test.go` | 3 | expect `acta config:` |
| `plugin/hooks/default-rules.md` | 3 | build line says `acta config show` |
| `internal/plugincheck/no_old_names_test.go` | 4 | new guard regex for `acta voice` |
| `internal/plugincheck/models_test.go`, `plugin_test.go`, `skill_build_test.go`, `skill_plan_test.go`, `skill_setup_test.go` | 4 | Must strings say `acta config` |
| `plugin/skills/{brainstorm,build,debug,dispatch,plan,review,setup}/SKILL.md`, `plugin/skills/build/implementer-prompt.md`, `plugin/README.md`, `plugin/evals/FACTS.md` | 4 | every `acta voice` becomes `acta config` |

## Waves

- Wave 1: Task 1, Task 2 (no shared file).
- Wave 2: Task 3 (needs Task 2: both touch `cmd/acta/hook_test.go`, and the test runs `acta config set`).
- Wave 3: Task 4 (needs Task 1: same skill and test files; needs Task 3: the guard also scans the hook text and `default-rules.md`).

---

### Task 1: omp never runs the dispatch executor

**Files:**
- Modify: `plugin/skills/build/SKILL.md` (the `dispatch` row of the `## Executors` table, and the paragraph that starts "Before you pick one, run `acta voice show`.")
- Modify: `plugin/skills/dispatch/SKILL.md` (new section right above `## Step -2 — Detect herdr FIRST`)
- Test: `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/skill_dispatch_test.go`

**verify:** No path in the build or dispatch skill lets an omp session open a dispatch tab. List every path checked: `build_executor: dispatch` read in omp, no `build_executor` line in omp with the user answering "dispatch", the user asking omp to dispatch directly, and an omp dispatch recipient running `acta:build`. Claude Code with `build_executor: dispatch` still dispatches.

**Interfaces:**
- Consumes: nothing.
- Produces: the exact sentences below; Task 4 renames `acta voice show` inside them and keeps the rest.

- [x] **Step 1: Write the failing test**

In `internal/plugincheck/skill_build_test.go`, `TestSkillBuild`, add to `Must`:

```go
			"On omp, `dispatch` runs as `subagent`",
			"Dispatch is only for harnesses other than omp.",
```

In `internal/plugincheck/skill_dispatch_test.go`, `TestSkillDispatch`, add to `Must`:

```go
			"## Step -3 — Refuse inside omp",
			"Running in omp: STOP.",
			"Dispatch is only for harnesses other than omp.",
```

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/plugincheck/ -run 'TestSkillBuild$|TestSkillDispatch$'`
Expected: FAIL, both tests report the missing strings.

- [x] **Step 3: Write minimal implementation**

In `plugin/skills/build/SKILL.md`, the `dispatch` row becomes:

```markdown
| `dispatch` | an omp agent in its own herdr tab | follow `acta:dispatch`; it refuses without herdr. On omp, `dispatch` runs as `subagent` |
```

Append to the paragraph that starts "Before you pick one, run `acta voice show`.":

```markdown
On omp, `dispatch` runs as `subagent`: use `agent()` with `agent="task"` and do not ask. Dispatch is only for harnesses other than omp.
```

In `plugin/skills/dispatch/SKILL.md`, right above `## Step -2 — Detect herdr FIRST`, add:

```markdown
## Step -3 — Refuse inside omp

Running in omp: STOP. Do not open a tab. Go back to `acta:build` and run the `subagent` executor (`agent()` with `agent="task"`). Dispatch is only for harnesses other than omp.
```

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/plugincheck/`
Expected: PASS (size caps and every other skill rule still hold).

- [x] **Step 5: Commit**

```bash
git add plugin/skills/build/SKILL.md plugin/skills/dispatch/SKILL.md internal/plugincheck/skill_build_test.go internal/plugincheck/skill_dispatch_test.go
git commit -m "fix: omp runs build subagents instead of dispatch"
```

### Task 2: rename the `acta voice` command to `acta config`

**Files:**
- Rename: `internal/cli/voice.go` to `internal/cli/config_cmd.go` (`git mv`)
- Rename: `cmd/acta/voice_test.go` to `cmd/acta/config_test.go` (`git mv`)
- Modify: `internal/cli/cli.go` (the `case "voice":` line in `Run`)
- Modify: `internal/cli/cli_test.go` (every `mustRun`/`runCode` call whose first arg is `"voice"`, and the comment "The voice command needs no repo")
- Modify: `internal/doctor/doctor.go` (`r.Fix = "acta voice set --theme " + theme.Default`), `internal/doctor/doctor_test.go` (`fix = "acta voice set --theme"`)
- Modify: `cmd/acta/hook_test.go` (only the `acta(t, dir, "", "voice", "set", "--language", "Korean")` call and its `t.Fatal("voice set failed")`; the expected prompt line is Task 3)

**verify:** Every former `acta voice` path works as `acta config` with the same output, exit codes and file writes, and no `acta voice` form still runs. List every path checked: `config show`, `config show --json`, `config set` with each flag, each bad-input case of the old `TestVoiceBadInput`, the unreadable-file case, and `voice`, `voice show`, `voice set --language X` all exiting 1 with `unknown command`.

**Interfaces:**
- Consumes: nothing.
- Produces: `acta config show [--json]` and `acta config set [flags]` with the same flags as before. Function `cmdConfig(args []string, stdout, stderr io.Writer) int`, const `configUsage`.

- [x] **Step 1: Write the failing test**

`git mv cmd/acta/voice_test.go cmd/acta/config_test.go` and `git mv internal/cli/voice.go internal/cli/config_cmd.go`. In `cmd/acta/config_test.go`, change every `"voice"` command arg to `"config"`, rename `TestVoice*` functions to `TestConfig*`, and add this case to the bad-input test:

```go
func TestVoiceCommandIsGone(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PM_VOICE_FILE", filepath.Join(dir, "voice.yaml"))
	for _, args := range [][]string{{"voice"}, {"voice", "show"}, {"voice", "set", "--language", "Korean"}} {
		_, errOut, code := acta(t, dir, "", args...)
		if code != 1 || !strings.Contains(errOut, "unknown command") {
			t.Errorf("%v: exit %d, stderr %q; want 1 and unknown command", args, code, errOut)
		}
	}
}
```

In `internal/cli/cli_test.go`, change every `"voice"` first arg of `mustRun`/`runCode` to `"config"` and the comment to "The config command needs no repo, so no inDir here." In `internal/doctor/doctor_test.go`, the expected fix becomes `fix = "acta config set --theme"`. In `cmd/acta/hook_test.go`, the call becomes `acta(t, dir, "", "config", "set", "--language", "Korean")` and the fatal text `"config set failed"`.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./cmd/acta/ ./internal/cli/ ./internal/doctor/`
Expected: FAIL, `config` is an unknown command and the doctor fix text still says `acta voice set`.

- [x] **Step 3: Write minimal implementation**

In `internal/cli/cli.go`:

```go
	case "config":
		return cmdConfig(args[1:], stdout, stderr)
```

In `internal/cli/config_cmd.go`: rename `cmdVoice` to `cmdConfig`, `voiceUsage` to `configUsage`, the flag set names to `"config show"` and `"config set"`, and the usage string to:

```go
const configUsage = "usage: acta config show [--json] | acta config set [--language L] [--style adhd|plain] [--tone T] [--clear-tone] [--repo-language L] [--executor subagent|dispatch|inline] [--subagent-models split] [--clear-subagent-models] [--theme NAME] [--clear-theme]"
```

Any other `acta voice` text in that file becomes `acta config`. In `internal/doctor/doctor.go`: `r.Fix = "acta config set --theme " + theme.Default`.

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./cmd/acta/ ./internal/cli/ ./internal/doctor/`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./cmd/acta/ ./internal/cli/ ./internal/doctor/
git add -A cmd/acta internal/cli internal/doctor
git commit -m "feat: rename acta voice command to acta config"
```

### Task 3: hook text says `acta config`

**Files:**
- Modify: `internal/hook/hook.go` (the `build` entry of the skill list, and the three `acta voice:` strings in `Prompt`)
- Modify: `plugin/hooks/default-rules.md` (the `acta:build` line)
- Test: `internal/hook/hook_test.go`, `internal/cli/hook_session_test.go`, `cmd/acta/hook_test.go`

**verify:** No text a session is shown names the `acta voice` command. List every path checked: `SessionStart` first run, set, broken and conflicts; `Prompt` set, missing and broken; the `default-rules.md` fallback; the `prompt-reminder` hook through the binary.

**Interfaces:**
- Consumes: Task 2's `acta config set`.
- Produces: prompt lines `acta config: reply in <L>, <style> style.`, `acta config: not set up yet; reply in English, adhd style, and run /acta:setup once (see the session rules).`, `acta config: the config file could not be read; reply in English, adhd style.`; build skill line `run an approved plan in a worktree; executor from `acta config show`, else ask: subagent, dispatch or inline`.

- [x] **Step 1: Write the failing test**

In `internal/hook/hook_test.go`, the expected strings become:

```go
		"- acta:build: run an approved plan in a worktree; executor from `acta config show`, else ask: subagent, dispatch or inline",
```

```go
		"set":     {korean(), "acta config: reply in Korean, adhd style."},
		"missing": {Input{Voice: voice.Default()}, "acta config: not set up yet; reply in English, adhd style, and run /acta:setup once (see the session rules)."},
		"broken":  {Input{Voice: voice.Default(), VoiceExists: true, VoiceErr: errors.New("x")}, "acta config: the config file could not be read; reply in English, adhd style."},
```

In `internal/cli/hook_session_test.go`, `strings.Contains(voiceOut, "acta voice:")` becomes `strings.Contains(voiceOut, "acta config:")`. In `cmd/acta/hook_test.go`, the expected line becomes `"acta config: reply in Korean, adhd style."`.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/hook/ ./internal/cli/ ./cmd/acta/`
Expected: FAIL on the `acta voice` strings.

- [x] **Step 3: Write minimal implementation**

In `internal/hook/hook.go`, make the four strings match Step 1 exactly. In `plugin/hooks/default-rules.md`, the build line becomes:

```markdown
- acta:build: run an approved plan in a worktree; executor from `acta config show`, else ask: subagent, dispatch or inline
```

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/hook/ ./internal/cli/ ./cmd/acta/ ./internal/plugincheck/`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./internal/hook/ ./internal/cli/ ./cmd/acta/
git add internal/hook plugin/hooks/default-rules.md internal/cli/hook_session_test.go cmd/acta/hook_test.go
git commit -m "feat: hook text names acta config"
```

### Task 4: plugin text says `acta config`, with a guard

**Files:**
- Modify: `internal/plugincheck/no_old_names_test.go` (new regex in `oldNameProblems` checks)
- Modify: `internal/plugincheck/models_test.go` (`modelsPara`), `plugin_test.go` (the `"acta voice set"` Must), `skill_build_test.go` (the two `acta voice show` Musts), `skill_plan_test.go` (`"run `acta voice show`"`), `skill_setup_test.go` (`"acta voice set"`, `"acta voice show"`)
- Modify: every `acta voice` in `plugin/skills/brainstorm/SKILL.md`, `plugin/skills/build/SKILL.md`, `plugin/skills/build/implementer-prompt.md`, `plugin/skills/debug/SKILL.md`, `plugin/skills/dispatch/SKILL.md`, `plugin/skills/plan/SKILL.md`, `plugin/skills/review/SKILL.md`, `plugin/skills/setup/SKILL.md`, `plugin/README.md`, `plugin/evals/FACTS.md`

**verify:** The `acta voice` command name cannot come back anywhere a user or agent reads it: any file under `plugin/`, or any text the hooks show. List every file and hook text the guard scans, and show that the guard fails when one `acta voice` is put back.

**Interfaces:**
- Consumes: Task 1's sentences (their `acta voice show` becomes `acta config show`), Task 3's hook text.
- Produces: `oldVoiceRe` in `no_old_names_test.go`.

- [x] **Step 1: Write the failing test**

In `internal/plugincheck/no_old_names_test.go`, next to `oldRootRe`:

```go
// oldVoiceRe matches the old config command. The Go package and the
// PM_VOICE_FILE variable keep the voice name, so only "acta voice" fails.
var oldVoiceRe = regexp.MustCompile(`\bacta voice\b`)
```

and add to the `checks` list in `oldNameProblems`:

```go
		{"old acta voice command", oldVoiceRe},
```

In the five plugincheck test files named above, change every `acta voice` to `acta config` (including `modelsPara`).

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/plugincheck/`
Expected: FAIL, the guard lists each plugin file that still says `acta voice`, and the skill Musts miss `acta config`.

- [x] **Step 3: Write minimal implementation**

In each plugin file named above, replace `acta voice` with `acta config`, word for word, nothing else. In `plugin/evals/FACTS.md` line 119, `acta voice setup` becomes `acta config setup`; line 191 becomes `acta config: not set up yet; reply in English, adhd style, and run /acta:setup once (see the session rules).`

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/plugincheck/`
Expected: PASS. Then put one `acta voice` back in `plugin/README.md`, run the same command, see it FAIL with `old acta voice command`, and undo that edit by hand.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./internal/plugincheck/
git add internal/plugincheck plugin
git commit -m "feat: plugin text names acta config, guard the old name"
```
