package gitc

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseWorktrees(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	out := "worktree " + a + "\nHEAD 1111\nbranch refs/heads/main\n\n" +
		"worktree " + b + "\nHEAD 2222\ndetached\n\n" +
		"worktree /nowhere/at/all\nHEAD 3333\nbranch refs/heads/gone\n\n" +
		"worktree " + a + "/bare\nbare\n\n" +
		"worktree " + b + "/p\nHEAD 4444\nbranch refs/heads/p\nprunable gitdir file points to non-existent location\n"
	got := parseWorktrees(out)
	want := []Worktree{{Path: a, Branch: "main"}, {Path: b, Branch: "(detached)"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestWorktreesAndCommonDir(t *testing.T) {
	repo := setupRepo(t)
	wt := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-feat")
	git(t, repo, "worktree", "add", "-q", wt, "-b", "feat")
	t.Cleanup(func() { os.RemoveAll(wt) })
	got, err := Worktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Branch != "main" || got[1].Branch != "feat" {
		t.Fatalf("got %+v", got)
	}
	dir, err := CommonDir(wt)
	if err != nil || dir != filepath.Join(repo, ".git") {
		t.Fatalf("CommonDir = %q %v", dir, err)
	}
	if _, err := Worktrees(t.TempDir()); err == nil {
		t.Fatal("not a repo should be an error")
	}
}

func TestUnmergedBranchesAndBranchFiles(t *testing.T) {
	repo := setupRepo(t)
	git(t, repo, "checkout", "-q", "-b", "feat")
	if err := os.MkdirAll(filepath.Join(repo, ".pm", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# Branch story\nwith a second line\n"
	writeFile(t, filepath.Join(repo, ".pm", "specs", "2026-09-26-x.md"), body)
	writeFile(t, filepath.Join(repo, ".pm", "specs", "2026-09-27-empty.md"), "")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-q", "-m", "spec on feat")
	git(t, repo, "checkout", "-q", "main")
	git(t, repo, "branch", "merged-one")

	got, err := UnmergedBranches(repo)
	if err != nil || !reflect.DeepEqual(got, []string{"feat"}) {
		t.Fatalf("unmerged = %v %v", got, err)
	}
	files, err := BranchFiles(repo, "feat", ".pm")
	if err != nil {
		t.Fatal(err)
	}
	if string(files[".pm/specs/2026-09-26-x.md"]) != body || len(files) != 2 || len(files[".pm/specs/2026-09-27-empty.md"]) != 0 {
		t.Fatalf("files = %q", files)
	}
	if f, err := BranchFiles(repo, "main", ".pm"); err != nil || len(f) != 0 {
		t.Fatalf("branch with no .pm: %v %v", f, err)
	}
	if _, err := BranchFiles(repo, "nope", ".pm"); err == nil {
		t.Fatal("unknown branch should be an error")
	}
}
