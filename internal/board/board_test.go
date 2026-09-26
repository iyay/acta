package board

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"pm-board/internal/config"
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
		"plans/2026-09-21-alpha#task-1":          {KindTask, "done", "derived", 2, 2},
		"plans/2026-09-21-alpha#task-2":          {KindTask, "doing", "derived", 1, 2},
		"specs/2026-09-22-beta":                  {KindStory, "draft", "derived", 0, 0},
		"plans/2026-09-23-lonely":                {KindStory, "approved", "derived", 0, 1},
		"plans/2026-09-23-lonely#task-1":         {KindTask, "todo", "derived", 0, 0},
		"bugs/2026-09-24-crash":                  {KindBug, "fixed", "derived", 1, 1},
		"plans/2026-09-25-crash-fix#task-F1":     {KindTask, "done", "derived", 2, 2},
		"plans/2026-09-28-dotted-tasks":          {KindStory, "done", "derived", 2, 2},
		"plans/2026-09-28-dotted-tasks#task-2.1": {KindTask, "done", "derived", 2, 2},
		"plans/2026-09-28-dotted-tasks#task-2.2": {KindTask, "done", "derived", 2, 2},
		"plans/2026-09-26-dash-tasks":            {KindStory, "done", "derived", 2, 2},
		"plans/2026-09-26-dash-tasks#task-F-1":   {KindTask, "done", "derived", 2, 2},
		"plans/2026-09-26-dash-tasks#task-F-2":   {KindTask, "done", "derived", 2, 2},
		"bugs/2026-09-26-open":                   {KindBug, "open", "derived", 0, 0},
		"plans/2026-09-27-orphan":                {KindStory, "done", "derived", 1, 1},
		"specs/2026-09-19-dropped":               {KindStory, "dropped", "frontmatter", 0, 0},
		"specs/2026-09-18-weird":                 {KindStory, "bogus", "frontmatter", 0, 0},
		"specs/2026-09-17-broken":                {KindStory, "draft", "derived", 0, 0},
		"specs/2026-09-16-finished":              {KindStory, "done", "derived", 1, 1},
		"specs/2026-09-15-really-bug":            {KindBug, "open", "derived", 0, 0},
		"docs/superpowers/specs/2026-01-01-old":  {KindStory, "approved", "derived", 0, 1},
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
	if len(b.Items) != 25 {
		t.Errorf("got %d items, want 25: %v", len(b.Items), ids(b.Items))
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
		"plans/2026-09-27-orphan": "parent bugs/nope not found",
		"specs/2026-09-18-weird":  "unknown status bogus",
		"specs/2026-09-17-broken": "frontmatter:",
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
		"plans/2026-09-23-lonely", "specs/2026-09-22-beta", "specs/2026-09-20-alpha",
		"specs/2026-09-18-weird", "specs/2026-09-17-broken",
	})
	check("active bugs", ids(b.List(KindBug, false)), []string{"bugs/2026-09-26-open", "specs/2026-09-15-really-bug"})
	check("active tasks", ids(b.List(KindTask, false)), []string{"plans/2026-09-23-lonely#task-1", "plans/2026-09-21-alpha#task-2"})
	check("untyped", ids(b.Untyped(false)), []string{"docs/superpowers/specs/2026-01-01-old"})
	if got := len(b.List(KindStory, true)); got != 10 {
		t.Errorf("all non-legacy stories = %d, want 10", got)
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
