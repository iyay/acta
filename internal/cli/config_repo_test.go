package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
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

// With nothing set anywhere, plan_depth still shows, as full, marked
// (default) so setup knows it was never asked.
func TestConfigShowPlanDepthDefault(t *testing.T) {
	repoWithGlobal(t, "", "")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "plan_depth: full (default)\n") {
		t.Errorf("show lacks plan_depth: full (default):\n%s", out)
	}
	if out := mustRun(t, "config", "show", "--json"); strings.Contains(out, "(default)") {
		t.Errorf("json carries the mark:\n%s", out)
	}
}

// A plan_depth set in either file, even to full, gets no (default) mark.
func TestConfigShowPlanDepthSetHasNoDefaultMark(t *testing.T) {
	cases := []struct{ global, repo, want string }{
		{"plan_depth: full\n", "", "plan_depth: full\n"},
		{"plan_depth: minimal\n", "", "plan_depth: minimal\n"},
		{"", "plan_depth: full\n", "plan_depth: full (repo)\n"},
		{"", "plan_depth: minimal\n", "plan_depth: minimal (repo)\n"},
	}
	for _, c := range cases {
		repoWithGlobal(t, c.global, c.repo)
		out := mustRun(t, "config", "show")
		// Only the plan_depth line is judged: an unset coding_guide shows (default) too.
		var line string
		for _, l := range strings.Split(out, "\n") {
			if strings.HasPrefix(l, "plan_depth:") {
				line = l
			}
		}
		if !strings.Contains(out, c.want) || strings.Contains(line, "(default)") {
			t.Errorf("global %q repo %q: want %q, no (default):\n%s", c.global, c.repo, c.want, out)
		}
	}
}

// With nothing set anywhere, coding_guide still shows, as lean, marked
// (default). The JSON carries the plain value.
func TestConfigShowCodingGuideDefault(t *testing.T) {
	repoWithGlobal(t, "", "")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "coding_guide: lean (default)\n") {
		t.Errorf("show lacks coding_guide: lean (default):\n%s", out)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if got["coding_guide"] != "lean" {
		t.Errorf("json coding_guide %v, want lean", got["coding_guide"])
	}
	if from, _ := got["from_repo"].([]any); len(from) != 0 {
		t.Errorf("json from_repo %v, want nothing", got["from_repo"])
	}
}

// A coding_guide set in either file, even to lean, gets no (default) mark. A
// repo value beats the user value both ways and shows (repo), in the text and
// in from_repo.
func TestConfigShowCodingGuideSet(t *testing.T) {
	cases := []struct {
		global, repo, want, json string
		fromRepo                 bool
	}{
		{"coding_guide: lean\n", "", "coding_guide: lean\n", "lean", false},
		{"coding_guide: off\n", "", "coding_guide: off\n", "off", false},
		{"", "coding_guide: lean\n", "coding_guide: lean (repo)\n", "lean", true},
		{"", "coding_guide: off\n", "coding_guide: off (repo)\n", "off", true},
		{"coding_guide: lean\n", "coding_guide: off\n", "coding_guide: off (repo)\n", "off", true},
		{"coding_guide: off\n", "coding_guide: lean\n", "coding_guide: lean (repo)\n", "lean", true},
	}
	for _, c := range cases {
		repoWithGlobal(t, c.global, c.repo)
		if out := mustRun(t, "config", "show"); !strings.Contains(out, c.want) {
			t.Errorf("global %q repo %q: show lacks %q:\n%s", c.global, c.repo, c.want, out)
		}
		var got map[string]any
		if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
			t.Fatal(err)
		}
		if got["coding_guide"] != c.json {
			t.Errorf("global %q repo %q: json coding_guide %v, want %s", c.global, c.repo, got["coding_guide"], c.json)
		}
		from, _ := got["from_repo"].([]any)
		if (len(from) == 1 && from[0] == "coding_guide") != c.fromRepo {
			t.Errorf("global %q repo %q: json from_repo %v, want coding_guide=%v", c.global, c.repo, got["from_repo"], c.fromRepo)
		}
	}
}

// Without --repo the questions value goes to the global file, one or probe,
// and never to .acta.yaml. A bad value is refused before any write.
func TestConfigSetQuestionsGlobal(t *testing.T) {
	dir, globalPath := repoWithGlobal(t, "", "")
	mustRun(t, "config", "set", "--questions", "probe")
	if v, _, err := config.LoadUser(globalPath); err != nil || v.Questions != "probe" {
		t.Errorf("global file holds %+v, %v; want questions probe", v, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); err == nil {
		t.Error("set without --repo wrote .acta.yaml")
	}
	mustRun(t, "config", "set", "--questions", "one")
	if v, _, err := config.LoadUser(globalPath); err != nil || v.Questions != "one" {
		t.Errorf("global file holds %+v, %v; want questions one", v, err)
	}
	before, _ := os.ReadFile(globalPath)
	for _, bad := range []string{"maybe", "Probe", "one at a time"} {
		if code, _, errs := runCodeOut("config", "set", "--questions", bad); code != exitBadInput || !strings.Contains(errs, "questions") {
			t.Errorf("%q: exit %d stderr %q, want bad input naming questions", bad, code, errs)
		}
	}
	if after, _ := os.ReadFile(globalPath); string(before) != string(after) {
		t.Errorf("a bad value changed the global file:\n%s", after)
	}
}

// With nothing set anywhere, questions still shows, as one, marked (default)
// so setup knows it was never asked. A set value gets no mark, and no file
// can put it there: questions is a user key only.
func TestConfigShowQuestions(t *testing.T) {
	repoWithGlobal(t, "", "")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "questions: one (default)\n") {
		t.Errorf("show lacks questions: one (default):\n%s", out)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if got["questions"] != "one" {
		t.Errorf("json questions %v, want one", got["questions"])
	}
	for _, c := range []struct{ global, repo, want string }{
		{"questions: probe\n", "", "questions: probe\n"},
		{"questions: one\n", "", "questions: one\n"},
	} {
		repoWithGlobal(t, c.global, c.repo)
		if out := mustRun(t, "config", "show"); !strings.Contains(out, c.want) {
			t.Errorf("show lacks %q:\n%s", c.want, out)
		}
	}
}

// --repo takes no questions: how a person likes to be asked is their taste,
// not a rule the repo puts on everyone who clones it.
func TestConfigSetRepoQuestionsRefused(t *testing.T) {
	dir, globalPath := repoWithGlobal(t, "chat_language: English\nstyle: adhd\nrepo_language: English\n", "")
	before, _ := os.ReadFile(globalPath)
	if code, _, errs := runCodeOut("config", "set", "--repo", "--questions", "probe"); code != exitBadInput || !strings.Contains(errs, "without --repo") {
		t.Errorf("exit %d stderr %q, want bad input sending questions back to the user file", code, errs)
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); err == nil {
		t.Error("a refused --repo --questions wrote .acta.yaml")
	}
	if after, _ := os.ReadFile(globalPath); string(before) != string(after) {
		t.Errorf("the global file changed:\n%s", after)
	}
	// The usage line names the flag on the user side only.
	_, _, errs := runCodeOut("config", "set")
	if !strings.Contains(errs, "[--questions one|probe]") {
		t.Errorf("usage lacks [--questions one|probe]:\n%s", errs)
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
		{"config", "set", "--repo", "--coding-guide", "deep"},
		{"config", "set", "--repo", "--coding-guide", "Lean"},
		{"config", "set", "--repo", "--coding-guide", "off", "--language", "Korean"},
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

// set --repo --coding-guide writes .acta.yaml and never the global file, and
// show then marks the value (repo).
func TestConfigSetRepoCodingGuide(t *testing.T) {
	dir, globalPath := repoWithGlobal(t, "chat_language: Indonesian\nstyle: adhd\nrepo_language: English\ncoding_guide: lean\n", "")
	before, _ := os.ReadFile(globalPath)
	mustRun(t, "config", "set", "--repo", "--coding-guide", "off", "--plan-depth", "minimal")
	after, _ := os.ReadFile(globalPath)
	if string(before) != string(after) {
		t.Errorf("global file changed:\n%s", after)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, ".acta.yaml"))
	for _, want := range []string{"coding_guide: off", "plan_depth: minimal"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf(".acta.yaml lacks %q:\n%s", want, raw)
		}
	}
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "coding_guide: off (repo)\n") {
		t.Errorf("show:\n%s", out)
	}
}

// Without --repo the guide goes to the global file, alone or next to other
// flags, and never to .acta.yaml.
func TestConfigSetCodingGuideGlobal(t *testing.T) {
	dir, globalPath := repoWithGlobal(t, "", "")
	mustRun(t, "config", "set", "--coding-guide", "off")
	// Read the file back through LoadUser: the yaml writer quotes off, so a
	// text match on the raw file would test the writer, not the setting.
	if v, _, err := config.LoadUser(globalPath); err != nil || v.CodingGuide != "off" {
		t.Errorf("global file holds %+v, %v; want coding_guide off", v, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); err == nil {
		t.Error("set without --repo wrote .acta.yaml")
	}
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "coding_guide: off\n") {
		t.Errorf("show:\n%s", out)
	}
	mustRun(t, "config", "set", "--coding-guide", "lean", "--plan-depth", "minimal")
	if v, _, err := config.LoadUser(globalPath); err != nil || v.CodingGuide != "lean" || v.PlanDepth != "minimal" {
		t.Errorf("global file holds %+v, %v; want coding_guide lean and plan_depth minimal", v, err)
	}
}

// A bad guide is refused with its name, and .acta.yaml keeps the bytes it had.
func TestConfigSetRepoCodingGuideRefusesBadValue(t *testing.T) {
	for _, bad := range []string{"deep", "Lean", "on", "minimal", "lean off"} {
		dir, _ := repoWithGlobal(t, "", "plan_depth: minimal\n")
		path := filepath.Join(dir, ".acta.yaml")
		before, _ := os.ReadFile(path)
		if code, _, errs := runCodeOut("config", "set", "--repo", "--coding-guide", bad); code != exitBadInput || !strings.Contains(errs, "coding_guide must be lean or off") {
			t.Errorf("%q: exit %d stderr %q, want bad input saying coding_guide must be lean or off", bad, code, errs)
		}
		if after, _ := os.ReadFile(path); string(before) != string(after) {
			t.Errorf("%q changed .acta.yaml:\n%s", bad, after)
		}
	}
}

// A bad guide is refused before any write, so a file keeps the value it had
// and a missing file stays missing.
func TestConfigSetCodingGuideRefusesBadValue(t *testing.T) {
	for _, bad := range []string{"deep", "Lean", "on", "minimal"} {
		_, globalPath := repoWithGlobal(t, "", "")
		if code, _, errs := runCodeOut("config", "set", "--coding-guide", bad); code != exitBadInput || !strings.Contains(errs, "coding_guide") {
			t.Errorf("%q: exit %d stderr %q, want bad input naming coding_guide", bad, code, errs)
		}
		if _, err := os.Stat(globalPath); err == nil {
			t.Errorf("%q wrote the global file", bad)
		}
	}
	_, globalPath := repoWithGlobal(t, "coding_guide: off\n", "")
	before, _ := os.ReadFile(globalPath)
	if code, _, _ := runCodeOut("config", "set", "--coding-guide", "deep"); code != exitBadInput {
		t.Errorf("exit %d, want bad input", code)
	}
	if after, _ := os.ReadFile(globalPath); string(before) != string(after) {
		t.Errorf("global file changed:\n%s", after)
	}
}

// A bad guide already sitting in either file stops show, and the message names
// the key and the file. Nothing falls back to lean by itself.
func TestConfigShowRefusesBadCodingGuide(t *testing.T) {
	for _, c := range []struct{ global, repo, file string }{
		{"coding_guide: deep\n", "", "config.yaml"},
		{"coding_guide: Lean\n", "", "config.yaml"},
		{"", "coding_guide: deep\n", ".acta.yaml"},
		{"", "coding_guide: OFF\n", ".acta.yaml"},
	} {
		repoWithGlobal(t, c.global, c.repo)
		for _, args := range [][]string{{"config", "show"}, {"config", "show", "--json"}} {
			code, out, errs := runCodeOut(args...)
			if code != exitBadInput || out != "" || !strings.Contains(errs, "coding_guide") || !strings.Contains(errs, c.file) {
				t.Errorf("global %q repo %q %v: exit %d stdout %q stderr %q", c.global, c.repo, args, code, out, errs)
			}
		}
	}
}

// The usage line and the --repo refusal both name the new flag.
func TestConfigUsageNamesCodingGuide(t *testing.T) {
	repoWithGlobal(t, "", "")
	_, _, errs := runCodeOut("config", "set")
	for _, want := range []string{"[--coding-guide lean|off]", "[--coding-guide G]"} {
		if !strings.Contains(errs, want) {
			t.Errorf("usage lacks %q:\n%s", want, errs)
		}
	}
	_, _, errs = runCodeOut("config", "set", "--repo", "--language", "Korean")
	if !strings.Contains(errs, "--coding-guide") {
		t.Errorf("--repo refusal lacks --coding-guide:\n%s", errs)
	}
}

// A broken repo file stops show and names the file.
func TestConfigShowBrokenRepoFile(t *testing.T) {
	repoWithGlobal(t, "", "chat_language: Korean\n")
	if code, _, errs := runCodeOut("config", "show"); code != exitBadInput || !strings.Contains(errs, ".acta.yaml") {
		t.Errorf("exit %d stderr %q", code, errs)
	}
}
