package board

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
)

func loadFixture(t *testing.T) *Board {
	t.Helper()
	dir, err := filepath.Abs("testdata/basic")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Load(config.Default(dir))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func ids(items []*Item) []string {
	out := []string{}
	for _, it := range items {
		out = append(out, it.ID)
	}
	return out
}

func TestLoadDerivesEveryItem(t *testing.T) {
	t.Parallel()

	b := loadFixture(t)
	type want struct {
		kind           Kind
		status, source string
		done, total    int
	}
	cases := map[string]want{
		"specs/2026-09-20-alpha":                 {KindStory, "in-progress", "derived", 1, 2},
		"plans/2026-09-21-alpha":                 {KindPlan, "in-progress", "derived", 1, 2},
		"plans/2026-09-21-alpha#task-1":          {KindTask, "done", "derived", 2, 2},
		"plans/2026-09-21-alpha#task-2":          {KindTask, "in-progress", "derived", 1, 2},
		"specs/2026-09-22-beta":                  {KindStory, "draft", "derived", 0, 0},
		"plans/2026-09-23-lonely":                {KindPlan, "approved", "derived", 0, 1},
		"plans/2026-09-23-lonely#task-1":         {KindTask, "todo", "derived", 0, 0},
		"bugs/2026-09-24-crash":                  {KindBug, "fixed", "derived", 1, 1},
		"plans/2026-09-25-crash-fix":             {KindPlan, "done", "derived", 1, 1},
		"plans/2026-09-25-crash-fix#task-F1":     {KindTask, "done", "derived", 2, 2},
		"plans/2026-09-28-dotted-tasks":          {KindPlan, "done", "derived", 2, 2},
		"plans/2026-09-28-dotted-tasks#task-2.1": {KindTask, "done", "derived", 2, 2},
		"plans/2026-09-28-dotted-tasks#task-2.2": {KindTask, "done", "derived", 2, 2},
		"plans/2026-09-26-dash-tasks":            {KindPlan, "done", "derived", 2, 2},
		"plans/2026-09-26-dash-tasks#task-F-1":   {KindTask, "done", "derived", 2, 2},
		"plans/2026-09-26-dash-tasks#task-F-2":   {KindTask, "done", "derived", 2, 2},
		"bugs/2026-09-26-open":                   {KindBug, "open", "derived", 0, 0},
		"plans/2026-09-27-orphan":                {KindPlan, "done", "derived", 1, 1},
		"specs/2026-09-19-dropped":               {KindStory, "dropped", "frontmatter", 0, 0},
		"specs/2026-09-18-weird":                 {KindStory, "bogus", "frontmatter", 0, 0},
		"specs/2026-09-17-broken":                {KindStory, "draft", "derived", 0, 0},
		"specs/2026-09-16-finished":              {KindStory, "done", "derived", 1, 1},
		"plans/2026-09-16-finished":              {KindPlan, "done", "derived", 1, 1},
		"specs/2026-09-15-really-bug":            {KindBug, "open", "derived", 0, 0},
		"docs/superpowers/specs/2026-01-01-old":  {KindStory, "approved", "derived", 0, 1},
		"docs/superpowers/plans/2026-01-02-old":  {KindPlan, "approved", "derived", 0, 1},
		"specs/2026-09-28-from-scratch-design":   {KindStory, "draft", "derived", 0, 0},
		"scratch/2026-09-28-idea-raw":            {KindScratch, "raw", "frontmatter", 0, 0},
		"scratch/2026-09-28-idea-used":           {KindScratch, "specced", "derived", 0, 1},
		"scratch/2026-09-28-idea-dropped":        {KindScratch, "dropped", "frontmatter", 0, 0},
		"scratch/2026-09-28-idea-brainstorm":     {KindScratch, "brainstorming", "frontmatter", 0, 0},
		"bugs/2026-09-28-lag":                    {KindBug, "fixing", "frontmatter", 0, 0},
	}
	for id, w := range cases {
		it := b.Get(id)
		if it == nil {
			t.Errorf("%s: missing", id)
			continue
		}
		if it.Kind != w.kind || it.Status != w.status || it.StatusSource != w.source || it.Done != w.done || it.Total != w.total {
			t.Errorf("%s = kind %s status %s (%s) %d/%d, want %+v", id, it.Kind, it.Status, it.StatusSource, it.Done, it.Total, w)
		}
	}
	if len(b.Items) != 37 {
		t.Errorf("got %d items, want 37: %v", len(b.Items), ids(b.Items))
	}
}

// A task that has begun reads in-progress, whichever way it began: one ticked
// box of three, or a started record and nothing ticked yet.
func TestTaskUnderWayReadsInProgress(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"plans/2026-09-21-a.md": "# A plan\n\n### Task 1: One\n- [x] a\n- [ ] b\n- [ ] c\n",
	})
	it := b.Get("plans/2026-09-21-a#task-1")
	if it == nil || it.Status != "in-progress" || it.Done != 1 || it.Total != 3 {
		t.Fatalf("half ticked task = %+v, want in-progress at 1 of 3", it)
	}
}

// The word doing is gone for good, so no item the board holds can read it.
func TestNoItemOnTheBoardReadsDoing(t *testing.T) {
	t.Parallel()

	for _, it := range loadFixture(t).Items {
		if it.Status == "doing" {
			t.Errorf("%s reads doing, want in-progress", it.ID)
		}
	}
}

func TestLoadLinksPlans(t *testing.T) {
	t.Parallel()

	b := loadFixture(t)
	parents := map[string]string{
		"plans/2026-09-21-alpha#task-1":                "specs/2026-09-20-alpha",
		"plans/2026-09-25-crash-fix#task-F1":           "bugs/2026-09-24-crash",
		"plans/2026-09-23-lonely#task-1":               "plans/2026-09-23-lonely",
		"plans/2026-09-16-finished#task-1":             "specs/2026-09-16-finished",
		"docs/superpowers/plans/2026-01-02-old#task-1": "docs/superpowers/specs/2026-01-01-old",
	}
	for task, parent := range parents {
		if got := b.Get(task).Parent; got != parent {
			t.Errorf("%s parent = %s, want %s", task, got, parent)
		}
	}
	if got := b.Get("specs/2026-09-20-alpha").Children; !reflect.DeepEqual(got, []string{"plans/2026-09-21-alpha#task-1", "plans/2026-09-21-alpha#task-2"}) {
		t.Errorf("alpha children = %v", got)
	}
	if b.Get("plans/2026-09-21-alpha#task-9") != nil {
		t.Error("a task heading inside a code block became a task")
	}
	task := b.Get("plans/2026-09-25-crash-fix#task-F1")
	if task.Title != "(be): Guard the nil config" || task.Line != 6 || task.Ref != "B-1" {
		t.Errorf("F1 = title %q line %d ref %q", task.Title, task.Line, task.Ref)
	}
}

// Every plan file is its own item, with or without a spec, and a spec keeps
// the progress and status it had while the plan folded into it.
func TestPlansAreItems(t *testing.T) {
	t.Parallel()

	b := loadFixture(t)
	alpha := b.Get("plans/2026-09-21-alpha")
	if alpha == nil || alpha.Kind != KindPlan {
		t.Fatalf("alpha plan = %v", alpha)
	}
	if alpha.SpecID != "specs/2026-09-20-alpha" {
		t.Errorf("alpha plan spec = %q, want the spec it names", alpha.SpecID)
	}
	want := []string{"plans/2026-09-21-alpha#task-1", "plans/2026-09-21-alpha#task-2"}
	if !reflect.DeepEqual(alpha.Children, want) {
		t.Errorf("alpha plan children = %v, want %v", alpha.Children, want)
	}
	lonely := b.Get("plans/2026-09-23-lonely")
	if lonely == nil || lonely.Kind != KindPlan {
		t.Fatalf("lonely plan = %v", lonely)
	}
	if lonely.SpecID != "" {
		t.Errorf("lonely plan spec = %q, want empty", lonely.SpecID)
	}
	spec := b.Get("specs/2026-09-20-alpha")
	if spec.Done != 1 || spec.Total != 2 || spec.Status != "in-progress" {
		t.Errorf("alpha spec = %d/%d %s, want 1/2 in-progress", spec.Done, spec.Total, spec.Status)
	}
	for _, it := range b.List(KindTask, true) {
		if it.PlanID == "" {
			t.Errorf("task %s names no plan", it.ID)
		}
	}
}

// TestPlanSpecLineLinksBesideParent: a plan that works debt once hid its
// spec, so the spec stayed draft after the plan landed (SPEC-16).
func TestPlanSpecLineLinksBesideParent(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"specs/2026-09-29-s-design.md": "# S\n",
		"debt/2026-09-29-d.md":         "# D\n\n- [ ] one\n",
		"plans/2026-09-29-p.md":        "---\nparent: debt/2026-09-29-d\n---\n# P\n\n**Spec:** `.acta/specs/2026-09-29-s-design.md`\n\n### Task 1: A\n\n- [x] a\n",
	})
	spec := b.Get("specs/2026-09-29-s-design")
	plan := b.Get("plans/2026-09-29-p")
	if spec.Status != "done" || spec.Total != 1 {
		t.Errorf("spec = %s %d/%d, want done 1/1", spec.Status, spec.Done, spec.Total)
	}
	if plan.SpecID != "debt/2026-09-29-d" {
		t.Errorf("plan sits under %q, want the debt parent", plan.SpecID)
	}
	if len(plan.Problems) != 0 {
		t.Errorf("problems %v, want none", plan.Problems)
	}
}

// A spec path that names nothing is a problem on the plan, but the plan still
// hangs under the parent it names in frontmatter.
func TestPlanSpecLineMissingBesideParent(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"debt/2026-09-29-d.md":  "# D\n\n- [ ] one\n",
		"plans/2026-09-29-p.md": "---\nparent: debt/2026-09-29-d\n---\n# P\n\n**Spec:** `.acta/specs/missing.md`\n\n### Task 1: A\n\n- [x] a\n",
	})
	plan := b.Get("plans/2026-09-29-p")
	if plan.SpecID != "debt/2026-09-29-d" {
		t.Errorf("plan sits under %q, want the debt parent", plan.SpecID)
	}
	want := []string{"spec .acta/specs/missing.md not found"}
	if !reflect.DeepEqual(plan.Problems, want) {
		t.Errorf("problems %v, want %v", plan.Problems, want)
	}
}

// A parent that names nothing stays broken: the spec does not take the plan's
// place in the tree, but the spec still counts the plan.
func TestPlanSpecLineLinksWithBrokenParent(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"specs/2026-09-29-s-design.md": "# S\n",
		"plans/2026-09-29-p.md":        "---\nparent: debt/nope\n---\n# P\n\n**Spec:** `.acta/specs/2026-09-29-s-design.md`\n\n### Task 1: A\n\n- [x] a\n",
	})
	spec := b.Get("specs/2026-09-29-s-design")
	plan := b.Get("plans/2026-09-29-p")
	if plan.SpecID != "" {
		t.Errorf("plan sits under %q, want no parent", plan.SpecID)
	}
	want := []string{"parent debt/nope not found"}
	if !reflect.DeepEqual(plan.Problems, want) {
		t.Errorf("problems %v, want %v", plan.Problems, want)
	}
	if spec.Status != "done" || spec.Total != 1 {
		t.Errorf("spec = %s %d/%d, want done 1/1", spec.Status, spec.Done, spec.Total)
	}
}

// A plan whose parent is its own spec counts that spec once, not twice.
func TestPlanSpecLineSameItemAsParent(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"specs/2026-09-29-s-design.md": "# S\n",
		"plans/2026-09-29-p.md":        "---\nparent: specs/2026-09-29-s-design\n---\n# P\n\n**Spec:** `.acta/specs/2026-09-29-s-design.md`\n\n### Task 1: A\n\n- [x] a\n",
	})
	spec := b.Get("specs/2026-09-29-s-design")
	plan := b.Get("plans/2026-09-29-p")
	if plan.SpecID != "specs/2026-09-29-s-design" {
		t.Errorf("plan sits under %q, want the spec", plan.SpecID)
	}
	if spec.Status != "done" || spec.Done != 1 || spec.Total != 1 {
		t.Errorf("spec = %s %d/%d, want done 1/1 counted once", spec.Status, spec.Done, spec.Total)
	}
	if len(plan.Problems) != 0 {
		t.Errorf("problems %v, want none", plan.Problems)
	}
}

func TestLoadDashTasks(t *testing.T) {
	t.Parallel()

	b := loadFixture(t)
	t1 := b.Get("plans/2026-09-26-dash-tasks#task-F-1")
	t2 := b.Get("plans/2026-09-26-dash-tasks#task-F-2")
	if t1 == nil || t2 == nil {
		t.Fatalf("dash tasks missing: F-1=%v F-2=%v", t1 != nil, t2 != nil)
	}
	if t1.Title != "Drop the stored-column fallback" || t1.Line != 3 {
		t.Errorf("F-1 = title %q line %d", t1.Title, t1.Line)
	}
	if t2.Title != "Re-check the quantity math" || t2.Line != 8 {
		t.Errorf("F-2 = title %q line %d", t2.Title, t2.Line)
	}
	parent := b.Get("plans/2026-09-26-dash-tasks")
	if parent.Done != 2 || parent.Total != 2 {
		t.Errorf("parent progress = %d/%d, want 2/2", parent.Done, parent.Total)
	}
}

func TestLoadDottedTasks(t *testing.T) {
	t.Parallel()

	b := loadFixture(t)
	t1 := b.Get("plans/2026-09-28-dotted-tasks#task-2.1")
	t2 := b.Get("plans/2026-09-28-dotted-tasks#task-2.2")
	if t1 == nil || t2 == nil {
		t.Fatalf("dotted tasks missing: 2.1=%v 2.2=%v", t1 != nil, t2 != nil)
	}
	if t1.Title != "First dotted step" || t1.Line != 3 {
		t.Errorf("2.1 = title %q line %d", t1.Title, t1.Line)
	}
	if t2.Title != "Second dotted step" || t2.Line != 8 {
		t.Errorf("2.2 = title %q line %d", t2.Title, t2.Line)
	}
	parent := b.Get("plans/2026-09-28-dotted-tasks")
	if parent.Done != 2 || parent.Total != 2 {
		t.Errorf("parent progress = %d/%d, want 2/2", parent.Done, parent.Total)
	}
}

func TestLoadRecordsProblems(t *testing.T) {
	t.Parallel()

	b := loadFixture(t)
	cases := map[string]string{
		"plans/2026-09-27-orphan":     "parent bugs/nope not found",
		"specs/2026-09-18-weird":      "unknown status bogus",
		"specs/2026-09-17-broken":     "frontmatter:",
		"debt/2026-09-27-orphan-debt": "parent plans/nope not found",
	}
	for id, part := range cases {
		it := b.Get(id)
		if !strings.Contains(strings.Join(it.Problems, "\n"), part) {
			t.Errorf("%s problems = %v, want one containing %q", id, it.Problems, part)
		}
	}
	if got := b.Get("specs/2026-09-17-broken").Title; got != "Broken spec" {
		t.Errorf("broken title = %q", got)
	}
	if p := b.Get("specs/2026-09-20-alpha").Problems; len(p) != 0 {
		t.Errorf("alpha should have no problems, got %v", p)
	}
}

func TestLoadMarksLegacy(t *testing.T) {
	t.Parallel()

	b := loadFixture(t)
	if !b.Get("docs/superpowers/specs/2026-01-01-old").Legacy || !b.Get("docs/superpowers/plans/2026-01-02-old#task-1").Legacy {
		t.Error("legacy items not marked")
	}
	if b.Get("specs/2026-09-20-alpha").Legacy {
		t.Error("root item marked legacy")
	}
}

func TestLists(t *testing.T) {
	t.Parallel()

	b := loadFixture(t)
	check := func(name string, got, want []string) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s = %v\nwant %v", name, got, want)
		}
	}
	check("active stories", ids(b.List(KindStory, false)), []string{
		"specs/2026-09-28-from-scratch-design", "specs/2026-09-22-beta", "specs/2026-09-20-alpha",
		"specs/2026-09-18-weird", "specs/2026-09-17-broken",
	})
	check("active plans", ids(b.List(KindPlan, false)), []string{
		"plans/2026-09-23-lonely", "plans/2026-09-21-alpha",
	})
	check("active bugs", ids(b.List(KindBug, false)), []string{"bugs/2026-09-28-lag", "bugs/2026-09-26-open", "specs/2026-09-15-really-bug"})
	check("active tasks", ids(b.List(KindTask, false)), []string{"plans/2026-09-23-lonely#task-1", "plans/2026-09-21-alpha#task-2"})
	check("untyped", ids(b.Untyped(false)), []string{"docs/superpowers/specs/2026-01-01-old"})
	if got := len(b.List(KindStory, true)); got != 7 {
		t.Errorf("all non-legacy stories = %d, want 7", got)
	}
	if got := len(b.List(KindPlan, true)); got != 7 {
		t.Errorf("all non-legacy plans = %d, want 7", got)
	}
	for _, it := range b.List(KindTask, true) {
		if it.Legacy {
			t.Errorf("List gave legacy task %s", it.ID)
		}
	}
}

func TestSearch(t *testing.T) {
	t.Parallel()

	b := loadFixture(t)
	if got := ids(b.Search("b-1")); !contains(got, "bugs/2026-09-24-crash") {
		t.Errorf("search by ref = %v", got)
	}
	if !contains(ids(b.Search("OLD STORY")), "docs/superpowers/specs/2026-01-01-old") {
		t.Error("search misses legacy items or is case sensitive")
	}
	if len(b.Search("  ")) != 0 {
		t.Error("blank search should find nothing")
	}
}

func TestLoadMissingRoot(t *testing.T) {
	t.Parallel()

	b, err := Load(config.Default(t.TempDir()))
	if err != nil || len(b.Items) != 0 {
		t.Fatalf("got %v items, err %v", len(b.Items), err)
	}
}

func TestAllowedAndClosed(t *testing.T) {
	t.Parallel()

	if !reflect.DeepEqual(Allowed(KindBug), []string{"open", "fixing", "fixed", "wontfix"}) {
		t.Error("bug statuses")
	}
	if !reflect.DeepEqual(Allowed(KindStory), []string{"draft", "approved", "in-progress", "done", "dropped"}) {
		t.Error("story statuses")
	}
	if !reflect.DeepEqual(Allowed(KindTask), []string{"todo", "in-progress", "done"}) {
		t.Error("task statuses")
	}
	for _, s := range []string{"done", "fixed", "dropped", "wontfix"} {
		if !Closed(s) {
			t.Errorf("%s should be closed", s)
		}
	}
	if Closed("in-progress") || Closed("bogus") {
		t.Error("open statuses marked closed")
	}
}

func TestLoadDebtFileAndLines(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"plans/2026-09-26-short-ids.md": "---\nid: PLAN-3\nhash: k3f2\n---\n# Short IDs\n",
		"debt/2026-09-27-short-ids.md":  "---\nid: DEBT-3\nhash: t9qe\nparent: plans/2026-09-26-short-ids\n---\n# Review NOTEs: Short IDs\n\n- [ ] a\n- [x] b\n- [-] c\n",
	})
	f := b.Get("DBT-0003")
	if f == nil || f.Kind != KindDebt || f.Status != "open" || f.Done != 2 || f.Total != 3 {
		t.Fatalf("file = %+v", f)
	}
	if !f.OldForm {
		t.Error("an old id and an old 4-char hash must mark the file for rewriting")
	}
	for _, id := range []string{"DBT-0003", "dbt-0003", "dbt-t9qe", "DBT-T9QE", "debt/2026-09-27-short-ids"} {
		if b.Get(id) != f {
			t.Errorf("%s does not resolve to the file", id)
		}
	}
	cases := map[string]string{"DBT-0003.01": "open", "dbt-0003.01": "open", "dbt-t9qe.02": "done", "DBT-T9QE.02": "done", "debt/2026-09-27-short-ids#item-3": "wontfix"}
	for id, status := range cases {
		it := b.Get(id)
		if it == nil || it.Kind != KindDebtItem || it.Status != status {
			t.Errorf("%s = %+v, want status %s", id, it, status)
		}
	}
}

func TestDebtFileDoneWhenAllLinesClosed(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"debt/2026-09-27-x.md": "---\nid: DEBT-1\n---\n# R\n\n- [x] a\n- [-] b\n",
	})
	if got := b.Get("DBT-0001").Status; got != "done" {
		t.Fatalf("status = %s, want done", got)
	}
}

// The author of an item is the person whose commit first added its file. A
// task lives in the plan file and takes its author, a debt item in the debt
// file. A file git has no commit for belongs to whoever commits here now, and
// a folder outside git leaves the field empty.
func TestAuthorComesFromTheFirstCommit(t *testing.T) {
	t.Parallel()

	dir := authorRepo(t, "Ana", map[string]string{
		"specs/2026-09-20-a.md": "---\nid: SPEC-1\n---\n# Spec A\n",
		"plans/2026-09-21-a.md": "---\nid: PLAN-1\n---\n# Plan A\n\n**Spec:** `.acta/specs/2026-09-20-a.md`\n\n### Task 1: One\n- [ ] x\n\n### Task 2: Two\n- [ ] y\n",
	})
	b := loadDir(t, dir)
	b.FillAuthors()
	for _, id := range []string{"SPC-0001", "PLN-0001", "PLN-0001.01", "PLN-0001.02"} {
		it := b.Get(id)
		if it == nil {
			t.Errorf("the board holds no %s", id)
			continue
		}
		if it.Author != "Ana" {
			t.Errorf("%s author = %q, want Ana", id, it.Author)
		}
	}

	late := filepath.Join(dir, ".acta", "bugs", "2026-09-22-b.md")
	if err := os.MkdirAll(filepath.Dir(late), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(late, []byte("---\nid: BUG-1\n---\n# Bug B\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "config", "user.name", "Sari")
	again := loadDir(t, dir)
	again.FillAuthors()
	if got := again.Get("BUG-0001").Author; got != "Sari" {
		t.Errorf("a file with no commit has author %q, want the user.name Sari", got)
	}

	outside := boardWith(t, map[string]string{"plans/2026-09-21-a.md": "# Plan A\n"})
	outside.FillAuthors()
	if got := outside.Get("plans/2026-09-21-a").Author; got != "" {
		t.Errorf("a board outside git has author %q, want empty", got)
	}
}

// A debt item lives in its debt file, so it takes that file's author and the
// text around the checklist as its body, not the whole file again.
func TestDebtItemTakesTheDebtFile(t *testing.T) {
	t.Parallel()

	dir := authorRepo(t, "Budi", map[string]string{
		"debt/2026-09-24-notes.md": "---\nid: DEBT-1\n---\n# Review NOTEs\n\nProse about the review.\n\n- [ ] first note\n- [x] second note\n",
	})
	b := loadDir(t, dir)
	b.FillAuthors()
	if got := b.Get("DBT-0001").Author; got != "Budi" {
		t.Errorf("the debt file has author %q, want Budi", got)
	}
	it := b.Get("DBT-0001.01")
	if it == nil || it.Author != "Budi" {
		t.Fatalf("debt item = %+v, want the debt file's author Budi", it)
	}
	got := it.Body
	if !strings.Contains(got, "Prose about the review.") {
		t.Errorf("debt item body = %q, want the prose of the file", got)
	}
	if strings.Contains(got, "first note") {
		t.Errorf("debt item body = %q, want no checklist line in it", got)
	}
}

// authorAsk is one question the stubbed git got: the checkout it was asked in
// and the files it was asked about.
type authorAsk struct {
	repo  string
	paths []string
}

// stubAuthorQuestions swaps the two questions a board asks git for counters,
// so a test can count them without a real repo, and puts the real ones back
// when the test ends. It swaps package variables, so the test cannot run in
// parallel.
func stubAuthorQuestions(t *testing.T) (asks *[]authorAsk, names *[]string) {
	t.Helper()
	asks, names = &[]authorAsk{}, &[]string{}
	gitAuthors = func(repo string, paths []string) map[string]string {
		*asks = append(*asks, authorAsk{repo, append([]string{}, paths...)})
		return map[string]string{}
	}
	gitUserName = func(repo string) string {
		*names = append(*names, repo)
		return "Sari"
	}
	t.Cleanup(func() { gitAuthors, gitUserName = gitc.Authors, gitc.UserName })
	return asks, names
}

// writeUnder writes each file at its path inside dir, making the folders.
func writeUnder(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// Loading a board is for showing it, and only the detail box reads an author,
// so a load must never ask git for one: not for the main tree, not for another
// worktree, not for a branch read from git.
func TestLoadNeverAsksForAuthors(t *testing.T) {
	asks, names := stubAuthorQuestions(t)

	main, side := t.TempDir(), t.TempDir()
	writeUnder(t, main, map[string]string{
		".acta/specs/2026-09-20-a.md":              "---\nid: SPEC-1\n---\n# Spec A\n",
		".acta/plans/2026-09-21-a.md":              "---\nid: PLAN-1\n---\n# Plan A\n\n### Task 1: One\n- [ ] x\n",
		"docs/superpowers/specs/2026-09-19-old.md": "# Old spec\n",
	})
	writeUnder(t, side, map[string]string{".acta/specs/2026-09-22-b.md": "---\nid: SPEC-2\n---\n# Spec B\n"})
	gone := Tree{Cfg: config.Default(filepath.Join(t.TempDir(), "gone")), Branch: "gone",
		Files: map[string][]byte{".acta/specs/2026-09-23-c.md": []byte("# Spec C\n")}}

	if _, err := Load(config.Default(main)); err != nil {
		t.Fatal(err)
	}
	b, err := LoadTrees(config.Default(main), []Tree{{Cfg: config.Default(side), Branch: "side"}, gone})
	if err != nil {
		t.Fatal(err)
	}
	if len(*asks) != 0 || len(*names) != 0 {
		t.Errorf("a load asked git for authors %v and for user.name %v, want neither", *asks, *names)
	}
	for _, it := range b.Items {
		if it.Author != "" {
			t.Errorf("%s has author %q straight after a load, want it empty", it.ID, it.Author)
		}
	}
}

// Git is asked once per checkout, not once per file, per item or per folder.
// The main checkout holds two specs, a plan with three tasks, a debt file with
// two lines and a legacy spec, so git gets one question that names each file
// once. A worktree nested inside the main folder is a checkout of its own: its
// file goes to it, not to the main checkout. The name git commits under is
// read once, in the main checkout.
func TestAuthorIsAskedOncePerCheckout(t *testing.T) {
	asks, names := stubAuthorQuestions(t)

	dir := t.TempDir()
	nested := filepath.Join(dir, ".worktrees", "feat")
	writeUnder(t, dir, map[string]string{
		".acta/specs/2026-09-20-a.md":              "---\nid: SPEC-1\n---\n# Spec A\n",
		".acta/specs/2026-09-20-b.md":              "---\nid: SPEC-2\n---\n# Spec B\n",
		".acta/plans/2026-09-21-a.md":              "---\nid: PLAN-1\n---\n# Plan A\n\n### Task 1: One\n- [ ] x\n\n### Task 2: Two\n- [ ] y\n\n### Task 3: Three\n- [ ] z\n",
		".acta/debt/2026-09-24-notes.md":           "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] one\n- [ ] two\n",
		"docs/superpowers/specs/2026-09-19-old.md": "# Old spec\n",
	})
	writeUnder(t, nested, map[string]string{".acta/specs/2026-09-25-n.md": "---\nid: SPEC-3\n---\n# Spec N\n"})

	b, err := LoadTrees(config.Default(dir), []Tree{{Cfg: config.Default(nested), Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	b.FillAuthors()

	if len(*asks) != 2 {
		t.Fatalf("git was asked %d times, want 2 (main checkout, nested worktree): %v", len(*asks), *asks)
	}
	byRepo := map[string][]string{}
	for _, a := range *asks {
		byRepo[a.repo] = a.paths
	}
	wantMain := []string{
		filepath.Join(dir, ".acta", "debt", "2026-09-24-notes.md"),
		filepath.Join(dir, ".acta", "plans", "2026-09-21-a.md"),
		filepath.Join(dir, ".acta", "specs", "2026-09-20-a.md"),
		filepath.Join(dir, ".acta", "specs", "2026-09-20-b.md"),
		filepath.Join(dir, "docs", "superpowers", "specs", "2026-09-19-old.md"),
	}
	got := append([]string{}, byRepo[dir]...)
	slices.Sort(got)
	if !slices.Equal(got, wantMain) {
		t.Errorf("the main checkout question named %v, want each of its files once: %v", got, wantMain)
	}
	wantNested := []string{filepath.Join(nested, ".acta", "specs", "2026-09-25-n.md")}
	if !slices.Equal(byRepo[nested], wantNested) {
		t.Errorf("the nested worktree question named %v, want only its own file %v", byRepo[nested], wantNested)
	}
	if len(*names) != 1 || (*names)[0] != dir {
		t.Errorf("user.name was read in %v, want once, in the main checkout %s", *names, dir)
	}
	for _, id := range []string{"SPC-0001", "SPC-0002", "SPC-0003", "PLN-0001", "PLN-0001.03", "DBT-0001.01", "DBT-0001.02"} {
		it := b.Get(id)
		if it == nil {
			t.Errorf("the board holds no %s", id)
			continue
		}
		if it.Author != "Sari" {
			t.Errorf("%s author = %q, want the one name the whole call read", id, it.Author)
		}
	}
}

// A file whose folder is under no checkout the board knows is asked about in
// its own folder, as it always was, so a root set outside the repo still gets
// its authors.
func TestAuthorOfAFileOutsideEveryCheckoutIsAskedByItsFolder(t *testing.T) {
	asks, _ := stubAuthorQuestions(t)

	repo, away := t.TempDir(), t.TempDir()
	cfg := config.Default(repo)
	cfg.Root = filepath.Join(away, ".acta")
	writeUnder(t, away, map[string]string{".acta/specs/2026-09-20-a.md": "# Spec A\n"})
	b, err := Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	b.FillAuthors()

	folder := filepath.Join(away, ".acta", "specs")
	if len(*asks) != 1 || (*asks)[0].repo != folder {
		t.Fatalf("git was asked %v, want one question in %s", *asks, folder)
	}
}

// A checkout owns the files below it, matched on whole path parts, and the
// deepest one wins, so a worktree inside the main folder owns its own files.
func TestCheckoutOfPicksTheDeepestWholePathMatch(t *testing.T) {
	t.Parallel()

	for _, roots := range [][]string{{"/r/a", "/r/a/wt"}, {"/r/a/wt", "/r/a"}} {
		for path, want := range map[string]string{
			"/r/a/x.md":    "/r/a",
			"/r/a/wt/y.md": "/r/a/wt",
			"/r/ab/x.md":   "",
			"/r/b/x.md":    "",
		} {
			if got := checkoutOf(roots, filepath.FromSlash(path)); got != filepath.FromSlash(want) {
				t.Errorf("checkoutOf(%v, %s) = %q, want %q", roots, path, got, want)
			}
		}
	}
}

// Every kind of item gets the author the old per-folder lookup gave: a file
// of the main tree, a legacy folder, a worktree beside the repo, a worktree
// nested inside the repo folder, and a file nobody committed. A branch read
// only from git is not on disk, so it gets none.
func TestAuthorsReachEveryKindOfItem(t *testing.T) {
	t.Parallel()

	dir := authorRepo(t, "Ana", map[string]string{
		"specs/2026-09-20-main.md":                    "# Main spec\n",
		"../docs/superpowers/specs/2026-09-19-old.md": "# Old spec\n",
	})
	gitRun(t, dir, "config", "user.name", "Sari")
	beside, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	beside = filepath.Join(beside, "wt")
	nested := filepath.Join(dir, ".worktrees", "feat")
	gitRun(t, dir, "worktree", "add", "-q", "-b", "side", beside)
	gitRun(t, dir, "worktree", "add", "-q", "-b", "feat", nested)
	commitFile(t, beside, "Budi", ".acta/specs/2026-09-21-beside.md", "# Beside spec\n")
	commitFile(t, nested, "Citra", ".acta/specs/2026-09-22-nested.md", "# Nested spec\n")
	writeUnder(t, nested, map[string]string{".acta/bugs/2026-09-23-new.md": "# Never committed\n"})
	gone := Tree{Cfg: config.Default(filepath.Join(t.TempDir(), "gone")), Branch: "gone",
		Files: map[string][]byte{".acta/specs/2026-09-24-branch.md": []byte("# Branch spec\n")}}

	b, err := LoadTrees(config.Default(dir), []Tree{
		{Cfg: config.Default(beside), Branch: "side"},
		{Cfg: config.Default(nested), Branch: "feat"},
		gone,
	})
	if err != nil {
		t.Fatal(err)
	}
	b.FillAuthors()

	for _, c := range []struct{ kind, path, want string }{
		{"main tree", filepath.Join(dir, ".acta", "specs", "2026-09-20-main.md"), "Ana"},
		{"legacy folder", filepath.Join(dir, "docs", "superpowers", "specs", "2026-09-19-old.md"), "Ana"},
		{"worktree beside the repo", filepath.Join(beside, ".acta", "specs", "2026-09-21-beside.md"), "Budi"},
		{"worktree inside the repo folder", filepath.Join(nested, ".acta", "specs", "2026-09-22-nested.md"), "Citra"},
		{"file nobody committed", filepath.Join(nested, ".acta", "bugs", "2026-09-23-new.md"), "Sari"},
		{"branch read from git", "gone:.acta/specs/2026-09-24-branch.md", ""},
	} {
		found := false
		for _, it := range b.Items {
			if it.Path == c.path {
				found = true
				if it.Author != c.want {
					t.Errorf("%s: author = %q, want %q", c.kind, it.Author, c.want)
				}
			}
		}
		if !found {
			t.Errorf("%s: the board holds no item at %s", c.kind, c.path)
		}
	}
}

// One folder can hold a committed file and a file nobody committed yet. The
// first keeps its author and the second gets the name this checkout commits
// under.
func TestAuthorMixesCommittedAndNewFilesInOneFolder(t *testing.T) {
	t.Parallel()

	dir := authorRepo(t, "Ana", map[string]string{
		"specs/2026-09-20-a.md": "---\nid: SPEC-1\n---\n# Spec A\n",
	})
	late := filepath.Join(dir, ".acta", "specs", "2026-09-21-b.md")
	if err := os.WriteFile(late, []byte("---\nid: SPEC-2\n---\n# Spec B\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "config", "user.name", "Sari")
	b := loadDir(t, dir)
	b.FillAuthors()
	if got := b.Get("SPC-0001").Author; got != "Ana" {
		t.Errorf("the committed spec has author %q, want Ana", got)
	}
	if got := b.Get("SPC-0002").Author; got != "Sari" {
		t.Errorf("the new spec has author %q, want the user.name Sari", got)
	}
}

// commitFile writes one file into a checkout and commits it as name.
func commitFile(t *testing.T, checkout, name, rel, body string) {
	t.Helper()
	writeUnder(t, checkout, map[string]string{rel: body})
	gitRun(t, checkout, "add", rel)
	cmd := exec.Command("git", "-C", checkout, "commit", "-q", "-m", "add "+rel)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME="+name, "GIT_COMMITTER_NAME="+name,
		"GIT_AUTHOR_EMAIL="+name+"@example.com", "GIT_COMMITTER_EMAIL="+name+"@example.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v %s", err, out)
	}
}

// authorRepo writes the files into a fresh checkout and commits them as name,
// so every file on the board has an author to find.
func authorRepo(t *testing.T, name string, files map[string]string) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "init", "-q", "-b", "main")
	for rel, body := range files {
		p := filepath.Join(dir, ".acta", rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitRun(t, dir, "add", ".")
	cmd := exec.Command("git", "-C", dir, "commit", "-q", "-m", "add the board")
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME="+name, "GIT_COMMITTER_NAME="+name,
		"GIT_AUTHOR_EMAIL="+name+"@example.com", "GIT_COMMITTER_EMAIL="+name+"@example.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v %s", err, out)
	}
	return dir
}

// TestSpecWithPlanIgnoresWrittenStatus: SPEC-6 stayed approved with every
// task done because the written status won.
func TestSpecWithPlanIgnoresWrittenStatus(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"specs/2026-09-29-s-design.md": "---\nstatus: approved\n---\n# S\n",
		"plans/2026-09-29-p.md":        "# P\n\n**Spec:** `.acta/specs/2026-09-29-s-design.md`\n\n### Task 1: A\n\n- [x] a\n",
	}
	s := boardWith(t, files).Get("specs/2026-09-29-s-design")
	if s.Status != "done" || s.StatusSource != "derived" {
		t.Errorf("spec = %s (%s), want done (derived)", s.Status, s.StatusSource)
	}
	if want := []string{"written status approved ignored, derived done"}; !slices.Equal(s.Problems, want) {
		t.Errorf("problems %v, want %v", s.Problems, want)
	}
	files["specs/2026-09-29-s-design.md"] = "---\nstatus: done\n---\n# S\n"
	if p := boardWith(t, files).Get("specs/2026-09-29-s-design").Problems; len(p) != 0 {
		t.Errorf("same written and derived value still warns: %v", p)
	}
	delete(files, "plans/2026-09-29-p.md")
	files["specs/2026-09-29-s-design.md"] = "---\nstatus: approved\n---\n# S\n"
	if s := boardWith(t, files).Get("specs/2026-09-29-s-design"); s.Status != "approved" || s.StatusSource != "frontmatter" {
		t.Errorf("spec with no plan = %s (%s), want approved (frontmatter)", s.Status, s.StatusSource)
	}
}

func TestFileItemTitleFallback(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, src, want string
	}{
		{"title only", "---\ntitle: Severity or priority\n---\nsome words\n", "Severity or priority"},
		{"heading wins", "---\ntitle: From front\n---\n# From body\n", "From body"},
		{"empty title", "---\ntitle: \"\"\n---\nwords\n", "my-slug"},
		{"spaces title", "---\ntitle: \"   \"\n---\nwords\n", "my-slug"},
		{"null title", "---\ntitle:\n---\nwords\n", "my-slug"},
		{"neither", "---\nstatus: raw\n---\nwords\n", "my-slug"},
		{"unicode title", "---\ntitle: Catat ide 日本語 ✓\n---\nwords\n", "Catat ide 日本語 ✓"},
	}
	for _, c := range cases {
		it := fileItem(KindScratch, "scratch/2026-09-30-my-slug", "scratch/2026-09-30-my-slug.md", "2026-09-30", "my-slug", false, Parse([]byte(c.src)))
		if it.Title != c.want {
			t.Errorf("%s: title = %q, want %q", c.name, it.Title, c.want)
		}
	}
}

// TestPlanWrittenStatusFollowsBoxes: BUG-0013, a plan stayed approved after
// every task was ticked because the written word beat the boxes.
func TestPlanWrittenStatusFollowsBoxes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, written, boxes, status, source string
	}{
		{"approved and all ticked", "approved", "- [x] a", "done", "derived"},
		{"approved and half ticked", "approved", "- [x] a\n- [ ] b", "in-progress", "derived"},
		{"in-progress and none ticked", "in-progress", "- [ ] a", "approved", "derived"},
		{"draft keeps its word", "draft", "- [x] a", "draft", "frontmatter"},
		{"dropped keeps its word", "dropped", "- [x] a", "dropped", "frontmatter"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := map[string]string{
				"plans/2026-10-01-p.md": "---\nstatus: " + c.written + "\n---\n# P\n\n### Task 1: A\n\n" + c.boxes + "\n",
			}
			it := boardWith(t, files).Get("plans/2026-10-01-p")
			if it.Status != c.status || it.StatusSource != c.source {
				t.Errorf("written %s with %q = %s (%s), want %s (%s)",
					c.written, c.boxes, it.Status, it.StatusSource, c.status, c.source)
			}
			if len(it.Problems) != 0 {
				t.Errorf("written %s problems %v, want none", c.written, it.Problems)
			}
		})
	}
}
