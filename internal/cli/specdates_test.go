package cli

import (
	"strings"
	"testing"
)

// spec-dates commits what it writes, so the parent is clean right after and
// the next land never meets a dirty checkout. A one-plan spec gets nothing.
func TestSpecDatesCommitsAndLeavesNoChange(t *testing.T) {
	split := func(id, started, box string) string {
		return "---\nid: " + id + "\nparent: specs/2026-10-01-s\nstarted: \"" + started + "\"\n---\n# Plan " + id + "\n\n### Task 1: Do it\n- [" + box + "] a\n"
	}
	dir := readyRepo(t, map[string]string{
		"specs/2026-10-01-s.md": "---\nid: SPC-0009\n---\n# S\n",
		"plans/2026-10-01-a.md": split("PLN-0001", "2026-10-01 10:00:00", "x"),
		"plans/2026-10-01-b.md": split("PLN-0002", "2026-10-01 11:00:00", "x"),
	})
	before := tidyGit(t, dir, "rev-parse", "HEAD")
	var stdout, stderr strings.Builder
	var code int
	inDir(t, dir, func() {
		code = Run([]string{"spec-dates", "PLN-0002"}, strings.NewReader(""), false, &stdout, &stderr)
	})
	if code != exitOK {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if tidyGit(t, dir, "rev-parse", "HEAD") == before {
		t.Error("no commit made")
	}
	if st := tidyGit(t, dir, "status", "--porcelain", "--untracked-files=no"); st != "" {
		t.Errorf("checkout has tracked changes after spec-dates:\n%s", st)
	}
	spec := tidyGit(t, dir, "show", "HEAD:.acta/specs/2026-10-01-s.md")
	if !strings.Contains(spec, "started: \"2026-10-01 10:00:00\"") || !strings.Contains(spec, "finished:") {
		t.Errorf("committed spec lacks the dates:\n%s", spec)
	}
}
