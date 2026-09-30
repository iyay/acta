package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The repo language in .acta.yaml reaches the session rules, so an agent in
// that repo writes files in the repo's language.
func TestHookUsesRepoLanguage(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("PM_VOICE_FILE", global)
	if err := os.WriteFile(global, []byte("chat_language: Indonesian\nstyle: adhd\nrepo_language: English\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte("repo_language: Korean\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	out := mustRun(t, "hook", "session-start")
	if !strings.Contains(out, "(code, comments, commits, specs, plans) in Korean") {
		t.Errorf("session-start ignores the repo language:\n%s", out)
	}
	// A broken repo file is named, and the chat language the user picked
	// stays, since a repo file must never change how the agent talks.
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte("chat_language: Korean\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out = mustRun(t, "hook", "session-start")
	if !strings.Contains(out, ".acta.yaml") || !strings.Contains(out, "in Indonesian") || !strings.Contains(out, "plans) in English") {
		t.Errorf("session-start with a bad .acta.yaml:\n%s", out)
	}
	if out := mustRun(t, "hook", "prompt"); !strings.Contains(out, "reply in Indonesian") {
		t.Errorf("prompt with a bad .acta.yaml: %q", out)
	}
}

// A repo file that holds only keys the agent does not print must not change
// the voice lines, and no file shape at all may crash the hook.
func TestHookRepoFileEdgeCases(t *testing.T) {
	global := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("PM_VOICE_FILE", global)
	if err := os.WriteFile(global, []byte("chat_language: Indonesian\nstyle: adhd\nrepo_language: English\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A repo file that sets only a key the voice does not print changes
	// nothing and is not an error.
	for _, body := range []string{"build_executor: subagent\n", "plan_depth: minimal\n"} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Chdir(dir)
		out := mustRun(t, "hook", "session-start")
		if !strings.Contains(out, "(code, comments, commits, specs, plans) in English") {
			t.Errorf("voice changed for %q:\n%s", body, out)
		}
		if strings.Contains(out, "The repo settings could not be read") {
			t.Errorf("reported %q as a broken repo file:\n%s", body, out)
		}
	}
	// No file shape at all may crash the hook or change the voice.
	for _, tc := range []struct{ name, body string }{
		{"empty file", ""},
		{"unparseable file", "repo_language: [unclosed\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			t.Chdir(dir)
			if out := mustRun(t, "hook", "session-start"); !strings.Contains(out, "plans) in English") {
				t.Errorf("voice changed:\n%s", out)
			}
		})
	}
	// A folder where the file should be is still no crash and no bad exit.
	t.Run("repo file is a folder", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, ".acta.yaml"), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(dir)
		if out := mustRun(t, "hook", "session-start"); !strings.Contains(out, "plans) in English") {
			t.Errorf("voice changed:\n%s", out)
		}
	})
}
