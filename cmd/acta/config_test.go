package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigSetAndShow(t *testing.T) {
	dir := t.TempDir()
	vf := filepath.Join(dir, "voice.yaml")
	t.Setenv("PM_VOICE_FILE", vf)

	out, _, code := acta(t, dir, "", "config", "show", "--json")
	if code != 0 || !strings.Contains(out, `"exists": false`) || !strings.Contains(out, `"chat_language": "English"`) {
		t.Fatalf("show before set: exit %d %s", code, out)
	}

	if out, errOut, code := acta(t, dir, "", "config", "set", "--language", "Korean", "--tone", "Short."); code != 0 || strings.TrimSpace(out) != vf {
		t.Fatalf("set: exit %d out %q err %q", code, out, errOut)
	}
	if _, _, code := acta(t, dir, "", "config", "set", "--style", "plain"); code != 0 {
		t.Fatalf("second set exit %d", code)
	}
	out, _, _ = acta(t, dir, "", "config", "show", "--json")
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got["chat_language"] != "Korean" || got["style"] != "plain" || got["tone"] != "Short." || got["exists"] != true {
		t.Fatalf("show after set: %v", got)
	}

	if _, _, code := acta(t, dir, "", "config", "set", "--clear-tone"); code != 0 {
		t.Fatalf("clear-tone exit %d", code)
	}
	b, _ := os.ReadFile(vf)
	if strings.Contains(string(b), "tone") {
		t.Fatalf("tone still in file: %s", b)
	}
}

func TestConfigBadInput(t *testing.T) {
	dir := t.TempDir()
	vf := filepath.Join(dir, "voice.yaml")
	t.Setenv("PM_VOICE_FILE", vf)
	for _, args := range [][]string{
		{"config"},
		{"config", "set"},
		{"config", "set", "--style", "loud"},
		{"config", "set", "--nope"},
		{"config", "frob"},
		{"config", "show", "extra"},
	} {
		if _, _, code := acta(t, dir, "", args...); code != 1 {
			t.Errorf("%v: exit %d, want 1", args, code)
		}
	}
	if _, err := os.Stat(vf); !os.IsNotExist(err) {
		t.Fatal("bad input wrote a voice file")
	}

	if err := os.WriteFile(vf, []byte("style: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, errOut, code := acta(t, dir, "", "config", "set", "--language", "Korean"); code != 1 || !strings.Contains(errOut, "fix or delete") {
		t.Fatalf("set over a broken file: exit %d %q", code, errOut)
	}
	if b, _ := os.ReadFile(vf); string(b) != "style: [\n" {
		t.Fatal("broken file was overwritten")
	}
}

func TestVoiceCommandIsGone(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PM_VOICE_FILE", filepath.Join(dir, "voice.yaml"))
	for _, args := range [][]string{{"voice"}, {"voice", "show"}, {"voice", "set", "--language", "Korean"}} {
		_, errOut, code := acta(t, dir, "", args...)
		if code != 1 || !strings.Contains(errOut, "unknown command") {
			t.Errorf("%v: exit %d, stderr %q; want 1 and unknown command", args, code, errOut)
		}
	}
}
