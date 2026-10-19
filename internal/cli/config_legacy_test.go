package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An old config file can still hold repo_language and subagent_models.
// Show must not print them, and their flags must be gone.
func TestConfigDroppedSettings(t *testing.T) {
	t.Chdir(t.TempDir())
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("PM_VOICE_FILE", path)
	if err := os.WriteFile(path, []byte("chat_language: Korean\nstyle: adhd\nrepo_language: German\nsubagent_models: split\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	surfaces := map[string]string{
		"show": mustRun(t, "config", "show"),
		"json": mustRun(t, "config", "show", "--json"),
	}
	var errs strings.Builder
	Run([]string{"config", "set"}, strings.NewReader(""), false, &strings.Builder{}, &errs)
	surfaces["usage"] = errs.String()
	surfaces["help"] = mustRun(t, "help", "config")
	for name, out := range surfaces {
		for _, bad := range []string{"repo_language", "subagent_models", "repo-language", "subagent-models"} {
			if strings.Contains(out, bad) {
				t.Errorf("%s names %q:\n%s", name, bad, out)
			}
		}
	}
	for _, args := range [][]string{
		{"config", "set", "--repo-language", "German"},
		{"config", "set", "--subagent-models", "split"},
		{"config", "set", "--clear-subagent-models"},
		{"config", "set", "--repo", "--repo-language", "German"},
	} {
		if code := runCode(t, args...); code != exitBadInput {
			t.Errorf("%v: exit %d, want %d", args, code, exitBadInput)
		}
	}
}
