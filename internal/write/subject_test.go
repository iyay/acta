package write

import (
	"testing"
)

// A scratch write must commit under the chore form with the scratch scope,
// so the log reads as a conventional commit and never names the tool.
func TestNewScratchCommitsChoreScope(t *testing.T) {
	// No Parallel here because fixNow swaps the shared clock.

	fixNow(t)
	cfg := repoWith(t, baseFiles)
	if _, err := NewScratch(cfg, "newest-first", "", []byte("catet aja dulu\n")); err != nil {
		t.Fatal(err)
	}
	want := "chore(scratch): new 2026-09-26-newest-first"
	if msg := gitRun(t, cfg.RepoRoot, "log", "-1", "--format=%s"); msg != want {
		t.Errorf("commit subject %q, want %q", msg, want)
	}
}

// A commit over files of more than one kind carries no scope, so a mixed
// run never claims to be one kind.
func TestSubjectDropsScopeForMixedKinds(t *testing.T) {
	// No Parallel here because fixNow swaps the shared clock.

	fixNow(t)
	cfg := repoWith(t, baseFiles)
	got := Subject(cfg, []string{
		cfg.Root + "/scratch/2026-09-26-idea.md",
		cfg.Root + "/plans/2026-09-25-crash-fix.md",
	}, "assign short ids")
	if want := "chore: assign short ids"; got != want {
		t.Errorf("subject %q, want %q", got, want)
	}
}
