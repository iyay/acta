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

// TestBuildDispatchDelivery checks the delivery rules that came from the old
// dispatch skill. Each file is read on its own.
func TestBuildDispatchDelivery(t *testing.T) {
	for file, wants := range map[string][]string{
		"dispatch.md": {
			"/goal", "REPLY-BACK", "references/house-rules.md", "SKILL: load build", "Never wait",
			"herdr-delivery.md", "PROPERTY", "ultrathink orchestrate", ".acta/plans/",
			"GATES (from the worktree): <the plan's fast test command>",
			"exactly one read", "checkpoint unconfirmed", "gets its own one read", "outside this rule",
			"HARD RULE", "New session started", "🎯 Goal", "never in one prompt", "then `/goal` only",
			"acta:review", "acta:land", "herdr pane close", "Bugs found by recipient",
			"Without `HERDR_ENV=1` there is no pane to hand the plan to: `dispatch` runs as `subagent`",
			"literal skill name, a review keyword, and the range", "sorts them",
		},
		"herdr-delivery.md": {
			"exactly one read", "checkpoint unconfirmed", "HARD RULE", "New session started", "🎯 Goal",
			"never in one prompt", "then `/goal` only", "--agent omp", "acta dispatch init", "herdr pane close",
			"Reference for a dispatch already running",
		},
	} {
		txt := readBuildFile(t, file)
		for _, want := range wants {
			if !strings.Contains(txt, want) {
				t.Errorf("build/%s missing %q", file, want)
			}
		}
		if strings.Contains(txt, "Back-to-back, no settle-wait") {
			t.Errorf("build/%s still says \"Back-to-back, no settle-wait\"", file)
		}
	}
	txt := readBuildFile(t, "dispatch.md")
	if n := strings.Count(txt, "checkpoint unconfirmed"); n < 2 {
		t.Errorf("build/dispatch.md says \"checkpoint unconfirmed\" %d times; the loop item and the section both need it", n)
	}
	// These belong to build, acta:review and acta:land now. A second copy
	// drifts, and drift is how duplicate ids once reached the parent branch.
	for _, bad := range []string{
		"acta:dispatch", "Step -3", "Refuse inside omp", "dispatch requires herdr", "## Entry gate",
		"## Autonomous loop", "## Landing", "NOTEs go to memory", "--no-ff", "git branch -d",
		"files them with acta debt new", "files them with `acta debt new`",
		"superpowers:", "git-bug", "/Users/", "<new-head-sha>", "fix round R of 5", "until herdr agent read",
		"read it again", "<one-shot test runner>",
	} {
		if strings.Contains(txt, bad) {
			t.Errorf("build/dispatch.md still carries %q", bad)
		}
	}
}

// TestBuildDispatchGatesNoBareGoTest reads dispatch.md on its own. The GATES
// line of the brief is where the recipient learns how to run tests, so the
// bare `go test` ban and the hook that enforces it belong there.
func TestBuildDispatchGatesNoBareGoTest(t *testing.T) {
	txt := readBuildFile(t, "dispatch.md")
	if !strings.Contains(txt, "never run bare `go test`") {
		t.Error("build/dispatch.md missing \"never run bare `go test`\"")
	}
	if !strings.Contains(txt, "the pre-tool hook blocks it") {
		t.Error("build/dispatch.md does not say the pre-tool hook blocks it")
	}
}
