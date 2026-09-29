package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// hookRepo makes a git repo holding a planning root, which is all the hook
// needs to find where its state lives.
func hookRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".acta"), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	return dir
}

// runHook calls the real command line, because the hooks reach acta this way.
func runHook(t *testing.T, sub, stdin string) (int, string, string) {
	t.Helper()
	var out, errb strings.Builder
	code := Run([]string{"hook", sub}, strings.NewReader(stdin), false, &out, &errb)
	return code, out.String(), errb.String()
}

func hookEvent(session, command string) string {
	return fmt.Sprintf(`{"session_id":%q,"tool_input":{"command":%q}}`, session, command)
}

const brainstormOne = "acta set scratch/one status brainstorming"

func readState(t *testing.T, dir string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".acta", "state", "sessions.json"))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var state map[string]string
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("state file is not a map of session to item: %v (%s)", err, data)
	}
	return state
}

// TestHookSessionPostToolRecords covers the one write the post-tool hook makes:
// the item a session brainstormed, kept out of git the first time round.
func TestHookSessionPostToolRecords(t *testing.T) {
	dir := hookRepo(t)
	t.Chdir(dir)

	if code, out, errb := runHook(t, "post-tool", hookEvent("s1", brainstormOne)); code != exitOK || out != "" || errb != "" {
		t.Fatalf("recording a brainstorm: exit %d, stdout %q, stderr %q", code, out, errb)
	}
	if got := readState(t, dir)["s1"]; got != "one" {
		t.Fatalf("state holds %q for s1, want the stem one", got)
	}
	ignore, err := os.ReadFile(filepath.Join(dir, ".acta", ".gitignore"))
	if err != nil {
		t.Fatalf("read .acta/.gitignore: %v", err)
	}
	if !strings.Contains(string(ignore), "state/") {
		t.Fatalf(".acta/.gitignore is %q, want a line for state/", ignore)
	}
}

// TestHookSessionPostToolQuietPaths covers every post-tool input that must
// leave the session alone: the hook runs on every single Bash call.
func TestHookSessionPostToolQuietPaths(t *testing.T) {
	cases := []struct {
		name  string
		stdin string
	}{
		{"bad json", "{bad"},
		{"no session", `{"tool_input":{"command":"` + brainstormOne + `"}}`},
		{"no command", `{"session_id":"s1"}`},
		{"empty stdin", ""},
		{"another command", hookEvent("s1", "git status")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := hookRepo(t)
			t.Chdir(dir)
			if code, out, errb := runHook(t, "post-tool", c.stdin); code != exitOK || out != "" || errb != "" {
				t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", code, out, errb)
			}
			if _, err := os.Stat(filepath.Join(dir, ".acta", "state", "sessions.json")); err == nil {
				t.Fatal("a command that is not a brainstorm wrote state")
			}
		})
	}
}

// TestHookSessionPostToolUnwritableRoot keeps a planning folder the hook cannot
// write to from turning into a failed hook.
func TestHookSessionPostToolUnwritableRoot(t *testing.T) {
	dir := hookRepo(t)
	t.Chdir(dir)
	root := filepath.Join(dir, ".acta")
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })
	if code, out, errb := runHook(t, "post-tool", hookEvent("s1", brainstormOne)); code != exitOK || out != "" || errb != "" {
		t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", code, out, errb)
	}
}

// TestHookSessionPreToolBlocksSecondItem is the one input that stops work: a
// second, different brainstorm in a session that already has one.
func TestHookSessionPreToolBlocksSecondItem(t *testing.T) {
	dir := hookRepo(t)
	t.Chdir(dir)
	if code, _, _ := runHook(t, "post-tool", hookEvent("s1", brainstormOne)); code != exitOK {
		t.Fatalf("recording the first brainstorm: exit %d", code)
	}

	code, out, errb := runHook(t, "pre-tool", hookEvent("s1", "acta set scratch/two status brainstorming"))
	if code != exitBlock {
		t.Fatalf("a second, different brainstorm: exit %d, want %d", code, exitBlock)
	}
	if out != "" {
		t.Fatalf("the block message must go to stderr only, stdout was %q", out)
	}
	if !strings.Contains(errb, "already brainstormed") {
		t.Fatalf("stderr %q, want the block message", errb)
	}
}

// TestHookSessionPreToolQuietPaths covers every pre-tool input that must let
// the command run untouched.
func TestHookSessionPreToolQuietPaths(t *testing.T) {
	cases := []struct {
		name  string
		stdin string
	}{
		{"the same item again", hookEvent("s1", brainstormOne)},
		{"another command", hookEvent("s1", "git status")},
		{"another session", hookEvent("s2", "acta set scratch/two status brainstorming")},
		{"bad json", "{bad"},
		{"no session", `{"tool_input":{"command":"acta set scratch/two status brainstorming"}}`},
		{"no command", `{"session_id":"s1"}`},
		{"empty stdin", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := hookRepo(t)
			t.Chdir(dir)
			if code, _, _ := runHook(t, "post-tool", hookEvent("s1", brainstormOne)); code != exitOK {
				t.Fatalf("recording the first brainstorm: exit %d", code)
			}
			if code, out, errb := runHook(t, "pre-tool", c.stdin); code != exitOK || out != "" || errb != "" {
				t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", code, out, errb)
			}
		})
	}
}

// TestHookSessionOutsideAnyRepo keeps the hooks from writing planning folders
// into a folder that is not a project.
func TestHookSessionOutsideAnyRepo(t *testing.T) {
	for _, sub := range []string{"pre-tool", "post-tool"} {
		t.Run(sub, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			if code, out, errb := runHook(t, sub, hookEvent("s1", brainstormOne)); code != exitOK || out != "" || errb != "" {
				t.Fatalf("exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", code, out, errb)
			}
			if _, err := os.Stat(filepath.Join(dir, ".acta")); err == nil {
				t.Fatal("the hook made a planning folder outside a project")
			}
		})
	}
}

// TestHookSessionPromptReminder checks the prompt hook keeps its voice line on
// every path and names the brainstormed item only for a session that has one.
func TestHookSessionPromptReminder(t *testing.T) {
	dir := hookRepo(t)
	t.Chdir(dir)
	// A voice file that is not there makes the voice line the same on every
	// machine, so the test does not read the developer's own voice.
	t.Setenv("PM_VOICE_FILE", filepath.Join(dir, "voice.yaml"))
	if code, _, _ := runHook(t, "post-tool", hookEvent("s1", brainstormOne)); code != exitOK {
		t.Fatalf("recording the first brainstorm: exit %d", code)
	}

	voiceCode, voiceOut, voiceErr := runHook(t, "prompt", "")
	if voiceCode != exitOK || voiceErr != "" || !strings.Contains(voiceOut, "acta voice:") {
		t.Fatalf("prompt with empty stdin: exit %d, stdout %q, stderr %q", voiceCode, voiceOut, voiceErr)
	}

	// A session that brainstormed gets the same voice line plus one more.
	code, out, errb := runHook(t, "prompt", hookEvent("s1", "git status"))
	if code != exitOK || errb != "" {
		t.Fatalf("prompt for a session that brainstormed: exit %d, stderr %q", code, errb)
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 2 || lines[0] != strings.TrimSuffix(voiceOut, "\n") {
		t.Fatalf("prompt output %q, want the voice line %q and one more line", out, voiceOut)
	}
	if !strings.Contains(lines[1], "already brainstormed") {
		t.Fatalf("second line %q, want the reminder naming the item", lines[1])
	}

	// A session that never brainstormed, and a project with no state file, get
	// the voice line on its own.
	code, out, errb = runHook(t, "prompt", hookEvent("s9", "git status"))
	if code != exitOK || errb != "" || out != voiceOut {
		t.Fatalf("prompt for a fresh session: exit %d, stdout %q, stderr %q, want %q", code, out, errb, voiceOut)
	}

	// Outside a project the reminder has nowhere to read, so the voice line
	// is all there is.
	t.Chdir(t.TempDir())
	if code, out, errb := runHook(t, "prompt", hookEvent("s1", "git status")); code != exitOK || errb != "" || out != voiceOut {
		t.Fatalf("prompt outside a project: exit %d, stdout %q, stderr %q, want %q", code, out, errb, voiceOut)
	}
}

// TestHookSessionArgs covers the wrong command lines, which are the user's
// typing and must be answered with the usage line.
func TestHookSessionArgs(t *testing.T) {
	dir := hookRepo(t)
	t.Chdir(dir)
	for _, args := range [][]string{{}, {"pre-tool", "extra"}, {"post-tool", "extra"}} {
		t.Run(strings.Join(append([]string{"hook"}, args...), " "), func(t *testing.T) {
			var out, errb strings.Builder
			code := Run(append([]string{"hook"}, args...), strings.NewReader(""), false, &out, &errb)
			if code != exitBadInput {
				t.Fatalf("exit %d, want %d", code, exitBadInput)
			}
			if !strings.Contains(errb.String(), hookUsage) || out.String() != "" {
				t.Fatalf("stderr %q, stdout %q, want the usage line on stderr", errb.String(), out.String())
			}
		})
	}
}
