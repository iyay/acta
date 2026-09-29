---
created: "2026-09-30"
id: PLN-0041
hash: ndygdhp
started: "2026-09-30"
finished: "2026-09-30"
---
# Config File Name Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The user setting lives in `~/.acta/config.yaml` (an old `~/.acta/voice.yaml` moves there on the first read), `acta:setup` makes a `CLAUDE.md` when a repo has none, and `acta:plan` takes the build executor from the config instead of asking.

**Architecture:** `voice.Path()` points at `config.yaml`. `voice.Resolve()` renames `~/.acta/voice.yaml` to `config.yaml` once when `config.yaml` is missing and `PM_VOICE_FILE` is not set, and reads the old file in place when the rename fails. The setup and plan skills change their text, and the session index line for `build` in `internal/hook/hook.go` stops calling `subagent` the default.

**Tech Stack:** Go (`os.Rename`), skill markdown checked by `internal/plugincheck`.

**Spec:** `.acta/specs/2026-09-30-config-file-name-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- `PM_VOICE_FILE` keeps its name and still wins when set. No `ACTA_` env name, no `acta config` alias.
- `acta voice show` and `acta voice set` keep their names.
- `~/.pm/voice.yaml` stays a read-only fallback. It is never moved or written.
- No setting may be lost: when the move fails, the old file is read where it is.
- Old specs, plans, debt and scratch items under `.acta/` are not edited.
- No new dependency in go.mod.
- Comments use plain English that a 10-year-old can read back: short words, and they say why.
- Before each commit, run `gofmt -l cmd internal` (it must print nothing), `go vet ./...` and `go test ./...`.

## File Map

- `internal/voice/voice.go`: `Path` returns `config.yaml`; new `voicePath`; `Resolve` does the one-time move; comments name the new file.
- `internal/voice/voice_acta_test.go`: move tests; existing home-path tests switch to `config.yaml`.
- `internal/voice/voice_test.go`: `TestPath` default path switches to `config.yaml`.
- `plugin/README.md`, `plugin/evals/*/scaffold.sh` (4 files): name `~/.acta/config.yaml`.
- `plugin/skills/setup/SKILL.md`, `internal/plugincheck/skill_setup_test.go`: config path and the new acta block rules.
- `plugin/skills/plan/SKILL.md`, `internal/plugincheck/skill_plan_test.go`: plan reads the executor.
- `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`: new `build` index line.

## Waves

- Wave 1: Task 1, Task 2, Task 3. No two tasks share a file.

---

### Task 1: Config lives in config.yaml, old voice.yaml moves on first read

**Files:**
- Modify: `internal/voice/voice.go` (package comment, `Voice` comment, `Path`, new `voicePath`, `Resolve`)
- Modify: `plugin/README.md` (the two lines that name `~/.acta/voice.yaml`)
- Modify: `plugin/evals/answers-appended/scaffold.sh`, `plugin/evals/brainstorm-files-scratch-first/scaffold.sh`, `plugin/evals/note-to-scratch/scaffold.sh`, `plugin/evals/side-idea-to-scratch/scaffold.sh` (`$HOME/.acta/voice.yaml` becomes `$HOME/.acta/config.yaml`)
- Test: `internal/voice/voice_acta_test.go`, `internal/voice/voice_test.go`

**verify:** No read path loses a setting or leaves two live files. List every case `Resolve` can meet (only `config.yaml`; only `~/.acta/voice.yaml`; both; neither; only `~/.pm/voice.yaml`; `PM_VOICE_FILE` set with `~/.acta/voice.yaml` present; rename fails; two readers racing) and what each returns and leaves on disk. `~/.pm/voice.yaml` is never moved or changed on any path. No live file outside `.acta/` still names `~/.acta/voice.yaml` as the home path (`grep -rn '\.acta/voice\.yaml' --exclude-dir=.git --exclude-dir=.acta .` prints only test lines that set up the old file on purpose).

**Interfaces:**
- Consumes: `Load(path string) (Voice, bool, error)`, `Default() Voice`, `oldPath()`, the test helpers `withHome(t)` and `writeVoice(t, path, body)` in `voice_acta_test.go`.
- Produces: `Path() (string, error)` returns `PM_VOICE_FILE` or `~/.acta/config.yaml`. `Resolve() (Voice, bool, error)` keeps its signature. New unexported `voicePath() (string, error)` returns `~/.acta/voice.yaml`.

- [x] **Step 1: Write the failing tests**

In `internal/voice/voice_test.go`, inside `TestPath`, change the default-path line to:

```go
	if p, _ := Path(); p != filepath.Join(home, ".acta", "config.yaml") {
```

In `internal/voice/voice_acta_test.go`, change `TestPathIsActaDefault` to want `filepath.Join(home, ".acta", "config.yaml")`. In `TestResolvePrefersNewFile` write the new file to `filepath.Join(home, ".acta", "config.yaml")`. In `TestSaveWritesNewFileOnly` load `filepath.Join(home, ".acta", "config.yaml")`. At the end of `TestResolveFallsBackToOldFile` add:

```go
	if _, err := os.Stat(filepath.Join(home, ".pm", "voice.yaml")); err != nil {
		t.Fatalf("the ~/.pm file must stay where it is: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "config.yaml")); !os.IsNotExist(err) {
		t.Fatalf("reading the ~/.pm file must not make config.yaml: %v", err)
	}
```

Then add these tests to `internal/voice/voice_acta_test.go`:

```go
func TestResolveMovesVoiceFile(t *testing.T) {
	home := withHome(t)
	oldFile := filepath.Join(home, ".acta", "voice.yaml")
	body := "chat_language: Korean\nstyle: plain\nbuild_executor: dispatch\n"
	writeVoice(t, oldFile, body)

	v, exists, err := Resolve()
	if err != nil || !exists {
		t.Fatalf("got %+v %v %v", v, exists, err)
	}
	if v.ChatLanguage != "Korean" || v.BuildExecutor != "dispatch" {
		t.Fatalf("got %+v, want the voice.yaml values", v)
	}
	got, err := os.ReadFile(filepath.Join(home, ".acta", "config.yaml"))
	if err != nil || string(got) != body {
		t.Fatalf("config.yaml = %q, %v; want the old bytes %q", got, err, body)
	}
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatalf("voice.yaml must be gone after the move: %v", err)
	}
}

func TestResolveConfigWinsOverVoiceFile(t *testing.T) {
	home := withHome(t)
	oldFile := filepath.Join(home, ".acta", "voice.yaml")
	writeVoice(t, oldFile, "chat_language: Korean\nstyle: plain\n")
	writeVoice(t, filepath.Join(home, ".acta", "config.yaml"), "chat_language: German\nstyle: adhd\n")

	v, _, err := Resolve()
	if err != nil || v.ChatLanguage != "German" {
		t.Fatalf("got %+v %v, want the config.yaml values", v, err)
	}
	got, err := os.ReadFile(oldFile)
	if err != nil || string(got) != "chat_language: Korean\nstyle: plain\n" {
		t.Fatalf("voice.yaml must stay as it was: %q %v", got, err)
	}
}

func TestResolveDoesNotMoveWhenEnvIsSet(t *testing.T) {
	home := withHome(t)
	oldFile := filepath.Join(home, ".acta", "voice.yaml")
	writeVoice(t, oldFile, "chat_language: Korean\nstyle: plain\n")
	t.Setenv("PM_VOICE_FILE", filepath.Join(t.TempDir(), "mine.yaml"))

	v, exists, err := Resolve()
	if err != nil || exists || v != Default() {
		t.Fatalf("got %+v %v %v, want defaults from the missing env file", v, exists, err)
	}
	if _, err := os.Stat(oldFile); err != nil {
		t.Fatalf("voice.yaml must stay when PM_VOICE_FILE is set: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "config.yaml")); !os.IsNotExist(err) {
		t.Fatalf("config.yaml must not appear when PM_VOICE_FILE is set: %v", err)
	}
}

func TestResolveWithNoFilesGivesDefaults(t *testing.T) {
	home := withHome(t)
	v, exists, err := Resolve()
	if err != nil || exists || v != Default() {
		t.Fatalf("got %+v %v %v, want defaults", v, exists, err)
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "config.yaml")); !os.IsNotExist(err) {
		t.Fatalf("a read must not make config.yaml: %v", err)
	}
}

func TestResolveReadsVoiceFileWhenMoveFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can rename in a read-only folder")
	}
	home := withHome(t)
	dir := filepath.Join(home, ".acta")
	oldFile := filepath.Join(dir, "voice.yaml")
	writeVoice(t, oldFile, "chat_language: Korean\nstyle: plain\n")
	// A folder we cannot write to makes the rename fail, but the file can
	// still be read.
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })

	v, exists, err := Resolve()
	if err != nil || !exists || v.ChatLanguage != "Korean" {
		t.Fatalf("got %+v %v %v, want the voice.yaml values", v, exists, err)
	}
	if _, err := os.Stat(oldFile); err != nil {
		t.Fatalf("voice.yaml must stay after a failed move: %v", err)
	}
}
```

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/voice/`
Expected: FAIL. `TestPath`, `TestPathIsActaDefault`, `TestResolveMovesVoiceFile`, `TestResolvePrefersNewFile` and `TestSaveWritesNewFileOnly` fail, because `Path` still returns `voice.yaml`.

- [x] **Step 3: Write the code**

In `internal/voice/voice.go`, change the `Voice` comment to say the setting lives in `~/.acta/config.yaml`, that an old `~/.acta/voice.yaml` is moved there on the first read, and that `~/.pm/voice.yaml` is only read when both are missing. Replace `Path` and add `voicePath` below it:

```go
// Path is PM_VOICE_FILE when set, else ~/.acta/config.yaml. Writes always go
// here, so one place holds the truth.
func Path() (string, error) {
	if p := os.Getenv("PM_VOICE_FILE"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".acta", "config.yaml"), nil
}

// voicePath is the name the file had before it held more than the voice.
// Resolve moves it to Path once.
func voicePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".acta", "voice.yaml"), nil
}
```

Replace `Resolve`:

```go
// Resolve reads config.yaml. When it is missing, an old ~/.acta/voice.yaml
// is renamed to config.yaml first, so the user ends up with one file. When
// both are missing, the ~/.pm file from before the acta rename is read.
func Resolve() (Voice, bool, error) {
	path, err := Path()
	if err != nil {
		return Default(), false, err
	}
	v, exists, err := Load(path)
	if exists || err != nil || os.Getenv("PM_VOICE_FILE") != "" {
		return v, exists, err
	}
	voiceFile, err := voicePath()
	if err != nil {
		return Default(), false, err
	}
	// A missing voice.yaml is fine: there was none, or another hook moved it
	// a moment ago. Any other failure means the file is still there, so read
	// it where it is and lose nothing. The next read tries the move again.
	if err := os.Rename(voiceFile, path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Load(voiceFile)
	}
	if v, exists, err := Load(path); exists || err != nil {
		return v, exists, err
	}
	old, err := oldPath()
	if err != nil {
		return Default(), false, err
	}
	return Load(old)
}
```

In `plugin/README.md`, change both `~/.acta/voice.yaml` to `~/.acta/config.yaml`. In the four `plugin/evals/*/scaffold.sh` files, change `$HOME/.acta/voice.yaml` to `$HOME/.acta/config.yaml` (two places per file).

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/voice/ && go test ./...`
Expected: PASS, no failures in any package.

Run: `grep -rn '\.acta/voice\.yaml' --exclude-dir=.git --exclude-dir=.acta .`
Expected: only lines in `internal/voice/voice_acta_test.go` that set up the old file on purpose.

- [x] **Step 5: Commit**

```bash
gofmt -l cmd internal && go vet ./... && go test ./...
git add internal/voice/voice.go internal/voice/voice_acta_test.go internal/voice/voice_test.go plugin/README.md plugin/evals/answers-appended/scaffold.sh plugin/evals/brainstorm-files-scratch-first/scaffold.sh plugin/evals/note-to-scratch/scaffold.sh plugin/evals/side-idea-to-scratch/scaffold.sh
git commit -m "Move the user config to ~/.acta/config.yaml; an old voice.yaml moves on first read"
```

---

### Task 2: Setup names config.yaml and makes CLAUDE.md when the repo has none

**Files:**
- Modify: `plugin/skills/setup/SKILL.md` (line 8 path, "The acta block" section, the last "Limits" line)
- Test: `internal/plugincheck/skill_setup_test.go`

**verify:** No wording in the setup skill can lead an agent to write a CLAUDE.md or AGENTS.md before the user said yes, to create a CLAUDE.md when an AGENTS.md already exists, or to touch text outside the acta markers in a file that was there before. List each file state (both exist; only CLAUDE.md; only AGENTS.md; neither, in Claude Code; neither, in another harness) and what the skill tells the agent to do. The old "never create" rule cannot come back in any form.

**Interfaces:**
- Consumes: `CheckSkill(t, SkillRule{Name, MaxLines, Must, MustNot})` from `internal/plugincheck`.
- Produces: nothing other tasks use.

- [x] **Step 1: Write the failing test**

Replace `internal/plugincheck/skill_setup_test.go` with:

```go
package plugincheck

import "testing"

func TestSkillSetup(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "setup",
		MaxLines: 66,
		Must: []string{
			"acta voice set", "--language", "--style", "--tone", "--clear-tone", "--repo-language",
			"acta voice show", "adhd", "plain", "full English name",
			"acta doctor", "acta doctor --fix", "--executor", "HERDR_ENV=1", "herdr",
			"--subagent-models split", "Claude Code only",
			"<!-- acta:begin -->", "<!-- acta:end -->",
			"only after a yes", "never edits settings",
			"which part to change",
			"`~/.acta/config.yaml`",
			"When only AGENTS.md exists, write the block there",
			"run `/init` first",
			"create a CLAUDE.md that holds only the block",
		},
		MustNot: []string{"superpowers:", "It never edits CLAUDE.md",
			"~/.acta/voice.yaml", "only to files that already exist", "never create a CLAUDE.md",
		},
	})
}
```

- [x] **Step 2: Run the test to see it fail**

Run: `go test ./internal/plugincheck/ -run TestSkillSetup`
Expected: FAIL, naming the missing `config.yaml`, `/init` and AGENTS.md phrases and the banned `~/.acta/voice.yaml` and "never create" phrases.

- [x] **Step 3: Change the skill text**

In `plugin/skills/setup/SKILL.md`, line 8 becomes:

```markdown
The chat language, style and tone live in `~/.acta/config.yaml` (or the file `PM_VOICE_FILE` names). An old `~/.acta/voice.yaml` moves there on its own. `acta` reads it at the start of every session and before every message.
```

In "### The acta block", the first line becomes:

```markdown
Ask: add the acta block to CLAUDE.md / AGENTS.md? When the repo has neither, ask to create a CLAUDE.md for it. Show this exact block first, before any yes:
```

The paragraph after the block (starting "Write only after a yes") becomes one line:

```markdown
Write only after a yes. When both files exist, ask which one. When only CLAUDE.md exists, write the block there. When only AGENTS.md exists, write the block there and make no CLAUDE.md. When neither exists: in Claude Code, run `/init` first, then add the block to the CLAUDE.md it made; in any other harness, create a CLAUDE.md that holds only the block. Write only between the two markers; a re-run replaces the text inside them and leaves the rest of the file alone.
```

The last "Limits" line becomes:

```markdown
- This skill edits CLAUDE.md or AGENTS.md only between the acta markers, and only after a yes; the one exception is the new CLAUDE.md that `/init` writes. It never edits settings.
```

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/plugincheck/ && go test ./...`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l cmd internal && go vet ./... && go test ./...
git add plugin/skills/setup/SKILL.md internal/plugincheck/skill_setup_test.go
git commit -m "Setup names config.yaml and makes a CLAUDE.md when the repo has none"
```

---

### Task 3: Plan takes the build executor from the config

**Files:**
- Modify: `plugin/skills/plan/SKILL.md` ("## Hand-off", the "Then run it with `acta:build`" line)
- Modify: `internal/hook/hook.go:24` (the `build` entry in `Skills`)
- Modify: `plugin/hooks/default-rules.md` (regenerated, not hand-edited)
- Test: `internal/plugincheck/skill_plan_test.go`, `internal/hook/hook_test.go`

**verify:** No text the agent reads after a plan is approved calls `subagent` the default or tells it to ask for the executor when the config names one. List every surface checked: the plan skill hand-off, the hook session index in `hook.go`, and `plugin/hooks/default-rules.md`.

**Interfaces:**
- Consumes: `SessionStart(Input) string`, `korean()` test helper, `Skills` in `internal/hook/hook.go`; `CheckSkill` in `internal/plugincheck`.
- Produces: nothing other tasks use.

- [x] **Step 1: Write the failing tests**

In `internal/plugincheck/skill_plan_test.go`, add to `Must`:

```go
			"run `acta voice show`", "When it prints `build_executor: <name>`, that executor is chosen",
			"do not ask",
```

and add to `MustNot`:

```go
			"Ask which executor only if the user has not said",
```

In `internal/hook/hook_test.go`, inside `TestSessionStartListsSkillsAndRules`, add to the `want` list:

```go
		"- acta:build: run an approved plan in a worktree; executor from `acta voice show`, else ask: subagent, dispatch or inline",
```

and after the `want` loop add:

```go
	if strings.Contains(out, "executor subagent (default)") {
		t.Error("the build line must not call subagent the default; the config picks the executor")
	}
```

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/hook/ ./internal/plugincheck/ -run 'TestSessionStartListsSkillsAndRules|TestSkillPlan'`
Expected: FAIL, naming the missing build line, the old "(default)" text, and the missing plan skill phrases.

- [x] **Step 3: Change the text**

In `internal/hook/hook.go`, the `build` entry becomes:

```go
	{"build", "run an approved plan in a worktree; executor from `acta voice show`, else ask: subagent, dispatch or inline"},
```

In `plugin/skills/plan/SKILL.md`, the line starting "Then run it with `acta:build`." becomes:

```markdown
Then run it with `acta:build`. First run `acta voice show`. When it prints `build_executor: <name>`, that executor is chosen: name it and do not ask. When the line is missing and the user has not said, ask which one: `subagent`, `dispatch` (an omp agent in its own herdr tab), or `inline` (you write the code yourself).
```

Regenerate the default rules file:

Run: `go test ./internal/hook -run TestDefaultRulesFile -update`

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/hook/ ./internal/plugincheck/ && go test ./...`
Expected: PASS.

Run: `grep -rn 'executor subagent (default)' --exclude-dir=.git --exclude-dir=.acta .`
Expected: no output.

- [x] **Step 5: Commit**

```bash
gofmt -l cmd internal && go vet ./... && go test ./...
git add plugin/skills/plan/SKILL.md internal/hook/hook.go internal/hook/hook_test.go plugin/hooks/default-rules.md internal/plugincheck/skill_plan_test.go
git commit -m "Plan takes the build executor from the config instead of asking"
```
