package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVoiceSetAndShow(t *testing.T) {
	dir := t.TempDir()
	vf := filepath.Join(dir, "voice.yaml")
	t.Setenv("PM_VOICE_FILE", vf)

	out, _, code := pmb(t, dir, "", "voice", "show", "--json")
	if code != 0 || !strings.Contains(out, `"exists": false`) || !strings.Contains(out, `"chat_language": "English"`) {
		t.Fatalf("show before set: exit %d %s", code, out)
	}

	if out, errOut, code := pmb(t, dir, "", "voice", "set", "--language", "Korean", "--tone", "Short."); code != 0 || strings.TrimSpace(out) != vf {
		t.Fatalf("set: exit %d out %q err %q", code, out, errOut)
	}
	if _, _, code := pmb(t, dir, "", "voice", "set", "--style", "plain"); code != 0 {
		t.Fatalf("second set exit %d", code)
	}
	out, _, _ = pmb(t, dir, "", "voice", "show", "--json")
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got["chat_language"] != "Korean" || got["style"] != "plain" || got["tone"] != "Short." || got["exists"] != true {
		t.Fatalf("show after set: %v", got)
	}

	if _, _, code := pmb(t, dir, "", "voice", "set", "--clear-tone"); code != 0 {
		t.Fatalf("clear-tone exit %d", code)
	}
	b, _ := os.ReadFile(vf)
	if strings.Contains(string(b), "tone") {
		t.Fatalf("tone still in file: %s", b)
	}
}

func TestVoiceBadInput(t *testing.T) {
	dir := t.TempDir()
	vf := filepath.Join(dir, "voice.yaml")
	t.Setenv("PM_VOICE_FILE", vf)
	for _, args := range [][]string{
		{"voice"},
		{"voice", "set"},
		{"voice", "set", "--style", "loud"},
		{"voice", "set", "--nope"},
		{"voice", "frob"},
		{"voice", "show", "extra"},
	} {
		if _, _, code := pmb(t, dir, "", args...); code != 1 {
			t.Errorf("%v: exit %d, want 1", args, code)
		}
	}
	if _, err := os.Stat(vf); !os.IsNotExist(err) {
		t.Fatal("bad input wrote a voice file")
	}

	if err := os.WriteFile(vf, []byte("style: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, errOut, code := pmb(t, dir, "", "voice", "set", "--language", "Korean"); code != 1 || !strings.Contains(errOut, "fix or delete") {
		t.Fatalf("set over a broken file: exit %d %q", code, errOut)
	}
	if b, _ := os.ReadFile(vf); string(b) != "style: [\n" {
		t.Fatal("broken file was overwritten")
	}
}
