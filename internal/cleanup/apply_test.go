package cleanup

import (
	"path/filepath"
	"strings"
	"testing"
)

var applyFiles = map[string]string{
	".acta/plans/2026-01-01-base.md":     "---\nid: PLN-0001\nhash: aaaaaaa\n---\n# Base\n\n### Task 1: A\n- [x] a\n",
	".acta/debt/2026-01-02-notes.md":     "---\nid: DBT-0001\nparent: plans/2026-01-01-base\n---\n# Notes\n\n- [ ] (medium) first\n- [ ] (high) second\n- [ ] (medium) third\n",
	".acta/scratch/2026-01-03-idea.md":   "---\nid: SCR-0001\nstatus: raw\n---\n# Idea\n",
	".acta/bugs/2026-01-04-crash.md":     "---\nid: BUG-0001\n---\n# Crash\n",
	".acta/specs/2026-01-05-alpha.md":    "---\nid: SPC-0001\n---\n# Alpha\n",
	".acta/plans/2026-01-05-alpha.md":    "---\nid: PLN-0002\nhash: bbbbbbb\nparent: specs/2026-01-05-alpha\n---\n# Alpha plan\n\n### Task 1: A\n- [x] a\n\n### Task 2: B\n- [ ] b\n",
	".acta/specs/2026-01-06-beta.md":     "---\nid: SPC-0002\n---\n# Beta\n",
	".acta/plans/2026-01-06-beta-one.md": "---\nid: PLN-0003\nhash: ccccccc\nparent: specs/2026-01-06-beta\n---\n# Beta one\n\n### Task 1: A\n- [x] a\n\n### Task 2: B\n- [ ] b\n",
	".acta/plans/2026-01-06-beta-two.md": "---\nid: PLN-0004\nhash: ddddddd\nparent: specs/2026-01-06-beta\n---\n# Beta two\n\n### Task 1: A\n- [ ] a\n",
}

func (r testRepo) apply(stdin string) error {
	r.t.Helper()
	ds, err := ParseDecisions(strings.NewReader(stdin))
	if err != nil {
		return err
	}
	cfg, b := r.load()
	_, err = Apply(cfg, b, ds)
	return err
}

func TestApplyWritesTheBatchInOneCommit(t *testing.T) {
	t.Parallel()
	r := newRepo(t, 0, applyFiles)
	err := r.apply(`{"id":"DBT-0001.01","status":"done","reason":"done in 4dcc44c"}
{"id":"debt/2026-01-02-notes#item-3","status":"wontfix","reason":"stale"}
{"id":"SCR-0001","status":"done","reason":"built in 4dcc44c"}
{"id":"BUG-0001","status":"fixed","fixed_in":"cb1f58b","reason":"runner stops at the final reply"}
{"id":"PLN-0002","status":"dropped","reason":"no work for 30 days"}
{"id":"PLN-0003","status":"dropped","reason":"merged into PLN-0004"}
`)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.out("rev-list", "--count", "HEAD"); got != "2" {
		t.Fatalf("commits = %s, want one new commit", got)
	}
	if got := r.out("status", "--porcelain"); got != "" {
		t.Fatalf("left changes: %s", got)
	}
	// Six named items plus the spec that goes with PLN-0002.
	if got := r.out("log", "-1", "--format=%s"); !strings.Contains(got, "7 items") {
		t.Fatalf("subject = %q, want the 7 items written", got)
	}
	_, b := r.load()
	checks := []struct{ id, status, reason string }{
		{"SCR-0001", "done", "built in 4dcc44c"},
		{"BUG-0001", "fixed", "runner stops at the final reply"},
		{"PLN-0002", "dropped", "no work for 30 days"},
		// The spec had only this plan, so it goes with it.
		{"SPC-0001", "dropped", "plan PLN-0002 dropped: no work for 30 days"},
		{"PLN-0003", "dropped", "merged into PLN-0004"},
	}
	for _, c := range checks {
		it := b.Get(c.id)
		if it == nil || it.Status != c.status || it.Reason != c.reason {
			t.Errorf("%s = %+v, want %s %q", c.id, it, c.status, c.reason)
		}
	}
	if got := b.Get("BUG-0001").FixedIn; got != "cb1f58b" {
		t.Errorf("fixed_in = %q", got)
	}
	// Beta still has a live plan, so it stays open.
	if got := b.Get("SPC-0002"); got.Status == "dropped" || got.Reason != "" {
		t.Errorf("SPC-0002 = %s %q, want it left open", got.Status, got.Reason)
	}
	notes := r.read(".acta/debt/2026-01-02-notes.md")
	for _, line := range []string{"- [x] (medium) first (done: done in 4dcc44c)\n", "- [ ] (high) second\n", "- [-] (medium) third (wontfix: stale)\n"} {
		if !strings.Contains(notes, line) {
			t.Errorf("debt file lacks %q:\n%s", line, notes)
		}
	}
}

// One bad line stops the whole batch before any file is written, and the error
// names the line, so a batch never lands half done.
func TestApplyRefusesTheWholeBatch(t *testing.T) {
	t.Parallel()
	good := `{"id":"SCR-0001","status":"done","reason":"built in 4dcc44c"}` + "\n"
	cases := []struct{ name, bad string }{
		{"malformed json", `{"id":"BUG-0001",`},
		{"two objects on one line", `{"id":"BUG-0001","status":"wontfix","reason":"x"} {"id":"PLN-0002","status":"dropped","reason":"y"}`},
		{"stray brace after the object", `{"id":"BUG-0001","status":"wontfix","reason":"x"}}`},
		{"unknown field", `{"id":"BUG-0001","status":"wontfix","reason":"x","owner":"rian"}`},
		{"unknown id", `{"id":"BUG-0099","status":"wontfix","reason":"x"}`},
		{"status not for the kind", `{"id":"BUG-0001","status":"dropped","reason":"x"}`},
		{"keep is not a write", `{"id":"BUG-0001","status":"keep","reason":"x"}`},
		{"missing reason", `{"id":"BUG-0001","status":"wontfix"}`},
		{"multi-line reason", `{"id":"BUG-0001","status":"wontfix","reason":"a\nb"}`},
		{"fixed without sha", `{"id":"BUG-0001","status":"fixed","reason":"x"}`},
		{"bad sha", `{"id":"BUG-0001","status":"fixed","fixed_in":"XYZ","reason":"x"}`},
		{"sha on a wontfix", `{"id":"BUG-0001","status":"wontfix","fixed_in":"cb1f58b","reason":"x"}`},
		{"same id twice", good[:len(good)-1]},
		{"debt line as fixed", `{"id":"DBT-0001.01","status":"fixed","fixed_in":"cb1f58b","reason":"x"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := newRepo(t, 0, applyFiles)
			err := r.apply(good + c.bad + "\n")
			if err == nil || !strings.Contains(err.Error(), "line 2") {
				t.Fatalf("err = %v, want one naming line 2", err)
			}
			if got := r.out("status", "--porcelain"); got != "" {
				t.Fatalf("refused batch changed files: %s", got)
			}
			if got := r.out("rev-list", "--count", "HEAD"); got != "1" {
				t.Fatalf("refused batch made a commit")
			}
		})
	}
}

// A plan a linked worktree is building cannot be dropped from main.
func TestApplyRefusesABuildingPlan(t *testing.T) {
	t.Parallel()
	r := newRepo(t, 0, applyFiles)
	wt := filepath.Join(t.TempDir(), "wt")
	r.git(0, "worktree", "add", "-q", wt, "-b", "building")
	plan := ".acta/plans/2026-01-05-alpha.md"
	testRepo{t, wt}.commit(0, "chore(plan): tick", map[string]string{plan: applyFiles[plan] + "\n"})
	err := r.apply(`{"id":"PLN-0002","status":"dropped","reason":"x"}` + "\n")
	if err == nil || !strings.Contains(err.Error(), "line 1") || !strings.Contains(err.Error(), "worktree") {
		t.Fatalf("err = %v", err)
	}
	if got := r.out("status", "--porcelain"); got != "" {
		t.Fatalf("refused batch changed files: %s", got)
	}
}

// A spec in a legacy folder is read-only to acta, so dropping its plan leaves
// that spec as it is.
func TestApplyLeavesALegacySpecAlone(t *testing.T) {
	t.Parallel()
	spec := "docs/superpowers/specs/2026-01-07-old-design.md"
	r := newRepo(t, 0, map[string]string{
		spec:                              "# Old\n",
		".acta/plans/2026-01-07-gamma.md": "---\nid: PLN-0005\nhash: eeeeeee\n---\n# Gamma\n\n**Spec:** `docs/superpowers/specs/2026-01-07-old-design.md`\n\n### Task 1: A\n- [ ] a\n",
	})
	if _, b := r.load(); b.Get("PLN-0005").SpecID == "" {
		t.Fatal("the plan does not point at the legacy spec, so this test proves nothing")
	}
	if err := r.apply(`{"id":"PLN-0005","status":"dropped","reason":"gone"}` + "\n"); err != nil {
		t.Fatal(err)
	}
	if got := r.read(spec); got != "# Old\n" {
		t.Fatalf("legacy spec changed:\n%s", got)
	}
}
