package doctor

import (
	"strings"
	"testing"
)

// The worktrees line never offers a fix that cannot fix what it counts:
// acta worktrees --clean removes clean and missing strays, never dirty ones.
func TestWorktreesCheck(t *testing.T) {
	ok := env(t)
	wantLevel(t, byName(Run(ok), "worktrees"), OK, "")

	clean := env(t)
	clean.CleanStrays = []string{"/x/repo-headchk", "/x/repo-premerge"}
	r := byName(Run(clean), "worktrees")
	wantLevel(t, r, Warn, "acta worktrees --clean")
	if !strings.HasPrefix(r.Msg, "2 stray worktrees --clean can remove: ") || !strings.Contains(r.Msg, "repo-headchk") {
		t.Fatalf("clean msg %q", r.Msg)
	}
	if Failed(Run(clean)) != Failed(Run(ok)) {
		t.Fatal("strays failed a run that was not failing before")
	}

	one := env(t)
	one.CleanStrays = []string{"/x/repo-headchk"}
	if got := byName(Run(one), "worktrees").Msg; !strings.HasPrefix(got, "1 stray worktree --clean can remove: ") {
		t.Fatalf("one msg %q", got)
	}

	dirty := env(t)
	dirty.DirtyStrays = []string{"/x/repo-dirty"}
	r = byName(Run(dirty), "worktrees")
	wantLevel(t, r, Warn, "")
	if !strings.HasPrefix(r.Msg, "1 dirty stray worktree: ") || strings.Contains(r.Fix, "acta ") || r.Fix == "" {
		t.Fatalf("dirty result %+v", r)
	}

	both := env(t)
	both.CleanStrays = []string{"/x/repo-headchk"}
	both.DirtyStrays = []string{"/x/repo-dirty", "/x/repo-dirty2"}
	r = byName(Run(both), "worktrees")
	wantLevel(t, r, Warn, "acta worktrees --clean")
	if !strings.Contains(r.Msg, "1 stray worktree --clean can remove") || !strings.Contains(r.Msg, "2 dirty stray worktrees") || !strings.Contains(r.Fix, "dirty") {
		t.Fatalf("both result %+v", r)
	}

	outside := env(t)
	outside.RepoRoot = ""
	for _, r := range Run(outside) {
		if r.Name == "worktrees" {
			t.Fatalf("outside git got %+v", r)
		}
	}
}
