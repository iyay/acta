package board

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
)

// boardWith writes files under a temp .pm root and loads the board.
func boardWith(t *testing.T, files map[string]string) *Board {
	t.Helper()
	dir := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(dir, ".pm", rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	b, err := Load(config.Default(dir))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

var idFiles = map[string]string{
	"specs/2026-09-20-a-design.md": "---\nid: SPEC-4\nhash: m2x9\n---\n# A\n",
	"plans/2026-09-21-a.md":        "---\nid: PLAN-12\nhash: k3f2\n---\n# A plan\n\n**Spec:** `.pm/specs/2026-09-20-a-design.md`\n\n### Task 3: Three\n- [ ] x\n\n### Task F1: Fix\n- [ ] y\n",
	"plans/2026-09-22-solo.md":     "---\nid: PLAN-1234\nhash: q8d1\n---\n# Solo\n\n**Spec:** none (Bounded, approved in chat on 2026-09-22)\n\n### Task 1: One\n- [ ] z\n",
	"bugs/2026-09-23-crash.md":     "---\nid: BUG-7\nhash: b7aa\n---\n# Crash\n",
	"plans/2026-09-24-plain.md":    "# Plain\n\n### Task 1: One\n- [ ] w\n",
}

func TestGetResolvesEveryIDForm(t *testing.T) {
	b := boardWith(t, idFiles)
	cases := map[string]string{
		"specs/2026-09-20-a-design": "specs/2026-09-20-a-design",
		"SPEC-4":                    "specs/2026-09-20-a-design",
		"spec-m2x9":                 "specs/2026-09-20-a-design",
		"PLAN-12":                   "plans/2026-09-21-a",
		"PLAN-12.3":                 "plans/2026-09-21-a#task-3",
		"plan-k3f2.f1":              "plans/2026-09-21-a#task-F1",
		"PLAN-1234":                 "plans/2026-09-22-solo",
		"PLAN-q8d1.1":               "plans/2026-09-22-solo#task-1",
		"BUG-7":                     "bugs/2026-09-23-crash",
		"bug-b7aa":                  "bugs/2026-09-23-crash",
		"plans/2026-09-24-plain":    "plans/2026-09-24-plain",
	}
	for in, want := range cases {
		it := b.Get(in)
		if it == nil || it.ID != want {
			t.Errorf("Get(%q) = %v, want %s", in, it, want)
		}
	}
	for _, miss := range []string{"PLAN-99", "PLAN-zzzz", "PLAN-12.9", "SPEC-", "PLAN-k3f", "BUG-7.1"} {
		if it := b.Get(miss); it != nil {
			t.Errorf("Get(%q) = %s, want nil", miss, it.ID)
		}
	}
}

func TestItemsCarryShortIDs(t *testing.T) {
	b := boardWith(t, idFiles)
	task := b.Get("plans/2026-09-21-a#task-3")
	if task.ShortID != "PLAN-12.3" || task.Hash != "PLAN-k3f2.3" || !strings.HasSuffix(task.PlanPath, "2026-09-21-a.md") {
		t.Fatalf("task ids = %q %q %q", task.ShortID, task.Hash, task.PlanPath)
	}
	if s := b.Get("specs/2026-09-20-a-design"); s.ShortID != "SPEC-4" || s.Hash != "SPEC-m2x9" {
		t.Fatalf("spec ids = %q %q", s.ShortID, s.Hash)
	}
	if p := b.Get("plans/2026-09-24-plain"); p.ShortID != "" || p.Hash != "" {
		t.Fatalf("plain plan got ids %q %q", p.ShortID, p.Hash)
	}
}

func TestIsHash(t *testing.T) {
	for s, want := range map[string]bool{"k3f2": true, "abcd": true, "1234": false, "K3f2": false, "k3f": false, "k3f2a": false, "k-f2": false, "": false} {
		if IsHash(s) != want {
			t.Errorf("IsHash(%q) = %v", s, !want)
		}
	}
}

func TestDuplicateShortIDsAreProblems(t *testing.T) {
	b := boardWith(t, map[string]string{
		"bugs/2026-09-23-a.md": "---\nid: BUG-7\nhash: b7aa\n---\n# A\n",
		"bugs/2026-09-24-b.md": "---\nid: BUG-7\nhash: b7aa\n---\n# B\n",
	})
	for _, id := range []string{"bugs/2026-09-23-a", "bugs/2026-09-24-b"} {
		got := strings.Join(b.Get(id).Problems, "; ")
		if !strings.Contains(got, "duplicate id BUG-7") || !strings.Contains(got, "duplicate hash BUG-b7aa") {
			t.Errorf("%s problems = %q", id, got)
		}
	}
}

// The prefix a file is read with is the kind it ends up as, so a file that
// says type: bug in specs/ reads as BUG and a plan file reads as PLAN. The
// folder a file sits in is not the kind it is.
func TestPrefixFollowsTheKindTheFileIs(t *testing.T) {
	b := boardWith(t, map[string]string{
		"specs/2026-09-20-a-design.md": "---\ntype: bug\nid: BUG-3\nhash: m2x9\n---\n# A\n",
		"plans/2026-09-21-b.md":        "---\nid: PLAN-12\nhash: k3f2\n---\n# B\n\n### Task 1: T\n- [ ] a\n",
	})
	spec := b.Get("specs/2026-09-20-a-design")
	if spec.Kind != KindBug || spec.ShortID != "BUG-3" || spec.Hash != "BUG-m2x9" {
		t.Fatalf("bug in specs = kind %q id %q hash %q", spec.Kind, spec.ShortID, spec.Hash)
	}
	if got := strings.Join(spec.Problems, "; "); got != "" {
		t.Fatalf("clean file has problems: %q", got)
	}
	plan := b.Get("plans/2026-09-21-b")
	if plan.ShortID != "PLAN-12" || plan.Hash != "PLAN-k3f2" {
		t.Fatalf("plan ids = %q %q", plan.ShortID, plan.Hash)
	}
	if got := strings.Join(plan.Problems, "; "); got != "" {
		t.Fatalf("clean plan has problems: %q", got)
	}
}

// A value in the wrong shape stays a problem and never becomes an ID, and the
// file's own id is kept so a tool can see it without reading the file again.
func TestIDsKeepTheWrittenValue(t *testing.T) {
	b := boardWith(t, map[string]string{
		"specs/2026-09-20-a-design.md": "---\nid: SPEC-zz\nhash: toolong\n---\n# A\n",
	})
	it := b.Get("specs/2026-09-20-a-design")
	if it.ShortID != "" || it.Hash != "" {
		t.Fatalf("bad values became ids: %q %q", it.ShortID, it.Hash)
	}
	if it.RawID != "SPEC-zz" || it.RawHash != "toolong" {
		t.Fatalf("written values lost: %q %q", it.RawID, it.RawHash)
	}
	got := strings.Join(it.Problems, "; ")
	if !strings.Contains(got, "bad id SPEC-zz") || !strings.Contains(got, "bad hash toolong") {
		t.Fatalf("problems = %q", got)
	}
}

// A plan whose tasks belong to a spec is an item of its own, and the plan
// file is where its own id lives.
func TestPlanKeepsItsOwnFile(t *testing.T) {
	b := boardWith(t, map[string]string{
		"specs/2026-09-20-a-design.md": "# A\n",
		"plans/2026-09-21-b.md":        "---\nparent: specs/2026-09-20-a-design\nid: PLAN-12\nhash: k3f2\n---\n# B\n\n### Task 1: T\n- [ ] a\n",
	})
	plan := b.Get("plans/2026-09-21-b")
	if plan == nil || plan.Kind != KindPlan {
		t.Fatalf("plan = %v", plan)
	}
	if !strings.HasSuffix(plan.Path, "plans/2026-09-21-b.md") || plan.ShortID != "PLAN-12" || plan.Hash != "PLAN-k3f2" {
		t.Fatalf("plan = %+v", plan)
	}
	if plan.SpecID != "specs/2026-09-20-a-design" {
		t.Fatalf("plan spec = %q", plan.SpecID)
	}
	if holder := b.Get("specs/2026-09-20-a-design"); len(holder.Children) != 1 ||
		holder.Children[0] != "plans/2026-09-21-b#task-1" {
		t.Fatalf("spec children = %v", holder.Children)
	}
}
