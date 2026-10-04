package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The hook reads HERDR_ENV, so the session text can name a herdr tab as the
// extra way to run a second brainstorm. Without it, the text stays silent.
// The phrase belongs to the herdr block only; the skill index has its own line
// about the dispatch skill.
const herdrLine = "this session runs in a herdr tab"

func TestHookSessionStartNamesHerdrOnlyInsideHerdr(t *testing.T) {
	t.Chdir(hookRepo(t))
	t.Setenv("PM_VOICE_FILE", "")

	// Unset, not empty: an unset variable must not count as herdr either.
	saved, had := os.LookupEnv("HERDR_ENV")
	if err := os.Unsetenv("HERDR_ENV"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("HERDR_ENV", saved)
		}
	})

	if _, out, _ := runHook(t, "session-start", ""); strings.Contains(out, herdrLine) {
		t.Errorf("session start outside herdr offers a herdr tab:\n%s", out)
	}
	t.Setenv("HERDR_ENV", "1")
	if _, out, _ := runHook(t, "session-start", ""); !strings.Contains(out, herdrLine) {
		t.Errorf("herdr session start missing the herdr tab sentence:\n%s", out)
	}
	t.Setenv("HERDR_ENV", "0")
	if _, out, _ := runHook(t, "session-start", ""); strings.Contains(out, herdrLine) {
		t.Errorf("HERDR_ENV=0 counts as herdr:\n%s", out)
	}
}

// TestHookPreToolBlocksBareGoTest runs the real command line in a repo that
// keeps its tests in scripts/test, which is the one place a bare go test is
// stopped. A repo without that script is left alone, and a broken hook must
// never stand in the way of any Bash call.
func TestHookPreToolBlocksBareGoTest(t *testing.T) {
	withScript := goTestRepoCLI(t)
	code, out, errb := runGoTestHook(t, withScript, "go test ./... -count=1")
	if code != exitBlock {
		t.Fatalf("a bare go test: exit %d, want %d", code, exitBlock)
	}
	if out != "" {
		t.Fatalf("the block message must go to stderr only, stdout was %q", out)
	}
	want := "acta: run tests with scripts/test, not go test. It adds -short and the machine-wide lock. Use: scripts/test ./..."
	if !strings.Contains(errb, want) {
		t.Fatalf("stderr %q, want it to end in %q", errb, want)
	}

	// Without the script there is no lock to join, so the command runs.
	noScript := hookRepo(t)
	if code, out, errb := runGoTestHook(t, noScript, "go test ./..."); code != exitOK || out != "" || errb != "" {
		t.Fatalf("a repo with no scripts/test: exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", code, out, errb)
	}

	// A repo with a scripts/test but no planning root is still a repo the
	// block applies to: the script is what matters, not the .acta folder.
	noPlanning := t.TempDir()
	if err := os.MkdirAll(filepath.Join(noPlanning, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(noPlanning, "scripts", "test"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", noPlanning, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	if code, _, errb := runGoTestHook(t, noPlanning, "go test ./..."); code != exitBlock {
		t.Fatalf("a bare go test with no planning root: exit %d, stderr %q, want %d", code, errb, exitBlock)
	}
}

// goTestRepoCLI is a git repo whose git top-level holds scripts/test.
func goTestRepoCLI(t *testing.T) string {
	t.Helper()
	dir := hookRepo(t)
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "test"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func runGoTestHook(t *testing.T, dir, command string) (int, string, string) {
	t.Helper()
	t.Chdir(dir)
	return runHook(t, "pre-tool", hookEvent("s1", command))
}

// tuiHint is what Claude Code reads on stdout when a tool call touches a file
// that the page tui-wrap covers: one JSON object on one line.
const tuiHint = `{"hookSpecificOutput":{"hookEventName":"PreToolUse","additionalContext":` +
	`"wiki: .acta/wiki/tui-wrap.md: Use Hardwrap(Wordwrap(...)) for wide lines"}}` + "\n"

// hintRepo is a repo whose wiki has one page, tui-wrap, for the given paths.
func hintRepo(t *testing.T, paths string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := hookRepo(t)
	page := "---\ntype: Gotcha\ntitle: Wrap overflows\ndescription: Use Hardwrap(Wordwrap(...)) for wide lines\n" +
		"paths: [" + paths + "]\ntimestamp: 2026-10-04T10:00:00Z\n---\nbody\n"
	if err := os.MkdirAll(filepath.Join(dir, ".acta", "wiki"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".acta", "wiki", "tui-wrap.md"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// toolEvent is the payload Claude Code sends a PreToolUse hook. agent is empty
// for the main thread, which sends no agent_id at all.
func toolEvent(t *testing.T, session, agent, tool string, input map[string]string) string {
	t.Helper()
	ev := map[string]any{"session_id": session, "tool_name": tool, "tool_input": input}
	if agent != "" {
		ev["agent_id"] = agent
	}
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestHookPreToolHintsAWikiPage runs the real command line for every tool the
// hook is wired to. A hint is the one JSON object on stdout, and the hook
// still exits 0.
func TestHookPreToolHintsAWikiPage(t *testing.T) {
	for _, tool := range []string{"Read", "Edit", "Write", "MultiEdit"} {
		t.Run(tool, func(t *testing.T) {
			dir := hintRepo(t, "internal/tui/")
			t.Chdir(dir)
			stdin := toolEvent(t, "s1", "", tool, map[string]string{"file_path": filepath.Join(dir, "internal", "tui", "model.go")})
			if code, out, errb := runHook(t, "pre-tool", stdin); code != exitOK || out != tuiHint || errb != "" {
				t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0 and %q", code, out, errb, tuiHint)
			}
		})
	}
	t.Run("Bash", func(t *testing.T) {
		dir := hintRepo(t, "internal/tui/")
		t.Chdir(dir)
		stdin := toolEvent(t, "s1", "", "Bash", map[string]string{"command": "go vet ./internal/tui/..."})
		if code, out, errb := runHook(t, "pre-tool", stdin); code != exitOK || out != tuiHint || errb != "" {
			t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0 and %q", code, out, errb, tuiHint)
		}
	})
	t.Run("Bash from a subfolder", func(t *testing.T) {
		dir := hintRepo(t, "internal/tui/")
		sub := filepath.Join(dir, "internal", "tui")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(sub)
		stdin := toolEvent(t, "s1", "", "Bash", map[string]string{"command": "cat model.go"})
		if code, out, errb := runHook(t, "pre-tool", stdin); code != exitOK || out != tuiHint || errb != "" {
			t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0 and %q", code, out, errb, tuiHint)
		}
	})
}

// TestHookPreToolHintsOncePerContext covers the repeat cases through the real
// command line: the same touch, a subagent, another session, and the state that
// has to stay out of git.
func TestHookPreToolHintsOncePerContext(t *testing.T) {
	dir := hintRepo(t, "internal/tui/")
	t.Chdir(dir)
	file := map[string]string{"file_path": filepath.Join(dir, "internal", "tui", "model.go")}
	steps := []struct {
		name, stdin, want string
	}{
		{"first touch", toolEvent(t, "s1", "", "Read", file), tuiHint},
		{"same touch again", toolEvent(t, "s1", "", "Read", file), ""},
		{"another tool, same page", toolEvent(t, "s1", "", "Edit", file), ""},
		{"a subagent", toolEvent(t, "s1", "agent-1", "Read", file), tuiHint},
		{"that subagent again", toolEvent(t, "s1", "agent-1", "Read", file), ""},
		{"another session", toolEvent(t, "s2", "", "Read", file), tuiHint},
	}
	for _, s := range steps {
		if code, out, errb := runHook(t, "pre-tool", s.stdin); code != exitOK || out != s.want || errb != "" {
			t.Errorf("%s: exit %d, stdout %q, stderr %q, want exit 0 and %q", s.name, code, out, errb, s.want)
		}
	}
	state, err := os.ReadFile(filepath.Join(dir, ".acta", "state", "wiki-hints.json"))
	if err != nil {
		t.Fatalf("read the shown pages: %v", err)
	}
	if want := `{"s1/":["tui-wrap"],"s1/agent-1":["tui-wrap"],"s2/":["tui-wrap"]}`; string(state) != want {
		t.Errorf("shown pages = %s, want %s", state, want)
	}
	ignore, err := os.ReadFile(filepath.Join(dir, ".acta", ".gitignore"))
	if err != nil || !strings.Contains(string(ignore), "state/") {
		t.Errorf(".acta/.gitignore = %q (%v), want a line for state/", ignore, err)
	}
}

// TestHookPreToolHintQuietPaths covers every input that must print nothing and
// write nothing, and still exit 0.
func TestHookPreToolHintQuietPaths(t *testing.T) {
	cases := []struct {
		name  string
		stdin func(t *testing.T, dir string) string
	}{
		{"a file no page covers", func(t *testing.T, dir string) string {
			return toolEvent(t, "s1", "", "Read", map[string]string{"file_path": filepath.Join(dir, "README.md")})
		}},
		{"a file outside the repo", func(t *testing.T, dir string) string {
			return toolEvent(t, "s1", "", "Read", map[string]string{"file_path": "/etc/hosts"})
		}},
		{"a command with no path word", func(t *testing.T, dir string) string {
			return toolEvent(t, "s1", "", "Bash", map[string]string{"command": "git status"})
		}},
		{"a file tool with no path", func(t *testing.T, dir string) string {
			return toolEvent(t, "s1", "", "Read", map[string]string{})
		}},
		{"no session", func(t *testing.T, dir string) string {
			return `{"tool_name":"Read","tool_input":{"file_path":"` + filepath.Join(dir, "internal", "tui", "a.go") + `"}}`
		}},
		{"bad json", func(*testing.T, string) string { return "{bad" }},
		{"empty stdin", func(*testing.T, string) string { return "" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := hintRepo(t, "internal/tui/")
			t.Chdir(dir)
			if code, out, errb := runHook(t, "pre-tool", c.stdin(t, dir)); code != exitOK || out != "" || errb != "" {
				t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", code, out, errb)
			}
			if _, err := os.Stat(filepath.Join(dir, ".acta", "state")); err == nil {
				t.Fatal("a call with no hint made a state folder")
			}
		})
	}

	// A repo with no wiki folder has nothing to hint, however the file is named.
	t.Run("a repo with no wiki", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		dir := hookRepo(t)
		t.Chdir(dir)
		stdin := toolEvent(t, "s1", "", "Read", map[string]string{"file_path": filepath.Join(dir, "internal", "tui", "a.go")})
		if code, out, errb := runHook(t, "pre-tool", stdin); code != exitOK || out != "" || errb != "" {
			t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", code, out, errb)
		}
	})
}

// TestHookPreToolHintWithUnwritableRoot keeps a planning folder the hook cannot
// write to from turning into a failed hook.
func TestHookPreToolHintWithUnwritableRoot(t *testing.T) {
	dir := hintRepo(t, "internal/tui/")
	t.Chdir(dir)
	root := filepath.Join(dir, ".acta")
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })
	stdin := toolEvent(t, "s1", "", "Read", map[string]string{"file_path": filepath.Join(dir, "internal", "tui", "a.go")})
	if code, out, errb := runHook(t, "pre-tool", stdin); code != exitOK || out != tuiHint || errb != "" {
		t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0 and %q", code, out, errb, tuiHint)
	}
}

// TestHookPreToolBlocksComeBeforeHints keeps the two blocks that exist today as
// they were: a blocked call exits 2 with its reason on stderr, prints no hint
// and records no page as shown, even when the call touches a covered file.
func TestHookPreToolBlocksComeBeforeHints(t *testing.T) {
	t.Run("a bare go test", func(t *testing.T) {
		dir := hintRepo(t, "scripts/")
		if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "scripts", "test"), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(dir)
		stdin := toolEvent(t, "s1", "", "Bash", map[string]string{"command": "go test ./scripts/..."})
		code, out, errb := runHook(t, "pre-tool", stdin)
		if code != exitBlock || out != "" || !strings.Contains(errb, "run tests with scripts/test") {
			t.Fatalf("exit %d, stdout %q, stderr %q, want exit 2, no stdout and the block reason", code, out, errb)
		}
		if _, err := os.Stat(filepath.Join(dir, ".acta", "state", "wiki-hints.json")); err == nil {
			t.Fatal("a blocked call was recorded as a shown page")
		}
	})
	t.Run("a second brainstorm", func(t *testing.T) {
		dir := hintRepo(t, "scratch/")
		t.Chdir(dir)
		if code, _, _ := runHook(t, "post-tool", hookEvent("s1", brainstormOne)); code != exitOK {
			t.Fatalf("recording the first brainstorm: exit %d", code)
		}
		stdin := toolEvent(t, "s1", "", "Bash", map[string]string{"command": "acta set scratch/two status brainstorming"})
		code, out, errb := runHook(t, "pre-tool", stdin)
		if code != exitBlock || out != "" || !strings.Contains(errb, "already brainstormed") {
			t.Fatalf("exit %d, stdout %q, stderr %q, want exit 2, no stdout and the block reason", code, out, errb)
		}
		if _, err := os.Stat(filepath.Join(dir, ".acta", "state", "wiki-hints.json")); err == nil {
			t.Fatal("a blocked call was recorded as a shown page")
		}
	})
}
