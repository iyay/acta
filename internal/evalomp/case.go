// Package evalomp runs the plugin's eval cases through omp, so the plugin is
// checked in omp the same way claude plugin eval checks it in Claude Code.
// The cases are the same files; only the runner differs.
package evalomp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// defaultTimeout is claude plugin eval's own default, so a case with no
// timeout gets the same cap in both harnesses.
const defaultTimeout = 300

// Target says what a grader looks at: the final reply, or one file.
type Target struct {
	Kind string
	Path string
}

// UnmarshalYAML takes both forms a grader file uses: a plain word like
// last_message, or a map with source and path.
func (t *Target) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		t.Kind = n.Value
		return nil
	}
	var m struct {
		Source string `yaml:"source"`
		Path   string `yaml:"path"`
	}
	if err := n.Decode(&m); err != nil {
		return err
	}
	t.Kind, t.Path = m.Source, m.Path
	return nil
}

// Grader is one file under graders/. Name is the file name without .md, and
// Body is the text after the frontmatter, which is the rubric of an llm grader.
type Grader struct {
	Name       string `yaml:"-"`
	Type       string `yaml:"type"`
	Path       string `yaml:"path"`
	Exists     *bool  `yaml:"exists"`
	Pattern    string `yaml:"pattern"`
	Flags      string `yaml:"flags"`
	Match      string `yaml:"match"`
	Target     Target `yaml:"target"`
	Tool       string `yaml:"tool"`
	InputMatch string `yaml:"input_match"`
	Min        *int   `yaml:"min"`
	Max        *int   `yaml:"max"`
	Body       string `yaml:"-"`
}

// Case is one folder under the eval folder.
type Case struct {
	Name           string
	Dir            string
	Prompt         string
	Tags           []string
	TimeoutSeconds int
	Scaffold       string
	Graders        []Grader
}

// ClaudeOnly says the case only makes sense in Claude Code. A tag and not a
// field, because claude plugin eval refuses frontmatter keys it does not know.
func (c Case) ClaudeOnly() bool {
	return slices.Contains(c.Tags, "claude-only")
}

// LoadCases reads every case under evalDir, in name order. A folder with no
// prompt.md is not a case, so a results folder next to the cases is skipped.
func LoadCases(evalDir string) ([]Case, error) {
	entries, err := os.ReadDir(evalDir)
	if err != nil {
		return nil, err
	}
	var cases []Case
	for _, e := range entries {
		dir := filepath.Join(evalDir, e.Name())
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "prompt.md")); err != nil {
			continue
		}
		c, err := loadCase(dir)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		cases = append(cases, c)
	}
	return cases, nil
}

func loadCase(dir string) (Case, error) {
	src, err := os.ReadFile(filepath.Join(dir, "prompt.md"))
	if err != nil {
		return Case{}, err
	}
	fm, body, ok := splitFrontmatter(string(src))
	if !ok {
		return Case{}, errors.New("prompt.md has no frontmatter")
	}
	var meta struct {
		Tags    []string `yaml:"tags"`
		Timeout int      `yaml:"timeout_seconds"`
	}
	if err := yaml.Unmarshal([]byte(fm), &meta); err != nil {
		return Case{}, fmt.Errorf("prompt.md: %w", err)
	}
	c := Case{
		Name:           filepath.Base(dir),
		Dir:            dir,
		Prompt:         strings.TrimSpace(body),
		Tags:           meta.Tags,
		TimeoutSeconds: meta.Timeout,
	}
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = defaultTimeout
	}
	if y, err := os.ReadFile(filepath.Join(dir, "case.yaml")); err == nil {
		var cy struct {
			Context struct {
				Scaffold string `yaml:"scaffold_script"`
			} `yaml:"context"`
		}
		if err := yaml.Unmarshal(y, &cy); err != nil {
			return Case{}, fmt.Errorf("case.yaml: %w", err)
		}
		if cy.Context.Scaffold != "" {
			c.Scaffold = filepath.Join(dir, cy.Context.Scaffold)
		}
	}
	files, err := filepath.Glob(filepath.Join(dir, "graders", "*.md"))
	if err != nil {
		return Case{}, err
	}
	for _, f := range files {
		g, err := loadGrader(f)
		if err != nil {
			return Case{}, fmt.Errorf("%s: %w", filepath.Base(f), err)
		}
		c.Graders = append(c.Graders, g)
	}
	return c, nil
}

func loadGrader(path string) (Grader, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return Grader{}, err
	}
	fm, body, ok := splitFrontmatter(string(src))
	if !ok {
		return Grader{}, errors.New("no frontmatter")
	}
	var g Grader
	if err := yaml.Unmarshal([]byte(fm), &g); err != nil {
		return Grader{}, err
	}
	g.Name = strings.TrimSuffix(filepath.Base(path), ".md")
	g.Body = strings.TrimSpace(body)
	return g, nil
}

// splitFrontmatter cuts a file into the YAML between the two --- lines and the
// text after them. A grader file is often only frontmatter, so the closing
// line may be the very end of the file.
func splitFrontmatter(src string) (fm, body string, ok bool) {
	rest, found := strings.CutPrefix(src, "---\n")
	if !found {
		return "", "", false
	}
	if fm, body, found = strings.Cut(rest, "\n---\n"); found {
		return fm, body, true
	}
	if rest == "---" || strings.HasPrefix(rest, "---\n") {
		// Empty frontmatter: the closing line comes right away.
		return "", strings.TrimPrefix(strings.TrimPrefix(rest, "---"), "\n"), true
	}
	fm, found = strings.CutSuffix(strings.TrimRight(rest, "\n"), "\n---")
	return fm, "", found
}
