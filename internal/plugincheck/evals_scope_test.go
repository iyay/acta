package plugincheck

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// tagsLine matches the tags line in the head of a case's prompt.md.
var tagsLine = regexp.MustCompile(`(?m)^tags: \[(.*)\]$`)

// caseTags reads the tags of one eval case.
func caseTags(t *testing.T, folder string) []string {
	t.Helper()
	m := tagsLine.FindStringSubmatch(readFile(t, "evals", folder, "prompt.md"))
	if m == nil {
		t.Fatalf("plugin/evals/%s/prompt.md has no tags line", folder)
	}
	var tags []string
	for _, tag := range strings.Split(m[1], ",") {
		tags = append(tags, strings.TrimSpace(tag))
	}
	return tags
}

// TestEvalCasesCarryTheirSkillTag keeps the narrow eval gate honest. A land
// that changed one skill runs `scripts/eval --tag=<skill folder>`, so a case
// that guards a skill but lacks that skill's tag would never run for it. A case
// in the evalCases table must carry the folder of the file it guards. Every
// other case must carry the name of at least one skill folder. A case that
// guards a file outside skills/ needs no skill tag: a change there runs the
// whole gate.
func TestEvalCasesCarryTheirSkillTag(t *testing.T) {
	skillDirs, err := os.ReadDir(filepath.Join(pluginRoot(t), "skills"))
	if err != nil {
		t.Fatal(err)
	}
	var skills []string
	for _, d := range skillDirs {
		if d.IsDir() {
			skills = append(skills, d.Name())
		}
	}
	inTable := map[string]bool{}
	for _, c := range evalCases {
		inTable[c.folder] = true
		parts := strings.Split(c.file, "/")
		if parts[0] != "skills" {
			continue
		}
		if !slices.Contains(caseTags(t, c.folder), parts[1]) {
			t.Errorf("case %s guards %s but lacks the tag %q, so a narrow run for that skill skips it", c.folder, c.file, parts[1])
		}
	}
	cases, err := os.ReadDir(filepath.Join(pluginRoot(t), "evals"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		if !c.IsDir() || inTable[c.Name()] {
			continue
		}
		tags := caseTags(t, c.Name())
		if !slices.ContainsFunc(tags, func(tag string) bool { return slices.Contains(skills, tag) }) {
			t.Errorf("case %s has tags %v and none is a skill folder, so no narrow run picks it", c.Name(), tags)
		}
	}
}

// TestRepoNamesItsEvalGate keeps the four parts of this repo's slow gate in
// CLAUDE.md, where slice reads them for a plan's **Gates:** line. The narrow
// form uses --tag= with the equals sign: with a space the eval tool takes the
// next word, the plugin folder, as a second tag and finds no case.
func TestRepoNamesItsEvalGate(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(pluginRoot(t), "..", "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, want := range []string{
		"## Slow gate",
		"`scripts/eval`",
		"Name: `eval`",
		"runs when the diff touches `plugin/skills/` or `plugin/hooks/`",
		"`scripts/eval --tag=<skill folder>`",
		"`scripts/eval --case <name>`",
		"whole when the diff touches `plugin/hooks/` or `plugin/references/`",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("CLAUDE.md does not name the eval gate part %q", want)
		}
	}
	if strings.Contains(text, "--tag <") {
		t.Error("CLAUDE.md shows --tag with a space; the eval tool then reads the plugin folder as a tag")
	}
}
