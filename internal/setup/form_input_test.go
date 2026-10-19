package setup_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/iyay/acta/internal/setup"
)

// The forms must read the terminal by its real name. Only a file named
// /dev/tty lets an old form stop reading, so the next form gets the key.
func TestTTYPathIsDevTTY(t *testing.T) {
	if setup.TTYPath != "/dev/tty" {
		t.Fatalf("TTYPath = %q, want /dev/tty", setup.TTYPath)
	}
}

func TestFormInputUsesOpenedTTY(t *testing.T) {
	want, err := os.Create(filepath.Join(t.TempDir(), "tty"))
	if err != nil {
		t.Fatal(err)
	}
	defer want.Close()
	opened := ""
	got := setup.FormInput(true, func(name string) (*os.File, error) {
		opened = name
		return want, nil
	})
	if got != want || opened != setup.TTYPath {
		t.Fatalf("got %v from %q, want the file opened from %q", got, opened, setup.TTYPath)
	}
}

func TestFormInputKeepsDefaultWhenOpenFails(t *testing.T) {
	got := setup.FormInput(true, func(string) (*os.File, error) {
		return nil, errors.New("no tty")
	})
	if got != nil {
		t.Fatalf("got %v, want nil so the form keeps its default input", got)
	}
}

func TestFormInputSkipsTTYWhenStdinIsNotTerminal(t *testing.T) {
	got := setup.FormInput(false, func(string) (*os.File, error) {
		t.Fatal("must not open the tty when stdin is not a terminal")
		return nil, nil
	})
	if got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

// Every screen, the override ones too, must read the same opened tty.
// One screen left on stdin brings the double Enter back.
func TestAskGivesEveryFormTheTTY(t *testing.T) {
	want, err := os.Create(filepath.Join(t.TempDir(), "tty"))
	if err != nil {
		t.Fatal(err)
	}
	defer want.Close()
	root := repoWithActaYAML(t, "build_executor: dispatch\n")
	e := setup.Env{TTY: true, RepoRoot: root, Current: overrideUser()}
	got, err := setup.AskRecordingInputs(e, want)
	if err != nil {
		t.Fatal(err)
	}
	questions := len(setup.FormGroups(e))
	if len(got) <= questions {
		t.Fatalf("ran %d forms, want the %d questions plus an override screen", len(got), questions)
	}
	for i, in := range got {
		if in != want {
			t.Errorf("form %d got input %v, want the opened tty", i, in)
		}
	}
}
