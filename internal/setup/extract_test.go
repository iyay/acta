package setup

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/iyay/acta/plugin"
)

func TestExtractFresh(t *testing.T) {
	home := t.TempDir()
	dir, err := ExtractPlugin(plugin.Files, home)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".acta", "plugin"); dir != want {
		t.Fatalf("dir = %q, want %q", dir, want)
	}
	// Every embedded file lands byte for byte, and mode bits match disk.
	err = fs.WalkDir(plugin.Files, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		want, _ := plugin.Files.ReadFile(p)
		got, rerr := os.ReadFile(filepath.Join(dir, p))
		if rerr != nil || string(got) != string(want) {
			t.Errorf("%s not extracted byte for byte: %v", p, rerr)
		}
		st, _ := os.Stat(filepath.Join("..", "..", "plugin", p))
		out, _ := os.Stat(filepath.Join(dir, p))
		if st != nil && out != nil && (st.Mode()&0o111 != 0) != (out.Mode()&0o111 != 0) {
			t.Errorf("%s: disk mode %v, extracted mode %v", p, st.Mode(), out.Mode())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(filepath.Join(dir, "hooks", "session-start")); err != nil || st.Mode()&0o111 == 0 {
		t.Fatalf("hook script not executable: %v %v", st, err)
	}
}

func TestExtractReplacesOldTree(t *testing.T) {
	home := t.TempDir()
	old := filepath.Join(home, ".acta", "plugin")
	if err := os.MkdirAll(filepath.Join(old, "gone"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "gone", "stale.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir, err := ExtractPlugin(plugin.Files, home)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "gone", "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("stale file survived: %v", err)
	}
	// No temp folder is left behind in ~/.acta.
	entries, _ := os.ReadDir(filepath.Join(home, ".acta"))
	if len(entries) != 1 {
		t.Fatalf("~/.acta holds %v, want only plugin", entries)
	}
	// Nothing outside ~/.acta was written.
	top, _ := os.ReadDir(home)
	if len(top) != 1 {
		t.Fatalf("home holds %v, want only .acta", top)
	}
}

func TestExtractFailureKeepsOldTree(t *testing.T) {
	home := t.TempDir()
	old := filepath.Join(home, ".acta", "plugin")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(old, "keep.txt")
	if err := os.WriteFile(keep, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A path that escapes the tree must be refused before any swap.
	bad := fstest.MapFS{"../evil": {Data: []byte("x")}}
	if _, err := ExtractPlugin(bad, home); err == nil {
		t.Fatal("want an error for an unsafe path")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("old tree lost: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "evil")); !os.IsNotExist(err) {
		t.Fatal("wrote outside the plugin folder")
	}
}

func TestExtractHomeBlocked(t *testing.T) {
	home := t.TempDir()
	// ~/.acta is a file, so no folder can be made.
	if err := os.WriteFile(filepath.Join(home, ".acta"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractPlugin(plugin.Files, home); err == nil {
		t.Fatal("want an error")
	}
	if _, err := ExtractPlugin(plugin.Files, ""); err == nil {
		t.Fatal("want an error for an empty home")
	}
}
