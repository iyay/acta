package write

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
)

var markFiles = map[string]string{
	".acta/specs/2026-09-30-s.md": "---\nid: SPC-0001\n---\n# S\n",
	".acta/plans/2026-09-30-p.md": "---\nid: PLN-0002\nparent: specs/2026-09-30-s\n---\n# P\n\n### Task 1: One\n- [ ] a\n- [ ] b\n\n### Task 2: Two\n- [ ] c\n",
	".acta/debt/2026-09-30-d.md":  "---\nid: DBT-0003\n---\n# D\n\n- [ ] x\n- [-] y\n",
}

// markRepo is markFiles plus one extra file, so a test can add a file without
// changing the files the other tests count on.
func markRepo(extra map[string]string) map[string]string {
	files := map[string]string{}
	for p, body := range markFiles {
		files[p] = body
	}
	for p, body := range extra {
		files[p] = body
	}
	return files
}

// firstOfKind gives the first item of that kind, because a debt line id ends
// in the line number of its own box.
func firstOfKind(t *testing.T, b *board.Board, kind board.Kind) string {
	t.Helper()
	for _, it := range b.Items {
		if it.Kind == kind {
			return it.ID
		}
	}
	t.Fatalf("no %s on the board", kind)
	return ""
}

func TestMarkItemDoneTicksTheWholeTaskDatesAndCommits(t *testing.T) {
	cfg := repoWith(t, markFiles)
	b := mustLoad(t, cfg)
	o, err := MarkItem(cfg, b, "plans/2026-09-30-p#task-1", true)
	if err != nil || !o.Committed {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	plan := readFile(t, filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-30-p.md"))
	if !strings.Contains(plan, "- [x] a\n- [x] b\n") || !strings.Contains(plan, "- [ ] c") {
		t.Fatalf("plan boxes: %q", plan)
	}
	if !strings.Contains(plan, "started:") {
		t.Fatalf("plan not dated: %q", plan)
	}
	spec := readFile(t, filepath.Join(cfg.RepoRoot, ".acta/specs/2026-09-30-s.md"))
	if !strings.Contains(spec, "started:") {
		t.Fatalf("spec not dated: %q", spec)
	}
	if st := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); st != "" {
		t.Fatalf("left uncommitted: %q", st)
	}
}

func TestMarkItemOpenClearsTaskAndDebtLine(t *testing.T) {
	cfg := repoWith(t, markFiles)
	if _, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-1", true); err != nil {
		t.Fatal(err)
	}
	if _, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-1", false); err != nil {
		t.Fatal(err)
	}
	plan := readFile(t, filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-30-p.md"))
	if !strings.Contains(plan, "- [ ] a\n- [ ] b\n") {
		t.Fatalf("task not open: %q", plan)
	}
	b := mustLoad(t, cfg)
	var wontfix string
	for _, it := range b.Items {
		if it.Path == filepath.Join(cfg.RepoRoot, ".acta/debt/2026-09-30-d.md") && it.Status == "wontfix" {
			wontfix = it.ID
		}
	}
	if wontfix == "" {
		t.Fatal("no wontfix debt line in the board")
	}
	if _, err := MarkItem(cfg, b, wontfix, false); err != nil {
		t.Fatal(err)
	}
	debt := readFile(t, filepath.Join(cfg.RepoRoot, ".acta/debt/2026-09-30-d.md"))
	if !strings.Contains(debt, "- [ ] x\n- [ ] y") {
		t.Fatalf("debt line not open: %q", debt)
	}
}

func TestMarkItemRefusesAndWritesNothing(t *testing.T) {
	cfg := repoWith(t, markFiles)
	b := mustLoad(t, cfg)
	// Each id names the one refusal it hits, so a board read that goes
	// wrong cannot pass by failing later for another reason.
	for id, want := range map[string]string{
		"nope":               "unknown id",
		"specs/2026-09-30-s": "not a task or a debt line",
		"plans/2026-09-30-p": "not a task or a debt line",
	} {
		_, err := MarkItem(cfg, b, id, true)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err = %v, want one saying %q", id, err, want)
		}
	}
	if st := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); st != "" {
		t.Fatalf("refusal wrote files: %q", st)
	}
}

func TestMarkItemLeavesADirtyFileUncommitted(t *testing.T) {
	cfg := repoWith(t, markFiles)
	path := filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-30-p.md")
	if err := os.WriteFile(path, []byte(readFile(t, path)+"\nnote\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	o, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-2", true)
	if err != nil || o.Committed || !o.Skipped {
		t.Fatalf("outcome %+v err %v", o, err)
	}
}

// A legacy task and a task read from another worktree both live in a file
// this press must not touch: one is waiting to be moved, the other is only
// read from a branch. Both are refused and nothing is written.
func TestMarkItemRefusesALegacyOrWorktreeTask(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".acta/plans/2026-09-30-p.md":              "# P\n\n### Task 1: One\n- [ ] a\n",
		"docs/superpowers/plans/2026-01-01-old.md": "# Old\n\n### Task 1: Old\n- [ ] b\n",
	})
	var legacy string
	for _, it := range mustLoad(t, cfg).Items {
		if it.Kind == board.KindTask && it.Legacy {
			legacy = it.ID
		}
	}
	if legacy == "" {
		t.Fatal("no legacy task on the board")
	}
	trees, err := board.LoadTrees(cfg, []board.Tree{{Cfg: cfg, Branch: "wt", Files: map[string][]byte{
		".acta/plans/2026-09-25-q.md": []byte("# Q\n\n### Task 1: One\n- [ ] c\n"),
	}}})
	if err != nil {
		t.Fatal(err)
	}
	var worktree string
	for _, it := range trees.Items {
		if it.Kind == board.KindTask && it.Worktree != "" {
			worktree = it.ID
		}
	}
	if worktree == "" {
		t.Fatal("no worktree task on the board")
	}
	// The message says which refusal it was, so a read that failed later
	// for another reason cannot pass here.
	if _, err := MarkItem(cfg, trees, legacy, true); err == nil || !strings.Contains(err.Error(), "legacy") {
		t.Errorf("%s: err = %v, want a legacy refusal", legacy, err)
	}
	if _, err := MarkItem(cfg, trees, worktree, true); err == nil || !strings.Contains(err.Error(), "worktree") {
		t.Errorf("%s: err = %v, want a worktree refusal", worktree, err)
	}
	if st := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); st != "" {
		t.Fatalf("refusal wrote files: %q", st)
	}
}

// Marking a task done twice is one press too many, not an error: the boxes
// are already ticked and the dates are already there, so nothing changes and
// nothing is committed.
func TestMarkItemDoneTwiceChangesNothing(t *testing.T) {
	cfg := repoWith(t, markFiles)
	if _, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-1", true); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-30-p.md")
	before, head := readFile(t, path), gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD")
	o, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-1", true)
	if err != nil || o.Committed {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	if after := readFile(t, path); after != before {
		t.Fatalf("plan changed on the second press: %q", after)
	}
	if now := gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD"); now != head {
		t.Fatalf("a second commit appeared: %s then %s", head, now)
	}
}

// An open task marked open stays open, and gets no start day either: only a
// tick starts the clock.
func TestMarkItemOpenOnAnOpenTaskChangesNothing(t *testing.T) {
	cfg := repoWith(t, markFiles)
	path := filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-30-p.md")
	before, head := readFile(t, path), gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD")
	o, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-1", false)
	if err != nil || o.Committed {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	if after := readFile(t, path); after != before || strings.Contains(after, "started:") {
		t.Fatalf("plan changed: %q", after)
	}
	if now := gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD"); now != head {
		t.Fatalf("a commit appeared: %s then %s", head, now)
	}
}

// One press is one commit, and it holds the files this press touched. A file
// dirty from before stays dirty and stays out of the commit.
func TestMarkItemCommitsOnlyTheFilesItTouched(t *testing.T) {
	cfg := repoWith(t, markRepo(map[string]string{"notes.md": "hello\n"}))
	notes := filepath.Join(cfg.RepoRoot, "notes.md")
	if err := os.WriteFile(notes, []byte("hello\nagain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	o, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-1", true)
	if err != nil || !o.Committed {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	// Only notes.md is left over: the plan and the spec this press wrote are
	// committed, and nothing else was waiting.
	if st := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); !strings.Contains(st, "notes.md") || strings.Contains(st, ".acta") {
		t.Fatalf("notes not left alone: %q", st)
	}
	if show := gitRun(t, cfg.RepoRoot, "show", "--name-only", "--format=", "HEAD"); strings.Contains(show, "notes.md") {
		t.Fatalf("commit took notes.md: %q", show)
	}
}

func TestMarkItemWithAutoCommitOffJustWrites(t *testing.T) {
	cfg := repoWith(t, markFiles)
	cfg.AutoCommit = false
	o, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-1", true)
	if err != nil || o.Committed || o.Reason != "auto_commit is off" {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	plan := readFile(t, filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-30-p.md"))
	if !strings.Contains(plan, "- [x] a\n- [x] b\n") || !strings.Contains(plan, "started:") {
		t.Fatalf("plan not written: %q", plan)
	}
	if st := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); st == "" {
		t.Fatal("wrote nothing git can see")
	}
}

// A debt line is one box, not work on the plan that raised it, so marking it
// done gives no plan and no spec a date.
func TestMarkItemOnADebtLineWritesNoDates(t *testing.T) {
	cfg := repoWith(t, markRepo(map[string]string{
		".acta/debt/2026-09-30-d.md": "---\nid: DBT-0003\nparent: plans/2026-09-30-p\n---\n# D\n\n- [ ] x\n",
	}))
	b := mustLoad(t, cfg)
	if _, err := MarkItem(cfg, b, firstOfKind(t, b, board.KindDebtItem), true); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{".acta/plans/2026-09-30-p.md", ".acta/specs/2026-09-30-s.md"} {
		if body := readFile(t, filepath.Join(cfg.RepoRoot, p)); strings.Contains(body, "started:") || strings.Contains(body, "finished:") {
			t.Fatalf("%s dated: %q", p, body)
		}
	}
	if st := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); st != "" {
		t.Fatalf("left uncommitted: %q", st)
	}
}

// TaskDates gives back the files it changed, so a caller can commit exactly
// those. A second call has nothing new to write and gives nothing back.
func TestTaskDatesGivesBackOnlyTheFilesItChanged(t *testing.T) {
	cfg := repoWith(t, markFiles)
	task := mustLoad(t, cfg).Get("plans/2026-09-30-p#task-1")
	if task == nil {
		t.Fatal("no task on the board")
	}
	plan := filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-30-p.md")
	spec := filepath.Join(cfg.RepoRoot, ".acta/specs/2026-09-30-s.md")
	got, err := TaskDates(cfg, task)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, " ") != plan+" "+spec {
		t.Fatalf("wrote %v, want only the plan and the spec", got)
	}
	again, err := TaskDates(cfg, task)
	if err != nil || len(again) != 0 {
		t.Fatalf("second call wrote %v err %v, want nothing", again, err)
	}
}
