package voice

import (
	"os"
	"path/filepath"
	"sync"
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
	if want := filepath.Join(home, ".acta", "config.yaml"); got != want {
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
	if _, err := os.Stat(filepath.Join(home, ".pm", "voice.yaml")); err != nil {
		t.Fatalf("the ~/.pm file must stay where it is: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "config.yaml")); !os.IsNotExist(err) {
		t.Fatalf("reading the ~/.pm file must not make config.yaml: %v", err)
	}
}

func TestResolvePrefersNewFile(t *testing.T) {
	home := withHome(t)
	writeVoice(t, filepath.Join(home, ".pm", "voice.yaml"), "chat_language: Korean\nstyle: plain\n")
	writeVoice(t, filepath.Join(home, ".acta", "config.yaml"), "chat_language: German\nstyle: adhd\n")
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
	got, exists, err := Load(filepath.Join(home, ".acta", "config.yaml"))
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

func TestResolveMovesVoiceFile(t *testing.T) {
	home := withHome(t)
	oldFile := filepath.Join(home, ".acta", "voice.yaml")
	body := "chat_language: Korean\nstyle: plain\nbuild_executor: dispatch\n"
	writeVoice(t, oldFile, body)

	v, exists, err := Resolve()
	if err != nil || !exists {
		t.Fatalf("got %+v %v %v", v, exists, err)
	}
	if v.ChatLanguage != "Korean" || v.BuildExecutor != "dispatch" {
		t.Fatalf("got %+v, want the voice.yaml values", v)
	}
	got, err := os.ReadFile(filepath.Join(home, ".acta", "config.yaml"))
	if err != nil || string(got) != body {
		t.Fatalf("config.yaml = %q, %v; want the old bytes %q", got, err, body)
	}
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatalf("voice.yaml must be gone after the move: %v", err)
	}
}

func TestResolveConfigWinsOverVoiceFile(t *testing.T) {
	home := withHome(t)
	oldFile := filepath.Join(home, ".acta", "voice.yaml")
	writeVoice(t, oldFile, "chat_language: Korean\nstyle: plain\n")
	writeVoice(t, filepath.Join(home, ".acta", "config.yaml"), "chat_language: German\nstyle: adhd\n")

	v, _, err := Resolve()
	if err != nil || v.ChatLanguage != "German" {
		t.Fatalf("got %+v %v, want the config.yaml values", v, err)
	}
	got, err := os.ReadFile(oldFile)
	if err != nil || string(got) != "chat_language: Korean\nstyle: plain\n" {
		t.Fatalf("voice.yaml must stay as it was: %q %v", got, err)
	}
}

func TestResolveDoesNotMoveWhenEnvIsSet(t *testing.T) {
	home := withHome(t)
	oldFile := filepath.Join(home, ".acta", "voice.yaml")
	writeVoice(t, oldFile, "chat_language: Korean\nstyle: plain\n")
	t.Setenv("PM_VOICE_FILE", filepath.Join(t.TempDir(), "mine.yaml"))

	v, exists, err := Resolve()
	if err != nil || exists || v != Default() {
		t.Fatalf("got %+v %v %v, want defaults from the missing env file", v, exists, err)
	}
	if _, err := os.Stat(oldFile); err != nil {
		t.Fatalf("voice.yaml must stay when PM_VOICE_FILE is set: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "config.yaml")); !os.IsNotExist(err) {
		t.Fatalf("config.yaml must not appear when PM_VOICE_FILE is set: %v", err)
	}
}

func TestResolveWithNoFilesGivesDefaults(t *testing.T) {
	home := withHome(t)
	v, exists, err := Resolve()
	if err != nil || exists || v != Default() {
		t.Fatalf("got %+v %v %v, want defaults", v, exists, err)
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "config.yaml")); !os.IsNotExist(err) {
		t.Fatalf("a read must not make config.yaml: %v", err)
	}
}

func TestResolveReadsVoiceFileWhenMoveFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can rename in a read-only folder")
	}
	home := withHome(t)
	dir := filepath.Join(home, ".acta")
	oldFile := filepath.Join(dir, "voice.yaml")
	writeVoice(t, oldFile, "chat_language: Korean\nstyle: plain\n")
	// A folder we cannot write to makes the rename fail, but the file can
	// still be read.
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })

	v, exists, err := Resolve()
	if err != nil || !exists || v.ChatLanguage != "Korean" {
		t.Fatalf("got %+v %v %v, want the voice.yaml values", v, exists, err)
	}
	if _, err := os.Stat(oldFile); err != nil {
		t.Fatalf("voice.yaml must stay after a failed move: %v", err)
	}
}

func TestResolveTwoReadersRacingBothSeeTheSetting(t *testing.T) {
	home := withHome(t)
	writeVoice(t, filepath.Join(home, ".acta", "voice.yaml"), "chat_language: Korean\nstyle: plain\n")

	const readers = 8
	var wg sync.WaitGroup
	results := make([]Voice, readers)
	errs := make([]error, readers)
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], _, errs[i] = Resolve()
		}(i)
	}
	wg.Wait()
	for i := range results {
		if errs[i] != nil || results[i].ChatLanguage != "Korean" {
			t.Fatalf("reader %d got %+v %v, want the voice.yaml values", i, results[i], errs[i])
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "voice.yaml")); !os.IsNotExist(err) {
		t.Fatalf("voice.yaml must be gone: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "config.yaml")); err != nil {
		t.Fatalf("config.yaml must exist: %v", err)
	}
}
