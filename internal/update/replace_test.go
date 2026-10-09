package update

import (
	"os"
	"path/filepath"
	"testing"
)

// leftovers lists any temp file Replace may have forgotten in dir.
func leftovers(t *testing.T, dir string) []string {
	t.Helper()
	found, err := filepath.Glob(filepath.Join(dir, ".acta.*"))
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func TestReplaceWritesContentAndMode(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "acta")
	if err := os.WriteFile(exe, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Replace(exe, []byte("new binary")); err != nil {
		t.Fatalf("Replace: %v", err)
	}

	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new binary" {
		t.Fatalf("content = %q, want %q", got, "new binary")
	}
	info, err := os.Stat(exe)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("mode = %v, want 0755", info.Mode().Perm())
	}
	if left := leftovers(t, dir); len(left) != 0 {
		t.Fatalf("temp files left behind: %v", left)
	}
}

func TestReplaceFollowsSymlink(t *testing.T) {
	realDir := t.TempDir()
	linkDir := t.TempDir()
	target := filepath.Join(realDir, "acta-real")
	link := filepath.Join(linkDir, "acta")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks not available: %v", err)
	}

	if err := Replace(link, []byte("fresh")); err != nil {
		t.Fatalf("Replace: %v", err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("the link was replaced by a plain file")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "fresh" {
		t.Fatalf("target content = %q, want %q", got, "fresh")
	}
	if left := leftovers(t, realDir); len(left) != 0 {
		t.Fatalf("temp files in target dir: %v", left)
	}
	if left := leftovers(t, linkDir); len(left) != 0 {
		t.Fatalf("temp files in link dir: %v", left)
	}
}

func TestReplaceReadOnlyDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	dir := filepath.Join(t.TempDir(), "bin")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(dir, "acta")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	// EvalSymlinks may turn the temp path into a different spelling (macOS
	// /var vs /private/var), so the expected text uses the resolved dir.
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = Replace(exe, []byte("new"))
	if err == nil {
		t.Fatal("want an error for a read-only dir")
	}
	if want := "no write access to " + resolved; err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}

	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "old" {
		t.Fatalf("old file changed: %q", got)
	}
	if left := leftovers(t, dir); len(left) != 0 {
		t.Fatalf("temp files left behind: %v", left)
	}
}

func TestReplaceMissingExe(t *testing.T) {
	dir := t.TempDir()
	if err := Replace(filepath.Join(dir, "nope"), []byte("x")); err == nil {
		t.Fatal("want an error when the binary does not exist")
	}
	if left := leftovers(t, dir); len(left) != 0 {
		t.Fatalf("temp files left behind: %v", left)
	}
}
