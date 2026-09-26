package board

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
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
		"plans/2026-09-21-alpha#task-2":          {KindTask, "doing", "derived", 1, 2},
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
	if len(b.Items) != 31 {
		t.Errorf("got %d items, want 31: %v", len(b.Items), ids(b.Items))
	}
}

func TestLoadLinksPlans(t *testing.T) {
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

func TestLoadDashTasks(t *testing.T) {
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
	b := loadFixture(t)
	if !b.Get("docs/superpowers/specs/2026-01-01-old").Legacy || !b.Get("docs/superpowers/plans/2026-01-02-old#task-1").Legacy {
		t.Error("legacy items not marked")
	}
	if b.Get("specs/2026-09-20-alpha").Legacy {
		t.Error("root item marked legacy")
	}
}

func TestLists(t *testing.T) {
	b := loadFixture(t)
	check := func(name string, got, want []string) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s = %v\nwant %v", name, got, want)
		}
	}
	check("active stories", ids(b.List(KindStory, false)), []string{
		"specs/2026-09-22-beta", "specs/2026-09-20-alpha",
		"specs/2026-09-18-weird", "specs/2026-09-17-broken",
	})
	check("active plans", ids(b.List(KindPlan, false)), []string{
		"plans/2026-09-23-lonely", "plans/2026-09-21-alpha",
	})
	check("active bugs", ids(b.List(KindBug, false)), []string{"bugs/2026-09-26-open", "specs/2026-09-15-really-bug"})
	check("active tasks", ids(b.List(KindTask, false)), []string{"plans/2026-09-23-lonely#task-1", "plans/2026-09-21-alpha#task-2"})
	check("untyped", ids(b.Untyped(false)), []string{"docs/superpowers/specs/2026-01-01-old"})
	if got := len(b.List(KindStory, true)); got != 6 {
		t.Errorf("all non-legacy stories = %d, want 6", got)
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
	b, err := Load(config.Default(t.TempDir()))
	if err != nil || len(b.Items) != 0 {
		t.Fatalf("got %v items, err %v", len(b.Items), err)
	}
}

func TestAllowedAndClosed(t *testing.T) {
	if !reflect.DeepEqual(Allowed(KindBug), []string{"open", "fixing", "fixed", "wontfix"}) {
		t.Error("bug statuses")
	}
	if !reflect.DeepEqual(Allowed(KindStory), []string{"draft", "approved", "in-progress", "done", "dropped"}) {
		t.Error("story statuses")
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
	b := boardWith(t, map[string]string{
		"plans/2026-09-26-short-ids.md": "---\nid: PLAN-3\nhash: k3f2\n---\n# Short IDs\n",
		"debt/2026-09-27-short-ids.md":  "---\nid: DEBT-3\nhash: t9qe\nparent: plans/2026-09-26-short-ids\n---\n# Review NOTEs: Short IDs\n\n- [ ] a\n- [x] b\n- [-] c\n",
	})
	f := b.Get("DEBT-3")
	if f == nil || f.Kind != KindDebt || f.Status != "open" || f.Done != 2 || f.Total != 3 {
		t.Fatalf("file = %+v", f)
	}
	for _, id := range []string{"DEBT-3", "debt-3", "debt-t9qe", "DEBT-T9QE", "debt/2026-09-27-short-ids"} {
		if b.Get(id) != f {
			t.Errorf("%s does not resolve to the file", id)
		}
	}
	cases := map[string]string{"DEBT-3.1": "open", "debt-3.1": "open", "debt-t9qe.2": "done", "DEBT-T9QE.2": "done", "debt/2026-09-27-short-ids#item-3": "wontfix"}
	for id, status := range cases {
		it := b.Get(id)
		if it == nil || it.Kind != KindDebtItem || it.Status != status {
			t.Errorf("%s = %+v, want status %s", id, it, status)
		}
	}
}

func TestDebtFileDoneWhenAllLinesClosed(t *testing.T) {
	b := boardWith(t, map[string]string{
		"debt/2026-09-27-x.md": "---\nid: DEBT-1\n---\n# R\n\n- [x] a\n- [-] b\n",
	})
	if got := b.Get("DEBT-1").Status; got != "done" {
		t.Fatalf("status = %s, want done", got)
	}
}
