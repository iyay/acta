package board

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
)

// boardWith writes files under a temp .acta root and loads the board.
func boardWith(t *testing.T, files map[string]string) *Board {
	t.Helper()
	dir := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(dir, ".acta", rel)
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
	"plans/2026-09-21-a.md":        "---\nid: PLAN-12\nhash: k3f2\n---\n# A plan\n\n**Spec:** `.acta/specs/2026-09-20-a-design.md`\n\n### Task 3: Three\n- [ ] x\n\n### Task F1: Fix\n- [ ] y\n",
	"plans/2026-09-22-solo.md":     "---\nid: PLAN-1234\nhash: q8d1\n---\n# Solo\n\n**Spec:** none (Bounded, approved in chat on 2026-09-22)\n\n### Task 1: One\n- [ ] z\n",
	"bugs/2026-09-23-crash.md":     "---\nid: BUG-7\nhash: b7aa\n---\n# Crash\n",
	"plans/2026-09-24-plain.md":    "# Plain\n\n### Task 1: One\n- [ ] w\n",
}

func TestGetResolvesEveryIDForm(t *testing.T) {
	b := boardWith(t, idFiles)
	cases := map[string]string{
		"specs/2026-09-20-a-design": "specs/2026-09-20-a-design",
		"SPC-0004":                  "specs/2026-09-20-a-design",
		"spc-m2x9":                  "specs/2026-09-20-a-design",
		"PLN-0012":                  "plans/2026-09-21-a",
		"PLN-0012.03":               "plans/2026-09-21-a#task-3",
		"pln-k3f2.F1":               "plans/2026-09-21-a#task-F1",
		"PLN-1234":                  "plans/2026-09-22-solo",
		"PLN-q8d1.01":               "plans/2026-09-22-solo#task-1",
		"BUG-0007":                  "bugs/2026-09-23-crash",
		"bug-b7aa":                  "bugs/2026-09-23-crash",
		"plans/2026-09-24-plain":    "plans/2026-09-24-plain",
	}
	for in, want := range cases {
		it := b.Get(in)
		if it == nil || it.ID != want {
			t.Errorf("Get(%q) = %v, want %s", in, it, want)
		}
	}
	for _, miss := range []string{"PLN-0099", "PLN-zzzz", "PLN-0012.9", "SPC-", "PLN-k3f", "BUG-0007.1"} {
		if it := b.Get(miss); it != nil {
			t.Errorf("Get(%q) = %s, want nil", miss, it.ID)
		}
	}
}

func TestItemsCarryShortIDs(t *testing.T) {
	b := boardWith(t, idFiles)
	task := b.Get("plans/2026-09-21-a#task-3")
	if task.ShortID != "PLN-0012.03" || task.Hash != "PLN-k3f2.03" || !strings.HasSuffix(task.PlanPath, "2026-09-21-a.md") {
		t.Fatalf("task ids = %q %q %q", task.ShortID, task.Hash, task.PlanPath)
	}
	if s := b.Get("specs/2026-09-20-a-design"); s.ShortID != "SPC-0004" || s.Hash != "SPC-m2x9" || !s.OldForm {
		t.Fatalf("spec ids = %q %q old=%v", s.ShortID, s.Hash, s.OldForm)
	}
	if p := b.Get("plans/2026-09-24-plain"); p.ShortID != "" || p.Hash != "" {
		t.Fatalf("plain plan got ids %q %q", p.ShortID, p.Hash)
	}
}

// A 4-char hash is no longer a hash, but setIDs still reads one so an old file
// loads. isHashLen is what lets it.
func TestIsHash(t *testing.T) {
	for s, want := range map[string]bool{"k3f2abc": true, "abcd": false, "1234": false, "K3f2": false, "k3f": false, "k3f2ab": false, "k-f2": false, "": false} {
		if IsHash(s) != want {
			t.Errorf("IsHash(%q) = %v", s, !want)
		}
	}
	for s, want := range map[string]bool{"k3f2": true, "abcd": true, "k3f": false, "k3f2a": false, "k-f2": false, "": false} {
		if isHashLen(s, 4) != want {
			t.Errorf("isHashLen(%q, 4) = %v", s, !want)
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
		if !strings.Contains(got, "duplicate id BUG-0007") || !strings.Contains(got, "duplicate hash BUG-b7aa") {
			t.Errorf("%s problems = %q", id, got)
		}
	}
}

// The prefix a file is read with is the kind it ends up as, so a file that
// says type: bug in specs/ reads as BUG and a plan file reads as PLN. The
// folder a file sits in is not the kind it is.
func TestPrefixFollowsTheKindTheFileIs(t *testing.T) {
	b := boardWith(t, map[string]string{
		"specs/2026-09-20-a-design.md": "---\ntype: bug\nid: BUG-3\nhash: m2x9\n---\n# A\n",
		"plans/2026-09-21-b.md":        "---\nid: PLAN-12\nhash: k3f2\n---\n# B\n\n### Task 1: T\n- [ ] a\n",
	})
	spec := b.Get("specs/2026-09-20-a-design")
	if spec.Kind != KindBug || spec.ShortID != "BUG-0003" || spec.Hash != "BUG-m2x9" {
		t.Fatalf("bug in specs = kind %q id %q hash %q", spec.Kind, spec.ShortID, spec.Hash)
	}
	if got := strings.Join(spec.Problems, "; "); got != "" {
		t.Fatalf("clean file has problems: %q", got)
	}
	plan := b.Get("plans/2026-09-21-b")
	if plan.ShortID != "PLN-0012" || plan.Hash != "PLN-k3f2" {
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
		"specs/2026-09-20-a-design.md": "---\nid: SPEC-zz\nhash: toolongg\n---\n# A\n",
	})
	it := b.Get("specs/2026-09-20-a-design")
	if it.ShortID != "" || it.Hash != "" {
		t.Fatalf("bad values became ids: %q %q", it.ShortID, it.Hash)
	}
	if it.RawID != "SPEC-zz" || it.RawHash != "toolongg" {
		t.Fatalf("written values lost: %q %q", it.RawID, it.RawHash)
	}
	got := strings.Join(it.Problems, "; ")
	if !strings.Contains(got, "bad id SPEC-zz") || !strings.Contains(got, "bad hash toolongg") {
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
	if !strings.HasSuffix(plan.Path, "plans/2026-09-21-b.md") || plan.ShortID != "PLN-0012" || plan.Hash != "PLN-k3f2" {
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

// hasProblem says if the item carries that problem line.
func hasProblem(it *Item, p string) bool {
	for _, pr := range it.Problems {
		if pr == p {
			return true
		}
	}
	return false
}

func TestIsIDNewFormat(t *testing.T) {
	for id, want := range map[string]bool{
		"PLN-0030": true, "PLN-0001": true, "PLN-10000": true,
		"PLN-30": false, "PLN-0000": false, "PLN-00300": false, "PLAN-30": false,
		"PLN-": false, "PLN-00a1": false, "PLN-٠٠٣٠": false, "": false,
	} {
		if got := IsID(id, "PLN"); got != want {
			t.Errorf("IsID(%q) = %v, want %v", id, got, want)
		}
	}
}

func TestIsHashSevenChars(t *testing.T) {
	for h, want := range map[string]bool{
		"oxoqk2m": true, "a000000": true,
		"oxoq": false, "oxoqk2mz": false, "1xoqk2m": false, "Oxoqk2m": false, "oxoqk2é": false, "": false,
	} {
		if got := IsHash(h); got != want {
			t.Errorf("IsHash(%q) = %v, want %v", h, got, want)
		}
	}
}

func TestBoardShowsNewFormForOldFiles(t *testing.T) {
	b := boardWith(t, map[string]string{
		"plans/2026-09-21-a.md":   "---\nid: PLAN-12\nhash: k3f2\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n\n### Task 3: Three\n- [ ] x\n\n### Task F1: Fix\n- [ ] y\n",
		"plans/2026-09-22-b.md":   "---\nid: PLN-0013\nhash: k3f2abc\n---\n# B\n\n**Spec:** none (Bounded, approved in chat on 2026-09-22)\n\n### Task 1: One\n- [ ] z\n",
		"scratch/2026-09-23-c.md": "---\nid: SCRATCH-4\nhash: m2x9\ntitle: C\nstatus: raw\n---\nbody\n",
	})
	cases := map[string]struct {
		short, hash string
		old         bool
	}{
		"plans/2026-09-21-a":         {"PLN-0012", "PLN-k3f2", true},
		"plans/2026-09-21-a#task-3":  {"PLN-0012.03", "PLN-k3f2.03", true},
		"plans/2026-09-21-a#task-F1": {"PLN-0012.F1", "PLN-k3f2.F1", true},
		"plans/2026-09-22-b":         {"PLN-0013", "PLN-k3f2abc", false},
		"scratch/2026-09-23-c":       {"SCR-0004", "SCR-m2x9", true},
	}
	for id, want := range cases {
		it := b.Get(id)
		if it == nil {
			t.Fatalf("%s not loaded", id)
		}
		if it.ShortID != want.short || it.Hash != want.hash || it.OldForm != want.old {
			t.Errorf("%s: got %q %q old=%v, want %q %q old=%v", id, it.ShortID, it.Hash, it.OldForm, want.short, want.hash, want.old)
		}
		if len(it.Problems) != 0 {
			t.Errorf("%s: unexpected problems %v", id, it.Problems)
		}
	}
}

func TestBadIDShapesStayProblems(t *testing.T) {
	for _, raw := range []string{"PLN-30", "PLN-0000", "PLAN-030", "SPEC-4", "PLN-00300"} {
		b := boardWith(t, map[string]string{
			"plans/2026-09-21-a.md": "---\nid: " + raw + "\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n",
		})
		it := b.Get("plans/2026-09-21-a")
		if it.ShortID != "" || !hasProblem(it, "bad id "+raw) {
			t.Errorf("%s: ShortID %q problems %v", raw, it.ShortID, it.Problems)
		}
	}
}

func TestOversizeNumbersGetProblems(t *testing.T) {
	var tasks strings.Builder
	tasks.WriteString("### Task 100: Big\n- [ ] x\n")
	var notes strings.Builder
	for i := 1; i <= 100; i++ {
		notes.WriteString("- [ ] note " + strconv.Itoa(i) + "\n")
	}
	b := boardWith(t, map[string]string{
		"plans/2026-09-21-a.md": "---\nid: PLN-10000\nhash: k3f2abc\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n\n" + tasks.String(),
		"debt/2026-09-27-d.md":  "---\nid: DBT-0002\n---\n# Review NOTEs\n\n" + notes.String(),
	})
	plan := b.Get("plans/2026-09-21-a")
	if plan.ShortID != "PLN-10000" || !hasProblem(plan, "id PLN-10000 is wider than 4 digits") {
		t.Errorf("plan: %q %v", plan.ShortID, plan.Problems)
	}
	if !hasProblem(plan, "task 100 is wider than 2 digits") {
		t.Errorf("plan: %v", plan.Problems)
	}
	debt := b.Get("debt/2026-09-27-d")
	if !hasProblem(debt, "item 100 is wider than 2 digits") {
		t.Errorf("debt: %v", debt.Problems)
	}
}

func TestCanon(t *testing.T) {
	for in, want := range map[string]string{
		"PLAN-30.3": "pln-30.3", "pln-0030.03": "pln-30.3", "PLN-30.3": "pln-30.3",
		"SCRATCH-14": "scr-14", "scr-0014": "scr-14", "DEBT-22.4": "dbt-22.4",
		"SPEC-4": "spc-4", "BUG-0007": "bug-7", "PLN-0030.F1": "pln-30.f1",
		"PLN-k3f2abc": "pln-k3f2abc", "": "",
	} {
		if got := Canon(in); got != want {
			t.Errorf("Canon(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGetFindsEveryForm(t *testing.T) {
	b := boardWith(t, map[string]string{
		"plans/2026-09-21-a.md": "---\nid: PLAN-12\nhash: k3f2abc\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n\n### Task 3: Three\n- [ ] x\n\n### Task F1: Fix\n- [ ] y\n",
		"plans/2026-09-22-b.md": "---\nid: PLN-0013\nhash: k3f2xyz\n---\n# B\n\n**Spec:** none (Bounded, approved in chat on 2026-09-22)\n",
		"debt/2026-09-27-d.md":  "---\nid: DBT-0002\n---\n# Review NOTEs\n\n- [ ] first\n",
	})
	for _, id := range []string{"PLN-0012", "pln-0012", "PLAN-12", "PLN-12", "plans/2026-09-21-a", "PLN-k3f2abc", "k3f2a", "k3f2abc"} {
		if it := b.Get(id); it == nil || it.ID != "plans/2026-09-21-a" {
			t.Errorf("Get(%q) = %v", id, it)
		}
	}
	for _, id := range []string{"PLN-0012.03", "PLAN-12.3", "pln-12.03"} {
		if it := b.Get(id); it == nil || it.ID != "plans/2026-09-21-a#task-3" {
			t.Errorf("Get(%q) = %v", id, it)
		}
	}
	if it := b.Get("PLN-0012.F1"); it == nil || it.ID != "plans/2026-09-21-a#task-F1" {
		t.Errorf("Get F1 = %v", it)
	}
	for _, id := range []string{"DBT-0002.01", "DEBT-2.1", "dbt-2.01", "debt/2026-09-27-d#item-1"} {
		if it := b.Get(id); it == nil || it.ID != "debt/2026-09-27-d#item-1" {
			t.Errorf("Get(%q) = %v", id, it)
		}
	}
	for _, id := range []string{"PLX-0012", "", "k3f", "k3f2", "SPC-0012", "PLAN-12.9", "PLN-0012.04"} {
		if it := b.Get(id); it != nil {
			t.Errorf("Get(%q) = %s, want nil", id, it.ID)
		}
	}
	if !b.Ambiguous("k3f2") || b.Ambiguous("k3f2a") {
		t.Error("Ambiguous wrong for k3f2 / k3f2a")
	}
	for _, id := range []string{"PLX-0012", "", "k3f", "SPC-0012", "PLAN-12", "plans/2026-09-21-a"} {
		if b.Ambiguous(id) {
			t.Errorf("Ambiguous(%q) = true, want false", id)
		}
	}
}
