package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readBuildFile reads one file of the build skill on its own. CheckSkill joins
// the whole folder, so a revert of one file could stay green.
func readBuildFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestNoDispatchSkill keeps build the only skill that runs a plan. A model
// once loaded the dispatch skill on its own and refused with a wrong reason.
func TestNoDispatchSkill(t *testing.T) {
	if _, err := os.Stat(filepath.Join(pluginRoot(t), "skills", "dispatch")); !os.IsNotExist(err) {
		t.Error("plugin/skills/dispatch still exists; dispatch lives in skills/build/dispatch.md")
	}
	err := filepath.WalkDir(pluginRoot(t), func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "node_modules" || d.Name() == ".git") {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), "acta:dispatch") {
			rel, _ := filepath.Rel(pluginRoot(t), p)
			t.Errorf("%s still names acta:dispatch", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestBuildDispatchDelivery checks the rules the one short dispatch.md still
// owns. The brief, the herdr steps and the goal text moved into Go, where
// internal/cli tests them. The file is read on its own.
func TestBuildDispatchDelivery(t *testing.T) {
	txt := readBuildFile(t, "dispatch.md")
	for _, want := range []string{
		"Without `HERDR_ENV=1` there is no pane to hand the plan to: `dispatch` runs as `subagent`",
		"acta dispatch send --plan .acta/plans/<stem>.md --rules <abs path>",
		"this skill's base dir plus `../../references/house-rules.md`",
		"Read those lines and the exit code",
		"`checkpoint: ok`", "`unconfirmed`", "drift",
		"## Never wait for the recipient — HARD RULE", "end your turn",
		"herdr agent wait <slug> --until idle --until done", "run_in_background",
		"1. A reply-back already arrived", "2. Every task of the plan is ticked",
		"3. Tasks still open", "4. Idle again with no new commit",
		"git log --oneline <base>..HEAD", "git diff --name-only <base>..HEAD      # only plan files",
		"the plan's fast test command", "git diff --name-only <base>..HEAD -- .acta/bugs",
		"Bugs found by recipient", "Harvested from omp", "omp memory: nothing to harvest", "learned.md",
		"acta:review", "acta:land",
		"Its small-change self-review never applies to a dispatch",
		"literal skill name, a review keyword, and the range",
		"NOTEs are never written to memory", "sorts them into `[fix]`, `[debt]` or `[note]`",
		"## Fix rounds", "Append `## Fix round <n>`", "--round fix-<n>", "--round polish",
		"The polish goes out the same way",
		"`## After a CLEAN round` owns what happens on reply-back",
		"Clean and complete: run `## After a CLEAN round`; its `acta:land` is build's `## Close`",
		"## Close the tab", "acta dispatch close", "## Advisor", "`advisor` role",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("build/dispatch.md missing %q", want)
		}
	}
	// These belong to build, acta:review, acta:land or Go now. A second copy
	// drifts, and drift is how duplicate ids once reached the parent branch.
	for _, bad := range []string{
		"acta:dispatch", "Step -3", "Refuse inside omp", "dispatch requires herdr", "## Entry gate",
		"## Autonomous loop", "## Landing", "NOTEs go to memory", "--no-ff", "git branch -d",
		"files them with acta debt new", "files them with `acta debt new`",
		"superpowers:", "git-bug", "/Users/", "<new-head-sha>", "fix round R of 5", "until herdr agent read",
		"one NOTE (memory)",
		"on reply-back the two reviewers review the polish range; it is not a fix round",
		"orchestrator sorts them through",
		"orchestrator sorts it through",
		"read it again", "<one-shot test runner>",
		"then build's `## Close` runs `acta:land`",
		// The new send never sends /new: a fresh omp process is the new
		// session. The two spellings are banned instead of "/new", which
		// would also hit words like "/newest".
		"herdr-delivery.md", "New session started", "`/new`", "send /new", "Back-to-back, no settle-wait",
	} {
		if strings.Contains(txt, bad) {
			t.Errorf("build/dispatch.md still carries %q", bad)
		}
	}
	if _, err := os.Stat(filepath.Join(pluginRoot(t), "skills", "build", "herdr-delivery.md")); !os.IsNotExist(err) {
		t.Error("build/herdr-delivery.md still exists; dispatch.md is the one dispatch file")
	}
}
