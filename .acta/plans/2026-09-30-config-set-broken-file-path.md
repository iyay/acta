---
parent: bugs/2026-09-30-config-set-names-wrong-broken-file
id: PLN-0060
created: "2026-09-30 23:21:23"
hash: rg7bji1
---
# Config Set Broken File Path Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** When `acta config set` cannot read the setting, its "fix or delete" line names the file that is really broken.

**Architecture:** The `set` case switches from `config.ResolveUser` to `config.ResolveUserFile` (added in PLN-0059) and prints the path it returns.

**Tech Stack:** Go.

**Spec:** `.acta/specs/2026-09-30-config-set-broken-file-path-design.md`

**Tests:** fast `scripts/test`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- `acta config set` still writes only to `config.UserPath()` and never overwrites a file it could not read.
- Tests never touch the real HOME: set `HOME` to a temp dir and `PM_VOICE_FILE=""`.
- Every file written is in English. Comments use short, plain words and say why.

## File map

| File | Task |
|---|---|
| `internal/cli/config_cmd.go` | 1 |
| `internal/cli/cli_test.go` | 1 |

## Waves

- Wave 1: Task 1.

---

### Task 1: config set names the broken file

**Files:**
- Modify: `internal/cli/config_cmd.go` (the `set` case, the `config.ResolveUser()` call and its error line)
- Test: `internal/cli/cli_test.go`

**verify:** Whenever `config set` stops on a file it cannot read, the "fix or delete" line names that exact file, and `set` writes nothing. List every file `ResolveUserFile` can fail on (`~/.acta/config.yaml`, `PM_VOICE_FILE`, a `~/.acta/voice.yaml` that could not move, `~/.pm/voice.yaml`) and the path the line prints for each.

**Interfaces:**
- Consumes: `config.ResolveUserFile() (User, string, bool, error)`.
- Produces: nothing.

- [ ] **Step 1: Write the failing test**

In `internal/cli/cli_test.go`, add:

```go
// TestConfigSetNamesBrokenOldFile covers a broken ~/.pm/voice.yaml: set once
// told the user to fix ~/.acta/config.yaml, a file that was not there.
func TestConfigSetNamesBrokenOldFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", "")
	old := filepath.Join(home, ".pm", "voice.yaml")
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("chat_language: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	newFile := filepath.Join(home, ".acta", "config.yaml")
	var stdout, stderr strings.Builder
	code := Run([]string{"config", "set", "--language", "Korean"}, strings.NewReader(""), false, &stdout, &stderr)
	if code != exitBadInput {
		t.Fatalf("exit %d, want %d; stderr %q", code, exitBadInput, stderr.String())
	}
	if want := "fix or delete " + old + " first"; !strings.Contains(stderr.String(), want) {
		t.Errorf("stderr %q lacks %q", stderr.String(), want)
	}
	if strings.Contains(stderr.String(), newFile) {
		t.Errorf("stderr %q names %s, which does not exist", stderr.String(), newFile)
	}
	if _, err := os.Stat(newFile); !os.IsNotExist(err) {
		t.Errorf("set wrote %s after a read error: %v", newFile, err)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/cli/ -run TestConfigSetNamesBrokenOldFile`
Expected: FAIL: stderr names `.acta/config.yaml` instead of `.pm/voice.yaml`.

- [ ] **Step 3: Write the code**

In `internal/cli/config_cmd.go`, `set` case:

```go
		v, read, _, err := config.ResolveUserFile()
		if err != nil {
			// Never overwrite a file the user may still want to repair by hand.
			// Name the file that failed: it can be an old one, not config.yaml.
			fmt.Fprintf(stderr, "%v\nfix or delete %s first\n", err, read)
			return exitBadInput
		}
```

When `ResolveUserFile` fails before it knows any path (`UserPath` error), `read` is empty; `cmdConfig` already returned on that error at its top, so this line never prints an empty path.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/cli/ ./internal/config/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/ && go vet ./internal/cli/
git add internal/cli/config_cmd.go internal/cli/cli_test.go
git commit -m "config set names the broken file it read"
```
