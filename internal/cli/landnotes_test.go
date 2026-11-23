package cli

import (
	"strings"
	"testing"
)

// landNotesRepo is main right after a plan merged with --no-ff. The plan
// closes a bug and a debt file and names spec s.
func landNotesRepo(t *testing.T, closes string) string {
	t.Helper()
	dir := readyRepo(t, map[string]string{
		"specs/2026-10-01-s.md": "---\nid: SPC-0009\n---\n# S\n",
		"bugs/2026-10-01-b.md":  "---\nid: BUG-0001\n---\n# B\n",
		"debt/2026-10-01-d.md":  "---\nid: DBT-0001\n---\n# D\n\n- [ ] one\n",
	})
	tidyGit(t, dir, "checkout", "-q", "-b", "build")
	writeReady(t, dir, "plans/2026-10-01-a.md", "---\nid: PLN-0001\nhash: abcd123\nparent: specs/2026-10-01-s\ncloses: ["+closes+"]\nstarted: \"2026-10-01 10:00:00\"\n---\n# Plan\n\n### Task 1: Do it\n- [x] a\n")
	tidyGit(t, dir, "add", "-A")
	tidyGit(t, dir, "commit", "-q", "-m", "feat: plan")
	tidyGit(t, dir, "checkout", "-q", "main")
	tidyGit(t, dir, "merge", "-q", "--no-ff", "build", "-m", "land")
	return dir
}

// land-notes puts every note LandNotes writes in one commit of its own on
// top of the merge, and an error from LandNotes commits nothing.
func TestLandNotesMakesOneCommit(t *testing.T) {
	dir := landNotesRepo(t, "BUG-0001, DBT-0001")
	// The repo sets its own HOME, so the telemetry home comes after it.
	s, _ := runEnv(t, false)
	merge := tidyGit(t, dir, "rev-parse", "HEAD")
	mustCLI(t, dir, "land-notes", "abcd123")
	// Full mode runs land-notes once per land, so it counts the landed plan.
	if got := todayCounts(t, s).Workflow; got["plan_landed"] != 1 || len(got) != 1 {
		t.Fatalf("workflow %v, want plan_landed 1", got)
	}
	if got := tidyGit(t, dir, "rev-parse", "HEAD^"); got != merge {
		t.Fatalf("HEAD^ = %s, want the merge %s: not exactly one commit", got, merge)
	}
	if got := tidyGit(t, dir, "log", "-1", "--format=%s"); got != "chore(acta): land of PLN-abcd123" {
		t.Fatalf("subject %q", got)
	}
	want := ".acta/bugs/2026-10-01-b.md\n.acta/debt/2026-10-01-d.md\n.acta/specs/2026-10-01-s.md"
	if got := tidyGit(t, dir, "diff", "--name-only", "HEAD^", "HEAD"); got != want {
		t.Fatalf("commit holds %q, want %q", got, want)
	}
	if st := tidyGit(t, dir, "status", "--porcelain", "--untracked-files=no"); st != "" {
		t.Fatalf("checkout left dirty:\n%s", st)
	}

	bad := landNotesRepo(t, "BUG-0001, DBT-0009")
	s, _ = runEnv(t, false)
	head := tidyGit(t, bad, "rev-parse", "HEAD")
	var stdout, stderr strings.Builder
	var code int
	inDir(t, bad, func() {
		code = Run([]string{"land-notes", "abcd123"}, strings.NewReader(""), false, &stdout, &stderr)
	})
	if code == exitOK || !strings.Contains(stderr.String(), "DBT-0009") {
		t.Fatalf("exit %d, stderr %q: want a failure naming DBT-0009", code, stderr.String())
	}
	if tidyGit(t, bad, "rev-parse", "HEAD") != head {
		t.Fatal("a failed land-notes made a commit")
	}
	if got := todayCounts(t, s).Workflow; len(got) != 0 {
		t.Fatalf("a failed land-notes counted %v", got)
	}
	if st := tidyGit(t, bad, "status", "--porcelain", "--untracked-files=no"); st != "" {
		t.Fatalf("a failed land-notes changed files:\n%s", st)
	}
}
