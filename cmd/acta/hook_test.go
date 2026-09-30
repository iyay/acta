package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookCommands(t *testing.T) {
	dir := t.TempDir()
	claude := t.TempDir()
	t.Setenv("PM_VOICE_FILE", filepath.Join(dir, "voice.yaml"))
	t.Setenv("CLAUDE_CONFIG_DIR", claude)
	if err := os.WriteFile(filepath.Join(claude, "settings.json"),
		[]byte(`{"enabledPlugins": {"superpowers@superpowers-dev": true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	known := filepath.Join(dir, "known.txt")
	if err := os.WriteFile(known, []byte("superpowers\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, _, code := acta(t, dir, "", "hook", "prompt")
	if code != 0 || !strings.Contains(out, "not set up yet") {
		t.Fatalf("prompt before setup: %d %q", code, out)
	}
	out, _, code = acta(t, dir, "", "hook", "session-start", "--known", known)
	if code != 0 || !strings.Contains(out, "Voice: not set up yet.") || !strings.Contains(out, "superpowers@superpowers-dev") {
		t.Fatalf("session-start before setup: %d\n%s", code, out)
	}

	if _, _, code := acta(t, dir, "", "config", "set", "--language", "Korean"); code != 0 {
		t.Fatal("config set failed")
	}
	out, _, _ = acta(t, dir, "", "hook", "prompt")
	if strings.TrimSpace(out) != "acta voice: reply in Korean, adhd style." {
		t.Fatalf("prompt after setup: %q", out)
	}
	out, _, _ = acta(t, dir, "", "hook", "session-start")
	if !strings.Contains(out, "in Korean.") || strings.Contains(out, "Another workflow plugin") {
		t.Fatalf("session-start without --known: %s", out)
	}

	if err := os.WriteFile(filepath.Join(dir, "voice.yaml"), []byte("style: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, _, code := acta(t, dir, "", "hook", "session-start"); code != 0 || !strings.Contains(out, "could not be read") {
		t.Fatalf("broken voice must still exit 0 with a message: %d %s", code, out)
	}

	for _, args := range [][]string{{"hook"}, {"hook", "frob"}, {"hook", "session-start", "--nope"}} {
		if _, _, code := acta(t, dir, "", args...); code != 1 {
			t.Errorf("%v: exit %d, want 1", args, code)
		}
	}
}

// A session must leave the record file git-ignored, or it shows up in every
// git status. The line is added once and never committed.
func TestHookSessionStartIgnoresTheAgentFile(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	ignore := filepath.Join(dir, ".acta", ".gitignore")
	for range 2 {
		if _, _, code := acta(t, dir, "", "hook", "session-start"); code != 0 {
			t.Fatalf("session-start exit %d", code)
		}
	}
	b, err := os.ReadFile(ignore)
	if err != nil {
		t.Fatalf("%s: %v", ignore, err)
	}
	if n := strings.Count(string(b), ".agents.json"); n != 1 {
		t.Fatalf("gitignore holds the line %d times: %q", n, b)
	}
	if n := commitCount(t, dir); n != "1" {
		t.Errorf("session-start committed: %s commits, want 1", n)
	}
}
