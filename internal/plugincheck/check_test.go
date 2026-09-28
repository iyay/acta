package plugincheck

import (
	"os"
	"path/filepath"
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

func TestSkillDescriptionPrefix(t *testing.T) {
	for _, c := range []struct {
		name, frontmatter, want string
	}{
		{"no prefix", "---\nname: x\ndescription: Use when...\n---\n", `description must start with "acta: "`},
		{"no space", "---\nname: x\ndescription: \"acta:Use when...\"\n---\n", `description must start with "acta: "`},
		{"capital", "---\nname: x\ndescription: \"Acta: Use when...\"\n---\n", `description must start with "acta: "`},
		{"middle", "---\nname: x\ndescription: \"Use acta: when...\"\n---\n", `description must start with "acta: "`},
		{"good", "---\nname: x\ndescription: \"acta: Use when...\"\n---\n", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "skills", "x")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(c.frontmatter+"\n# x\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			p := joined(SkillProblems(root, SkillRule{Name: "x", MaxLines: 100}))
			if c.want == "" {
				if p != "" {
					t.Errorf("want no problem, got:\n%s", p)
				}
				return
			}
			if p != c.want {
				t.Errorf("want problem %q, got %q", c.want, p)
			}
		})
	}
}
