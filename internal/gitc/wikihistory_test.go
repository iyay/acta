package gitc

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// historyRepo builds one repo that holds every history shape the old per-page
// queries handle, so the one-pass reader can be compared with them.
func historyRepo(t *testing.T) string {
	t.Helper()
	repo := setupRepo(t)
	for _, dir := range []string{"wiki", "src", "lib"} {
		if err := os.MkdirAll(filepath.Join(repo, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	const checked = "wiki/.checked"
	commitFileAt(t, repo, "src/x.go", "one\n", "", "change src")
	commitFileAt(t, repo, "wiki/p.md", "p\n", "", "add page p")
	commitFileAt(t, repo, checked, "z 2026-01-03T00:00:00Z\np 2026-01-01T00:00:00Z\n", "", "check z and p")
	// A re-sort only moves the lines, so it is no check.
	commitFileAt(t, repo, checked, "p 2026-01-01T00:00:00Z\nz 2026-01-03T00:00:00Z\n", "", "sorted rewrite")
	// Removing a line is no check, and a longer id is no check of a shorter one.
	commitFileAt(t, repo, checked, "p-two 2026-01-04T00:00:00Z\nz 2026-01-03T00:00:00Z\n", "", "drop p, check p-two")
	commitFileAt(t, repo, "lib/y.go", "one\n", "", "change lib")
	// A side branch with a change and a check, merged back.
	git(t, repo, "checkout", "-q", "-b", "side")
	commitFileAt(t, repo, "src/x.go", "side\n", "2020-01-01T00:00:00Z", "change src on side")
	commitFileAt(t, repo, "wiki/q.md", "q\n", "", "add page q on side")
	git(t, repo, "checkout", "-q", "main")
	commitFileAt(t, repo, checked, "p-two 2026-01-04T00:00:00Z\nq 2026-01-05T00:00:00Z\nz 2026-01-03T00:00:00Z\n", "2025-01-01T00:00:00Z", "check q on main")
	mergeSide(t, repo, "side")
	// A rename takes a file out of src, which still counts as a src change.
	git(t, repo, "mv", "src/x.go", "lib/x.go")
	git(t, repo, "commit", "-q", "-m", "move x to lib")
	// A branch that never merged is not in the history of HEAD.
	git(t, repo, "checkout", "-q", "-b", "lost")
	commitFileAt(t, repo, checked, "p 2026-02-01T00:00:00Z\n", "", "check p on lost")
	git(t, repo, "checkout", "-q", "main")
	return repo
}

func TestWikiHistoryMatchesPerPageQueries(t *testing.T) {
	repo := historyRepo(t)
	const checked = "wiki/.checked"
	paths := [][]string{{"src/"}, {"lib/"}, {"lib/y.go"}, {"src/", "lib/"}, {"nothing/"}}
	h, err := ReadWikiHistory(repo, "wiki/", checked)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"p", "q", "z", "p-two", "none"} {
		page := "wiki/" + id + ".md"
		want, err := CheckCommits(repo, page, checked, id)
		if err != nil {
			t.Fatal(err)
		}
		got := h.CheckCommits(page, id)
		if !slices.Equal(got, want) {
			t.Errorf("checks of %s: got %v, want %v", id, got, want)
		}
		for _, ps := range paths {
			wantSha, wantSubject, err := Unseen(repo, want, ps)
			if err != nil {
				t.Fatal(err)
			}
			gotSha, gotSubject := h.Unseen(got, func(file string) bool {
				for _, p := range ps {
					if file == p || strings.HasSuffix(p, "/") && strings.HasPrefix(file, p) {
						return true
					}
				}
				return false
			})
			if gotSha != wantSha || gotSubject != wantSubject {
				t.Errorf("unseen of %s under %v: got %q %q, want %q %q", id, ps, gotSha, gotSubject, wantSha, wantSubject)
			}
		}
	}
}

func TestWikiHistoryDirty(t *testing.T) {
	repo := historyRepo(t)
	writeFile(t, filepath.Join(repo, "wiki/p.md"), "edited\n")
	writeFile(t, filepath.Join(repo, "wiki/new.md"), "new\n")
	h, err := ReadWikiHistory(repo, "wiki/", "wiki/.checked")
	if err != nil {
		t.Fatal(err)
	}
	for file, want := range map[string]bool{"wiki/p.md": true, "wiki/new.md": true, "wiki/q.md": false} {
		if got := h.Dirty(file); got != want {
			t.Errorf("dirty %s: got %v, want %v", file, got, want)
		}
	}
}

// TestWikiHistoryGitCalls counts git processes with a wrapper first on PATH,
// so one read stays a fixed number of git commands.
func TestWikiHistoryGitCalls(t *testing.T) {
	repo := historyRepo(t)
	count := countGitCalls(t)
	if _, err := ReadWikiHistory(repo, "wiki/", "wiki/.checked"); err != nil {
		t.Fatal(err)
	}
	if n := count(); n > 4 {
		t.Errorf("ran git %d times, want at most 4", n)
	}
}

func TestWikiHistoryNoCheckout(t *testing.T) {
	if _, err := ReadWikiHistory(t.TempDir(), "wiki/", "wiki/.checked"); err == nil {
		t.Error("got no error for a folder that is no checkout")
	}
}

// countGitCalls puts a git wrapper first on PATH that logs each call, and
// gives back a function that says how many calls were logged so far.
func countGitCalls(t *testing.T) func() int {
	t.Helper()
	real, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	logFile := filepath.Join(dir, "calls")
	script := "#!/bin/sh\necho x >> '" + logFile + "'\nexec '" + real + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return func() int {
		b, _ := os.ReadFile(logFile)
		return strings.Count(string(b), "\n")
	}
}
