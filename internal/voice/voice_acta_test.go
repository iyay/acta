package voice

import (
	"os"
	"path/filepath"
	"testing"
)

// withHome points HOME at a fresh folder so UserHomeDir lands there.
func withHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", "")
	return home
}

func writeVoice(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPathIsActaDefault(t *testing.T) {
	home := withHome(t)
	got, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".acta", "voice.yaml"); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestResolveFallsBackToOldFile(t *testing.T) {
	home := withHome(t)
	writeVoice(t, filepath.Join(home, ".pm", "voice.yaml"), "chat_language: Korean\nstyle: plain\n")
	v, exists, err := Resolve()
	if err != nil || !exists {
		t.Fatalf("got %+v %v %v", v, exists, err)
	}
	if v.ChatLanguage != "Korean" || v.Style != "plain" {
		t.Fatalf("got %+v, want the old file values", v)
	}
}

func TestResolvePrefersNewFile(t *testing.T) {
	home := withHome(t)
	writeVoice(t, filepath.Join(home, ".pm", "voice.yaml"), "chat_language: Korean\nstyle: plain\n")
	writeVoice(t, filepath.Join(home, ".acta", "voice.yaml"), "chat_language: German\nstyle: adhd\n")
	v, exists, err := Resolve()
	if err != nil || !exists {
		t.Fatalf("got %+v %v %v", v, exists, err)
	}
	if v.ChatLanguage != "German" {
		t.Fatalf("got %+v, want the new file values", v)
	}
}

func TestSaveWritesNewFileOnly(t *testing.T) {
	home := withHome(t)
	writeVoice(t, filepath.Join(home, ".pm", "voice.yaml"), "chat_language: Korean\nstyle: plain\n")
	old, _ := os.ReadFile(filepath.Join(home, ".pm", "voice.yaml"))

	v := Default()
	v.ChatLanguage = "German"
	if err := SaveResolved(v); err != nil {
		t.Fatal(err)
	}
	got, exists, err := Load(filepath.Join(home, ".acta", "voice.yaml"))
	if err != nil || !exists {
		t.Fatalf("got %+v %v %v", got, exists, err)
	}
	if got.ChatLanguage != "German" {
		t.Fatalf("new file holds %+v, want German", got)
	}
	now, _ := os.ReadFile(filepath.Join(home, ".pm", "voice.yaml"))
	if string(now) != string(old) {
		t.Fatalf("old file changed: was %q, now %q", old, now)
	}
}
