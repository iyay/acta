package board

import (
	"slices"
	"testing"
)

func TestClosesLinksEveryAllowedKind(t *testing.T) {
	b := boardWith(t, map[string]string{
		"scratch/2026-09-29-a.md":      "---\nid: SCRATCH-1\n---\n# A\n",
		"scratch/2026-09-29-b.md":      "# B\n",
		"specs/2026-09-29-t-design.md": "# T\n",
		"debt/2026-09-29-d.md":         "---\nid: DEBT-1\n---\n# D\n\n- [ ] one\n",
		"specs/2026-09-29-s-design.md": "---\ncloses: [SCR-0001, scratch/2026-09-29-b, specs/2026-09-29-t-design, DBT-0001.01]\n---\n# S\n",
	})
	s := b.Get("specs/2026-09-29-s-design")
	want := []string{"scratch/2026-09-29-a", "scratch/2026-09-29-b", "specs/2026-09-29-t-design", "debt/2026-09-29-d#item-1"}
	if !slices.Equal(s.Closes, want) || len(s.Problems) != 0 {
		t.Fatalf("closes %v problems %v, want %v and none", s.Closes, s.Problems, want)
	}
	for _, id := range want {
		if got := b.Get(id).ClosedBy; !slices.Equal(got, []string{s.ID}) {
			t.Errorf("%s closed by %v, want [%s]", id, got, s.ID)
		}
	}
	if st := b.Get("scratch/2026-09-29-b").Status; st != "specced" {
		t.Errorf("closed scratch = %s, want specced", st)
	}
	if st := b.Get("debt/2026-09-29-d#item-1").Status; st != "todo" && st != "open" {
		t.Errorf("debt item changed to %s from closes", st)
	}
}

func TestClosesBadEntriesAreProblems(t *testing.T) {
	b := boardWith(t, map[string]string{
		"plans/2026-09-29-q.md":        "# Q\n\n### Task 1: A\n\n- [ ] a\n",
		"bugs/2026-09-29-x.md":         "# X\n",
		"specs/2026-09-29-s-design.md": "---\ncloses: [SCRATCH-99, plans/2026-09-29-q, plans/2026-09-29-q#task-1, bugs/2026-09-29-x]\n---\n# S\n",
	})
	want := []string{
		"closes SCRATCH-99 not found",
		"closes plans/2026-09-29-q: a plan cannot be closed",
		"closes plans/2026-09-29-q#task-1: a task cannot be closed",
		"closes bugs/2026-09-29-x: a bug cannot be closed",
	}
	if got := b.Get("specs/2026-09-29-s-design").Problems; !slices.Equal(got, want) {
		t.Errorf("problems %v, want %v", got, want)
	}
}

func TestClosesSingleStringLinksLikeAList(t *testing.T) {
	b := boardWith(t, map[string]string{
		"scratch/2026-09-29-a.md":      "---\nid: SCRATCH-1\n---\n# A\n",
		"specs/2026-09-29-s-design.md": "---\ncloses: SCR-0001\n---\n# S\n",
	})
	s := b.Get("specs/2026-09-29-s-design")
	want := []string{"scratch/2026-09-29-a"}
	if !slices.Equal(s.Closes, want) || len(s.Problems) != 0 {
		t.Fatalf("closes %v problems %v, want %v and none", s.Closes, s.Problems, want)
	}
	if got := b.Get("scratch/2026-09-29-a").ClosedBy; !slices.Equal(got, []string{s.ID}) {
		t.Errorf("closed by %v, want [%s]", got, s.ID)
	}
}

func TestClosesResolvesAHashID(t *testing.T) {
	// A hash is the one name an item keeps for good, so a closes list may use
	// it like any other id.
	b := boardWith(t, map[string]string{
		"scratch/2026-09-29-a.md":      "---\nid: SCRATCH-1\nhash: k3f2\n---\n# A\n",
		"specs/2026-09-29-s-design.md": "---\ncloses: [SCR-k3f2]\n---\n# S\n",
	})
	s := b.Get("specs/2026-09-29-s-design")
	want := []string{"scratch/2026-09-29-a"}
	if !slices.Equal(s.Closes, want) || len(s.Problems) != 0 {
		t.Fatalf("closes %v problems %v, want %v and none", s.Closes, s.Problems, want)
	}
	if got := b.Get("scratch/2026-09-29-a").ClosedBy; !slices.Equal(got, []string{s.ID}) {
		t.Errorf("closed by %v, want [%s]", got, s.ID)
	}
}

func TestClosesEmptyListLinksNothing(t *testing.T) {
	b := boardWith(t, map[string]string{
		"scratch/2026-09-29-a.md":      "---\nid: SCRATCH-1\n---\n# A\n",
		"specs/2026-09-29-s-design.md": "---\ncloses: []\n---\n# S\n",
	})
	s := b.Get("specs/2026-09-29-s-design")
	if len(s.Closes) != 0 || len(s.Problems) != 0 {
		t.Errorf("closes %v problems %v, want none and none", s.Closes, s.Problems)
	}
	if got := b.Get("scratch/2026-09-29-a").ClosedBy; len(got) != 0 {
		t.Errorf("closed by %v, want nothing", got)
	}
}

func TestClosesSameIDTwiceLinksOnce(t *testing.T) {
	b := boardWith(t, map[string]string{
		"scratch/2026-09-29-a.md":      "---\nid: SCRATCH-1\n---\n# A\n",
		"specs/2026-09-29-s-design.md": "---\ncloses: [SCR-0001, scratch/2026-09-29-a]\n---\n# S\n",
	})
	s := b.Get("specs/2026-09-29-s-design")
	want := []string{"scratch/2026-09-29-a"}
	if !slices.Equal(s.Closes, want) || len(s.Problems) != 0 {
		t.Fatalf("closes %v problems %v, want %v and none", s.Closes, s.Problems, want)
	}
	if got := b.Get("scratch/2026-09-29-a").ClosedBy; !slices.Equal(got, []string{s.ID}) {
		t.Errorf("closed by %v, want [%s]", got, s.ID)
	}
}

func TestClosesDroppedScratchStaysDropped(t *testing.T) {
	b := boardWith(t, map[string]string{
		"scratch/2026-09-29-d.md":      "---\nid: SCRATCH-2\nstatus: dropped\n---\n# D\n",
		"specs/2026-09-29-s-design.md": "---\ncloses: [SCR-0002]\n---\n# S\n",
	})
	d := b.Get("scratch/2026-09-29-d")
	if len(d.ClosedBy) != 1 {
		t.Fatalf("dropped scratch closed by %v, want the one spec", d.ClosedBy)
	}
	if d.Status != "dropped" || d.StatusSource != "frontmatter" {
		t.Errorf("dropped scratch = %s from %s, want dropped from frontmatter", d.Status, d.StatusSource)
	}
}

func TestClosesSpecCountsPlanAndItsTasks(t *testing.T) {
	b := boardWith(t, map[string]string{
		"specs/2026-09-29-s-design.md": "# S\n",
		"plans/2026-09-29-q.md":        "---\ncloses: [specs/2026-09-29-s-design]\n---\n# Q\n\n### Task 1: A\n\n- [x] a\n",
	})
	s := b.Get("specs/2026-09-29-s-design")
	if got := b.Get("plans/2026-09-29-q").Closes; !slices.Equal(got, []string{s.ID}) {
		t.Errorf("plan closes %v, want [%s]", got, s.ID)
	}
	if s.plans != 1 {
		t.Errorf("spec counts %d plans, want 1", s.plans)
	}
	want := []string{"plans/2026-09-29-q#task-1"}
	if !slices.Equal(s.Children, want) {
		t.Errorf("spec children %v, want %v", s.Children, want)
	}
	if s.Status != "done" {
		t.Errorf("spec status %s, want done with its only task ticked", s.Status)
	}
}

func TestClosesSpecFollowsTheBugThatClosesIt(t *testing.T) {
	// The bug is the item that finishes the spec, so the bug writes closes:
	// and the spec follows the bug's status.
	files := map[string]string{
		"specs/2026-09-29-s-design.md": "# S\n",
		"bugs/2026-09-29-x.md":         "---\ncloses: [specs/2026-09-29-s-design]\n---\n# X\n",
	}
	cases := map[string]string{"open": "approved", "fixing": "in-progress", "fixed": "done"}
	for status, want := range cases {
		files["bugs/2026-09-29-x.md"] = "---\nstatus: " + status + "\ncloses: [specs/2026-09-29-s-design]\n---\n# X\n"
		b := boardWith(t, files)
		if got := b.Get("specs/2026-09-29-s-design").Status; got != want {
			t.Errorf("bug %s gives spec %s, want %s", status, got, want)
		}
	}
}

func TestClosesSpecFollowsASpecThatClosesIt(t *testing.T) {
	// A spec may close a spec, so the answer can depend on a spec whose own
	// answer is not settled yet. The board reads the bug before the specs,
	// but the specs sort the other way round from each other.
	files := map[string]string{
		"bugs/2026-09-29-x.md":         "---\nstatus: fixed\ncloses: [specs/2026-09-29-a-design]\n---\n# X\n",
		"specs/2026-09-29-a-design.md": "---\ncloses: [specs/2026-09-29-b-design]\n---\n# A\n",
		"specs/2026-09-29-b-design.md": "# B\n",
	}
	b := boardWith(t, files)
	for _, id := range []string{"specs/2026-09-29-a-design", "specs/2026-09-29-b-design"} {
		if got := b.Get(id).Status; got != "done" {
			t.Errorf("%s = %s, want done all the way down the chain", id, got)
		}
	}
}

func TestClosesSpecWithAPlanOfItsOwnIgnoresItsClosers(t *testing.T) {
	// A spec with a plan of its own reads that plan, so a bug closing the
	// spec as well must not overwrite what the plan's boxes say.
	b := boardWith(t, map[string]string{
		"specs/2026-09-29-s-design.md": "# S\n",
		"plans/2026-09-29-q.md":        "# Q\n\n**Spec:** `.acta/specs/2026-09-29-s-design.md`\n\n### Task 1: A\n\n- [ ] a\n",
		"bugs/2026-09-29-x.md":         "---\nstatus: fixed\ncloses: [specs/2026-09-29-s-design]\n---\n# X\n",
	})
	s := b.Get("specs/2026-09-29-s-design")
	if len(s.ClosedBy) != 1 {
		t.Fatalf("spec closed by %v, want the one bug", s.ClosedBy)
	}
	if s.Status != "approved" {
		t.Errorf("spec = %s, want approved from its plan with no box ticked", s.Status)
	}
}

func TestClosesSpecWithAWrittenStatusKeepsIt(t *testing.T) {
	// A spec with no plan of its own may still carry a written status, and a
	// closer does not take that away.
	b := boardWith(t, map[string]string{
		"specs/2026-09-29-s-design.md": "---\nstatus: draft\ncloses: [bugs/2026-09-29-x]\n---\n# S\n",
		"bugs/2026-09-29-x.md":         "---\nstatus: fixed\ncloses: [specs/2026-09-29-s-design]\n---\n# X\n",
	})
	s := b.Get("specs/2026-09-29-s-design")
	if len(s.ClosedBy) != 1 {
		t.Fatalf("spec closed by %v, want the one bug", s.ClosedBy)
	}
	if s.Status != "draft" || s.StatusSource != "frontmatter" {
		t.Errorf("spec = %s from %s, want draft from frontmatter", s.Status, s.StatusSource)
	}
}

func TestClosesSpecChainSettlesFromTheEnd(t *testing.T) {
	// The spec at the end of the chain is read first, since the board sorts
	// by date and keeps file order inside one date, so the answer has to be
	// reached in more than one go.
	files := map[string]string{
		"bugs/2026-09-29-x.md":         "---\nstatus: fixed\ncloses: [specs/2026-09-29-z-design]\n---\n# X\n",
		"specs/2026-09-29-z-design.md": "---\ncloses: [specs/2026-09-29-y-design]\n---\n# Z\n",
		"specs/2026-09-29-y-design.md": "---\ncloses: [specs/2026-09-29-a-design]\n---\n# Y\n",
		"specs/2026-09-29-a-design.md": "# A\n",
	}
	b := boardWith(t, files)
	for _, id := range []string{"specs/2026-09-29-a-design", "specs/2026-09-29-y-design", "specs/2026-09-29-z-design"} {
		if got := b.Get(id).Status; got != "done" {
			t.Errorf("%s = %s, want done all the way down the chain", id, got)
		}
	}
}

func TestClosesOnAScratchFileIsIgnored(t *testing.T) {
	b := boardWith(t, map[string]string{
		"scratch/2026-09-29-a.md": "---\nid: SCRATCH-1\n---\n# A\n",
		"scratch/2026-09-29-w.md": "---\nid: SCRATCH-2\ncloses: [SCR-0001]\n---\n# W\n",
		"debt/2026-09-29-d.md":    "---\nid: DEBT-1\ncloses: [SCR-0001]\n---\n# D\n\n- [ ] one\n",
		"plans/2026-09-29-q.md":   "# Q\n\n### Task 1: A\n\n- [ ] a\n",
	})
	for _, id := range []string{"scratch/2026-09-29-w", "debt/2026-09-29-d", "plans/2026-09-29-q#task-1"} {
		if got := b.Get(id).Closes; len(got) != 0 {
			t.Errorf("%s closes %v, want nothing", id, got)
		}
	}
	a := b.Get("scratch/2026-09-29-a")
	if len(a.ClosedBy) != 0 {
		t.Errorf("scratch closed by %v, want nothing", a.ClosedBy)
	}
	for _, id := range []string{"scratch/2026-09-29-w", "debt/2026-09-29-d"} {
		if got := b.Get(id).Problems; len(got) != 0 {
			t.Errorf("%s problems %v, want none", id, got)
		}
	}
}

// A plan can point at a spec three ways: a parent: in frontmatter, a Spec
// line in the body, and a closes: list. However many of them name the same
// spec, the spec counts the plan once and holds each of its tasks once, so
// the progress reads right and a task is never shown twice.
func TestSpecCountsEachPlanOnce(t *testing.T) {
	const (
		specS = "specs/2026-09-29-s-design"
		specT = "specs/2026-09-29-t-design"
		debtD = "debt/2026-09-29-d"
		planP = "plans/2026-09-29-p"
		planR = "plans/2026-09-29-r"
	)
	pTasks := []string{planP + "#task-1", planP + "#task-2"}
	rTasks := []string{planR + "#task-1", planR + "#task-2"}
	// plan builds a plan file with two tasks, the first one ticked, so a spec
	// that counts the plan twice reads 2/4 instead of 1/2.
	plan := func(parent, specLine, closes string) string {
		fm := ""
		if parent != "" || closes != "" {
			fm = "---\n"
			if parent != "" {
				fm += "parent: " + parent + "\n"
			}
			if closes != "" {
				fm += "closes: [" + closes + "]\n"
			}
			fm += "---\n"
		}
		body := "# P\n\n"
		if specLine != "" {
			body += "**Spec:** `.acta/specs/" + specLine + ".md`\n\n"
		}
		return fm + body + "### Task 1: One\n\n- [x] one\n\n### Task 2: Two\n\n- [ ] two\n"
	}
	base := func() map[string]string {
		return map[string]string{
			specS + ".md": "# S\n",
			specT + ".md": "# T\n",
			debtD + ".md": "# D\n\n- [ ] one\n",
		}
	}
	type want struct {
		plans    int
		children []string
		done     int
		total    int
		status   string
	}
	// onePlan is what a spec looks like when a single plan points at it once:
	// one plan, its two tasks once each, one box ticked.
	onePlan := func() want { return want{1, pTasks, 1, 2, "in-progress"} }
	cases := []struct {
		name  string
		files func() map[string]string
		want  map[string]want
	}{
		{
			name:  "parent names the spec",
			files: func() map[string]string { f := base(); f[planP+".md"] = plan(specS, "", ""); return f },
			want:  map[string]want{specS: onePlan()},
		},
		{
			name:  "Spec line only",
			files: func() map[string]string { f := base(); f[planP+".md"] = plan("", "2026-09-29-s-design", ""); return f },
			want:  map[string]want{specS: onePlan()},
		},
		{
			name:  "closes only",
			files: func() map[string]string { f := base(); f[planP+".md"] = plan("", "", specS); return f },
			want:  map[string]want{specS: onePlan()},
		},
		{
			name: "Spec line and closes",
			files: func() map[string]string {
				f := base()
				f[planP+".md"] = plan("", "2026-09-29-s-design", specS)
				return f
			},
			want: map[string]want{specS: onePlan()},
		},
		{
			name:  "parent and closes",
			files: func() map[string]string { f := base(); f[planP+".md"] = plan(specS, "", specS); return f },
			want:  map[string]want{specS: onePlan()},
		},
		{
			name: "Spec line, debt parent and closes",
			files: func() map[string]string {
				f := base()
				f[planP+".md"] = plan(debtD, "2026-09-29-s-design", specS)
				return f
			},
			want: map[string]want{specS: onePlan()},
		},
		{
			name:  "debt parent and closes",
			files: func() map[string]string { f := base(); f[planP+".md"] = plan(debtD, "", specS); return f },
			want:  map[string]want{specS: onePlan()},
		},
		{
			name:  "closes names the spec twice",
			files: func() map[string]string { f := base(); f[planP+".md"] = plan("", "", specS+", "+specS); return f },
			want:  map[string]want{specS: onePlan()},
		},
		{
			name: "Spec line names one spec, closes another",
			files: func() map[string]string {
				f := base()
				f[planP+".md"] = plan("", "2026-09-29-s-design", specT)
				return f
			},
			want: map[string]want{specS: onePlan(), specT: onePlan()},
		},
		{
			name: "parent and Spec line name the same spec",
			files: func() map[string]string {
				f := base()
				f[planP+".md"] = plan(specS, "2026-09-29-s-design", "")
				return f
			},
			want: map[string]want{specS: onePlan()},
		},
		{
			name: "parent, Spec line and closes all name the same spec",
			files: func() map[string]string {
				f := base()
				f[planP+".md"] = plan(specS, "2026-09-29-s-design", specS)
				return f
			},
			want: map[string]want{specS: onePlan()},
		},
		{
			name: "two plans, one by Spec line and one by closes",
			files: func() map[string]string {
				f := base()
				f[planP+".md"] = plan("", "2026-09-29-s-design", "")
				f[planR+".md"] = plan("", "", specS)
				return f
			},
			want: map[string]want{specS: {2, append(append([]string{}, pTasks...), rTasks...), 2, 4, "in-progress"}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := boardWith(t, c.files())
			for id, w := range c.want {
				it := b.Get(id)
				if it == nil {
					t.Fatalf("%s missing from the board", id)
				}
				if it.plans != w.plans {
					t.Errorf("counts %d plans, want %d", it.plans, w.plans)
				}
				if it.Done != w.done || it.Total != w.total {
					t.Errorf("progress %d/%d, want %d/%d", it.Done, it.Total, w.done, w.total)
				}
				if it.Status != w.status {
					t.Errorf("status %s, want %s", it.Status, w.status)
				}
				for _, task := range w.children {
					if !slices.Contains(it.Children, task) {
						t.Errorf("holds no %s, want every task once", task)
					}
				}
				// Every wanted task is there and the list is no longer, so no
				// task sits in it twice.
				if len(it.Children) != len(w.children) {
					t.Errorf("children %v, want %v", it.Children, w.children)
				}
			}
		})
	}
}
