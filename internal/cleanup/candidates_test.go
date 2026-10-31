package cleanup

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// testRepo is a git repo whose first commit is backdated by old, so each test
// can make one file look quiet or busy with a later commit.
type testRepo struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T, old time.Duration, files map[string]string) testRepo {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	r := testRepo{t, dir}
	r.git(0, "init", "-q", "-b", "main")
	r.git(0, "config", "user.name", "test")
	r.git(0, "config", "user.email", "test@example.com")
	r.write(files)
	r.git(0, "add", ".")
	r.git(old, "commit", "-q", "-m", "init")
	return r
}

func (r testRepo) write(files map[string]string) {
	r.t.Helper()
	for p, body := range files {
		full := filepath.Join(r.dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			r.t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
}

// git runs git with its author and committer dates set ago in the past.
func (r testRepo) git(ago time.Duration, args ...string) {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", r.dir}, args...)...)
	when := time.Now().Add(-ago).Format(time.RFC3339)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+when, "GIT_COMMITTER_DATE="+when)
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("git %v: %v %s", args, err, out)
	}
}

// commit adds the files in one commit made ago in the past.
func (r testRepo) commit(ago time.Duration, msg string, files map[string]string) {
	r.t.Helper()
	r.write(files)
	r.git(ago, "add", ".")
	r.git(ago, "commit", "-q", "-m", msg)
}

func (r testRepo) candidates() map[string]Candidate {
	r.t.Helper()
	cfg, err := config.Load(r.dir, "")
	if err != nil {
		r.t.Fatal(err)
	}
	b, err := board.Load(cfg)
	if err != nil {
		r.t.Fatal(err)
	}
	list, err := Candidates(cfg, b, time.Now())
	if err != nil {
		r.t.Fatal(err)
	}
	out := map[string]Candidate{}
	for _, c := range list {
		out[c.ID] = c
	}
	return out
}

const day = 24 * time.Hour

func dateAgo(d time.Duration) string { return time.Now().Add(-d).Format("2006-01-02") }

func TestCandidatesPicksEachRule(t *testing.T) {
	t.Parallel()

	old := dateAgo(40 * day)
	young := dateAgo(5 * day)
	r := newRepo(t, 30*day, map[string]string{
		".acta/plans/2026-01-01-base.md":      "---\nid: PLN-0001\nhash: aaaaaaa\n---\n# Base\n\n### Task 1: A\n- [x] a\n",
		".acta/debt/2026-01-02-notes.md":      "---\nparent: plans/2026-01-01-base\n---\n# Notes\n\n- [ ] (medium) open line\n- [x] closed line\n",
		".acta/scratch/2026-01-03-raw.md":     "---\nstatus: raw\ncreated: \"" + old + "\"\n---\n# Raw\n",
		".acta/scratch/2026-01-04-stuck.md":   "---\nstatus: brainstorming\n---\n# Stuck\n",
		".acta/scratch/2026-01-05-going.md":   "---\nstatus: brainstorming\n---\n# Going\n",
		".acta/scratch/2026-01-06-specced.md": "---\nstatus: raw\n---\n# Specced\n",
		".acta/specs/2026-01-06-from.md":      "---\nparent: scratch/2026-01-06-specced\ncreated: \"" + young + "\"\n---\n# From\n",
		".acta/bugs/2026-01-07-open.md":       "# Open\n",
		".acta/bugs/2026-01-08-fixed.md":      "---\nstatus: fixed\n---\n# Fixed\n",
		".acta/specs/2026-01-09-lone.md":      "---\nstatus: draft\ncreated: \"" + old + "\"\n---\n# Lone\n",
		".acta/specs/2026-01-10-fresh.md":     "---\nstatus: draft\ncreated: \"" + young + "\"\n---\n# Fresh\n",
		".acta/specs/2026-01-11-planned.md":   "---\nstatus: approved\ncreated: \"" + old + "\"\n---\n# Planned\n",
		".acta/plans/2026-01-11-planned.md":   "---\nparent: specs/2026-01-11-planned\n---\n# Planned plan\n\n### Task 1: A\n- [ ] a\n",
		".acta/plans/2026-01-12-stale.md":     "---\nid: PLN-0002\nhash: bbbbbbb\n---\n# Stale\n\n### Task 1: A\n- [x] a\n\n### Task 2: B\n- [ ] b\n",
		".acta/plans/2026-01-13-ticked.md":    "---\nid: PLN-0003\nhash: ccccccc\n---\n# Ticked\n\n### Task 1: A\n- [x] a\n\n### Task 2: B\n- [ ] b\n",
		".acta/plans/2026-01-14-trailer.md":   "---\nid: PLN-0004\nhash: ddddddd\n---\n# Trailer\n\n### Task 1: A\n- [x] a\n\n### Task 2: B\n- [ ] b\n",
	})
	// The going brainstorm got a log line 13 days ago, one day short of quiet.
	r.commit(13*day, "chore(scratch): log", map[string]string{".acta/scratch/2026-01-05-going.md": "---\nstatus: brainstorming\n---\n# Going\n\nlog\n"})
	// A tick commit on the plan file is work on the plan.
	r.commit(2*day, "chore(plan): tick", map[string]string{".acta/plans/2026-01-13-ticked.md": "---\nid: PLN-0003\nhash: ccccccc\n---\n# Ticked\n\n### Task 1: A\n- [x] a\n\n### Task 2: B\n- [ ] b\n- [ ] c\n"})
	// A code commit names the plan only in its trailer.
	r.commit(1*day, "feat: code\n\nTask: PLN-ddddddd#2", map[string]string{"code.go": "package x\n"})

	got := r.candidates()
	want := map[string]string{
		"debt/2026-01-02-notes#item-1": "open debt line",
		"scratch/2026-01-03-raw":       "raw idea with no spec",
		"scratch/2026-01-04-stuck":     "brainstorm with no spec, quiet 30 days",
		"bugs/2026-01-07-open":         "open bug",
		"specs/2026-01-09-lone":        "draft spec with no plan",
		"plans/2026-01-12-stale":       "open tasks, no commit for 30 days",
	}
	for id, why := range want {
		if c, ok := got[id]; !ok || c.Why != why {
			t.Errorf("%s: got %+v, want why %q", id, c, why)
		}
	}
	for id := range got {
		if _, ok := want[id]; !ok {
			t.Errorf("%s should not be a candidate: %+v", id, got[id])
		}
	}
	if c := got["scratch/2026-01-03-raw"]; c.AgeDays != 40 || c.Path != ".acta/scratch/2026-01-03-raw.md" || c.Kind != "scratch" {
		t.Errorf("raw idea fields = %+v", c)
	}
}

// A plan file no commit has touched yet counts from started, then from
// created.
func TestCandidatesPlanClockFallsBack(t *testing.T) {
	t.Parallel()

	r := newRepo(t, 0, map[string]string{"README.md": "x\n"})
	plan := func(dates string) string {
		return "---\n" + dates + "---\n# P\n\n### Task 1: A\n- [x] a\n\n### Task 2: B\n- [ ] b\n"
	}
	r.write(map[string]string{
		".acta/plans/2026-01-01-started.md": plan("started: \"" + dateAgo(20*day) + "\"\ncreated: \"" + dateAgo(2*day) + "\"\n"),
		".acta/plans/2026-01-02-created.md": plan("created: \"" + dateAgo(20*day) + "\"\n"),
		".acta/plans/2026-01-03-young.md":   plan("started: \"" + dateAgo(2*day) + "\"\ncreated: \"" + dateAgo(20*day) + "\"\n"),
	})
	got := r.candidates()
	for _, id := range []string{"plans/2026-01-01-started", "plans/2026-01-02-created"} {
		if c := got[id]; c.Why != "open tasks, no commit for 20 days" {
			t.Errorf("%s: got %+v", id, c)
		}
	}
	if c, ok := got["plans/2026-01-03-young"]; ok {
		t.Errorf("a plan started 2 days ago is not stuck: %+v", c)
	}
}

// Work that lives in another worktree is still going on, so cleanup never
// asks about it: not an item only on that branch, not a plan it builds.
func TestCandidatesSkipsWorktreeWork(t *testing.T) {
	t.Parallel()

	plan := "---\nid: PLN-0001\nhash: aaaaaaa\n---\n# Building\n\n### Task 1: A\n- [x] a\n\n### Task 2: B\n- [ ] b\n"
	r := newRepo(t, 30*day, map[string]string{".acta/plans/2026-01-01-building.md": plan})
	wt := filepath.Join(t.TempDir(), "wt")
	r.git(30*day, "worktree", "add", "-q", wt, "-b", "building")
	w := testRepo{t, wt}
	w.commit(30*day, "chore(plan): tick", map[string]string{
		".acta/plans/2026-01-01-building.md": strings.Replace(plan, "- [ ] b", "- [ ] b\n- [ ] c", 1),
		".acta/scratch/2026-01-02-branch.md": "---\nstatus: raw\n---\n# Branch only\n",
	})
	got := r.candidates()
	for _, id := range []string{"plans/2026-01-01-building", "scratch/2026-01-02-branch"} {
		if _, ok := got[id]; ok {
			t.Errorf("%s belongs to a worktree and should be left out", id)
		}
	}
}

func TestCandidatesGroupsByKindOldestFirst(t *testing.T) {
	t.Parallel()

	r := newRepo(t, 0, map[string]string{
		".acta/bugs/2026-01-01-bug.md":       "---\ncreated: \"" + dateAgo(3*day) + "\"\n---\n# Bug\n",
		".acta/scratch/2026-01-02-young.md":  "---\nstatus: raw\ncreated: \"" + dateAgo(2*day) + "\"\n---\n# Young\n",
		".acta/scratch/2026-01-03-oldest.md": "---\nstatus: raw\ncreated: \"" + dateAgo(9*day) + "\"\n---\n# Oldest\n",
	})
	cfg, _ := config.Load(r.dir, "")
	b, _ := board.Load(cfg)
	list, err := Candidates(cfg, b, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, c := range list {
		ids = append(ids, c.ID)
	}
	if want := []string{"scratch/2026-01-03-oldest", "scratch/2026-01-02-young", "bugs/2026-01-01-bug"}; !slices.Equal(ids, want) {
		t.Fatalf("order = %v, want %v", ids, want)
	}
}
