package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoWithGlobal makes a temp folder to run in and a temp global config, so
// no test here ever reads or writes the real ~/.acta/config.yaml.
func repoWithGlobal(t *testing.T, global, repo string) (dir, globalPath string) {
	t.Helper()
	dir = t.TempDir()
	globalPath = filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("PM_VOICE_FILE", globalPath)
	if global != "" {
		if err := os.WriteFile(globalPath, []byte(global), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if repo != "" {
		if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte(repo), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	return dir, globalPath
}

// runCodeOut runs the CLI and hands back the exit code with both streams, so a
// test can look at what a failing command printed. It is named apart from the
// runCode in cli_test.go, which returns only the code.
func runCodeOut(args ...string) (int, string, string) {
	var out, errs strings.Builder
	code := Run(args, strings.NewReader(""), false, &out, &errs)
	return code, out.String(), errs.String()
}

// show labels only the values the repo set, and always names plan_depth.
func TestConfigShowRepoLayer(t *testing.T) {
	repoWithGlobal(t, "chat_language: Indonesian\nstyle: adhd\nrepo_language: English\nbuild_executor: dispatch\n", "plan_depth: minimal\n")
	out := mustRun(t, "config", "show")
	for _, want := range []string{"plan_depth: minimal (repo)\n", "build_executor: dispatch\n", "repo_language: English\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("show lacks %q:\n%s", want, out)
		}
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if got["plan_depth"] != "minimal" {
		t.Errorf("json plan_depth %v", got["plan_depth"])
	}
	if from, _ := got["from_repo"].([]any); len(from) != 1 || from[0] != "plan_depth" {
		t.Errorf("json from_repo %v", got["from_repo"])
	}
}

// With nothing set anywhere, plan_depth still shows, as full.
func TestConfigShowPlanDepthDefault(t *testing.T) {
	repoWithGlobal(t, "", "")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "plan_depth: full\n") {
		t.Errorf("show lacks plan_depth: full:\n%s", out)
	}
}

// set --repo writes .acta.yaml and never the global file.
func TestConfigSetRepoWritesRepoFile(t *testing.T) {
	dir, globalPath := repoWithGlobal(t, "chat_language: Indonesian\nstyle: adhd\nrepo_language: English\n", "")
	before, _ := os.ReadFile(globalPath)
	mustRun(t, "config", "set", "--repo", "--plan-depth", "minimal", "--executor", "inline", "--repo-language", "Korean")
	after, _ := os.ReadFile(globalPath)
	if string(before) != string(after) {
		t.Errorf("global file changed:\n%s", after)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, ".acta.yaml"))
	for _, want := range []string{"plan_depth: minimal", "build_executor: inline", "repo_language: Korean"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf(".acta.yaml lacks %q:\n%s", want, raw)
		}
	}
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "plan_depth: minimal (repo)") {
		t.Errorf("show:\n%s", out)
	}
}

// Personal flags, bad values and an old .pm.yaml are refused before any write.
func TestConfigSetRepoRefuses(t *testing.T) {
	for _, args := range [][]string{
		{"config", "set", "--repo", "--language", "Korean"},
		{"config", "set", "--repo", "--style", "plain"},
		{"config", "set", "--repo", "--tone", "short"},
		{"config", "set", "--repo", "--subagent-models", "split"},
		{"config", "set", "--repo", "--theme", "x"},
		{"config", "set", "--repo", "--plan-depth", "deep"},
		{"config", "set", "--repo"},
	} {
		dir, _ := repoWithGlobal(t, "", "")
		if code, _, errs := runCodeOut(args...); code != exitBadInput || errs == "" {
			t.Errorf("%v: exit %d stderr %q, want bad input", args, code, errs)
		}
		if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); err == nil {
			t.Errorf("%v wrote .acta.yaml", args)
		}
	}
	dir, _ := repoWithGlobal(t, "", "")
	if err := os.WriteFile(filepath.Join(dir, ".pm.yaml"), []byte("root: .pm\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, errs := runCodeOut("config", "set", "--repo", "--plan-depth", "minimal"); code != exitBadInput || !strings.Contains(errs, ".pm.yaml") {
		t.Errorf(".pm.yaml repo: exit %d stderr %q", code, errs)
	}
}

// Without --repo the depth goes to the global file.
func TestConfigSetPlanDepthGlobal(t *testing.T) {
	dir, globalPath := repoWithGlobal(t, "", "")
	mustRun(t, "config", "set", "--plan-depth", "minimal")
	raw, _ := os.ReadFile(globalPath)
	if !strings.Contains(string(raw), "plan_depth: minimal") {
		t.Errorf("global file:\n%s", raw)
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); err == nil {
		t.Error("set without --repo wrote .acta.yaml")
	}
}

// A broken repo file stops show and names the file.
func TestConfigShowBrokenRepoFile(t *testing.T) {
	repoWithGlobal(t, "", "chat_language: Korean\n")
	if code, _, errs := runCodeOut("config", "show"); code != exitBadInput || !strings.Contains(errs, ".acta.yaml") {
		t.Errorf("exit %d stderr %q", code, errs)
	}
}
