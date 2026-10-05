package write

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
)

// statePlan is a plan with two tasks and no State section yet.
const statePlan = "---\nstatus: in-progress\n---\n# Live work state\n\n### Task 1: acta state set\n- [ ] test\n- [ ] code\n"

// stateRepo makes a git repo with one plan and gives the config and the path
// the plan file has on disk.
func stateRepo(t *testing.T, plan string) (config.Config, string) {
	t.Helper()
	cfg := repoWith(t, map[string]string{".acta/plans/2026-10-05-live-work-state.md": plan})
	return cfg, filepath.Join(cfg.RepoRoot, ".acta", "plans", "2026-10-05-live-work-state.md")
}

func stateBody(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// A plan with no State section gets one, holding only the part that was set.
func TestSetStateCreatesMissingSection(t *testing.T) {
	cfg, path := stateRepo(t, statePlan)
	o, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "next", []byte("write the tests\n"))
	if err != nil || !o.Committed {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	want := statePlan + "\n## State\n\n### Next\n\nwrite the tests\n"
	if got := stateBody(t, path); got != want {
		t.Fatalf("file = %q\nwant %q", got, want)
	}
}

// A missing subsection inside a State section that is there goes in the slot
// the order fixes, and the part already written stays byte for byte.
func TestSetStateAddsMissingSubsectionInOrder(t *testing.T) {
	cfg, path := stateRepo(t, statePlan)
	if _, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "next", []byte("first\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "findings", []byte("second\n")); err != nil {
		t.Fatal(err)
	}
	want := statePlan + "\n## State\n\n### Next\n\nfirst\n\n### Findings\n\nsecond\n"
	if got := stateBody(t, path); got != want {
		t.Fatalf("file = %q\nwant %q", got, want)
	}
}

// A hand edit left a State section with only the last part. Setting the first
// one puts it in front, not behind.
func TestSetStatePutsAMissingFirstPartInFront(t *testing.T) {
	plan := statePlan + "\n## State\n\n### Open rulings\n\nask the user\n"
	cfg, path := stateRepo(t, plan)
	if _, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "next", []byte("run the gate\n")); err != nil {
		t.Fatal(err)
	}
	want := statePlan + "\n## State\n\n### Next\n\nrun the gate\n\n### Open rulings\n\nask the user\n"
	if got := stateBody(t, path); got != want {
		t.Fatalf("file = %q\nwant %q", got, want)
	}
}

// A subsection that is there is replaced whole, and the tasks below it and
// the other subsections stay exactly as they were.
func TestSetStateReplacesOnlyThatSubsection(t *testing.T) {
	plan := statePlan + "\n## State\n\n### Next\n\nold next\n\n### Findings\n\nold findings\n\n### Open rulings\n\nold rulings\n\n## After\n\nkeep me\n"
	cfg, path := stateRepo(t, plan)
	if _, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "findings", []byte("new findings\n")); err != nil {
		t.Fatal(err)
	}
	want := statePlan + "\n## State\n\n### Next\n\nold next\n\n### Findings\n\nnew findings\n\n### Open rulings\n\nold rulings\n\n## After\n\nkeep me\n"
	if got := stateBody(t, path); got != want {
		t.Fatalf("file = %q\nwant %q", got, want)
	}
}

// An empty body clears the subsection: the heading stays, no body lines are
// left, and the other parts are not touched.
func TestSetStateEmptyBodyClearsTheSubsection(t *testing.T) {
	plan := statePlan + "\n## State\n\n### Next\n\nold next\n\n### Findings\n\nkeep\n"
	cfg, path := stateRepo(t, plan)
	if _, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "next", []byte("\n  \n")); err != nil {
		t.Fatal(err)
	}
	want := statePlan + "\n## State\n\n### Next\n\n### Findings\n\nkeep\n"
	if got := stateBody(t, path); got != want {
		t.Fatalf("file = %q\nwant %q", got, want)
	}
}

// The commit says which plan the state belongs to, and the file is the only
// thing in it.
func TestSetStateCommitsWithTheStateSubject(t *testing.T) {
	cfg, path := stateRepo(t, statePlan)
	before := gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD")
	if _, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "rulings", []byte("which side?\n")); err != nil {
		t.Fatal(err)
	}
	if got := gitRun(t, cfg.RepoRoot, "log", "-1", "--format=%s"); got != "acta: state plans/2026-10-05-live-work-state" {
		t.Fatalf("commit subject %q", got)
	}
	if got := gitRun(t, cfg.RepoRoot, "show", "--name-only", "--format=", "HEAD"); got != ".acta/plans/2026-10-05-live-work-state.md" {
		t.Fatalf("commit holds %q", got)
	}
	if gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD") == before {
		t.Fatal("nothing was committed")
	}
	if got := stateBody(t, path); !strings.Contains(got, "### Open rulings\n\nwhich side?\n") {
		t.Fatalf("rulings not written under the Open rulings heading: %q", got)
	}
}

// Ten lines fit. The eleventh is refused with the limit named, and the file on
// disk stays byte for byte as it was.
func TestSetStateTenLinesFitAndElevenDoNot(t *testing.T) {
	cfg, path := stateRepo(t, statePlan)
	ten := strings.TrimSuffix(strings.Repeat("line\n", 10), "\n")
	if _, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "next", []byte(ten)); err != nil {
		t.Fatal(err)
	}
	if got := stateBody(t, path); !strings.Contains(got, "### Next\n\n"+ten+"\n") {
		t.Fatalf("ten lines not written: %q", got)
	}

	before := stateBody(t, path)
	head := gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD")
	_, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "next", []byte(ten+"\neleven\n"))
	if err == nil || !strings.Contains(err.Error(), "at most 10 lines") {
		t.Fatalf("err = %v, want one naming the 10 line limit", err)
	}
	if got := stateBody(t, path); got != before {
		t.Fatalf("file changed: %q", got)
	}
	if gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD") != head {
		t.Fatal("a refused body made a commit")
	}
}

// A part name outside the three is refused with all three named, and the file
// is left alone.
func TestSetStateUnknownPartIsRefused(t *testing.T) {
	t.Parallel()

	for _, part := range []string{"ruling", "Next", "", "findings "} {
		cfg, path := stateRepo(t, statePlan)
		before := stateBody(t, path)
		_, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", part, []byte("x\n"))
		if err == nil {
			t.Fatalf("part %q was accepted", part)
		}
		msg := err.Error()
		for _, want := range []string{"next", "findings", "rulings"} {
			if !strings.Contains(msg, want) {
				t.Errorf("part %q: error %q does not name %q", part, msg, want)
			}
		}
		if got := stateBody(t, path); got != before {
			t.Errorf("part %q changed the file: %q", part, got)
		}
	}
}

// A plan the board does not know is refused, and so is an id that names some
// other kind of item. Nothing is written either way.
func TestSetStateUnknownPlanIsRefused(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".acta/plans/2026-10-05-live-work-state.md": statePlan,
		".acta/bugs/2026-10-05-crash.md":            "---\nid: BUG-1\n---\n# Crash\n",
	})
	path := filepath.Join(cfg.RepoRoot, ".acta", "plans", "2026-10-05-live-work-state.md")
	before := stateBody(t, path)
	for _, id := range []string{"plans/2026-10-06-nope", "BUG-1", "PLAN-9"} {
		_, err := SetState(cfg, mustLoad(t, cfg), id, "next", []byte("x\n"))
		if err == nil {
			t.Fatalf("id %q was accepted", id)
		}
		if !strings.Contains(err.Error(), "unknown id "+id) && !strings.Contains(err.Error(), "is not a plan") {
			t.Errorf("id %q: error %q", id, err)
		}
	}
	if got := stateBody(t, path); got != before {
		t.Fatalf("file changed: %q", got)
	}
}

// With auto_commit off the file is written and left uncommitted, the way the
// other write commands behave.
func TestSetStateAutoCommitOffWritesNoCommit(t *testing.T) {
	t.Parallel()

	cfg, path := stateRepo(t, statePlan)
	cfg.AutoCommit = false
	head := gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD")
	o, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "next", []byte("x\n"))
	if err != nil || o.Committed || o.Skipped || o.Reason == "" {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	if gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD") != head {
		t.Fatal("auto_commit off still committed")
	}
	if got := stateBody(t, path); !strings.Contains(got, "### Next\n\nx\n") {
		t.Fatalf("file not written: %q", got)
	}
}

// Every byte of the file outside the one subsection that was named stays
// where it was. The plan keeps its CRLF line endings, its trailing
// whitespace, and a section a person wrote by hand after the State section.
func TestSetStateTouchesNothingOutsideTheSubsection(t *testing.T) {
	plan := "---\r\nstatus: in-progress\r\n---\r\n# Live work state\r\n\r\n" +
		"### Task 1: acta state set\r\n- [ ] test   \r\n\r\n" +
		"## State\r\n\r\n### Next\r\n\r\nold next\r\n\r\n### Findings\r\n\r\nkeep this\r\n\r\n" +
		"## Notes\r\n\r\nhand written   \r\n"
	cfg, path := stateRepo(t, plan)
	if _, err := SetState(cfg, mustLoad(t, cfg), "plans/2026-10-05-live-work-state", "next", []byte("new next\n")); err != nil {
		t.Fatal(err)
	}
	want := "---\r\nstatus: in-progress\r\n---\r\n# Live work state\r\n\r\n" +
		"### Task 1: acta state set\r\n- [ ] test   \r\n\r\n" +
		"## State\r\n\r\n### Next\r\n\r\nnew next\r\n\r\n### Findings\r\n\r\nkeep this\r\n\r\n" +
		"## Notes\r\n\r\nhand written   \r\n"
	if got := stateBody(t, path); got != want {
		t.Fatalf("file = %q\nwant %q", got, want)
	}
}
