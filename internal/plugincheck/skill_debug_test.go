package plugincheck

import (
	"strings"
	"testing"
)

// TestSkillDebug pins the evidence contract of the debug skill. Text checks
// cannot prove behaviour; the debug-* eval cases do that. These only keep the
// four phases, the safety rules and the recipe triggers from being cut.
func TestSkillDebug(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name: "debug",
		// race.md holds the copyable race probe, which made the folder longer.
		// The known-flake rerun rule in Phase 1 adds two more lines.
		MaxLines: 350,
		Must: []string{
			// The four phases and their exit evidence.
			"Symptom and feedback loop", "Narrow and distinguish", "Prove the mechanism", "Fix, verify and clean up",
			"exact symptom", "not proven", "observed", "unverified",
			// Safety and the acta gates.
			"outside the repo", "Redact secrets", "never push", "live systems",
			"acta:bug", "acta:shape", "acta:tdd", "acta:land", "Three failed fixes",
			// Recipes, each with the evidence that matches it.
			"force the interleaving", "smallest command that keeps the failing order",
			"count the work", "unmeasured", "failure count",
			"root-cause-tracing.md", "condition-based-waiting.md", "defense-in-depth.md",
			"trust boundar",
		},
		MustNot: []string{
			"superpowers:", "CREATION-LOG", "test-pressure", "test-academic", "git-bug",
			"add diagnostic instrumentation", "SMALLEST possible change", "When you can't trace manually, add instrumentation:",
			// Old advice the design replaced: blanket checks, made-up numbers,
			// and printing secrets to find a bug.
			"Validate at EVERY layer", "Add validation at each layer", "Bug impossible",
			"95%", "1847 tests", "60% → 100%", "40% faster",
			"security find-identity", "env | grep",
		},
	})
}

// TestDebugRerunsKnownFlake reads SKILL.md on its own. A test that an open
// bug already marks as flaky gets one lone rerun before any root-cause work,
// so a known flake is logged instead of debugged again. The rule has to sit
// in Phase 1, before the narrowing starts.
func TestDebugRerunsKnownFlake(t *testing.T) {
	txt := readSkill(t, "skills/debug/SKILL.md")
	for _, want := range []string{
		"`flaky_tests`", "acta list --type bug --json",
		"runs once more alone first", "a pass runs `acta flake record <name> --plan <plan id>`",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("debug/SKILL.md missing %q", want)
		}
	}
	rule, phase2 := strings.Index(txt, "runs once more alone first"), strings.Index(txt, "## Phase 2")
	if rule < 0 || phase2 < 0 || rule > phase2 {
		t.Errorf("debug/SKILL.md must give the flake rerun before Phase 2 (rule %d, Phase 2 %d)", rule, phase2)
	}
}

// TestRaceProbeStartsFresh reads race.md on its own. A probe that reuses one
// state file adds to the last run's count, so a rerun or the lock control can
// print the right number on code that is still racy.
func TestRaceProbeStartsFresh(t *testing.T) {
	txt := readSkill(t, "skills/debug/race.md")
	if !strings.Contains(txt, "tempfile.mkdtemp()") {
		t.Error("debug/race.md must give each run a fresh folder with tempfile.mkdtemp()")
	}
	if strings.Contains(txt, "/tmp/probe") {
		t.Error("debug/race.md still uses a fixed /tmp/probe path")
	}
}
