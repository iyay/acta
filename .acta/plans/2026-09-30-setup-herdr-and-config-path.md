---
parent: debt/2026-09-30-build-owns-dispatch
closes: [DBT-0051.15]
id: PLN-0059
created: "2026-09-30 22:47:34"
hash: ya759gz
started: "2026-09-30 22:51:02"
finished: "2026-09-30 22:54:44"
---
# Setup Herdr Check and Config Show Path Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Setup offers `dispatch` only inside a herdr pane, and `acta config show` names the file it really read.

**Architecture:** One sentence changes in the setup skill. In Go, `config.ResolveUserFile` returns the path it read next to the values; `ResolveUser` wraps it, and `acta config show` prints that path plus where writes go.

**Tech Stack:** Markdown skill text, Go.

**Spec:** `.acta/specs/2026-09-30-setup-herdr-and-config-path-design.md`

**Tests:** fast `scripts/test`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- `acta config set` still writes to `config.UserPath()` only. No other caller of `config.ResolveUser` changes.
- Tests never touch the real HOME: set `HOME` (and `PM_VOICE_FILE=""`) to a temp dir, as `withHome` in `internal/config/user_acta_test.go` does.
- Every file written is in English. Comments use short, plain words and say why.

## File map

| File | Task |
|---|---|
| `plugin/skills/setup/SKILL.md` | 1 |
| `internal/plugincheck/skill_setup_test.go` | 1 |
| `internal/config/user.go` | 2 |
| `internal/config/user_acta_test.go` | 2 |
| `internal/cli/config_cmd.go` | 2 |
| `internal/cli/cli_test.go` | 2 |

## Waves

- Wave 1: Task 1 and Task 2 (no shared file).

---

### Task 1: Setup offers dispatch only when HERDR_ENV=1

**Files:**
- Modify: `plugin/skills/setup/SKILL.md` (the `### Default build executor` paragraph)
- Test: `internal/plugincheck/skill_setup_test.go`

**verify:** No text in `plugin/skills/setup/` offers `dispatch` unless `HERDR_ENV=1` is set; `herdr` on PATH alone never makes setup offer it. List every sentence in the setup skill that decides whether `dispatch` is offered, and what each says.

**Interfaces:**
- Consumes: nothing.
- Produces: nothing.

- [x] **Step 1: Write the failing test**

In `TestSkillSetup`, add to `Must`:

```go
			"Offer `dispatch` only when `HERDR_ENV=1` is in the environment",
```

and add to `MustNot`:

```go
			"or `herdr` on PATH",
```

- [x] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/plugincheck/ -run TestSkillSetup`
Expected: FAIL: missing the new sentence, and the old `or `herdr` on PATH` wording is found.

- [x] **Step 3: Change the sentence**

In `plugin/skills/setup/SKILL.md`, the paragraph under `### Default build executor` becomes:

```markdown
`subagent` (default) or `inline`. Offer `dispatch` only when `HERDR_ENV=1` is in the environment, which means this session runs inside a herdr pane. `herdr` on PATH is not enough: dispatch needs this session's own pane. Without `HERDR_ENV=1`, do not offer `dispatch` at all.
```

- [x] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/plugincheck/`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l internal/ && go vet ./internal/plugincheck/
git add plugin/skills/setup/SKILL.md internal/plugincheck/skill_setup_test.go
git commit -m "Setup offers dispatch only inside a herdr pane"
```

### Task 2: config show names the file it read

**Files:**
- Modify: `internal/config/user.go` (`ResolveUser`, new `ResolveUserFile`)
- Modify: `internal/cli/config_cmd.go` (the `show` case)
- Test: `internal/config/user_acta_test.go`, `internal/cli/cli_test.go`

**verify:** Whatever file the values come from, `acta config show` (text and `--json`) names that file, and says where `config set` writes whenever the two differ. List every path `ResolveUserFile` can read from (config.yaml, voice.yaml moved, voice.yaml that cannot move, `~/.pm/voice.yaml`, no file, `PM_VOICE_FILE`) and the path it returns for each.

**Interfaces:**
- Consumes: nothing.
- Produces: `func ResolveUserFile() (User, string, bool, error)` in package `config`: values, the path read (`UserPath()` when no file exists), exists, error.

- [x] **Step 1: Write the failing tests**

In `internal/config/user_acta_test.go`, add:

```go
// TestResolveUserFileNamesTheFileRead checks every place the setting can come
// from. config show prints this path, so a wrong one sends the user to a file
// that is not there.
func TestResolveUserFileNamesTheFileRead(t *testing.T) {
	t.Run("config.yaml", func(t *testing.T) {
		home := withHome(t)
		want := filepath.Join(home, ".acta", "config.yaml")
		writeVoice(t, want, "chat_language: Korean\nstyle: plain\n")
		if _, got, exists, err := ResolveUserFile(); err != nil || !exists || got != want {
			t.Fatalf("got %q %v %v, want %q", got, exists, err, want)
		}
	})
	t.Run("voice.yaml moved to config.yaml", func(t *testing.T) {
		home := withHome(t)
		writeVoice(t, filepath.Join(home, ".acta", "voice.yaml"), "chat_language: Korean\nstyle: plain\n")
		want := filepath.Join(home, ".acta", "config.yaml")
		if _, got, exists, err := ResolveUserFile(); err != nil || !exists || got != want {
			t.Fatalf("got %q %v %v, want %q", got, exists, err, want)
		}
	})
	t.Run("old ~/.pm file", func(t *testing.T) {
		home := withHome(t)
		want := filepath.Join(home, ".pm", "voice.yaml")
		writeVoice(t, want, "chat_language: Korean\nstyle: plain\n")
		if v, got, exists, err := ResolveUserFile(); err != nil || !exists || got != want || v.ChatLanguage != "Korean" {
			t.Fatalf("got %+v %q %v %v, want %q", v, got, exists, err, want)
		}
	})
	t.Run("no file", func(t *testing.T) {
		home := withHome(t)
		want := filepath.Join(home, ".acta", "config.yaml")
		if _, got, exists, err := ResolveUserFile(); err != nil || exists || got != want {
			t.Fatalf("got %q %v %v, want %q and exists false", got, exists, err, want)
		}
	})
	t.Run("PM_VOICE_FILE", func(t *testing.T) {
		withHome(t)
		want := filepath.Join(t.TempDir(), "mine.yaml")
		t.Setenv("PM_VOICE_FILE", want)
		if _, got, _, err := ResolveUserFile(); err != nil || got != want {
			t.Fatalf("got %q %v, want %q", got, err, want)
		}
	})
}
```

In `TestResolveReadsVoiceFileWhenMoveFails`, change the call to `ResolveUserFile` and check the path:

```go
	v, got, exists, err := ResolveUserFile()
	if err != nil || !exists || v.ChatLanguage != "Korean" {
		t.Fatalf("got %+v %v %v, want the voice.yaml values", v, exists, err)
	}
	if got != oldFile {
		t.Fatalf("path = %q, want the voice.yaml it read, %q", got, oldFile)
	}
```

In `internal/cli/cli_test.go`, add:

```go
// TestConfigShowNamesOldFile covers the user who moved config.yaml away: show
// once named the missing file and printed the values of ~/.pm/voice.yaml.
func TestConfigShowNamesOldFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", "")
	old := filepath.Join(home, ".pm", "voice.yaml")
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("chat_language: Korean\nstyle: plain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writes := filepath.Join(home, ".acta", "config.yaml")
	out := mustRun(t, "config", "show")
	want := "file: " + old + " (exists: true, old file; config set writes " + writes + ")"
	if !strings.Contains(out, want) {
		t.Errorf("show:\n%s\nwant line %q", out, want)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if got["path"] != old || got["writes"] != writes {
		t.Errorf("json path %v writes %v, want %q and %q", got["path"], got["writes"], old, writes)
	}
}
```

Add `encoding/json` and `os` to the imports of `cli_test.go` if they are missing.

- [x] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/config/ -run 'TestResolveUserFile|TestResolveReadsVoiceFileWhenMoveFails' && go test ./internal/cli/ -run TestConfigShowNamesOldFile`
Expected: FAIL: `undefined: ResolveUserFile`, then the old `file:` line in the CLI test.

- [x] **Step 3: Write the code**

In `internal/config/user.go`, replace `ResolveUser` with the two functions below. Keep the existing comment inside the rename branch as it is.

```go
// ResolveUser reads config.yaml. When it is missing, an old ~/.acta/voice.yaml
// is renamed to config.yaml first, so the user ends up with one file. When
// both are missing, the ~/.pm file from before the acta rename is read.
func ResolveUser() (User, bool, error) {
	v, _, exists, err := ResolveUserFile()
	return v, exists, err
}

// ResolveUserFile is ResolveUser plus the path of the file it read. With no
// file at all the path is UserPath, where the next write goes.
func ResolveUserFile() (User, string, bool, error) {
	path, err := UserPath()
	if err != nil {
		return UserDefault(), "", false, err
	}
	v, exists, err := LoadUser(path)
	if exists || err != nil || os.Getenv("PM_VOICE_FILE") != "" {
		return v, path, exists, err
	}
	voiceFile, err := voicePath()
	if err != nil {
		return UserDefault(), path, false, err
	}
	// A missing voice.yaml is fine: there was none, or another hook moved it
	// a moment ago. Any other failure means the file is still there, so read
	// it where it is and lose nothing. The next read tries the move again.
	if err := os.Rename(voiceFile, path); err != nil && !errors.Is(err, os.ErrNotExist) {
		v, exists, err := LoadUser(voiceFile)
		return v, voiceFile, exists, err
	}
	if v, exists, err := LoadUser(path); exists || err != nil {
		return v, path, exists, err
	}
	old, err := oldPath()
	if err != nil {
		return UserDefault(), path, false, err
	}
	v, exists, err = LoadUser(old)
	if !exists {
		return v, path, false, err
	}
	return v, old, true, err
}
```

In `internal/cli/config_cmd.go`, `show` case: call `v, read, exists, err := config.ResolveUserFile()`. In the JSON map, `"path": read` and a new `"writes": path`. The text line becomes:

```go
		// The values can come from an old file. Name it, and say where the
		// next config set goes, so the user edits the right file.
		state := fmt.Sprintf("exists: %v", exists)
		if read != path {
			state += ", old file; config set writes " + path
		}
		fmt.Fprintf(stdout, "file: %s (%s)\nchat_language: %s\nstyle: %s\nrepo_language: %s\n",
			read, state, v.ChatLanguage, v.Style, v.RepoLanguage)
```

- [x] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/config/ ./internal/cli/`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l internal/ && go vet ./internal/config/ ./internal/cli/
git add internal/config/user.go internal/config/user_acta_test.go internal/cli/config_cmd.go internal/cli/cli_test.go
git commit -m "config show names the file it read and where writes go"
```
