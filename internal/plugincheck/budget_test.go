package plugincheck

import (
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
)

// The caps below are sizes in bytes. About 4 bytes make one token. A byte
// count is the same on every run and needs no API, so the test stays steady.
//
// Each cap starts at the size the thing has today. To make a skill leaner,
// lower its number in the same commit, so the gain cannot slip away. To add a
// file, add its line here on purpose.

// fileCaps holds the cap of every .md file under skills/, references/ and
// output-styles/. The keys are paths as walkPlugin names them.
var fileCaps = map[string]int{
	"output-styles/acta.md":                   1849,
	"references/house-rules.md":               3787,
	"references/unslop.md":                    3752,
	"skills/bug/SKILL.md":                     2327,
	"skills/build/SKILL.md":                   12651,
	"skills/build/dispatch.md":                6633,
	"skills/build/implementer-prompt.md":      7557,
	"skills/build/wiki.md":                    2750,
	"skills/debug/SKILL.md":                   6231,
	"skills/debug/condition-based-waiting.md": 3582,
	"skills/debug/defense-in-depth.md":        1505,
	"skills/debug/root-cause-tracing.md":      2721,
	"skills/debug/race.md":                    1751,
	"skills/frame/SKILL.md":                   4591,
	"skills/frame/startup.md":                 4337,
	"skills/land/SKILL.md":                    8885,
	"skills/land/baseline.md":                 2434,
	"skills/land/tidy.md":                     2450,
	"skills/lean/SKILL.md":                    1918,
	"skills/migrate/SKILL.md":                 3207,
	"skills/review/SKILL.md":                  13330,
	"skills/review/code-reviewer.md":          4042,
	"skills/review/depth.md":                  2085,
	"skills/review/fact-check.md":             1850,
	"skills/scratch/SKILL.md":                 2358,
	"skills/setup/SKILL.md":                   5536,
	"skills/shape/SKILL.md":                   7841,
	"skills/shape/probe.md":                   1717,
	"skills/slice/SKILL.md":                   12768,
	"skills/slice/split.md":                   2123,
	"skills/tdd/SKILL.md":                     10702,
	"skills/tdd/writing-good-tests.md":        8292,
}

// descriptionCaps holds the cap of the description line in each skill's
// SKILL.md. Every session pays for these lines, so they stay short.
var descriptionCaps = map[string]int{
	"skills/bug/SKILL.md":     311,
	"skills/build/SKILL.md":   541,
	"skills/debug/SKILL.md":   277,
	"skills/frame/SKILL.md":   345,
	"skills/land/SKILL.md":    304,
	"skills/lean/SKILL.md":    139,
	"skills/migrate/SKILL.md": 258,
	"skills/review/SKILL.md":  386,
	"skills/scratch/SKILL.md": 223,
	"skills/setup/SKILL.md":   230,
	"skills/shape/SKILL.md":   222,
	"skills/slice/SKILL.md":   234,
	"skills/tdd/SKILL.md":     208,
}

// sessionStartCap is the cap of the text the session start hook prints for
// the default user config with no herdr. Every session pays for it too. It
// grew by 207 bytes when the inline default executor line joined it.
const sessionStartCap = 3068

// budgetProblems lists what is wrong between sizes and caps: a size over its
// cap, a size with no cap, and a cap for something that is gone. The first two
// show the size in bytes and in tokens, so it is easy to judge the fix. The
// suffix goes after each name, so a description is not mistaken for its whole
// file.
func budgetProblems(suffix string, sizes, caps map[string]int) []string {
	var out []string
	for _, name := range slices.Sorted(maps.Keys(sizes)) {
		n := sizes[name]
		c, ok := caps[name]
		switch {
		case !ok:
			out = append(out, fmt.Sprintf("%s%s: %d bytes (~%d tokens), has no cap", name, suffix, n, n/4))
		case n > c:
			out = append(out, fmt.Sprintf("%s%s: %d bytes (~%d tokens), cap %d", name, suffix, n, n/4, c))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(caps)) {
		if _, ok := sizes[name]; !ok {
			out = append(out, fmt.Sprintf("%s%s: file is gone, remove its cap of %d", name, suffix, caps[name]))
		}
	}
	return out
}

// TestBudgetFiles fails when a skill, reference or output style file grows past
// its cap, has no cap, or has a cap but is gone. Every problem shows in one run.
func TestBudgetFiles(t *testing.T) {
	sizes := map[string]int{}
	walkPlugin(t, func(rel, text string) {
		if strings.HasSuffix(rel, ".md") && (strings.HasPrefix(rel, "skills/") || strings.HasPrefix(rel, "references/") || strings.HasPrefix(rel, "output-styles/")) {
			sizes[rel] = len(text)
		}
	})
	for _, p := range budgetProblems("", sizes, fileCaps) {
		t.Error(p)
	}
}

// TestBudgetDescriptions does the same for the description line of each skill.
// It measures the text the model reads, without the quotes of the YAML.
func TestBudgetDescriptions(t *testing.T) {
	sizes := map[string]int{}
	// If a skill's frontmatter cannot be read, its description has no size. Its
	// cap is dropped from this copy, so the test does not also say it is gone.
	caps := maps.Clone(descriptionCaps)
	walkPlugin(t, func(rel, text string) {
		// The star matches one folder name, so a nested SKILL.md is not a skill.
		if ok, _ := path.Match("skills/*/SKILL.md", rel); ok {
			_, desc, fmOK := frontmatter([]byte(text))
			if !fmOK {
				t.Errorf("%s: frontmatter cannot be read, so the description cannot be measured", rel)
				delete(caps, rel)
				return
			}
			sizes[rel] = len(desc)
		}
	})
	for _, p := range budgetProblems(" description", sizes, caps) {
		t.Error(p)
	}
}

// TestBudgetSessionStart fails when the session start text grows past its cap.
func TestBudgetSessionStart(t *testing.T) {
	n := len(hook.SessionStart(hook.Input{Voice: config.UserDefault(), VoiceExists: true}))
	if n > sessionStartCap {
		t.Errorf("session start text: %d bytes (~%d tokens), cap %d", n, n/4, sessionStartCap)
	}
}
