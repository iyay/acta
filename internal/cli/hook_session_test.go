package cli

import (
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

// hookScratch writes the two scratch items the brainstorm fixtures name, so
// the hooks see real items instead of mistyped stems.
func hookScratch(t *testing.T, dir string) {
	t.Helper()
	for stem, id := range map[string]string{"one": "SCRATCH-1", "two": "SCRATCH-2"} {
		path := filepath.Join(dir, ".acta", "scratch", stem+".md")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("---\nid: "+id+"\n---\n# "+stem+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
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

// sessionStart runs the session start hook in dir and hands back its text.
func sessionStart(t *testing.T, dir string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
	t.Chdir(dir)
	code, out, errb := runHook(t, "session-start", "")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, errb)
	}
	if !strings.Contains(out, "acta plugin is active.") {
		t.Fatalf("the normal session text is missing:\n%s", out)
	}
	return out
}

// A session that starts on a half done plan is told where the work stands,
// without asking the user.
func TestHookSessionStartNamesRunningPlans(t *testing.T) {
	dir, wt := stateWorktreeRepo(t, startedStatePlan)
	gitOut(t, wt, "commit", "-q", "--allow-empty", "-m", "state: show the view")
	out := sessionStart(t, dir)
	for _, want := range []string{
		"plans/2026-10-05-live-work-state",
		"#task-2",
		wt,
		"state: show the view",
		"next: run the state test",
		"acta state set <plan> next",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("session text does not hold %q:\n%s", want, out)
		}
	}
}

// A repo with a started plan but no worktree has no work going on, so the
// session text is the text it always got.
func TestHookSessionStartWithNoRunningPlanAddsNothing(t *testing.T) {
	out := sessionStart(t, stateRepo(t, startedStatePlan))
	for _, gone := range []string{"live-work-state#task", "worktree:", "acta state set <plan> next"} {
		if strings.Contains(out, gone) {
			t.Errorf("a repo with no worktree got %q in the session text:\n%s", gone, out)
		}
	}
}

// Another repo's plans are another repo's work. A plan running next door on
// the same machine must not turn up in this session.
func TestHookSessionStartLeavesOtherReposPlansOut(t *testing.T) {
	other, otherWt := stateWorktreeRepo(t, startedStatePlan)
	gitOut(t, otherWt, "commit", "-q", "--allow-empty", "-m", "other: state view")
	dir, _ := stateWorktreeRepo(t, startedStatePlan)

	if out := sessionStart(t, dir); !strings.Contains(out, "plans/2026-10-05-live-work-state") {
		t.Fatalf("this repo's own running plan is missing:\n%s", out)
	} else if strings.Contains(out, otherWt) || strings.Contains(out, "other: state view") {
		t.Errorf("the other repo's plan is in this session text:\n%s", out)
	}
	if out := sessionStart(t, other); !strings.Contains(out, "plans/2026-10-05-live-work-state") {
		t.Errorf("the other repo does not see its own plan:\n%s", out)
	}
}

// A folder that is not a planning repo, or one whose planning folder cannot
// be read, still starts its session: the hook exits 0 with the normal text
// and adds nothing.
func TestHookSessionStartQuietOnUnreadablePlanningRoot(t *testing.T) {
	dir := hookRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".acta", "plans"), []byte("not a folder\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := sessionStart(t, dir)
	for _, gone := range []string{"worktree:", "acta state set <plan> next", "more, run acta state"} {
		if strings.Contains(out, gone) {
			t.Errorf("a broken planning root got %q in the session text:\n%s", gone, out)
		}
	}
}

// A git repo with no planning folder at all is the same: a session starts.
func TestHookSessionStartQuietOutsideAPlanningRepo(t *testing.T) {
	dir := t.TempDir()
	gitOut(t, dir, "init", "-q")
	out := sessionStart(t, dir)
	if strings.Contains(out, "acta state set <plan> next") {
		t.Errorf("a repo with no planning folder got the running plans text:\n%s", out)
	}
}

// TestHookSessionManyTopicsInOneSession lets one session brainstorm several
// topics. An old state file from the per-session limit, sound or broken, must
// neither stop the next topic nor add a reminder, and post-tool writes nothing.
func TestHookSessionManyTopicsInOneSession(t *testing.T) {
	cases := []struct {
		name, state string
	}{
		{"no old state", ""},
		{"old state names another topic", `{"s1":"one"}`},
		{"old state is broken", "{bad"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := hookRepo(t)
			t.Chdir(dir)
			hookScratch(t, dir)
			t.Setenv("PM_VOICE_FILE", filepath.Join(dir, "voice.yaml"))
			statePath := filepath.Join(dir, ".acta", "state", "sessions.json")
			if c.state != "" {
				if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(statePath, []byte(c.state), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			for _, cmd := range []string{brainstormOne, "acta set scratch/two status brainstorming"} {
				if code, out, errb := runHook(t, "pre-tool", hookEvent("s1", cmd)); code != exitOK || out != "" || errb != "" {
					t.Fatalf("pre-tool %q: exit %d, stdout %q, stderr %q, want it to pass quietly", cmd, code, out, errb)
				}
				if code, out, errb := runHook(t, "post-tool", hookEvent("s1", cmd)); code != exitOK || out != "" || errb != "" {
					t.Fatalf("post-tool %q: exit %d, stdout %q, stderr %q", cmd, code, out, errb)
				}
			}
			got, err := os.ReadFile(statePath)
			if c.state == "" && err == nil {
				t.Fatalf("post-tool wrote brainstorm state %q", got)
			}
			if c.state != "" && string(got) != c.state {
				t.Fatalf("post-tool changed the old state to %q", got)
			}
			_, voice, _ := runHook(t, "prompt", "")
			if code, out, errb := runHook(t, "prompt", hookEvent("s1", "git status")); code != exitOK || errb != "" || out != voice {
				t.Fatalf("prompt: exit %d, stdout %q, stderr %q, want only the voice line %q", code, out, errb, voice)
			}
		})
	}
}
