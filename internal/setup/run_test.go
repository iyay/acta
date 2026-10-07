package setup_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/setup"
)

// tempHome points HOME, PM_VOICE_FILE and TMPDIR at fresh temp dirs, so no
// test touches the real home. It returns the voice file path.
func tempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	voice := filepath.Join(home, "config.yaml")
	t.Setenv("PM_VOICE_FILE", voice)
	t.Setenv("TMPDIR", home)
	return voice
}

type fakeRunner struct {
	calls [][]string
	fail  map[string]bool
}

func (f *fakeRunner) Run(argv []string) error {
	f.calls = append(f.calls, argv)
	if f.fail[strings.Join(argv, " ")] {
		return errors.New("exit 1")
	}
	return nil
}

func TestRunActions(t *testing.T) {
	u := config.User{ChatLanguage: "Korean", Style: "adhd", RepoLanguage: "English"}
	dir := t.TempDir()
	blockPath := filepath.Join(dir, "CLAUDE.md")

	t.Run("config install and block all run", func(t *testing.T) {
		voice := tempHome(t)
		f := &fakeRunner{}
		var out strings.Builder
		actions := []setup.Action{
			{Kind: "config", User: u},
			{Kind: "install", Harness: "claude", Argv: [][]string{{"claude", "plugin", "marketplace", "add", "/p"}}},
			{Kind: "block", Path: blockPath},
			{Kind: "print", Harness: "omp", Path: "omp plugin link <plugin dir>"},
		}
		if err := setup.Apply(actions, f, &out); err != nil {
			t.Fatal(err)
		}
		if len(f.calls) != 1 {
			t.Fatalf("runner calls = %v, want one install", f.calls)
		}
		if got, _, err := config.LoadUser(voice); err != nil || got.ChatLanguage != "Korean" {
			t.Fatalf("saved user = %+v err %v, want Korean", got, err)
		}
		raw, err := os.ReadFile(blockPath)
		if err != nil || string(raw) != setup.Block {
			t.Fatalf("block file = %q err %v, want the block only", raw, err)
		}
		if !strings.Contains(out.String(), "omp plugin link <plugin dir>") {
			t.Fatalf("output %q misses the print text", out.String())
		}
		if !strings.Contains(out.String(), "✓ claude\n") {
			t.Fatalf("output %q misses the short install line", out.String())
		}
		if strings.Contains(out.String(), "Ran `") || strings.Contains(out.String(), "Could not run") {
			t.Fatalf("output %q uses the long install lines", out.String())
		}
		if !strings.Contains(out.String(), "Saved") || !strings.Contains(out.String(), blockPath) {
			t.Fatalf("output %q misses the summary", out.String())
		}
	})

	t.Run("a failed install prints its command and carries on", func(t *testing.T) {
		f := &fakeRunner{fail: map[string]bool{"omp plugin link /p": true}}
		var out strings.Builder
		actions := []setup.Action{
			{Kind: "install", Harness: "omp", Argv: [][]string{{"omp", "plugin", "link", "/p"}}},
			{Kind: "block", Path: filepath.Join(dir, "second.md")},
		}
		if err := setup.Apply(actions, f, &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "│  ▲ omp: omp plugin link /p\n") {
			t.Fatalf("output %q misses the short failed install line", out.String())
		}
		if strings.Contains(out.String(), "Could not run") {
			t.Fatalf("output %q uses the long failure line", out.String())
		}
		if _, err := os.Stat(filepath.Join(dir, "second.md")); err != nil {
			t.Fatalf("block missing after failed install: %v", err)
		}
	})

	t.Run("bad user value stops before any write", func(t *testing.T) {
		voice := tempHome(t)
		bad := u
		bad.Style = "fancy"
		f := &fakeRunner{}
		var out strings.Builder
		err := setup.Apply([]setup.Action{{Kind: "config", User: bad}}, f, &out)
		if err == nil || !errors.Is(err, config.ErrBadUser) {
			t.Fatalf("Apply err = %v, want ErrBadUser", err)
		}
		if _, err := os.Stat(voice); !os.IsNotExist(err) {
			t.Fatalf("voice file written despite bad value")
		}
	})

	t.Run("unknown action kind fails", func(t *testing.T) {
		f := &fakeRunner{}
		var out strings.Builder
		if err := setup.Apply([]setup.Action{{Kind: "dance"}}, f, &out); err == nil {
			t.Fatal("Apply with unknown kind succeeded, want an error")
		}
	})
}

func TestFormDefaultsFromCurrent(t *testing.T) {
	cur := config.User{ChatLanguage: "Korean", Style: "plain", Tone: "short"}
	fields := setup.FormDefaults(setup.Env{Current: cur})
	if fields.Language != "Korean" || fields.Style != "plain" || fields.Tone != "short" {
		t.Fatalf("defaults = %+v, want the current values", fields)
	}
	empty := setup.FormDefaults(setup.Env{})
	if empty.Language != "English" || empty.Style != "adhd" {
		t.Fatalf("empty defaults = %+v, want the built-in ones", empty)
	}
}
