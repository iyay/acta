package write

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/board"
)

func TestAssignIDsGivesMissingOnly(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-20-a-design.md": "---\nid: SPEC-4\nhash: m2x9\n---\n# A\n",
		".pm/specs/2026-09-21-b-design.md": "# B\n",
		".pm/plans/2026-09-22-p.md":        "# P\n\n### Task 1: One\n- [ ] x\n",
		".pm/bugs/2026-09-23-c.md":         "# C\n",
	})
	before, _ := os.ReadFile(filepath.Join(cfg.Root, "specs/2026-09-20-a-design.md"))
	changes, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || !out.Committed || len(changes) != 3 {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	b := mustLoad(t, cfg)
	if b.Get("SPEC-5") == nil || b.Get("PLAN-1") == nil || b.Get("BUG-1") == nil {
		t.Fatal("new numbers do not resolve")
	}
	after, _ := os.ReadFile(filepath.Join(cfg.Root, "specs/2026-09-20-a-design.md"))
	if string(before) != string(after) {
		t.Fatal("an existing id was rewritten")
	}
	for _, it := range b.Items {
		if it.Kind != "task" && !board.IsHash(strings.TrimPrefix(it.Hash, board.Prefix(it.Kind, strings.Contains(it.Path, "/plans/"))+"-")) {
			t.Errorf("%s hash %q", it.ID, it.Hash)
		}
	}
	if c, _, _ := AssignIDs(cfg, b, nil); len(c) != 0 {
		t.Fatalf("second run changed %v", c)
	}
}

func TestAssignIDsOneCommit(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-21-b-design.md": "# B\n",
		".pm/bugs/2026-09-23-c.md":         "# C\n",
	})
	before, _ := strconv.Atoi(gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD"))
	changes, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || !out.Committed || len(changes) != 2 {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	after, _ := strconv.Atoi(gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD"))
	if after-before != 1 {
		t.Fatalf("commits went from %d to %d, want +1", before, after)
	}
}

func TestAssignIDsCountsOtherTrees(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/plans/2026-09-22-p.md": "# P\n\n### Task 1: One\n- [ ] x\n",
	})
	b, err := board.LoadTrees(cfg, []board.Tree{{Cfg: cfg, Branch: "x", Files: map[string][]byte{
		".pm/plans/2026-09-25-q.md": []byte("---\nid: PLAN-9\nhash: aaaa\n---\n# Q\n"),
	}}})
	changes, _, err := AssignIDs(cfg, b, nil)
	if err != nil || len(changes) != 1 {
		t.Fatalf("changes=%v err=%v", changes, err)
	}
	if b2 := mustLoad(t, cfg); b2.Get("PLAN-10") == nil {
		t.Fatal("new plan is not PLAN-10")
	}
}

func TestAssignIDsSkipsLegacy(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	legacy := filepath.Join(cfg.RepoRoot, "docs/superpowers/specs/2026-01-01-old.md")
	before, _ := os.ReadFile(legacy)
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(legacy)
	if string(before) != string(after) || strings.Contains(string(after), "id:") {
		t.Fatalf("legacy rewritten: %q", after)
	}
}

func TestAssignIDsOnlyNamed(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-21-b-design.md": "# B\n",
		".pm/bugs/2026-09-23-c.md":         "# C\n",
	})
	changes, _, err := AssignIDs(cfg, mustLoad(t, cfg), []string{"bugs/2026-09-23-c"})
	if err != nil || len(changes) != 1 {
		t.Fatalf("changes=%v err=%v", changes, err)
	}
	spec, _ := os.ReadFile(filepath.Join(cfg.Root, "specs/2026-09-21-b-design.md"))
	if strings.Contains(string(spec), "id:") {
		t.Fatalf("spec touched: %q", spec)
	}
	if b := mustLoad(t, cfg); b.Get("BUG-1") == nil {
		t.Fatal("bug did not get BUG-1")
	}
}

func TestAssignIDsHashClash(t *testing.T) {
	old := randHash
	calls := 0
	randHash = func() string {
		calls++
		if calls == 1 {
			return "m2x9"
		}
		return "zz00"
	}
	defer func() { randHash = old }()
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-20-a-design.md": "---\nid: SPEC-4\nhash: m2x9\n---\n# A\n",
		".pm/bugs/2026-09-23-c.md":         "# C\n",
	})
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	if got := mustLoad(t, cfg).Get("BUG-1"); got == nil || got.Hash != "BUG-zz00" {
		t.Fatalf("bug hash = %+v, want BUG-zz00", got)
	}
}

func TestFixDuplicates(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/bugs/2026-09-20-first.md": "---\nid: BUG-7\nhash: aaaa\n---\n# First\n",
	})
	time.Sleep(1100 * time.Millisecond)
	full := filepath.Join(cfg.Root, "bugs/2026-09-21-second.md")
	if err := os.WriteFile(full, []byte("---\nid: BUG-7\nhash: bbbb\n---\n# Second\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, cfg.RepoRoot, "add", ".")
	gitRun(t, cfg.RepoRoot, "commit", "-q", "-m", "second")
	changes, out, err := FixDuplicates(cfg, mustLoad(t, cfg))
	if err != nil || !out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if len(changes) != 1 || changes[0] != "BUG-7 -> BUG-8 (bugs/2026-09-21-second.md)" {
		t.Fatalf("changes=%v", changes)
	}
	b := mustLoad(t, cfg)
	if b.Get("BUG-7") == nil || b.Get("BUG-8") == nil {
		t.Fatal("renumber did not resolve")
	}
	if b.Get("BUG-7").Hash != "BUG-aaaa" || b.Get("BUG-8").Hash != "BUG-bbbb" {
		t.Fatalf("hashes moved: %q %q", b.Get("BUG-7").Hash, b.Get("BUG-8").Hash)
	}
}

func TestFixDuplicatesNoOp(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	before := gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD")
	changes, out, err := FixDuplicates(cfg, mustLoad(t, cfg))
	if err != nil || len(changes) != 0 || out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if after := gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD"); after != before {
		t.Fatal("commit made with nothing to fix")
	}
}

// A file whose frontmatter will not parse is left alone and named on stderr,
// the other files still get their IDs, and one commit holds the run.
func TestAssignIDsSkipsBrokenFrontmatterAndCommitsTheRest(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-17-broken.md":   "---\nref: [unclosed\n---\n# Broken\n",
		".pm/specs/2026-09-20-a-design.md": "# A\n",
		".pm/bugs/2026-09-23-c.md":         "# C\n",
	})
	before := gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD")
	changes, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil {
		t.Fatalf("broken file killed the run: %v", err)
	}
	if !out.Committed || len(changes) != 2 {
		t.Fatalf("changes=%v out=%+v", changes, out)
	}
	skips := strings.Join(out.Skips, "\n")
	if !strings.Contains(skips, "skip specs/2026-09-17-broken: frontmatter:") {
		t.Fatalf("skip lines = %q", skips)
	}
	broken := readFile(t, filepath.Join(cfg.Root, "specs/2026-09-17-broken.md"))
	if !strings.Contains(broken, "ref: [unclosed") || strings.Contains(broken, "id:") {
		t.Fatalf("broken file rewritten: %q", broken)
	}
	if after := gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD"); after != strconv.Itoa(mustAtoi(t, before)+1) {
		t.Fatalf("commits went from %s to %s, want +1", before, after)
	}
	again, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || len(again) != 0 || out.Committed {
		t.Fatalf("second run changes=%v out=%+v err=%v", again, out, err)
	}
}

// A file that says type: bug in specs/ is a bug, so it reads and writes as
// BUG both times, and the second run has nothing left to do.
func TestAssignIDsPrefixFollowsTheKindTheFileIs(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-15-really-bug.md": "---\ntype: bug\n---\n# Really a bug\n",
	})
	changes, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || !out.Committed || len(changes) != 1 {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if !strings.HasPrefix(changes[0], "BUG-") {
		t.Fatalf("change line = %q, want a BUG number", changes[0])
	}
	first := readFile(t, filepath.Join(cfg.Root, "specs/2026-09-15-really-bug.md"))
	again, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || len(again) != 0 {
		t.Fatalf("second run changes=%v err=%v", again, err)
	}
	if second := readFile(t, filepath.Join(cfg.Root, "specs/2026-09-15-really-bug.md")); second != first {
		t.Fatalf("second run rewrote the file:\n%s\n---\n%s", first, second)
	}
	b := mustLoad(t, cfg)
	it := b.Get("specs/2026-09-15-really-bug")
	if it == nil || it.ShortID == "" || !strings.HasPrefix(it.ShortID, "BUG-") {
		t.Fatalf("board reads %+v", it)
	}
	if got := strings.Join(it.Problems, "; "); got != "" {
		t.Fatalf("file pmb id wrote has problems: %q", got)
	}
}

// An id or hash in the wrong shape is a person's typo, not an empty field:
// it stays as written, the other field is still filled in, and the reason is
// on stderr.
func TestAssignIDsNeverRewritesAValueThatIsThere(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-20-a-design.md": "---\nid: SPEC-zz\n---\n# A\n",
		".pm/bugs/2026-09-23-c.md":         "---\nhash: toolong\n---\n# C\n",
	})
	_, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil {
		t.Fatal(err)
	}
	skips := strings.Join(out.Skips, "\n")
	if !strings.Contains(skips, "skip specs/2026-09-20-a-design: bad id SPEC-zz") {
		t.Fatalf("skip lines = %q", skips)
	}
	if !strings.Contains(skips, "skip bugs/2026-09-23-c: bad hash toolong") {
		t.Fatalf("skip lines = %q", skips)
	}
	spec := readFile(t, filepath.Join(cfg.Root, "specs/2026-09-20-a-design.md"))
	if !strings.Contains(spec, "id: SPEC-zz") || strings.Count(spec, "id:") != 1 {
		t.Fatalf("bad id rewritten: %q", spec)
	}
	if !strings.Contains(spec, "hash: ") {
		t.Fatalf("missing hash not filled in: %q", spec)
	}
	bug := readFile(t, filepath.Join(cfg.Root, "bugs/2026-09-23-c.md"))
	if !strings.Contains(bug, "hash: toolong") || strings.Count(bug, "hash:") != 1 {
		t.Fatalf("bad hash rewritten: %q", bug)
	}
	if !strings.Contains(bug, "id: BUG-") {
		t.Fatalf("missing id not filled in: %q", bug)
	}
	b := mustLoad(t, cfg)
	if it := b.Get("specs/2026-09-20-a-design"); it.ShortID != "" ||
		!strings.Contains(strings.Join(it.Problems, "; "), "bad id SPEC-zz") {
		t.Fatalf("board lost the bad id: %+v", it)
	}
}

// With auto_commit off the files are written and left for the person, and
// the run says so the way finish does.
func TestAssignIDsAutoCommitOff(t *testing.T) {
	cfg := repoWith(t, map[string]string{".pm/bugs/2026-09-23-c.md": "# C\n"})
	cfg.AutoCommit = false
	before := gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD")
	changes, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || len(changes) != 1 {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if out.Committed || out.Skipped || out.Reason != "auto_commit is off" {
		t.Fatalf("outcome = %+v", out)
	}
	if after := gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD"); after != before {
		t.Fatalf("commits went from %s to %s, want none", before, after)
	}
	if body := readFile(t, filepath.Join(cfg.Root, "bugs/2026-09-23-c.md")); !strings.Contains(body, "id: BUG-1") {
		t.Fatalf("file not written: %q", body)
	}
}

// A file written on a branch before the main file in clock time still gives
// the number up, because it only reaches the branch at the merge.
func TestFixDuplicatesKeepsTheFileThatReachedTheBranchFirst(t *testing.T) {
	cfg := repoWith(t, map[string]string{".pm/bugs/2026-09-20-seed.md": "# Seed\n"})
	put := func(rel, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(cfg.RepoRoot, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	at := func(date string, args ...string) {
		t.Helper()
		t.Setenv("GIT_AUTHOR_DATE", date)
		t.Setenv("GIT_COMMITTER_DATE", date)
		gitRun(t, cfg.RepoRoot, args...)
	}
	put(".pm/bugs/2026-09-21-main.md", "---\nid: BUG-7\nhash: bbbb\n---\n# Main\n")
	at("2030-01-01T00:00:00Z", "add", ".")
	at("2030-01-01T00:00:00Z", "commit", "-q", "-m", "main adds BUG-7")
	at("2020-01-01T00:00:00Z", "checkout", "-q", "-b", "side")
	put(".pm/bugs/2026-09-22-side.md", "---\nid: BUG-7\nhash: cccc\n---\n# Side\n")
	at("2020-01-01T00:00:00Z", "add", ".")
	at("2020-01-01T00:00:00Z", "commit", "-q", "-m", "side adds BUG-7")
	at("2024-01-01T00:00:00Z", "checkout", "-q", "main")
	at("2024-01-01T00:00:00Z", "merge", "-q", "--no-ff", "-m", "merge side", "side")

	changes, out, err := FixDuplicates(cfg, mustLoad(t, cfg))
	if err != nil || !out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if len(changes) != 1 || changes[0] != "BUG-7 -> BUG-8 (bugs/2026-09-22-side.md)" {
		t.Fatalf("changes=%v", changes)
	}
	main := readFile(t, filepath.Join(cfg.Root, "bugs/2026-09-21-main.md"))
	if !strings.Contains(main, "id: BUG-7") || !strings.Contains(main, "hash: bbbb") {
		t.Fatalf("main file rewritten: %q", main)
	}
}

// Two plans can hold the same number while their tasks sit on two different
// specs, so the plan files themselves are the ones to repair.
func TestFixDuplicatesRepairsPlansHeldBySpecs(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-20-a-design.md": "# A\n",
		".pm/specs/2026-09-21-b-design.md": "# B\n",
		".pm/plans/2026-09-25-a.md":        "---\nparent: specs/2026-09-20-a-design\nid: PLAN-13\nhash: aaaa\n---\n# Fix A\n\n### Task 1: T\n- [ ] a\n",
	})
	second := filepath.Join(cfg.RepoRoot, ".pm/plans/2026-09-26-b.md")
	if err := os.WriteFile(second, []byte("---\nparent: specs/2026-09-21-b-design\nid: PLAN-13\nhash: bbbb\n---\n# Fix B\n\n### Task 1: T\n- [ ] a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, cfg.RepoRoot, "add", ".")
	gitRun(t, cfg.RepoRoot, "commit", "-q", "-m", "second plan")
	changes, out, err := FixDuplicates(cfg, mustLoad(t, cfg))
	if err != nil || !out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if len(changes) != 1 || changes[0] != "PLAN-13 -> PLAN-14 (plans/2026-09-26-b.md)" {
		t.Fatalf("changes=%v", changes)
	}
	first := readFile(t, filepath.Join(cfg.Root, "plans/2026-09-25-a.md"))
	if !strings.Contains(first, "id: PLAN-13") || !strings.Contains(first, "hash: aaaa") {
		t.Fatalf("first plan rewritten: %q", first)
	}
	if again, _, err := FixDuplicates(cfg, mustLoad(t, cfg)); err != nil || len(again) != 0 {
		t.Fatalf("second run changes=%v err=%v", again, err)
	}
}

// A plan that belongs to a spec is a file with an id of its own, whether or
// not it has tasks to reach it.
func TestAssignIDsGivesHeldPlanWithoutTasksAnID(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-20-a-design.md": "# A\n",
		".pm/plans/2026-09-25-a.md":        "---\nparent: specs/2026-09-20-a-design\n---\n# Fix A\n\nNothing to tick yet.\n",
	})
	changes, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || !out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	plan := readFile(t, filepath.Join(cfg.Root, "plans/2026-09-25-a.md"))
	if !strings.Contains(plan, "id: PLAN-1") || !strings.Contains(plan, "hash: ") {
		t.Fatalf("held plan without tasks has no ids: %q", plan)
	}
	if len(changes) != 2 {
		t.Fatalf("changes=%v", changes)
	}
	if b := mustLoad(t, cfg); b.Get("PLAN-1") == nil {
		t.Fatal("the new plan id does not resolve")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		t.Fatal(err)
	}
	return n
}
