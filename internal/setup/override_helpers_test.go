package setup_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// repoWithActaYAML makes a temp git repo holding the given .acta.yaml, or
// none when body is empty. A real repo lets tests prove nothing is committed.
func repoWithActaYAML(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	if body != "" {
		if err := os.WriteFile(filepath.Join(root, ".acta.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// commitCount says how many commits the repo has. A fresh repo has none, and
// the wizard must leave it that way.
func commitCount(t *testing.T, root string) int {
	t.Helper()
	out, err := exec.Command("git", "-C", root, "rev-list", "--all", "--count").Output()
	if err != nil {
		t.Fatalf("git rev-list: %v", err)
	}
	if string(out) == "0\n" {
		return 0
	}
	return 1
}
