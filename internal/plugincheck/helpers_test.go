package plugincheck

import (
	"path/filepath"
	"testing"
)

// pluginRoot is the repo's plugin/ folder.
func pluginRoot(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "..", "plugin"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// CheckSkill fails the test once per problem in one skill folder.
func CheckSkill(t *testing.T, r SkillRule) {
	t.Helper()
	for _, p := range SkillProblems(pluginRoot(t), r) {
		t.Error(p)
	}
}
