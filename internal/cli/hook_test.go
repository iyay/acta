package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/hook"
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
		hookScratch(t, dir)
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

// branchHint is the hint of the page tui-wrap as the worktree branch has it.
const branchHint = `{"hookSpecificOutput":{"hookEventName":"PreToolUse","additionalContext":` +
	`"wiki: .acta/wiki/tui-wrap.md: On the branch: wrap is fixed"}}` + "\n"

// hintWorktree is hintRepo with its page committed, and a linked worktree beside
// it, the way acta:build makes one. The branch has changed the page text. It
// gives the main checkout and the worktree.
func hintWorktree(t *testing.T) (dir, tree string) {
	t.Helper()
	dir = hintRepo(t, "internal/tui/")
	git := func(args ...string) {
		t.Helper()
		full := append([]string{"-C", dir, "-c", "user.name=test", "-c", "user.email=test@example.com"}, args...)
		if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	git("add", "-A")
	git("commit", "-q", "-m", "init")
	tree = filepath.Join(t.TempDir(), "tree")
	git("worktree", "add", "-q", "-b", "feature", tree)
	page := "---\ntype: Gotcha\ntitle: Wrap overflows\ndescription: \"On the branch: wrap is fixed\"\n" +
		"paths: [internal/tui/]\ntimestamp: 2026-10-04T10:00:00Z\n---\nbody\n"
	if err := os.WriteFile(filepath.Join(tree, ".acta", "wiki", "tui-wrap.md"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, tree
}

// A Claude Code build keeps the session in the main checkout, so the hook runs
// there, while the implementers read and edit the worktree by its full path. The
// hint has to come from the checkout that holds the file.
func TestHookPreToolHintsFromTheFilesOwnCheckout(t *testing.T) {
	cases := []struct {
		name, tool string
		input      func(dir, tree string) map[string]string
		want       string
	}{
		{"a Read of a worktree file", "Read", func(dir, tree string) map[string]string {
			return map[string]string{"file_path": filepath.Join(tree, "internal", "tui", "model.go")}
		}, branchHint},
		{"a Write of a new file in a new worktree folder", "Write", func(dir, tree string) map[string]string {
			return map[string]string{"file_path": filepath.Join(tree, "internal", "tui", "new", "x.go")}
		}, branchHint},
		{"a Bash cd into the worktree and a relative word", "Bash", func(dir, tree string) map[string]string {
			return map[string]string{"command": "cd " + tree + " && cat internal/tui/model.go"}
		}, branchHint},
		{"a Bash word with the full path of a worktree file", "Bash", func(dir, tree string) map[string]string {
			return map[string]string{"command": "cat " + filepath.Join(tree, "internal", "tui", "model.go")}
		}, branchHint},
		{"a Read of a file in the checkout the hook runs in", "Read", func(dir, tree string) map[string]string {
			return map[string]string{"file_path": filepath.Join(dir, "internal", "tui", "model.go")}
		}, tuiHint},
		{"a Bash relative word with no cd", "Bash", func(dir, tree string) map[string]string {
			return map[string]string{"command": "cat internal/tui/model.go"}
		}, tuiHint},
		{"a path in no git checkout", "Read", func(dir, tree string) map[string]string {
			return map[string]string{"file_path": "/etc/hosts"}
		}, ""},
		{"a cd into no git checkout", "Bash", func(dir, tree string) map[string]string {
			return map[string]string{"command": "cd /etc && cat internal/tui/model.go"}
		}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir, tree := hintWorktree(t)
			t.Chdir(dir)
			code, out, errb := runHook(t, "pre-tool", toolEvent(t, "s1", "", c.tool, c.input(dir, tree)))
			if code != exitOK || out != c.want || errb != "" {
				t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0, stdout %q and no stderr", code, out, errb, c.want)
			}
			// What was shown is kept in the session checkout, whichever checkout
			// gave the hint. The worktree is left alone.
			keeps := map[string]bool{dir: c.want != "", tree: false}
			for checkout, want := range keeps {
				_, err := os.Stat(filepath.Join(checkout, ".acta", "state", "wiki-hints.json"))
				if (err == nil) != want {
					t.Errorf("%s keeps shown pages = %v, want %v", checkout, err == nil, want)
				}
			}
		})
	}
}

// The hook can run in a folder that is no checkout at all, like a session that
// was started above a few repos. It has no repository of its own to trust, so
// no file gets a hint and nothing is written in any repo.
func TestHookPreToolHintsWhenTheHookRunsOutsideEveryCheckout(t *testing.T) {
	dir, tree := hintWorktree(t)
	t.Chdir(t.TempDir())
	for _, c := range []struct{ name, file string }{
		{"a file of the main checkout", filepath.Join(dir, "internal", "tui", "model.go")},
		{"a file of the worktree", filepath.Join(tree, "internal", "tui", "model.go")},
	} {
		stdin := toolEvent(t, "s1", "", "Read", map[string]string{"file_path": c.file})
		if code, out, errb := runHook(t, "pre-tool", stdin); code != exitOK || out != "" || errb != "" {
			t.Errorf("%s: exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", c.name, code, out, errb)
		}
	}
	for _, d := range []string{dir, tree} {
		if _, err := os.Stat(filepath.Join(d, ".acta", "state")); err == nil {
			t.Errorf("a hook outside every checkout made a state folder in %s", d)
		}
	}
}

// A hint is an extra line. Whatever goes wrong while it is made, even a panic,
// must leave the call as it would be with no wiki: exit 0 and nothing printed.
// A Go panic exits 2, which is the code that blocks the tool.
func TestHookPreToolHintFailureNeverBlocks(t *testing.T) {
	saved := wikiHints
	wikiHints = func(string, string, hook.ToolEvent) string { panic("forced failure in the hint path") }
	t.Cleanup(func() { wikiHints = saved })

	// A panic that nothing caught fails this test with a message, instead of
	// killing the whole test run.
	run := func(t *testing.T, stdin string) (code int, out, errb string) {
		t.Helper()
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("pre-tool let a panic out: %v", r)
			}
		}()
		return runHook(t, "pre-tool", stdin)
	}

	dir := hintRepo(t, "internal/tui/")
	t.Chdir(dir)
	for _, tool := range []string{"Read", "Edit", "Write", "MultiEdit"} {
		stdin := toolEvent(t, "s1", "", tool, map[string]string{"file_path": filepath.Join(dir, "internal", "tui", "model.go")})
		if code, out, errb := run(t, stdin); code != exitOK || out != "" || errb != "" {
			t.Errorf("%s: exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", tool, code, out, errb)
		}
	}
	stdin := toolEvent(t, "s1", "", "Bash", map[string]string{"command": "go vet ./internal/tui/..."})
	if code, out, errb := run(t, stdin); code != exitOK || out != "" || errb != "" {
		t.Errorf("Bash: exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", code, out, errb)
	}

	// The two blocks run before the hint, so they still exit 2 with their reason.
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "test"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	stdin = toolEvent(t, "s1", "", "Bash", map[string]string{"command": "go test ./internal/tui/..."})
	if code, out, errb := run(t, stdin); code != exitBlock || out != "" || !strings.Contains(errb, "run tests with scripts/test") {
		t.Errorf("a bare go test: exit %d, stdout %q, stderr %q, want exit 2 and the block reason", code, out, errb)
	}
	hookScratch(t, dir)
	if code, _, _ := runHook(t, "post-tool", hookEvent("s1", brainstormOne)); code != exitOK {
		t.Fatalf("recording the first brainstorm: exit %d", code)
	}
	stdin = toolEvent(t, "s1", "", "Bash", map[string]string{"command": "acta set scratch/two status brainstorming"})
	if code, out, errb := run(t, stdin); code != exitBlock || out != "" || !strings.Contains(errb, "already brainstormed") {
		t.Errorf("a second brainstorm: exit %d, stdout %q, stderr %q, want exit 2 and the block reason", code, out, errb)
	}
}

// wikiLineHead is how the session start wiki line begins, up to its count.
const wikiLineHead = "Project wiki. Pages: "

// sessionStartEvent is the payload Claude Code sends a SessionStart hook. An
// empty source is left out, like a payload that has none.
func sessionStartEvent(t *testing.T, session, source string) string {
	t.Helper()
	ev := map[string]any{"session_id": session, "hook_event_name": "SessionStart"}
	if source != "" {
		ev["source"] = source
	}
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// putShownPages writes the shown-pages file by hand, with spaces and a line
// break that the hook would never write, so a file the hook left alone can be
// told from one it rewrote. It gives the path.
func putShownPages(t *testing.T, dir, body string) string {
	t.Helper()
	file := filepath.Join(dir, ".acta", "state", "wiki-hints.json")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return file
}

// s1 and its subagent, and s2, have been shown the page.
const shownByThree = "{\"s1/\": [\"tui-wrap\"], \"s1/agent-1\": [\"tui-wrap\"], \"s2/\": [\"tui-wrap\"]}\n"

// A session that starts after a clear or a compaction lost the hints it was
// shown, so its pages are forgotten for the main thread and every subagent. Any
// other session keeps what it was shown.
func TestHookSessionStartForgetsShownPagesAfterClearOrCompact(t *testing.T) {
	for _, source := range []string{"clear", "compact"} {
		t.Run(source, func(t *testing.T) {
			dir := hintRepo(t, "internal/tui/")
			t.Chdir(dir)
			file := putShownPages(t, dir, shownByThree)
			code, out, errb := runHook(t, "session-start", sessionStartEvent(t, "s1", source))
			if code != exitOK || errb != "" {
				t.Fatalf("exit %d, stderr %q, want exit 0 and no stderr", code, errb)
			}
			if !strings.Contains(out, wikiLineHead) {
				t.Error("the session text has no wiki line")
			}
			got, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if want := `{"s2/":["tui-wrap"]}`; string(got) != want {
				t.Errorf("shown pages = %s, want %s", got, want)
			}
		})
	}
}

// Anything that is not a clear or a compaction leaves the shown pages alone, and
// so does a payload the hook cannot use. The session text still prints and the
// hook still exits 0, because a hook must never get in the way of a session.
func TestHookSessionStartKeepsShownPagesOtherwise(t *testing.T) {
	cases := []struct{ name, stdin string }{
		{"startup", sessionStartEvent(t, "s1", "startup")},
		{"resume", sessionStartEvent(t, "s1", "resume")},
		{"fork", sessionStartEvent(t, "s1", "fork")},
		{"a source nobody knows", sessionStartEvent(t, "s1", "other")},
		{"a capital letter", sessionStartEvent(t, "s1", "Clear")},
		{"no source", sessionStartEvent(t, "s1", "")},
		{"clear with no session", `{"source":"clear"}`},
		{"clear with an empty session", sessionStartEvent(t, "", "clear")},
		{"a source that is not text", `{"session_id":"s1","source":5}`},
		{"bad json", "{bad"},
		{"empty stdin", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := hintRepo(t, "internal/tui/")
			t.Chdir(dir)
			file := putShownPages(t, dir, shownByThree)
			code, out, errb := runHook(t, "session-start", c.stdin)
			if code != exitOK || errb != "" {
				t.Fatalf("exit %d, stderr %q, want exit 0 and no stderr", code, errb)
			}
			if !strings.Contains(out, wikiLineHead) {
				t.Error("the session text has no wiki line")
			}
			if got, err := os.ReadFile(file); err != nil || string(got) != shownByThree {
				t.Errorf("shown pages = %q (%v), want them left as they were", got, err)
			}
		})
	}
}

// After a compaction the agent must hear a page again when it touches the files
// that page covers, in the main thread and in a subagent, and a session that was
// not compacted must stay quiet. This goes through the real command line from
// the first hint to the second.
func TestHookSessionStartAfterCompactHintsAgain(t *testing.T) {
	dir := hintRepo(t, "internal/tui/")
	t.Chdir(dir)
	file := map[string]string{"file_path": filepath.Join(dir, "internal", "tui", "model.go")}
	steps := []struct {
		name, hook, stdin, want string
	}{
		{"main thread", "pre-tool", toolEvent(t, "s1", "", "Read", file), tuiHint},
		{"a subagent", "pre-tool", toolEvent(t, "s1", "agent-1", "Read", file), tuiHint},
		{"another session", "pre-tool", toolEvent(t, "s2", "", "Read", file), tuiHint},
		{"main thread again", "pre-tool", toolEvent(t, "s1", "", "Read", file), ""},
		{"s1 compacts", "session-start", sessionStartEvent(t, "s1", "compact"), "*"},
		{"main thread after the compaction", "pre-tool", toolEvent(t, "s1", "", "Read", file), tuiHint},
		{"a subagent after the compaction", "pre-tool", toolEvent(t, "s1", "agent-1", "Read", file), tuiHint},
		{"main thread once more", "pre-tool", toolEvent(t, "s1", "", "Read", file), ""},
		{"the other session after it", "pre-tool", toolEvent(t, "s2", "", "Read", file), ""},
	}
	for _, s := range steps {
		code, out, errb := runHook(t, s.hook, s.stdin)
		if code != exitOK || errb != "" {
			t.Fatalf("%s: exit %d, stderr %q", s.name, code, errb)
		}
		if s.want == "*" {
			continue
		}
		if out != s.want {
			t.Errorf("%s: stdout %q, want %q", s.name, out, s.want)
		}
	}
}

// The line names the number of pages the wiki holds. A page that cannot load is
// for acta wiki check to report, so it is not counted, and a repo with no pages
// has no line at all.
func TestHookSessionStartCountsWikiPages(t *testing.T) {
	dir := hintRepo(t, "internal/tui/")
	t.Chdir(dir)
	wikiDir := filepath.Join(dir, ".acta", "wiki")
	put := func(name, body string) {
		t.Helper()
		file := filepath.Join(wikiDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	line := func(t *testing.T) string {
		t.Helper()
		code, out, errb := runHook(t, "session-start", "")
		if code != exitOK || errb != "" {
			t.Fatalf("exit %d, stderr %q", code, errb)
		}
		for _, l := range strings.Split(out, "\n") {
			if strings.HasPrefix(l, wikiLineHead) {
				return l
			}
		}
		return ""
	}
	if got := line(t); !strings.HasPrefix(got, wikiLineHead+"1.") {
		t.Errorf("one page: line %q", got)
	}
	put("sub/two.md", "---\ntype: Runbook\ntitle: Two\ndescription: A page in a folder\npaths: []\ntimestamp: 2026-10-04T10:00:00Z\n---\nbody\n")
	if got := line(t); !strings.HasPrefix(got, wikiLineHead+"2.") {
		t.Errorf("two pages, one in a folder: line %q", got)
	}
	put("broken.md", "no frontmatter at all\n")
	if got := line(t); !strings.HasPrefix(got, wikiLineHead+"2.") {
		t.Errorf("two pages and a broken one: line %q", got)
	}

	// No pages: a wiki folder with nothing loadable in it, and no wiki folder.
	for name, setup := range map[string]func(t *testing.T) string{
		"only a broken page": func(t *testing.T) string {
			d := hintRepo(t, "internal/tui/")
			if err := os.WriteFile(filepath.Join(d, ".acta", "wiki", "tui-wrap.md"), []byte("no frontmatter\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return d
		},
		"no wiki folder": func(t *testing.T) string {
			t.Setenv("HOME", t.TempDir())
			return hookRepo(t)
		},
		"outside a project": func(t *testing.T) string {
			t.Setenv("HOME", t.TempDir())
			return t.TempDir()
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Chdir(setup(t))
			if got := line(t); got != "" {
				t.Errorf("a repo with no pages got the line %q", got)
			}
		})
	}
}
