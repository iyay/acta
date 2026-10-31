package write

import (
	"errors"
	"strings"
	"testing"
)

// Every kind that can be closed by cleanup keeps a reason, and the board reads
// it back from the file.
func TestSetValueReasonRoundTrips(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		".acta/scratch/2026-09-24-idea.md": "---\nstatus: raw\n---\n# Idea\n",
		".acta/bugs/2026-09-24-crash.md":   "# Crash\n\n## Symptom\nx\n",
		".acta/specs/2026-09-20-alpha.md":  "# Alpha\n",
		".acta/plans/2026-09-25-fix.md":    "---\nparent: specs/2026-09-20-alpha\n---\n# Fix\n\n### Task 1: Fix\n- [ ] a\n",
	})
	ids := []string{"scratch/2026-09-24-idea", "bugs/2026-09-24-crash", "specs/2026-09-20-alpha", "plans/2026-09-25-fix"}
	for _, id := range ids {
		if _, err := SetValue(cfg, mustLoad(t, cfg), id, "reason", "built in abc1234"); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
	}
	b := mustLoad(t, cfg)
	for _, id := range ids {
		if got := b.Get(id).Reason; got != "built in abc1234" {
			t.Errorf("%s reason = %q", id, got)
		}
	}
}

func TestSetValueReasonBadInput(t *testing.T) {
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/scratch/2026-09-24-idea.md": "# Idea\n",
		".acta/plans/2026-09-25-fix.md":    "# Fix\n\n### Task 1: Fix\n- [ ] a\n",
		".acta/debt/2026-09-26-notes.md":   "---\nparent: plans/2026-09-25-fix\n---\n# Notes\n\n- [ ] (medium) a\n",
	})
	b := mustLoad(t, cfg)
	cases := []struct{ name, id, value string }{
		{"empty", "scratch/2026-09-24-idea", ""},
		{"blank", "scratch/2026-09-24-idea", "   "},
		{"newline", "scratch/2026-09-24-idea", "a\nb"},
		{"task", "plans/2026-09-25-fix#task-1", "x"},
		{"debt line", "debt/2026-09-26-notes#item-1", "x"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := SetValue(cfg, b, c.id, "reason", c.value); !errors.Is(err, ErrBadInput) {
				t.Fatalf("err = %v, want ErrBadInput", err)
			}
		})
	}
	if out := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); out != "" {
		t.Fatalf("bad input changed files: %s", out)
	}
}

// done is a choice the user writes, so it is allowed; specced still comes only
// from a spec link.
func TestSetValueScratchDone(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{".acta/scratch/2026-09-24-idea.md": "# Idea\n"})
	if _, err := SetValue(cfg, mustLoad(t, cfg), "scratch/2026-09-24-idea", "status", "done"); err != nil {
		t.Fatal(err)
	}
	if got := mustLoad(t, cfg).Get("scratch/2026-09-24-idea").Status; got != "done" {
		t.Fatalf("status = %q, want done", got)
	}
	_, err := SetValue(cfg, mustLoad(t, cfg), "scratch/2026-09-24-idea", "status", "specced")
	if !errors.Is(err, ErrBadInput) || !strings.Contains(err.Error(), "spec") {
		t.Fatalf("specced err = %v", err)
	}
}
