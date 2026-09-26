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

	out, _, code := pmb(t, dir, "", "hook", "prompt")
	if code != 0 || !strings.Contains(out, "not set up yet") {
		t.Fatalf("prompt before setup: %d %q", code, out)
	}
	out, _, code = pmb(t, dir, "", "hook", "session-start", "--known", known)
	if code != 0 || !strings.Contains(out, "Voice: not set up yet.") || !strings.Contains(out, "superpowers@superpowers-dev") {
		t.Fatalf("session-start before setup: %d\n%s", code, out)
	}

	if _, _, code := pmb(t, dir, "", "voice", "set", "--language", "Korean"); code != 0 {
		t.Fatal("voice set failed")
	}
	out, _, _ = pmb(t, dir, "", "hook", "prompt")
	if strings.TrimSpace(out) != "pm voice: reply in Korean, adhd style." {
		t.Fatalf("prompt after setup: %q", out)
	}
	out, _, _ = pmb(t, dir, "", "hook", "session-start")
	if !strings.Contains(out, "in Korean.") || strings.Contains(out, "Another workflow plugin") {
		t.Fatalf("session-start without --known: %s", out)
	}

	if err := os.WriteFile(filepath.Join(dir, "voice.yaml"), []byte("style: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, _, code := pmb(t, dir, "", "hook", "session-start"); code != 0 || !strings.Contains(out, "could not be read") {
		t.Fatalf("broken voice must still exit 0 with a message: %d %s", code, out)
	}

	for _, args := range [][]string{{"hook"}, {"hook", "frob"}, {"hook", "session-start", "--nope"}} {
		if _, _, code := pmb(t, dir, "", args...); code != 1 {
			t.Errorf("%v: exit %d, want 1", args, code)
		}
	}
}
