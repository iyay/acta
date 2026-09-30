package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// skillFiles returns every .md file under a folder of the plugin, so one
// guard covers the whole tree and not only the file that changed today.
func skillFiles(t *testing.T, rel string) []string {
	t.Helper()
	root := filepath.Join(pluginRoot(t), filepath.FromSlash(rel))
	var out []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".md") {
			out = append(out, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatalf("no .md file under %s", rel)
	}
	return out
}

func readSkill(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestBuildSkillEndsWithReplyBack checks the build skill itself. The
// recipient has to learn the reply-back from the skill it loads, not from
// a hand-filled line in a brief.
func TestBuildSkillEndsWithReplyBack(t *testing.T) {
	txt := readSkill(t, "skills/build/SKILL.md")
	for _, want := range []string{"acta reply-back", ".dispatch.json"} {
		if !strings.Contains(txt, want) {
			t.Errorf("build/SKILL.md missing %q", want)
		}
	}
}

// TestDispatchBriefLoadsBuild checks the brief template. A recipient that
// never loads the build skill never ticks --start and never replies back.
func TestDispatchBriefLoadsBuild(t *testing.T) {
	if txt := readSkill(t, "skills/build/dispatch.md"); !strings.Contains(txt, "SKILL: load build") {
		t.Error(`build/dispatch.md missing "SKILL: load build" in the brief format`)
	}
}

// TestNoSkillWhitelistsActaCommands stops a brief from overriding the
// build skill with its own list of commands. Case-insensitive on purpose.
func TestNoSkillWhitelistsActaCommands(t *testing.T) {
	for _, p := range skillFiles(t, "skills") {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if i := strings.Index(strings.ToLower(string(b)), "only allowed"); i >= 0 {
			rel, _ := filepath.Rel(pluginRoot(t), p)
			t.Errorf("%s whitelists acta commands: ...%s...", rel, string(b)[max(0, i-40):min(i+40, len(b))])
		}
	}
}

// TestNoHandFilledReplyBack stops a skill from telling a recipient to
// fill in a review command by hand. acta reply-back builds that command.
func TestNoHandFilledReplyBack(t *testing.T) {
	for _, p := range skillFiles(t, "skills/build") {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), "<new-head-sha>") {
			rel, _ := filepath.Rel(pluginRoot(t), p)
			t.Errorf("%s still carries a hand-filled reply-back (<new-head-sha>)", rel)
		}
	}
}

// TestDispatchInitBeforeGoal checks the delivery file. Without the record
// the recipient's reply-back has no pane and no range to send.
func TestDispatchInitBeforeGoal(t *testing.T) {
	if txt := readSkill(t, "skills/build/herdr-delivery.md"); !strings.Contains(txt, "acta dispatch init") {
		t.Error("build/herdr-delivery.md missing \"acta dispatch init\"")
	}
}
