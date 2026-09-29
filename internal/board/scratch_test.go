package board

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/iyay/acta/internal/config"
)

func TestScratchStatusFollowsOneRule(t *testing.T) {
	t.Parallel()

	b := loadFixture(t)
	cases := []struct {
		id, status, source string
	}{
		{"scratch/2026-09-28-idea-raw", "raw", "frontmatter"},
		{"scratch/2026-09-28-idea-used", "specced", "derived"},
		{"scratch/2026-09-28-idea-dropped", "dropped", "frontmatter"},
	}
	for _, c := range cases {
		it := b.Get(c.id)
		if it == nil {
			t.Fatalf("%s not loaded", c.id)
		}
		if it.Kind != KindScratch || it.Status != c.status {
			t.Errorf("%s: kind %s status %s, want scratch %s", c.id, it.Kind, it.Status, c.status)
		}
		if c.source != "" && it.StatusSource != c.source {
			t.Errorf("%s: source %s, want %s", c.id, it.StatusSource, c.source)
		}
	}
	used := b.Get("scratch/2026-09-28-idea-used")
	if !slices.Contains(used.Children, "specs/2026-09-28-from-scratch-design") {
		t.Errorf("used item children = %v, want the spec", used.Children)
	}
}

// Every written status against linked and not linked, so the rule holds on all
// eight pairs and not only on the ones the fixture happens to hold. The wants
// are written out, so a change to the rule itself has to fail here.
func TestScratchStatusEveryCombination(t *testing.T) {
	t.Parallel()

	cases := []struct {
		written, status, source string
		linked                  bool
	}{
		{"", "raw", "derived", false},
		{"", "specced", "derived", true},
		{"raw", "raw", "frontmatter", false},
		{"raw", "specced", "derived", true},
		{"brainstorming", "brainstorming", "frontmatter", false},
		{"brainstorming", "specced", "derived", true},
		{"dropped", "dropped", "frontmatter", false},
		{"dropped", "dropped", "frontmatter", true},
	}
	for _, c := range cases {
		name := c.written
		if name == "" {
			name = "none"
		}
		if c.linked {
			name += "+linked"
		}
		t.Run(name, func(t *testing.T) {
			item := "# Idea\n"
			if c.written != "" {
				item = "---\nstatus: " + c.written + "\n---\n# Idea\n"
			}
			files := map[string]string{"scratch/2026-09-28-idea.md": item}
			if c.linked {
				files["specs/2026-09-28-from-idea-design.md"] =
					"---\nparent: scratch/2026-09-28-idea\n---\n# From idea\n"
			}
			b := boardWith(t, files)
			it := b.Get("scratch/2026-09-28-idea")
			if it == nil || it.Kind != KindScratch {
				t.Fatalf("idea = %v, want a scratch item", it)
			}
			if it.Status != c.status || it.StatusSource != c.source {
				t.Errorf("written %q linked %v = %s (%s), want %s (%s)",
					c.written, c.linked, it.Status, it.StatusSource, c.status, c.source)
			}
			if got := len(it.Children) == 1; got != c.linked {
				t.Errorf("linked = %v, want the child list to match", got)
			}
			if p := it.Problems; len(p) != 0 {
				t.Errorf("problems = %v, want none", p)
			}
		})
	}
}

// A written dropped is a decision the user already made, so a spec that came
// out of the idea cannot turn it into specced.
func TestWrittenDroppedBeatsSpecLink(t *testing.T) {
	t.Parallel()

	files := map[string]string{"scratch/2026-09-28-idea.md": "---\nstatus: dropped\n---\n# Idea\n"}
	b := boardWith(t, files)
	it := b.Get("scratch/2026-09-28-idea")
	if it == nil {
		t.Fatal("idea missing")
	}
	if it.Status != "dropped" || it.StatusSource != "frontmatter" {
		t.Errorf("a dropped idea a spec points at = %s (%s), want dropped (frontmatter)", it.Status, it.StatusSource)
	}
	if !Closed(it.Status) {
		t.Error("a dropped idea must stay off the open lists")
	}

	files["specs/2026-09-28-from-idea-design.md"] = "---\nparent: scratch/2026-09-28-idea\n---\n# From idea\n"
	linked := boardWith(t, files).Get("scratch/2026-09-28-idea")
	if linked.Status != "dropped" {
		t.Errorf("a dropped idea with a spec naming it = %s, want dropped", linked.Status)
	}
	if !slices.Contains(linked.Children, "specs/2026-09-28-from-idea-design") {
		t.Errorf("the link is lost: children = %v", linked.Children)
	}
}

// A spec naming an idea that is not there says so once and the board still
// loads, so one typo never hides every other item.
func TestSpecParentNotFoundIsOneProblem(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"scratch/2026-09-28-idea.md":           "# Idea\n",
		"specs/2026-09-28-from-idea-design.md": "---\nparent: scratch/missing\n---\n# From idea\n",
	})
	it := b.Get("specs/2026-09-28-from-idea-design")
	if it == nil {
		t.Fatal("the spec did not load")
	}
	if want := []string{"parent scratch/missing not found"}; !slices.Equal(it.Problems, want) {
		t.Errorf("problems = %v, want exactly %v", it.Problems, want)
	}
	if b.Get("scratch/2026-09-28-idea") == nil {
		t.Error("one broken parent hid the rest of the board")
	}

	// The dir a repo names in .acta.yaml counts as a scratch parent too.
	custom := boardWithScratchDir(t, "ideas", map[string]string{
		"ideas/2026-09-28-idea.md":             "# Idea\n",
		"specs/2026-09-28-from-idea-design.md": "---\nparent: ideas/missing\n---\n# From idea\n",
	})
	if p := custom.Get("specs/2026-09-28-from-idea-design").Problems; !slices.Equal(p, []string{"parent ideas/missing not found"}) {
		t.Errorf("custom dir problems = %v, want the parent not found once", p)
	}
	if custom.Get("ideas/2026-09-28-idea") == nil {
		t.Error("the scratch dir from .acta.yaml was not read")
	}
}

// A status the file holds outside the scratch list gets the same problem line
// every other kind gives a status it does not know, specced included.
func TestScratchBadWrittenStatusIsOneProblem(t *testing.T) {
	t.Parallel()

	for _, bad := range []string{"bogus", "specced", "done"} {
		b := boardWith(t, map[string]string{
			"scratch/2026-09-28-idea.md": "---\nstatus: " + bad + "\n---\n# Idea\n",
		})
		it := b.Get("scratch/2026-09-28-idea")
		if it == nil {
			t.Fatalf("status %q: the idea did not load", bad)
		}
		if !slices.Equal(it.Problems, []string{"unknown status " + bad}) {
			t.Errorf("status %q problems = %v, want one unknown status line", bad, it.Problems)
		}
	}
}

func TestScratchStatusesAndPrefix(t *testing.T) {
	t.Parallel()

	if got := Allowed(KindScratch); !slices.Equal(got, []string{"raw", "brainstorming", "dropped"}) {
		t.Errorf("Allowed = %v", got)
	}
	if slices.Contains(Allowed(KindScratch), "specced") {
		t.Error("specced is derived, so no tool may offer it as a value to write")
	}
	for _, s := range []string{"specced", "dropped"} {
		if !Closed(s) {
			t.Errorf("Closed(%q) = false", s)
		}
	}
	for _, s := range []string{"raw", "brainstorming"} {
		if Closed(s) {
			t.Errorf("Closed(%q) = true", s)
		}
	}
	if Prefix(KindScratch, false) != "SCR" {
		t.Error("prefix")
	}
	if Prefix(KindScratch, true) != "PLN" {
		t.Error("a plan file is still a plan")
	}
	if Prefix(KindStory, false) != "SPC" || Prefix(KindBug, false) != "BUG" || Prefix(KindDebt, false) != "DBT" ||
		Prefix(KindPlan, false) != "PLN" || Prefix(KindTask, false) != "SPC" {
		t.Error("other prefixes changed")
	}
	for new, old := range map[string]string{"PLN": "PLAN", "SPC": "SPEC", "DBT": "DEBT", "SCR": "SCRATCH", "BUG": "BUG"} {
		if OldPrefix(new) != old {
			t.Errorf("OldPrefix(%q) = %q, want %q", new, OldPrefix(new), old)
		}
	}
	if got := FormatID("PLN", 30); got != "PLN-0030" {
		t.Errorf("FormatID = %q, want PLN-0030", got)
	}
	if got := Allowed(KindStory); !slices.Equal(got, []string{"draft", "approved", "in-progress", "done", "dropped"}) {
		t.Errorf("story statuses = %v", got)
	}
	if got := Allowed(KindTask); !slices.Equal(got, []string{"todo", "in-progress", "done"}) {
		t.Errorf("task statuses = %v", got)
	}
	if got := Allowed(KindBug); !slices.Equal(got, []string{"open", "fixing", "fixed", "wontfix"}) {
		t.Errorf("bug statuses = %v", got)
	}
	if got := Allowed(KindDebt); !slices.Equal(got, []string{"open", "done", "wontfix"}) {
		t.Errorf("debt statuses = %v", got)
	}
}

// boardWithScratchDir is boardWith with the scratch folder the repo names in
// .acta.yaml, so the board reads the same place the CLI writes.
func boardWithScratchDir(t *testing.T, dir string, files map[string]string) *Board {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(root, ".acta", rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default(root)
	cfg.Dirs.Scratch = dir
	b, err := Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
