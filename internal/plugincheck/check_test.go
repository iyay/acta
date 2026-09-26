package plugincheck

import (
	"strings"
	"testing"
)

func joined(p []string) string { return strings.Join(p, "\n") }

func TestSkillProblemsGood(t *testing.T) {
	p := SkillProblems("testdata/plugin", SkillRule{Name: "good", MaxLines: 20, Must: []string{"Needed phrase"}, MustNot: []string{"superpowers:"}})
	if len(p) != 0 {
		t.Fatalf("good skill has problems:\n%s", joined(p))
	}
}

func TestSkillProblemsBad(t *testing.T) {
	p := joined(SkillProblems("testdata/plugin", SkillRule{Name: "bad", MaxLines: 2, Must: []string{"Needed phrase"}, MustNot: []string{"superpowers:"}}))
	for _, want := range []string{
		`frontmatter name "wrong", want "bad"`,
		"description is empty",
		"lines of markdown, cap is 2",
		"missing required text: Needed phrase",
		"has forbidden text: superpowers:",
		"references/missing.md does not exist",
		"link target gone.md does not exist",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("missing problem %q in:\n%s", want, p)
		}
	}
}

func TestSkillProblemsNestedAndMissing(t *testing.T) {
	if p := joined(SkillProblems("testdata/plugin", SkillRule{Name: "nested", MaxLines: 100})); !strings.Contains(p, "nested SKILL.md") {
		t.Errorf("nested skill not caught:\n%s", p)
	}
	if p := joined(SkillProblems("testdata/plugin", SkillRule{Name: "absent", MaxLines: 100})); !strings.Contains(p, "missing skills/absent/SKILL.md") {
		t.Errorf("missing skill not caught:\n%s", p)
	}
}
