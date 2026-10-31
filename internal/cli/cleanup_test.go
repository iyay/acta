package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cleanupRepo is debtRepo plus one raw idea, so acta cleanup has a candidate.
func cleanupRepo(t *testing.T) string {
	t.Helper()
	dir := debtRepo(t)
	idea := filepath.Join(dir, ".pm", "scratch", "2026-09-27-idea.md")
	if err := os.MkdirAll(filepath.Dir(idea), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(idea, []byte("---\nstatus: raw\n---\n# Idea\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "add", ".")
	gitOut(t, dir, "commit", "-q", "-m", "idea")
	return dir
}

func TestCleanupListsCandidatesAndWritesNothing(t *testing.T) {
	dir := cleanupRepo(t)
	head := gitOut(t, dir, "rev-parse", "HEAD")
	inDir(t, dir, func() {
		out := mustRun(t, "cleanup")
		if !strings.Contains(out, "scratch/2026-09-27-idea") || !strings.Contains(out, "raw idea with no spec") {
			t.Fatalf("text output lacks the idea:\n%s", out)
		}
		lines := strings.Split(strings.TrimSpace(mustRun(t, "cleanup", "--json")), "\n")
		var c map[string]any
		if err := json.Unmarshal([]byte(lines[0]), &c); err != nil {
			t.Fatalf("first --json line %q: %v", lines[0], err)
		}
		for _, k := range []string{"id", "kind", "title", "path", "status", "age_days", "why"} {
			if _, ok := c[k]; !ok {
				t.Errorf("--json line lacks %q: %s", k, lines[0])
			}
		}
	})
	if got := gitOut(t, dir, "rev-parse", "HEAD"); got != head {
		t.Fatalf("cleanup made a commit")
	}
	if got := gitOut(t, dir, "status", "--porcelain"); got != "" {
		t.Fatalf("cleanup changed files: %s", got)
	}
}

func TestCleanupHelp(t *testing.T) {
	if out := mustRun(t, "help", "cleanup"); !strings.Contains(out, "acta cleanup [--json]") {
		t.Fatalf("help cleanup:\n%s", out)
	}
}

func TestCleanupFixWritesOneCommit(t *testing.T) {
	dir := cleanupRepo(t)
	plan := filepath.Join(dir, ".pm", "plans", "2026-09-28-ticked.md")
	if err := os.WriteFile(plan, []byte("---\nid: PLAN-4\nstatus: draft\n---\n# Ticked\n\n### Task 1: A\n- [x] a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "add", ".")
	gitOut(t, dir, "commit", "-q", "-m", "plan")
	head := gitOut(t, dir, "rev-parse", "HEAD")
	inDir(t, dir, func() {
		if out := mustRun(t, "cleanup"); !strings.Contains(out, "plans/2026-09-28-ticked") || !strings.Contains(out, "every task is ticked") {
			t.Fatalf("cleanup does not show the fix:\n%s", out)
		}
		if gitOut(t, dir, "rev-parse", "HEAD") != head {
			t.Fatal("plain cleanup wrote a commit")
		}
		mustRun(t, "cleanup", "--fix")
		if got := gitOut(t, dir, "rev-list", "--count", head+"..HEAD"); got != "1" {
			t.Fatalf("--fix made %s commits, want 1", got)
		}
		if out := mustRun(t, "cleanup", "--fix"); !strings.Contains(out, "nothing to fix") {
			t.Fatalf("second --fix:\n%s", out)
		}
	})
}

func TestCleanupApplyReadsStdin(t *testing.T) {
	dir := cleanupRepo(t)
	head := gitOut(t, dir, "rev-parse", "HEAD")
	inDir(t, dir, func() {
		var stdout, stderr strings.Builder
		bad := `{"id":"scratch/2026-09-27-idea","status":"fixed","reason":"x"}` + "\n"
		if code := Run([]string{"cleanup", "apply"}, strings.NewReader(bad), false, &stdout, &stderr); code != exitBadInput || !strings.Contains(stderr.String(), "line 1") {
			t.Fatalf("bad batch: exit %d stderr %q", code, stderr.String())
		}
		good := `{"id":"scratch/2026-09-27-idea","status":"dropped","reason":"not valid"}` + "\n"
		if code := Run([]string{"cleanup", "apply"}, strings.NewReader(good), false, &stdout, &stderr); code != exitOK {
			t.Fatalf("good batch: exit %d stderr %q", code, stderr.String())
		}
	})
	if got := gitOut(t, dir, "rev-list", "--count", head+"..HEAD"); got != "1" {
		t.Fatalf("apply made %s commits, want 1", got)
	}
	body, _ := os.ReadFile(filepath.Join(dir, ".pm", "scratch", "2026-09-27-idea.md"))
	if !strings.Contains(string(body), "status: dropped\n") || !strings.Contains(string(body), "reason: not valid\n") {
		t.Fatalf("idea file = %q", body)
	}
}
