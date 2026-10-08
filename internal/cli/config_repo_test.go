package cli

import (
	"encoding/json"
	"os"
	"os/exec"
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
	for _, want := range []string{"plan_depth: minimal (repo; yours: full)\n", "build_executor: dispatch\n", "repo_language: English\n"} {
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
		{"", "plan_depth: minimal\n", "plan_depth: minimal (repo; yours: full)\n"},
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
		{"", "coding_guide: off\n", "coding_guide: off (repo; yours: lean)\n", "off", true},
		{"coding_guide: lean\n", "coding_guide: off\n", "coding_guide: off (repo; yours: lean)\n", "off", true},
		{"coding_guide: off\n", "coding_guide: lean\n", "coding_guide: lean (repo; yours: off)\n", "lean", true},
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
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "plan_depth: minimal (repo; yours: full)") {
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

// A key that is not a repo key is refused before anything is written, even
// when the same call also sets a good key. Both files stay byte for byte.
func TestConfigSetRepoUnsetPersonalKeyRefused(t *testing.T) {
	repoBody := "# acta settings for this repo\nbuild_executor: dispatch\nroot: .acta\n"
	globalBody := "chat_language: Indonesian\nstyle: adhd\nrepo_language: English\n"
	dir, globalPath := repoWithGlobal(t, globalBody, repoBody)
	code, _, errs := runCodeOut("config", "set", "--repo", "--executor", "inline", "--unset", "style")
	if code != exitBadInput || !strings.Contains(errs, "style cannot be unset per repo") {
		t.Errorf("exit %d stderr %q, want bad input naming style", code, errs)
	}
	if raw, _ := os.ReadFile(filepath.Join(dir, ".acta.yaml")); string(raw) != repoBody {
		t.Errorf(".acta.yaml changed:\n%s", raw)
	}
	if raw, _ := os.ReadFile(globalPath); string(raw) != globalBody {
		t.Errorf("the global file changed:\n%s", raw)
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
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "coding_guide: off (repo; yours: lean)\n") {
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

// With nothing set anywhere, commit_history still shows, as tidy, marked
// (default). The JSON carries the plain value.
func TestConfigShowCommitHistoryDefault(t *testing.T) {
	repoWithGlobal(t, "", "")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "commit_history: tidy (default)\n") {
		t.Errorf("show lacks commit_history: tidy (default):\n%s", out)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if got["commit_history"] != "tidy" {
		t.Errorf("json commit_history %v, want tidy", got["commit_history"])
	}
}

// A commit_history set in either file, even to tidy, gets no (default) mark.
// A repo value beats the user value both ways and shows (repo), in the text
// and in from_repo.
func TestConfigShowCommitHistorySet(t *testing.T) {
	cases := []struct {
		global, repo, want, json string
		fromRepo                 bool
	}{
		{"commit_history: tidy\n", "", "commit_history: tidy\n", "tidy", false},
		{"commit_history: full\n", "", "commit_history: full\n", "full", false},
		{"", "commit_history: tidy\n", "commit_history: tidy (repo)\n", "tidy", true},
		{"", "commit_history: full\n", "commit_history: full (repo; yours: tidy)\n", "full", true},
		{"commit_history: tidy\n", "commit_history: full\n", "commit_history: full (repo; yours: tidy)\n", "full", true},
		{"commit_history: full\n", "commit_history: tidy\n", "commit_history: tidy (repo; yours: full)\n", "tidy", true},
	}
	for _, c := range cases {
		repoWithGlobal(t, c.global, c.repo)
		out := mustRun(t, "config", "show")
		if !strings.Contains(out, c.want) {
			t.Errorf("global %q repo %q: show lacks %q:\n%s", c.global, c.repo, c.want, out)
		}
		var got map[string]any
		if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
			t.Fatal(err)
		}
		if got["commit_history"] != c.json {
			t.Errorf("global %q repo %q: json commit_history %v, want %s", c.global, c.repo, got["commit_history"], c.json)
		}
		from, _ := got["from_repo"].([]any)
		if (len(from) == 1 && from[0] == "commit_history") != c.fromRepo {
			t.Errorf("global %q repo %q: json from_repo %v, want commit_history=%v", c.global, c.repo, got["from_repo"], c.fromRepo)
		}
	}
}

// Without --repo the history goes to the global file, alone or next to other
// flags, and never to .acta.yaml.
func TestConfigSetCommitHistoryGlobal(t *testing.T) {
	dir, globalPath := repoWithGlobal(t, "", "")
	mustRun(t, "config", "set", "--commit-history", "full")
	if v, _, err := config.LoadUser(globalPath); err != nil || v.CommitHistory != "full" {
		t.Errorf("global file holds %+v, %v; want commit_history full", v, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); err == nil {
		t.Error("set without --repo wrote .acta.yaml")
	}
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "commit_history: full\n") {
		t.Errorf("show:\n%s", out)
	}
}

// set --repo --commit-history writes .acta.yaml and never the global file, and
// show then marks the value (repo).
func TestConfigSetRepoCommitHistory(t *testing.T) {
	dir, globalPath := repoWithGlobal(t, "chat_language: Indonesian\nstyle: adhd\nrepo_language: English\ncommit_history: tidy\n", "")
	before, _ := os.ReadFile(globalPath)
	mustRun(t, "config", "set", "--repo", "--commit-history", "full")
	after, _ := os.ReadFile(globalPath)
	if string(before) != string(after) {
		t.Errorf("global file changed:\n%s", after)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, ".acta.yaml"))
	if !strings.Contains(string(raw), "commit_history: full") {
		t.Errorf(".acta.yaml lacks commit_history: full:\n%s", raw)
	}
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "commit_history: full (repo; yours: tidy)\n") {
		t.Errorf("show:\n%s", out)
	}
}

// A bad history is refused with its name, and .acta.yaml keeps the bytes it had.
func TestConfigSetRepoCommitHistoryRefusesBadValue(t *testing.T) {
	for _, bad := range []string{"deep", "Tidy", "minimal", "full off"} {
		dir, _ := repoWithGlobal(t, "", "plan_depth: minimal\n")
		path := filepath.Join(dir, ".acta.yaml")
		before, _ := os.ReadFile(path)
		if code, _, errs := runCodeOut("config", "set", "--repo", "--commit-history", bad); code != exitBadInput || !strings.Contains(errs, "commit_history must be tidy or full") {
			t.Errorf("%q: exit %d stderr %q, want bad input saying commit_history must be tidy or full", bad, code, errs)
		}
		if after, _ := os.ReadFile(path); string(before) != string(after) {
			t.Errorf("%q changed .acta.yaml:\n%s", bad, after)
		}
	}
}

// A bad history is refused before any write, so a file keeps the value it had
// and a missing file stays missing.
func TestConfigSetCommitHistoryRefusesBadValue(t *testing.T) {
	for _, bad := range []string{"deep", "Tidy", "minimal"} {
		_, globalPath := repoWithGlobal(t, "", "")
		if code, _, errs := runCodeOut("config", "set", "--commit-history", bad); code != exitBadInput || !strings.Contains(errs, "commit_history") {
			t.Errorf("%q: exit %d stderr %q, want bad input naming commit_history", bad, code, errs)
		}
		if _, err := os.Stat(globalPath); err == nil {
			t.Errorf("%q wrote the global file", bad)
		}
	}
	_, globalPath := repoWithGlobal(t, "commit_history: full\n", "")
	before, _ := os.ReadFile(globalPath)
	if code, _, _ := runCodeOut("config", "set", "--commit-history", "deep"); code != exitBadInput {
		t.Errorf("exit %d, want bad input", code)
	}
	if after, _ := os.ReadFile(globalPath); string(before) != string(after) {
		t.Errorf("global file changed:\n%s", after)
	}
}

// A bad history already sitting in either file stops show, and the message
// names the key and the file. Nothing falls back to tidy by itself.
func TestConfigShowRefusesBadCommitHistory(t *testing.T) {
	for _, c := range []struct{ global, repo, file string }{
		{"commit_history: deep\n", "", "config.yaml"},
		{"commit_history: Tidy\n", "", "config.yaml"},
		{"", "commit_history: deep\n", ".acta.yaml"},
		{"", "commit_history: TIDY\n", ".acta.yaml"},
	} {
		repoWithGlobal(t, c.global, c.repo)
		for _, args := range [][]string{{"config", "show"}, {"config", "show", "--json"}} {
			code, out, errs := runCodeOut(args...)
			if code != exitBadInput || out != "" || !strings.Contains(errs, "commit_history") || !strings.Contains(errs, c.file) {
				t.Errorf("global %q repo %q %v: exit %d stdout %q stderr %q", c.global, c.repo, args, code, out, errs)
			}
		}
	}
}

// readRepoFile gives the .acta.yaml text, or "" when there is no file.
func readRepoFile(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, ".acta.yaml"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(raw)
}

// --unset drops keys from .acta.yaml, can repeat, and prints the file path.
func TestConfigUnsetRepoRemovesKeys(t *testing.T) {
	dir, _ := repoWithGlobal(t, "", "plan_depth: minimal\nbuild_executor: inline\ncoding_guide: off\n")
	out := mustRun(t, "config", "set", "--repo", "--unset", "plan_depth")
	if strings.TrimSpace(out) != filepath.Join(dir, ".acta.yaml") && !strings.HasSuffix(strings.TrimSpace(out), ".acta.yaml") {
		t.Errorf("printed %q, want the .acta.yaml path", out)
	}
	if got := readRepoFile(t, dir); strings.Contains(got, "plan_depth") || !strings.Contains(got, "build_executor: inline") {
		t.Errorf("after one unset:\n%s", got)
	}
	mustRun(t, "config", "set", "--repo", "--unset", "build_executor", "--unset", "coding_guide")
	if got := readRepoFile(t, dir); strings.Contains(got, "build_executor") || strings.Contains(got, "coding_guide") {
		t.Errorf("after repeated unset:\n%s", got)
	}
}

// One call can set a key and unset another; both changes land.
func TestConfigUnsetRepoMixesWithSet(t *testing.T) {
	dir, _ := repoWithGlobal(t, "", "plan_depth: minimal\n")
	mustRun(t, "config", "set", "--repo", "--executor", "inline", "--unset", "plan_depth")
	got := readRepoFile(t, dir)
	if strings.Contains(got, "plan_depth") || !strings.Contains(got, "build_executor: inline") {
		t.Errorf("mix of set and unset:\n%s", got)
	}
}

// Every refused mix leaves both files as they were.
func TestConfigUnsetRefuses(t *testing.T) {
	const global = "chat_language: Indonesian\nstyle: adhd\nrepo_language: English\n"
	const repo = "plan_depth: minimal\nbuild_executor: inline\n"
	for _, args := range [][]string{
		// Same key set and unset in one call.
		{"config", "set", "--repo", "--plan-depth", "full", "--unset", "plan_depth"},
		{"config", "set", "--repo", "--executor", "inline", "--unset", "build_executor"},
		{"config", "set", "--repo", "--repo-language", "Korean", "--unset", "repo_language"},
		// Without --repo.
		{"config", "set", "--unset", "plan_depth"},
		{"config", "set", "--language", "Korean", "--unset", "plan_depth"},
		// Not a repo key, or no key at all.
		{"config", "set", "--repo", "--unset", "chat_language"},
		{"config", "set", "--repo", "--unset", "questions"},
		{"config", "set", "--repo", "--unset", "nonsense"},
		{"config", "set", "--repo", "--unset", "plan_depth", "--unset", "style"},
		{"config", "set", "--repo", "--unset", ""},
		// A bad value beside a good unset: nothing may land.
		{"config", "set", "--repo", "--plan-depth", "deep", "--unset", "build_executor"},
		// A personal flag beside --unset.
		{"config", "set", "--repo", "--language", "Korean", "--unset", "plan_depth"},
	} {
		dir, globalPath := repoWithGlobal(t, global, repo)
		if code, _, errs := runCodeOut(args...); code != exitBadInput || errs == "" {
			t.Errorf("%v: exit %d stderr %q, want bad input", args, code, errs)
		}
		if got := readRepoFile(t, dir); got != repo {
			t.Errorf("%v changed .acta.yaml:\n%s", args, got)
		}
		if raw, _ := os.ReadFile(globalPath); string(raw) != global {
			t.Errorf("%v changed the global file:\n%s", args, raw)
		}
	}
}

// Unset of a key that is not in the file, or with no file at all, is fine.
func TestConfigUnsetRepoMissingKeyIsFine(t *testing.T) {
	dir, _ := repoWithGlobal(t, "", "")
	mustRun(t, "config", "set", "--repo", "--unset", "plan_depth")
	if got := readRepoFile(t, dir); got != "" {
		t.Errorf("unset made a file:\n%s", got)
	}
}

// --unset only edits the file; it never commits.
func TestConfigUnsetRepoNeverCommits(t *testing.T) {
	dir, _ := repoWithGlobal(t, "", "plan_depth: minimal\n")
	git := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("add", ".acta.yaml")
	git("commit", "-q", "-m", "seed")
	head := git("rev-parse", "HEAD")
	mustRun(t, "config", "set", "--repo", "--unset", "plan_depth")
	if got := git("rev-parse", "HEAD"); got != head {
		t.Errorf("HEAD moved from %s to %s", head, got)
	}
	if got := git("status", "--porcelain"); !strings.Contains(got, ".acta.yaml") {
		t.Errorf("the unset should be left uncommitted, status:\n%s", got)
	}
}

// A repo key that differs from the user's value shows both; an equal one keeps
// the plain (repo) mark. Yours falls back to the default when the user set none.
func TestConfigShowNamesOverrides(t *testing.T) {
	repoWithGlobal(t, "chat_language: Indonesian\nstyle: adhd\nrepo_language: English\nbuild_executor: dispatch\ncoding_guide: lean\n",
		"build_executor: subagent\nplan_depth: minimal\ncoding_guide: lean\n")
	out := mustRun(t, "config", "show")
	for _, want := range []string{
		"build_executor: subagent (repo; yours: dispatch)\n",
		"plan_depth: minimal (repo; yours: full)\n",
		"coding_guide: lean (repo)\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("show lacks %q:\n%s", want, out)
		}
	}
	var got struct {
		FromRepo  []string `json:"from_repo"`
		Overrides []struct {
			Key, Repo, Yours string
		} `json:"overrides"`
	}
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.FromRepo) != 3 {
		t.Errorf("from_repo %v, want all three repo keys", got.FromRepo)
	}
	if len(got.Overrides) != 2 || got.Overrides[0].Key != "build_executor" || got.Overrides[0].Repo != "subagent" || got.Overrides[0].Yours != "dispatch" ||
		got.Overrides[1].Key != "plan_depth" || got.Overrides[1].Repo != "minimal" || got.Overrides[1].Yours != "full" {
		t.Errorf("overrides %+v", got.Overrides)
	}
}

// A key the user never set has no value to name, so show says so.
func TestConfigShowOverrideYoursNotSet(t *testing.T) {
	repoWithGlobal(t, "chat_language: Indonesian\nstyle: adhd\nrepo_language: English\n", "build_executor: inline\n")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "build_executor: inline (repo; yours: not set)\n") {
		t.Errorf("show:\n%s", out)
	}
	var got struct {
		Overrides []struct{ Key, Repo, Yours string } `json:"overrides"`
	}
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Overrides) != 1 || got.Overrides[0].Yours != "" {
		t.Errorf("overrides %+v", got.Overrides)
	}
}

// With no override the JSON still carries the field, as an empty list.
func TestConfigShowOverridesEmptyList(t *testing.T) {
	repoWithGlobal(t, "plan_depth: minimal\n", "plan_depth: minimal\n")
	var got map[string]any
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	list, ok := got["overrides"].([]any)
	if !ok || len(list) != 0 {
		t.Errorf("overrides %#v, want an empty list", got["overrides"])
	}
	if from, _ := got["from_repo"].([]any); len(from) != 1 {
		t.Errorf("from_repo %v", got["from_repo"])
	}
}
