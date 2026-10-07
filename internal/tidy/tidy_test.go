package tidy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/testguard"
)

func TestMain(m *testing.M) {
	testguard.Watch()
	os.Exit(m.Run())
}

// repoT is a temp repo with a few helpers.
type repoT struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T) *repoT {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	r := &repoT{t: t, dir: dir}
	r.git(nil, "init", "-q", "-b", "main")
	return r
}

func (r *repoT) git(env []string, args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", r.dir}, args...)...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// commit writes the files and commits them. date is RFC 3339; who is the author name.
func (r *repoT) commit(msg, date, who string, files map[string]string) string {
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
	r.git(nil, "add", "-A")
	env := []string{
		"GIT_AUTHOR_NAME=" + who, "GIT_AUTHOR_EMAIL=" + who + "@example.com",
		"GIT_COMMITTER_NAME=" + who, "GIT_COMMITTER_EMAIL=" + who + "@example.com",
		"GIT_AUTHOR_DATE=" + date, "GIT_COMMITTER_DATE=" + date,
	}
	r.git(env, "commit", "-q", "--allow-empty", "-m", msg)
	return r.git(nil, "rev-parse", "HEAD")
}

func (r *repoT) rev(ref string) string { return r.git(nil, "rev-parse", ref) }

func (r *repoT) subjects(rng string) []string {
	out := r.git(nil, "log", "--reverse", "--format=%s", rng)
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

func (r *repoT) refExists(ref string) bool {
	return exec.Command("git", "-C", r.dir, "rev-parse", "--verify", "-q", ref).Run() == nil
}

func d(n int) string {
	return time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC).Add(time.Duration(n) * time.Hour).Format(time.RFC3339)
}

func opt(base, branch string) Options {
	return Options{Base: base, Branch: branch, PlanningRoot: ".acta"}
}

// start makes main with one commit and a feature branch on top of it.
func start(t *testing.T) *repoT {
	r := newRepo(t)
	r.commit("init", d(0), "ann", map[string]string{"README.md": "hi\n", ".acta/seed.md": "seed\n"})
	r.git(nil, "checkout", "-q", "-b", "feat")
	return r
}

func TestKeepRuleAndFolds(t *testing.T) {
	r := start(t)
	base := r.rev("main")
	// Planning only commit before the first kept commit: rides forward.
	r.commit("chore(plan): start", d(1), "ann", map[string]string{".acta/early.md": "early\n"})
	r.commit("feat: a", d(2), "ann", map[string]string{"a.go": "a\n"})
	// Planning only after a kept commit: folds back.
	r.commit("chore(plan): tick a", d(3), "ann", map[string]string{".acta/tick.md": "tick\n"})
	// Code file, but type docs: folds back.
	r.commit("docs: note a", d(4), "ann", map[string]string{"README.md": "hi more\n"})
	r.commit("feat: b", d(5), "ann", map[string]string{"b.go": "b\n"})
	branchTip := r.rev("feat")

	res, err := Run(r.dir, opt("main", "feat"))
	if err != nil {
		t.Fatal(err)
	}
	if res.OldCount != 5 || res.NewCount != 2 {
		t.Fatalf("counts %+v", res)
	}
	got := r.subjects(base + ".." + "refs/acta/tidy/feat")
	if strings.Join(got, "|") != "feat: a|feat: b" {
		t.Fatalf("subjects %v", got)
	}
	if r.rev("refs/acta/tidy/feat") != res.Tip {
		t.Fatal("tip mismatch")
	}
	// Ride forward: the early planning file sits in the first new commit.
	first := r.rev(res.Tip + "~1")
	r.git(nil, "cat-file", "-e", first+":.acta/early.md")
	// Fold back: the tick and docs change sit in the first new commit, not the second.
	r.git(nil, "cat-file", "-e", first+":.acta/tick.md")
	if r.git(nil, "show", first+":README.md") != "hi more" {
		t.Fatal("docs change not folded into first commit")
	}
	if r.git(nil, "ls-tree", "-r", "--name-only", first) == r.git(nil, "ls-tree", "-r", "--name-only", res.Tip) {
		t.Fatal("first commit should not hold b.go")
	}
	// Tree equals the branch tree (no hashes in these files, so no remap edits).
	if r.rev(res.Tip+"^{tree}") != r.rev(branchTip+"^{tree}") {
		t.Fatal("final tree differs from branch tree")
	}
	// Nothing moved.
	if r.rev("feat") != branchTip || r.rev("main") != base {
		t.Fatal("a branch moved")
	}
}

func TestReviewMarkerFoldsRest(t *testing.T) {
	for _, marker := range []string{"chore(plan): review notes round 1", "polish: tidy names"} {
		t.Run(marker, func(t *testing.T) {
			r := start(t)
			r.commit("feat: a", d(1), "ann", map[string]string{"a.go": "a\n"})
			r.commit("feat: b", d(2), "ann", map[string]string{"b.go": "b\n"})
			r.commit(marker, d(3), "ann", map[string]string{"b.go": "b2\n"})
			r.commit("fix: review fix", d(4), "ann", map[string]string{"c.go": "c\n"})
			r.commit("feat: late", d(5), "ann", map[string]string{"e.go": "e\n"})

			res, err := Run(r.dir, opt("main", "feat"))
			if err != nil {
				t.Fatal(err)
			}
			got := r.subjects("main..refs/acta/tidy/feat")
			if strings.Join(got, "|") != "feat: a|feat: b" {
				t.Fatalf("subjects %v", got)
			}
			if res.OldCount != 5 || res.NewCount != 2 {
				t.Fatalf("counts %+v", res)
			}
			r.git(nil, "cat-file", "-e", res.Tip+":e.go")
		})
	}
}

func TestReplayClashFoldsIntoPrevious(t *testing.T) {
	r := start(t)
	r.commit("feat: a", d(1), "ann", map[string]string{"a.go": "a\n", "f.txt": "one\n"})
	r.git(nil, "checkout", "-q", "-b", "side")
	r.commit("side: f", d(2), "ann", map[string]string{"f.txt": "side\n"})
	r.git(nil, "checkout", "-q", "feat")
	r.commit("feat: b", d(3), "ann", map[string]string{"b.go": "b\n"})
	// The merge is skipped by the walk, so the replay below never sees "side".
	r.git([]string{"GIT_AUTHOR_DATE=" + d(4), "GIT_COMMITTER_DATE=" + d(4)}, "merge", "-q", "--no-ff", "-m", "merge side", "side")
	// Same lines as the side change: the replay on top of the chain clashes.
	r.commit("feat: c", d(5), "ann", map[string]string{"f.txt": "mine\n", "c.go": "c\n"})
	branchTip := r.rev("feat")

	res, err := Run(r.dir, opt("main", "feat"))
	if err != nil {
		t.Fatal(err)
	}
	got := r.subjects("main..refs/acta/tidy/feat")
	if strings.Join(got, "|") != "feat: a|feat: b" {
		t.Fatalf("subjects %v", got)
	}
	// c is folded into b, so the last commit holds c.go and the exact branch tree.
	if r.rev(res.Tip+"^{tree}") != r.rev(branchTip+"^{tree}") {
		t.Fatal("final tree differs from branch tree")
	}
	if r.git(nil, "show", res.Tip+":f.txt") != "mine" {
		t.Fatal("f.txt wrong")
	}
	// The first new commit stays clear of b and c.
	if strings.Contains(r.git(nil, "ls-tree", "-r", "--name-only", res.Tip+"~1"), "b.go") {
		t.Fatal("first commit holds b.go")
	}
}

func TestDatesNeverGoBackAndLastFoldedWins(t *testing.T) {
	r := start(t)
	r.commit("feat: a", d(5), "ann", map[string]string{"a.go": "a\n"})
	// Folded commit is later than its kept commit: the kept commit takes this date.
	r.commit("chore(plan): tick", d(9), "ann", map[string]string{".acta/t.md": "t\n"})
	// Kept commit with a date older than the one before it.
	r.commit("feat: b", d(3), "ann", map[string]string{"b.go": "b\n"})

	res, err := Run(r.dir, opt("main", "feat"))
	if err != nil {
		t.Fatal(err)
	}
	first := r.git(nil, "log", "-1", "--format=%aI|%cI", res.Tip+"~1")
	last := r.git(nil, "log", "-1", "--format=%aI|%cI", res.Tip)
	want9 := d(9) + "|" + d(9)
	if first != want9 {
		t.Fatalf("first date %s want %s", first, want9)
	}
	if last != want9 {
		t.Fatalf("last date %s must not go before %s", last, want9)
	}
}

func TestMessagesAndAuthorsUnchanged(t *testing.T) {
	r := start(t)
	r.commit("feat: a\n\nbody line\n\nCo-Authored-By: Bob <bob@example.com>", d(1), "zed", map[string]string{"a.go": "a\n"})
	r.commit("feat: b", d(2), "yan", map[string]string{"b.go": "b\n"})
	old := []string{
		r.git(nil, "log", "-1", "--format=%an|%ae|%cn|%ce|%B", "feat~1"),
		r.git(nil, "log", "-1", "--format=%an|%ae|%cn|%ce|%B", "feat"),
	}
	res, err := Run(r.dir, opt("main", "feat"))
	if err != nil {
		t.Fatal(err)
	}
	got := []string{
		r.git(nil, "log", "-1", "--format=%an|%ae|%cn|%ce|%B", res.Tip+"~1"),
		r.git(nil, "log", "-1", "--format=%an|%ae|%cn|%ce|%B", res.Tip),
	}
	if got[0] != old[0] || got[1] != old[1] {
		t.Fatalf("changed:\n%v\n%v", got, old)
	}
}

func TestNoKeptCommitMakesOneCommit(t *testing.T) {
	r := start(t)
	r.commit("chore(plan): one", d(1), "ann", map[string]string{".acta/1.md": "1\n"})
	r.commit("chore(plan): two", d(2), "ann", map[string]string{".acta/2.md": "2\n"})
	res, err := Run(r.dir, opt("main", "feat"))
	if err != nil {
		t.Fatal(err)
	}
	if res.OldCount != 2 || res.NewCount != 1 {
		t.Fatalf("counts %+v", res)
	}
	if got := r.subjects("main..refs/acta/tidy/feat"); len(got) != 1 || got[0] != "chore(plan): two" {
		t.Fatalf("subjects %v", got)
	}
	if r.rev(res.Tip+"^{tree}") != r.rev("feat^{tree}") {
		t.Fatal("tree differs")
	}
}

func TestTreeMismatchWritesNoRef(t *testing.T) {
	r := start(t)
	r.commit("feat: a", d(1), "ann", map[string]string{"a.go": "a\n"})
	tamperTree = func(tree string) string {
		// Hand back a tree that lacks a.go, so the proof must catch it.
		return r.rev("main^{tree}")
	}
	defer func() { tamperTree = nil }()
	branchTip := r.rev("feat")
	if _, err := Run(r.dir, opt("main", "feat")); err == nil {
		t.Fatal("want an error")
	}
	if r.refExists("refs/acta/tidy/feat") {
		t.Fatal("ref written")
	}
	if r.rev("feat") != branchTip {
		t.Fatal("branch moved")
	}
}

func TestOntoFoldsParentChoresForward(t *testing.T) {
	r := newRepo(t)
	r.commit("init", d(0), "ann", map[string]string{"README.md": "hi\n"})
	fold := r.rev("HEAD")
	r.commit("chore(plan): parent one", d(1), "ann", map[string]string{".acta/p1.md": "p1\n"})
	r.commit("chore(spec): parent two", d(2), "ann", map[string]string{".acta/p2.md": "p2\n"})
	r.git(nil, "checkout", "-q", "-b", "feat")
	r.commit("feat: a", d(3), "ann", map[string]string{"a.go": "a\n"})
	o := opt("main", "feat")
	o.Onto = fold
	res, err := Run(r.dir, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.OldCount != 1 || res.NewCount != 1 {
		t.Fatalf("counts %+v", res)
	}
	if r.rev(res.Tip+"^") != fold {
		t.Fatal("chain not built on the fold point")
	}
	r.git(nil, "cat-file", "-e", res.Tip+":.acta/p1.md")
	r.git(nil, "cat-file", "-e", res.Tip+":.acta/p2.md")
	if r.rev(res.Tip+"^{tree}") != r.rev("feat^{tree}") {
		t.Fatal("tree differs")
	}
	if r.git(nil, "log", "-1", "--format=%aI", res.Tip) != d(3) {
		t.Fatal("date should stay the kept commit date")
	}
}

func TestOntoIgnoredForNonChoreParentCommit(t *testing.T) {
	r := newRepo(t)
	r.commit("init", d(0), "ann", map[string]string{"README.md": "hi\n"})
	fold := r.rev("HEAD")
	r.commit("chore(plan): parent one", d(1), "ann", map[string]string{".acta/p1.md": "p1\n"})
	r.commit("feat: parent code", d(2), "ann", map[string]string{"p.go": "p\n"})
	base := r.rev("main")
	r.git(nil, "checkout", "-q", "-b", "feat")
	r.commit("feat: a", d(3), "ann", map[string]string{"a.go": "a\n"})
	o := opt("main", "feat")
	o.Onto = fold
	res, err := Run(r.dir, o)
	if err != nil {
		t.Fatal(err)
	}
	if r.rev(res.Tip+"^") != base {
		t.Fatal("chain must sit on the base")
	}
	if r.rev(res.Tip+"^{tree}") != r.rev("feat^{tree}") {
		t.Fatal("tree differs")
	}
}

func TestCheckGitVersion(t *testing.T) {
	for _, bad := range []string{"git version 2.39.5", "git version 1.9.0"} {
		err := checkVersion(bad)
		if err == nil || !strings.Contains(err.Error(), strings.Fields(bad)[2]) {
			t.Fatalf("%q: %v", bad, err)
		}
	}
	for _, ok := range []string{"git version 2.40.0", "git version 2.56.0 (Apple Git-1)", "git version 3.0.1"} {
		if err := checkVersion(ok); err != nil {
			t.Fatalf("%q: %v", ok, err)
		}
	}
}
