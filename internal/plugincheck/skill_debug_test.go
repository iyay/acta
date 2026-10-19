package plugincheck

import "testing"

// TestSkillDebug pins the evidence contract of the debug skill. Text checks
// cannot prove behaviour; the debug-* eval cases do that. These only keep the
// four phases, the safety rules and the recipe triggers from being cut.
func TestSkillDebug(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "debug",
		MaxLines: 330,
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
