package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/write"
)

// splitE2EPlan is a one-task plan of spec s, with the hash its task
// trailers name.
func splitE2EPlan(id, hash, after string) string {
	front := "id: " + id + "\nhash: " + hash + "\nparent: specs/2026-10-01-s\n"
	if after != "" {
		front = "id: " + id + "\nhash: " + hash + "\nparent: specs/2026-10-01-t\nafter: [" + after + "]\n"
	}
	return "---\n" + front + "---\n# Plan " + id + "\n\n### Task 1: Do it\n- [ ] Red: it fails\n"
}

// mustCLI runs one acta command in dir and fails the test on a non-zero exit.
func mustCLI(t *testing.T, dir string, args ...string) string {
	t.Helper()
	var stdout, stderr strings.Builder
	var code int
	inDir(t, dir, func() {
		code = Run(args, strings.NewReader(""), false, &stdout, &stderr)
	})
	if code != exitOK {
		t.Fatalf("acta %v in %s: exit %d\n%s%s", args, dir, code, stdout.String(), stderr.String())
	}
	return stdout.String()
}

// buildIn does one plan's whole build in its worktree: start the task at
// the given time, commit code with the task trailer, tick the task done and
// commit the plan file.
func buildIn(t *testing.T, wt, stem, hash string, at time.Time) {
	t.Helper()
	old := write.Now
	write.Now = func() time.Time { return at }
	defer func() { write.Now = old }()
	mustCLI(t, wt, "tick", "plans/"+stem+"#task-1", "--start")
	if err := os.WriteFile(filepath.Join(wt, stem+".txt"), []byte(stem), 0o644); err != nil {
		t.Fatal(err)
	}
	tidyGit(t, wt, "add", stem+".txt")
	tidyGit(t, wt, "commit", "-q", "-m", "feat: "+stem+"\n\nTask: PLN-"+hash+"#1")
	mustCLI(t, wt, "tick", "plans/"+stem+"#task-1", "--all")
	// Land needs a clean worktree, so every planning file a tick touched is
	// committed, the way an agent does it before landing.
	tidyGit(t, wt, "add", "-A", ".acta")
	tidyGit(t, wt, "commit", "-q", "-m", "chore(plan): tick "+stem)
}

// landFrom lands one worktree branch the tidy way, then writes the spec
// dates, and checks the main checkout is clean afterwards.
func landFrom(t *testing.T, main, branch, plan string) {
	t.Helper()
	mustCLI(t, main, "tidy", "main", branch)
	tidyGit(t, main, "merge", "-q", "--ff-only", "refs/acta/tidy/"+branch)
	mustCLI(t, main, "spec-dates", plan)
	if st := tidyGit(t, main, "status", "--porcelain", "--untracked-files=no"); st != "" {
		t.Fatalf("main has tracked changes after landing %s:\n%s", branch, st)
	}
}

// Two plans of one spec built in two real worktrees, started at different
// times, both land through tidy with no clash, in both land orders. The spec
// ends with the earliest started and one finished.
func TestSplitPlansLandFromTwoWorktrees(t *testing.T) {
	early := time.Date(2026, 10, 1, 10, 0, 0, 0, time.Local)
	late := early.Add(time.Hour)
	for _, firstLands := range []string{"a", "b"} {
		t.Run("first lands "+firstLands, func(t *testing.T) {
			main := readyRepo(t, map[string]string{
				"specs/2026-10-01-s.md": "---\nid: SPC-0009\n---\n# S\n",
				"plans/2026-10-01-a.md": splitE2EPlan("PLN-0001", "aaaaaaa", ""),
				"plans/2026-10-01-b.md": splitE2EPlan("PLN-0002", "bbbbbbb", ""),
			})
			wa, wb := filepath.Join(t.TempDir(), "wa"), filepath.Join(t.TempDir(), "wb")
			tidyGit(t, main, "worktree", "add", "-q", wa, "-b", "a", "main")
			tidyGit(t, main, "worktree", "add", "-q", wb, "-b", "b", "main")
			buildIn(t, wa, "2026-10-01-a", "aaaaaaa", early)
			buildIn(t, wb, "2026-10-01-b", "bbbbbbb", late)

			order := [][2]string{{"a", "PLN-0001"}, {"b", "PLN-0002"}}
			if firstLands == "b" {
				order[0], order[1] = order[1], order[0]
			}
			for _, o := range order {
				landFrom(t, main, o[0], o[1])
			}
			spec := tidyGit(t, main, "show", "HEAD:.acta/specs/2026-10-01-s.md")
			if !strings.Contains(spec, "started: \"2026-10-01 10:00:00\"") || strings.Count(spec, "finished:") != 1 {
				t.Errorf("spec after both lands, want started 10:00:00 and one finished:\n%s", spec)
			}
		})
	}
}

// A worktree made before its after plan landed is not ready with --head
// until it takes in the parent.
func TestSplitStaleWorktreeWaits(t *testing.T) {
	main := readyRepo(t, map[string]string{
		"specs/2026-10-01-s.md": "---\nid: SPC-0009\n---\n# S\n",
		"specs/2026-10-01-t.md": "---\nid: SPC-0010\n---\n# T\n",
		"plans/2026-10-01-a.md": splitE2EPlan("PLN-0001", "aaaaaaa", ""),
		"plans/2026-10-01-c.md": splitE2EPlan("PLN-0003", "ccccccc", "PLN-0001"),
	})
	wa, wc := filepath.Join(t.TempDir(), "wa"), filepath.Join(t.TempDir(), "wc")
	tidyGit(t, main, "worktree", "add", "-q", wa, "-b", "a", "main")
	tidyGit(t, main, "worktree", "add", "-q", wc, "-b", "c", "main")
	buildIn(t, wa, "2026-10-01-a", "aaaaaaa", time.Date(2026, 10, 1, 10, 0, 0, 0, time.Local))
	landFrom(t, main, "a", "PLN-0001")

	if code, out := runReady(t, main, "PLN-0003", "--parent", "main", "--head", "c"); code != readyWaiting {
		t.Errorf("stale worktree: exit %d, want %d\n%s", code, readyWaiting, out)
	}
	tidyGit(t, wc, "merge", "-q", "main")
	if code, out := runReady(t, main, "PLN-0003", "--parent", "main", "--head", "c"); code != exitOK {
		t.Errorf("worktree took in main: exit %d, want 0\n%s", code, out)
	}
}
