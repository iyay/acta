package plugincheck

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"pm-board/internal/hook"
)

func TestSkillFoldersMatchHookIndex(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join(pluginRoot(t), "skills"))
	if err != nil {
		t.Fatal(err)
	}
	var folders, index []string
	for _, e := range entries {
		if e.IsDir() {
			folders = append(folders, e.Name())
		}
	}
	for _, s := range hook.Skills {
		index = append(index, s.Name)
	}
	sort.Strings(folders)
	sort.Strings(index)
	if !reflect.DeepEqual(folders, index) {
		t.Fatalf("skill folders %v, hook index %v", folders, index)
	}
}

func walkPlugin(t *testing.T, fn func(rel, text string)) {
	t.Helper()
	root := pluginRoot(t)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		fn(filepath.ToSlash(rel), string(b))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoSuperpowersPrefixOrUserPaths(t *testing.T) {
	walkPlugin(t, func(rel, text string) {
		if rel != "NOTICE" && rel != "README.md" && strings.Contains(text, "superpowers:") {
			t.Errorf("%s still says superpowers:", rel)
		}
		if strings.Contains(text, "/Users/") && rel != "omp/FACTS.md" {
			t.Errorf("%s has an absolute user path", rel)
		}
	})
}

func TestTotalSkillSize(t *testing.T) {
	total := 0
	walkPlugin(t, func(rel, text string) {
		if strings.HasPrefix(rel, "skills/") && strings.HasSuffix(rel, ".md") {
			total += strings.Count(text, "\n")
		}
	})
	t.Logf("skills total: %d lines of markdown", total)
	if total > 4240 {
		t.Fatalf("skills total %d lines, cap is 4240", total)
	}
}

func TestHookScriptsAreTheOnlyExecutablesOutsideSkills(t *testing.T) {
	walkPlugin(t, func(rel, _ string) {
		st, err := os.Stat(filepath.Join(pluginRoot(t), rel))
		if err != nil {
			t.Fatal(err)
		}
		exec := st.Mode()&0o111 != 0
		allowed := rel == "hooks/session-start" || rel == "hooks/prompt-reminder" || rel == "skills/debug/find-polluter.sh"
		if exec && !allowed {
			t.Errorf("%s is executable but should not be", rel)
		}
		if allowed && !exec {
			t.Errorf("%s must be executable", rel)
		}
	})
}
