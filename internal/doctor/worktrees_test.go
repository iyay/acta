package doctor

import (
	"strings"
	"testing"
)

func TestWorktreesCheck(t *testing.T) {
	ok := env(t)
	wantLevel(t, byName(Run(ok), "worktrees"), OK, "")

	warn := env(t)
	warn.StrayWorktrees = []string{"/x/repo-headchk", "/x/repo-premerge"}
	r := byName(Run(warn), "worktrees")
	wantLevel(t, r, Warn, "acta worktrees --clean")
	if !strings.HasPrefix(r.Msg, "2 stray") {
		t.Fatalf("msg %q does not give the count", r.Msg)
	}
	if Failed(Run(warn)) != Failed(Run(ok)) {
		t.Fatal("strays failed a run that was not failing before")
	}

	one := env(t)
	one.StrayWorktrees = []string{"/x/repo-headchk"}
	if got := byName(Run(one), "worktrees").Msg; !strings.HasPrefix(got, "1 stray worktree: ") {
		t.Fatalf("msg %q", got)
	}

	outside := env(t)
	outside.RepoRoot = ""
	for _, r := range Run(outside) {
		if r.Name == "worktrees" {
			t.Fatalf("outside git got %+v", r)
		}
	}
	if !strings.Contains(r.Msg, "repo-headchk") {
		t.Fatalf("msg %q does not name the strays", r.Msg)
	}
}
