package write

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/board"
)

func TestAssignIDsGivesMissingOnly(t *testing.T) {
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-20-a-design.md": "---\nid: SPC-0004\nhash: m2x9abc\n---\n# A\n",
		".acta/specs/2026-09-21-b-design.md": "# B\n",
		".acta/plans/2026-09-22-p.md":        "# P\n\n### Task 1: One\n- [ ] x\n",
		".acta/bugs/2026-09-23-c.md":         "# C\n",
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
		if it.Kind != "task" && !it.OldForm && !board.IsHash(strings.TrimPrefix(it.Hash, board.Prefix(it.Kind, strings.Contains(it.Path, "/plans/"))+"-")) {
			t.Errorf("%s hash %q", it.ID, it.Hash)
		}
	}
	if c, _, _ := AssignIDs(cfg, b, nil); len(c) != 0 {
		t.Fatalf("second run changed %v", c)
	}
}

func TestAssignIDsOneCommit(t *testing.T) {
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-21-b-design.md": "# B\n",
		".acta/bugs/2026-09-23-c.md":         "# C\n",
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
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/plans/2026-09-22-p.md": "# P\n\n### Task 1: One\n- [ ] x\n",
	})
	b, err := board.LoadTrees(cfg, []board.Tree{{Cfg: cfg, Branch: "x", Files: map[string][]byte{
		".acta/plans/2026-09-25-q.md": []byte("---\nid: PLAN-9\nhash: aaaa\n---\n# Q\n"),
	}}})
	changes, _, err := AssignIDs(cfg, b, nil)
	if err != nil || len(changes) != 1 {
		t.Fatalf("changes=%v err=%v", changes, err)
	}
	if b2 := mustLoad(t, cfg); b2.Get("PLN-0010") == nil {
		t.Fatal("new plan is not PLN-0010")
	}
}

func TestAssignIDsSkipsLegacy(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-21-b-design.md": "# B\n",
		".acta/bugs/2026-09-23-c.md":         "# C\n",
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
			return "m2x9abc"
		}
		return "zz00000"
	}
	defer func() { randHash = old }()
	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-20-a-design.md": "---\nid: SPEC-4\nhash: m2x9\n---\n# A\n",
		".acta/scratch/2026-09-20-x.md":      "---\nid: SCR-1\nhash: m2x9abc\n---\nx\n",
		".acta/bugs/2026-09-23-c.md":         "# C\n",
	})
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	if got := mustLoad(t, cfg).Get("BUG-1"); got == nil || got.Hash != "BUG-zz00000" {
		t.Fatalf("bug hash = %+v, want BUG-zz00000", got)
	}
}

func TestFixDuplicates(t *testing.T) {
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/bugs/2026-09-20-first.md": "---\nid: BUG-7\nhash: aaaa\n---\n# First\n",
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
	if len(changes) != 1 || changes[0] != "BUG-7 -> BUG-0008 (bugs/2026-09-21-second.md)" {
		t.Fatalf("changes=%v", changes)
	}
	b := mustLoad(t, cfg)
	if b.Get("BUG-0007") == nil || b.Get("BUG-0008") == nil {
		t.Fatal("renumber did not resolve")
	}
	if b.Get("BUG-0007").Hash != "BUG-aaaa" || b.Get("BUG-0008").Hash != "BUG-bbbb" {
		t.Fatalf("hashes moved: %q %q", b.Get("BUG-0007").Hash, b.Get("BUG-0008").Hash)
	}
}

func TestFixDuplicatesNoOp(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-17-broken.md":   "---\nref: [unclosed\n---\n# Broken\n",
		".acta/specs/2026-09-20-a-design.md": "# A\n",
		".acta/bugs/2026-09-23-c.md":         "# C\n",
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
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-15-really-bug.md": "---\ntype: bug\n---\n# Really a bug\n",
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
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-20-a-design.md": "---\nid: SPEC-zz\n---\n# A\n",
		".acta/bugs/2026-09-23-c.md":         "---\nhash: toolongg\n---\n# C\n",
	})
	_, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil {
		t.Fatal(err)
	}
	skips := strings.Join(out.Skips, "\n")
	if !strings.Contains(skips, "skip specs/2026-09-20-a-design: bad id SPEC-zz") {
		t.Fatalf("skip lines = %q", skips)
	}
	if !strings.Contains(skips, "skip bugs/2026-09-23-c: bad hash toolongg") {
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
	if !strings.Contains(bug, "hash: toolongg") || strings.Count(bug, "hash:") != 1 {
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
	t.Parallel()

	cfg := repoWith(t, map[string]string{".acta/bugs/2026-09-23-c.md": "# C\n"})
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
	if body := readFile(t, filepath.Join(cfg.Root, "bugs/2026-09-23-c.md")); !strings.Contains(body, "id: BUG-0001") {
		t.Fatalf("file not written: %q", body)
	}
}

// A file written on a branch before the main file in clock time still gives
// the number up, because it only reaches the branch at the merge.
func TestFixDuplicatesKeepsTheFileThatReachedTheBranchFirst(t *testing.T) {
	t.Parallel()

	cfg := repoWith(t, map[string]string{".acta/bugs/2026-09-20-seed.md": "# Seed\n"})
	put := func(rel, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(cfg.RepoRoot, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	at := func(date string, args ...string) {
		t.Helper()
		// The dates go to this one git child only, not to the whole test process.
		cmd := exec.Command("git", append([]string{"-C", cfg.RepoRoot}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	put(".acta/bugs/2026-09-21-main.md", "---\nid: BUG-7\nhash: bbbb\n---\n# Main\n")
	at("2030-01-01T00:00:00Z", "add", ".")
	at("2030-01-01T00:00:00Z", "commit", "-q", "-m", "main adds BUG-7")
	at("2020-01-01T00:00:00Z", "checkout", "-q", "-b", "side")
	put(".acta/bugs/2026-09-22-side.md", "---\nid: BUG-7\nhash: cccc\n---\n# Side\n")
	at("2020-01-01T00:00:00Z", "add", ".")
	at("2020-01-01T00:00:00Z", "commit", "-q", "-m", "side adds BUG-7")
	at("2024-01-01T00:00:00Z", "checkout", "-q", "main")
	at("2024-01-01T00:00:00Z", "merge", "-q", "--no-ff", "-m", "merge side", "side")

	changes, out, err := FixDuplicates(cfg, mustLoad(t, cfg))
	if err != nil || !out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if len(changes) != 1 || changes[0] != "BUG-7 -> BUG-0008 (bugs/2026-09-22-side.md)" {
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
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-20-a-design.md": "# A\n",
		".acta/specs/2026-09-21-b-design.md": "# B\n",
		".acta/plans/2026-09-25-a.md":        "---\nparent: specs/2026-09-20-a-design\nid: PLAN-13\nhash: aaaa\n---\n# Fix A\n\n### Task 1: T\n- [ ] a\n",
	})
	second := filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-26-b.md")
	if err := os.WriteFile(second, []byte("---\nparent: specs/2026-09-21-b-design\nid: PLAN-13\nhash: bbbb\n---\n# Fix B\n\n### Task 1: T\n- [ ] a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, cfg.RepoRoot, "add", ".")
	gitRun(t, cfg.RepoRoot, "commit", "-q", "-m", "second plan")
	changes, out, err := FixDuplicates(cfg, mustLoad(t, cfg))
	if err != nil || !out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if len(changes) != 1 || changes[0] != "PLN-13 -> PLN-0014 (plans/2026-09-26-b.md)" {
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
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-20-a-design.md": "# A\n",
		".acta/plans/2026-09-25-a.md":        "---\nparent: specs/2026-09-20-a-design\n---\n# Fix A\n\nNothing to tick yet.\n",
	})
	changes, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || !out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	plan := readFile(t, filepath.Join(cfg.Root, "plans/2026-09-25-a.md"))
	if !strings.Contains(plan, "id: PLN-0001") || !strings.Contains(plan, "hash: ") {
		t.Fatalf("held plan without tasks has no ids: %q", plan)
	}
	if len(changes) != 2 {
		t.Fatalf("changes=%v", changes)
	}
	if b := mustLoad(t, cfg); b.Get("PLN-0001") == nil {
		t.Fatal("the new plan id does not resolve")
	}
}

func TestAssignIDsGivesDebtFileANumber(t *testing.T) {
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/debt/2026-09-27-x.md": "# Review NOTEs: X\n\n- [ ] a\n",
	})
	changes, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || !out.Committed || len(changes) != 1 {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	b := mustLoad(t, cfg)
	if b.Get("DEBT-1") == nil {
		t.Fatal("the new debt id does not resolve")
	}
}

// Scratch numbers count on their own, so a new idea never pushes a spec or a
// bug along.
func TestAssignIDsNumbersScratchOnItsOwn(t *testing.T) {
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-20-a-design.md": "---\nid: SPEC-4\nhash: m2x9\n---\n# A\n",
		".acta/bugs/2026-09-23-c.md":         "---\nid: BUG-7\nhash: b7aa\n---\n# C\n",
		".acta/scratch/2026-09-28-one.md":    "# One\n",
		".acta/scratch/2026-09-28-two.md":    "# Two\n",
	})
	changes, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	// Two new scratch files, plus the spec and the bug moved to the new form.
	if err != nil || len(changes) != 4 {
		t.Fatalf("changes=%v err=%v", changes, err)
	}
	b := mustLoad(t, cfg)
	for _, c := range []struct{ id, short string }{
		{"scratch/2026-09-28-one", "SCR-0001"},
		{"scratch/2026-09-28-two", "SCR-0002"},
	} {
		it := b.Get(c.id)
		if it == nil || it.Kind != board.KindScratch || it.ShortID != c.short {
			t.Errorf("%s = %+v, want %s", c.id, it, c.short)
		}
		if b.Get(c.short) != it {
			t.Errorf("Get(%q) does not answer to the scratch item", c.short)
		}
	}
	for _, c := range []struct{ id, short string }{
		{"SPEC-4", "SPC-0004"},
		{"BUG-7", "BUG-0007"},
	} {
		it := b.Get(c.id)
		if it == nil || it.ShortID != c.short {
			t.Errorf("%s = %+v, want the number it already had as %s", c.id, it, c.short)
		}
	}
	if b.Get("SPC-0005") != nil || b.Get("BUG-0008") != nil {
		t.Error("a new scratch file moved a spec or a bug number")
	}
}

// A file's first id is also the day the file was made. A kind that is not on
// yet only gets created.
func TestAssignIDsFirstIDWritesCreated(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-09-26-x-design.md": "# X\n",
	})
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	doc := board.Parse([]byte(readFile(t, filepath.Join(cfg.Root, "specs/2026-09-26-x-design.md"))))
	if doc.Front["created"] != "2026-09-26" {
		t.Errorf("created = %v", doc.Front["created"])
	}
	if board.HasSchema(doc.Front) {
		t.Error("a spec is not on yet, so it must not get schema: 1")
	}
}

// Scratch is the kind that is on, so a new idea is checked from now on and
// gets both fields on its first id.
func TestAssignIDsFirstIDWritesSchemaOnScratch(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		".acta/scratch/2026-09-26-idea.md": "# One\n",
	})
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	doc := board.Parse([]byte(readFile(t, filepath.Join(cfg.Root, "scratch/2026-09-26-idea.md"))))
	if doc.Front["created"] != "2026-09-26" {
		t.Errorf("created = %v", doc.Front["created"])
	}
	if !board.HasSchema(doc.Front) {
		t.Error("a new scratch file must get schema: 1")
	}
}

// A file that already had an id keeps every byte it came in with, so no date
// and no schema flag lands on an old file.
func TestAssignIDsLeavesAFileThatHasAnIDAlone(t *testing.T) {
	fixNow(t)
	in := "---\nid: SPC-0001\nhash: abcd123\n---\n# X\n"
	cfg := repoWith(t, map[string]string{".acta/specs/2026-09-01-x-design.md": in})
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(cfg.Root, "specs/2026-09-01-x-design.md")); got != in {
		t.Errorf("a file that already had an id changed:\n%s", got)
	}
}

// A file that already had an id only misses its hash. It takes the hash and
// nothing else, because created and the schema flag belong to a first id.
func TestAssignIDsWritesNoDatesToAFileThatAlreadyHasAnID(t *testing.T) {
	fixNow(t)
	for _, c := range []struct {
		name, file, in string
	}{
		{"a spec", "specs/2026-09-01-x-design.md", "---\nid: SPEC-1\n---\n# X\n"},
		{"a scratch", "scratch/2026-09-01-x.md", "---\nid: SCRATCH-1\n---\n# X\n\n## Words\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			cfg := repoWith(t, map[string]string{".acta/" + c.file: c.in})
			if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
				t.Fatal(err)
			}
			got := readFile(t, filepath.Join(cfg.Root, c.file))
			doc := board.Parse([]byte(got))
			if doc.Front["hash"] == nil {
				t.Error("the missing hash was not filled in")
			}
			if doc.Front["created"] != nil {
				t.Errorf("a file that already had an id gained created = %v", doc.Front["created"])
			}
			if hasField([]byte(got), "schema") {
				t.Errorf("a file that already had an id gained a schema flag:\n%s", got)
			}
		})
	}
}

// A scratch file that already carries schema: 1 keeps the line as it was
// written, so id does not turn the number into a string behind the tool that
// put it there.
func TestAssignIDsLeavesAnExistingSchemaFlagAlone(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		".acta/scratch/2026-09-26-idea.md": "---\nstatus: raw\nschema: 1\n---\n# One\n\n## Words\n",
	})
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, filepath.Join(cfg.Root, "scratch/2026-09-26-idea.md"))
	if !strings.Contains(got, "schema: 1\n") {
		t.Errorf("the schema flag was rewritten:\n%s", got)
	}
}

// A file that asked for the check and fails it takes no id and is named in
// the skips, while the good file next to it still gets one.
func TestAssignIDsSkipsASchemaFileThatFails(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		".acta/scratch/2026-09-26-bad.md":     "---\nstatus: raw\nschema: 1\n---\n# bad\n\n## Context\n",
		".acta/specs/2026-09-26-ok-design.md": "# OK\n",
	})
	_, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !containsLine(out.Skips, "scratch 2026-09-26-bad.md: missing ## Words") {
		t.Errorf("skips = %v", out.Skips)
	}
	if bad := readFile(t, filepath.Join(cfg.Root, "scratch/2026-09-26-bad.md")); strings.Contains(bad, "id:") {
		t.Errorf("a file that fails the check got an id:\n%s", bad)
	}
	if ok := readFile(t, filepath.Join(cfg.Root, "specs/2026-09-26-ok-design.md")); !strings.Contains(ok, "id: SPC-") {
		t.Errorf("the good file next to it got no id:\n%s", ok)
	}
}

// A spec that grows out of a scratch idea closes that idea, and the date
// rides the same commit as the id.
func TestAssignIDsFinishesTheScratchParent(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		".acta/scratch/2026-09-20-idea.md":      "---\nid: SCR-0001\nhash: aaaaaaa\nstatus: brainstorming\n---\nx\n",
		".acta/specs/2026-09-26-idea-design.md": "---\nparent: scratch/2026-09-20-idea\n---\n# Idea\n",
	})
	_, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || !out.Committed {
		t.Fatalf("outcome %+v err %v", out, err)
	}
	doc := board.Parse([]byte(readFile(t, filepath.Join(cfg.Root, "scratch/2026-09-20-idea.md"))))
	if doc.Front["finished"] != "2026-09-26" {
		t.Errorf("the scratch parent did not get finished = %v", doc.Front["finished"])
	}
	if n := gitRun(t, cfg.RepoRoot, "log", "--format=%s", "-1"); n != "acta: assign short ids" {
		t.Errorf("the date must ride the id commit, last commit %q", n)
	}
}

// A parent that already closed on an earlier day keeps that day, so a second
// spec grown from the same idea does not rewrite what people read.
func TestAssignIDsKeepsTheScratchParentCloseDay(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		".acta/scratch/2026-09-20-idea.md":      "---\nid: SCR-0001\nhash: aaaaaaa\nstatus: brainstorming\nfinished: \"2026-09-01\"\n---\nx\n",
		".acta/specs/2026-09-26-idea-design.md": "---\nparent: scratch/2026-09-20-idea\n---\n# Idea\n",
	})
	if _, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil || !out.Committed {
		t.Fatalf("outcome %+v err %v", out, err)
	}
	doc := board.Parse([]byte(readFile(t, filepath.Join(cfg.Root, "scratch/2026-09-20-idea.md"))))
	if doc.Front["finished"] != "2026-09-01" {
		t.Errorf("the scratch parent lost the day it closed = %v", doc.Front["finished"])
	}
}

// Only a spec closes a scratch parent here. A bug the spec hangs under keeps
// the day it closes, and a plan built on an idea does not close it either:
// both are a command of their own.
func TestAssignIDsClosesOnlyAScratchParentOfASpec(t *testing.T) {
	fixNow(t)
	for _, c := range []struct {
		name  string
		files map[string]string
		keep  string // the file that must not gain a finished
	}{
		{
			name: "a bug parent",
			files: map[string]string{
				".acta/bugs/2026-09-20-b.md":            "---\nid: BUG-1\nhash: bbbb\nstatus: raw\n---\n# B\n",
				".acta/specs/2026-09-26-idea-design.md": "---\nparent: bugs/2026-09-20-b\n---\n# Idea\n",
			},
			keep: "bugs/2026-09-20-b.md",
		},
		{
			name: "a plan on an idea",
			files: map[string]string{
				".acta/scratch/2026-09-20-idea.md": "---\nid: SCR-0001\nhash: aaaaaaa\nstatus: brainstorming\n---\nx\n",
				".acta/plans/2026-09-26-idea.md":   "---\nparent: scratch/2026-09-20-idea\n---\n# Idea\n",
			},
			keep: "scratch/2026-09-20-idea.md",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			cfg := repoWith(t, c.files)
			if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
				t.Fatal(err)
			}
			doc := board.Parse([]byte(readFile(t, filepath.Join(cfg.Root, c.keep))))
			if doc.Front["finished"] != nil {
				t.Errorf("%s gained finished = %v", c.keep, doc.Front["finished"])
			}
		})
	}
}

// containsLine says whether one of the lines holds the text.
func containsLine(lines []string, text string) bool {
	for _, l := range lines {
		if strings.Contains(l, text) {
			return true
		}
	}
	return false
}

func TestRandHashIsSevenChars(t *testing.T) {
	t.Parallel()

	for range 200 {
		if h := randHash(); !board.IsHash(h) {
			t.Fatalf("randHash() = %q", h)
		}
	}
}

// A plan still holding an old id keeps its number, so a new plan lands past
// it and not on top of it, and it lands in the new shape.
func TestAssignIDsWritesNewFormat(t *testing.T) {
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/plans/2026-09-21-a.md": "---\nid: PLAN-12\nhash: k3f2\n---\n# A\n\n### Task 1: One\n- [ ] x\n",
		".acta/plans/2026-09-22-b.md": "# B\n\n### Task 1: One\n- [ ] x\n",
	})
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	plan := filepath.Join(cfg.Root, "plans/2026-09-22-b.md")
	if got := frontField(t, plan, "id"); got != "PLN-0013" {
		t.Errorf("new plan id = %q, want PLN-0013", got)
	}
	if h := frontField(t, plan, "hash"); !board.IsHash(h) {
		t.Errorf("new plan hash = %q, want 7 characters", h)
	}
}

// An old id, an old 4-character hash and the old ids in a closes list are
// written in the new form. The parent, the status, the other keys and the
// body come out byte for byte the same, and a second run finds nothing old
// and changes no byte at all.
func TestAssignIDsRewritesOldFiles(t *testing.T) {
	t.Parallel()

	const body = "# A\n\nSee PLAN-12 and SCRATCH-14.\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n"
	cfg := repoWith(t, map[string]string{
		".acta/plans/2026-09-21-a.md":   "---\nparent: scratch/2026-09-20-x\nid: PLAN-12\nhash: k3f2\ncloses: [DEBT-17.1, SCRATCH-14]\nstatus: approved\n---\n" + body,
		".acta/scratch/2026-09-20-x.md": "---\nid: SCRATCH-14\nhash: m2x9abc\ntitle: X\nstatus: raw\n---\nx\n",
		".acta/debt/2026-09-19-d.md":    "---\nid: DBT-0017\nhash: d4d4abc\n---\n# D\n\n- [ ] one\n",
	})
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	plan := filepath.Join(cfg.Root, "plans/2026-09-21-a.md")
	got := readFile(t, plan)
	hash := frontField(t, plan, "hash")
	if !board.IsHash(hash) || !strings.HasPrefix(hash, "k3f2") {
		t.Errorf("hash = %q, want the old k3f2 and 7 characters", hash)
	}
	// Every line around the three rewritten ones, and the body, byte for byte.
	want := "---\nparent: scratch/2026-09-20-x\nid: PLN-0012\nhash: " + hash + "\ncloses: [DBT-0017.01, SCR-0014]\nstatus: approved\n---\n" + body
	if got != want {
		t.Errorf("plan came out as:\n%swant:\n%s", got, want)
	}
	// A file that already holds both values in the new form is not touched.
	debt := filepath.Join(cfg.Root, "debt/2026-09-19-d.md")
	if s := readFile(t, debt); s != "---\nid: DBT-0017\nhash: d4d4abc\n---\n# D\n\n- [ ] one\n" {
		t.Errorf("debt file touched: %q", s)
	}
	// An old id on a file that already has a new hash is still rewritten.
	if v := frontField(t, filepath.Join(cfg.Root, "scratch/2026-09-20-x.md"), "id"); v != "SCR-0014" {
		t.Errorf("scratch id = %q, want SCR-0014", v)
	}
	changes, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || len(changes) != 0 || readFile(t, plan) != got {
		t.Errorf("second run changed files: %v %v", changes, err)
	}
}

// A file already in the new form is left exactly as it is, closes list and
// all, so a run over a migrated tree writes nothing.
func TestAssignIDsLeavesNewFilesAlone(t *testing.T) {
	t.Parallel()

	src := "---\nid: PLN-0012\nhash: k3f2abc\ncloses: [SCR-0014]\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n"
	cfg := repoWith(t, map[string]string{
		".acta/plans/2026-09-21-a.md":   src,
		".acta/scratch/2026-09-20-x.md": "---\nid: SCR-0014\nhash: m2x9abc\ntitle: X\nstatus: raw\n---\nx\n",
	})
	changes, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || len(changes) != 0 || readFile(t, filepath.Join(cfg.Root, "plans/2026-09-21-a.md")) != src {
		t.Errorf("new file touched: %v %v", changes, err)
	}
}

// A closes: entry that names nothing stays as written, a file whose id and
// hash are already new is still rewritten for its list, and the rest of the
// run goes on past the entry it could not resolve.
func TestAssignIDsKeepsAClosesEntryItCannotResolve(t *testing.T) {
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/plans/2026-09-21-a.md":   "---\nid: PLN-0012\nhash: k3f2abc\ncloses: [SCRATCH-14, DEBT-99]\n---\n# A\n",
		".acta/scratch/2026-09-20-x.md": "---\nid: SCRATCH-14\nhash: m2x9abc\n---\nx\n",
		".acta/bugs/2026-09-23-c.md":    "---\nid: BUG-7\n---\n# C\n",
	})
	changes, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil {
		t.Fatal(err)
	}
	plan := filepath.Join(cfg.Root, "plans/2026-09-21-a.md")
	got := readFile(t, plan)
	if !strings.Contains(got, "closes: [SCR-0014, DEBT-99]") {
		t.Errorf("the entry that names nothing must stay as written:\n%s", got)
	}
	for _, keep := range []string{"id: PLN-0012\n", "hash: k3f2abc\n"} {
		if !strings.Contains(got, keep) {
			t.Errorf("lost %q", keep)
		}
	}
	// The rest of the run still rewrote the bug's old id and gave it a hash.
	bug := filepath.Join(cfg.Root, "bugs/2026-09-23-c.md")
	if v := frontField(t, bug, "id"); v != "BUG-0007" {
		t.Errorf("bug id = %q, want BUG-0007", v)
	}
	if h := frontField(t, bug, "hash"); !board.IsHash(h) {
		t.Errorf("bug hash = %q, want 7 characters", h)
	}
	if len(changes) != 3 {
		t.Errorf("changes = %v, want the plan, the scratch and the bug", changes)
	}
	again, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || len(again) != 0 {
		t.Errorf("second run changes = %v err = %v", again, err)
	}
}

// An old 4-character hash keeps its 4 characters and takes 3 more, and a
// clash is drawn again until one is free.
func TestAssignIDsExtendsAnOldHashAndAvoidsAClash(t *testing.T) {
	old := randHash
	calls := 0
	randHash = func() string {
		calls++
		if calls < 3 {
			return "aaaaaa0"
		}
		return "aaaaaa1"
	}
	defer func() { randHash = old }()
	cfg := repoWith(t, map[string]string{
		".acta/bugs/2026-09-20-a.md": "---\nid: BUG-3\nhash: k3f2\n---\n# A\n",
		".acta/bugs/2026-09-21-b.md": "---\nid: BUG-4\nhash: k3f2\n---\n# B\n",
	})
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		"bugs/2026-09-20-a.md": "k3f2aa0",
		"bugs/2026-09-21-b.md": "k3f2aa1",
	} {
		if v := frontField(t, filepath.Join(cfg.Root, path), "hash"); v != want {
			t.Errorf("%s hash = %q, want %q", path, v, want)
		}
	}
	if calls < 3 {
		t.Errorf("the clash was not drawn again: %d calls", calls)
	}
}

// A run that only fills in a missing value commits as an assignment, and a
// run that moved an old value commits as a migration, so the history says
// which of the two happened.
func TestAssignIDsCommitMessageSaysWhatChanged(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name, body, msg string
	}{
		{"new", "# C\n", "acta: assign short ids"},
		{"old", "---\nid: BUG-7\n---\n# C\n", "acta: migrate ids to 3-letter prefix"},
	} {
		cfg := repoWith(t, map[string]string{".acta/bugs/2026-09-23-c.md": c.body})
		if _, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil || !out.Committed {
			t.Fatalf("%s: out=%+v err=%v", c.name, out, err)
		}
		if msg := gitRun(t, cfg.RepoRoot, "log", "-1", "--format=%s"); msg != c.msg {
			t.Errorf("%s run committed as %q, want %q", c.name, msg, c.msg)
		}
	}
}

// One number written two ways is still one number, so the file that came
// second is the one that moves, and it moves to the new shape.
func TestFixDuplicatesCountsOldAndNewAsOneNumber(t *testing.T) {
	t.Parallel()

	cfg := repoWith(t, map[string]string{
		".acta/bugs/2026-09-20-first.md": "---\nid: BUG-5\nhash: aaaa\n---\n# First\n",
	})
	second := filepath.Join(cfg.Root, "bugs/2026-09-21-second.md")
	if err := os.WriteFile(second, []byte("---\nid: BUG-0005\nhash: b7b7b7c\n---\n# Second\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, cfg.RepoRoot, "add", ".")
	gitRun(t, cfg.RepoRoot, "commit", "-q", "-m", "second")
	changes, out, err := FixDuplicates(cfg, mustLoad(t, cfg))
	if err != nil || !out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if len(changes) != 1 || changes[0] != "BUG-5 -> BUG-0006 (bugs/2026-09-21-second.md)" {
		t.Fatalf("changes=%v", changes)
	}
	if got := frontField(t, second, "id"); got != "BUG-0006" {
		t.Errorf("renumbered id = %q, want BUG-0006", got)
	}
}

// frontField reads one value out of a file's own frontmatter, so a test sees
// what was written down and not what the board makes of it.
func frontField(t *testing.T, path, key string) string {
	t.Helper()
	for _, line := range strings.Split(readFile(t, path), "\n") {
		if v, ok := strings.CutPrefix(line, key+": "); ok {
			return strings.Trim(strings.TrimSpace(v), `"`)
		}
	}
	return ""
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
