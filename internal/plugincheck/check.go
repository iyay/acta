// Package plugincheck keeps the pm plugin folder in shape: every skill has
// valid frontmatter, stays under its size cap, keeps the rules it must carry,
// and only points at files that exist.
package plugincheck

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillRule is what one skill folder must satisfy.
type SkillRule struct {
	Name     string
	MaxLines int      // newlines across every .md file in the folder
	Must     []string // text that must appear somewhere in those files
	MustNot  []string // text that must not appear
}

var (
	refRe  = regexp.MustCompile(`references/([A-Za-z0-9_-]+\.md)`)
	linkRe = regexp.MustCompile(`\]\(([A-Za-z0-9_-]+\.md)\)`)
)

// SkillProblems lists everything wrong with root/skills/<name>. Empty means fine.
func SkillProblems(root string, r SkillRule) []string {
	dir := filepath.Join(root, "skills", r.Name)
	skill, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return []string{"missing skills/" + r.Name + "/SKILL.md"}
	}
	var probs []string
	name, desc, ok := frontmatter(skill)
	if !ok {
		probs = append(probs, "SKILL.md has no frontmatter block")
	}
	if name != r.Name {
		probs = append(probs, fmt.Sprintf("frontmatter name %q, want %q", name, r.Name))
	}
	if strings.TrimSpace(desc) == "" {
		probs = append(probs, "description is empty")
	}
	if !strings.HasPrefix(desc, "acta: ") {
		probs = append(probs, `description must start with "acta: "`)
	}
	if len(desc) > 1024 {
		probs = append(probs, "description is longer than 1024 characters")
	}

	var text strings.Builder
	lines := 0
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != dir {
				if _, err := os.Stat(filepath.Join(p, "SKILL.md")); err == nil {
					probs = append(probs, "nested SKILL.md in "+p+" (omp only finds skills one level under skills/)")
				}
			}
			return nil
		}
		if !strings.HasSuffix(p, ".md") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		lines += strings.Count(string(b), "\n")
		text.Write(b)
		text.WriteString("\n")
		return nil
	})
	if err != nil {
		probs = append(probs, err.Error())
	}
	if lines > r.MaxLines {
		probs = append(probs, fmt.Sprintf("%d lines of markdown, cap is %d", lines, r.MaxLines))
	}
	all := text.String()
	for _, m := range r.Must {
		if !strings.Contains(all, m) {
			probs = append(probs, "missing required text: "+m)
		}
	}
	for _, m := range r.MustNot {
		if strings.Contains(all, m) {
			probs = append(probs, "has forbidden text: "+m)
		}
	}
	for _, m := range refRe.FindAllStringSubmatch(all, -1) {
		if !exists(filepath.Join(root, "references", m[1])) {
			probs = append(probs, "references/"+m[1]+" does not exist")
		}
	}
	for _, m := range linkRe.FindAllStringSubmatch(all, -1) {
		if !exists(filepath.Join(dir, m[1])) && !exists(filepath.Join(root, "references", m[1])) {
			probs = append(probs, "link target "+m[1]+" does not exist")
		}
	}
	return probs
}

func frontmatter(src []byte) (name, desc string, ok bool) {
	text := strings.ReplaceAll(string(src), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return "", "", false
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return "", "", false
	}
	var fm struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if yaml.Unmarshal([]byte(text[4:4+end]), &fm) != nil {
		return "", "", false
	}
	return fm.Name, fm.Description, true
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
