package config

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeRepoYAML(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// A value in .acta.yaml wins over the global one, key by key, and the set
// says which keys came from the repo so show can label them.
func TestMergeRepoBeatsGlobal(t *testing.T) {
	global := User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "English", BuildExecutor: "dispatch", PlanDepth: "full"}
	cases := []struct {
		yaml string
		want User
		from []string
	}{
		{"repo_language: Korean\n", User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "Korean", BuildExecutor: "dispatch", PlanDepth: "full"}, []string{"repo_language"}},
		{"build_executor: inline\n", User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "English", BuildExecutor: "inline", PlanDepth: "full"}, []string{"build_executor"}},
		{"plan_depth: minimal\n", User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "English", BuildExecutor: "dispatch", PlanDepth: "minimal"}, []string{"plan_depth"}},
		{"root: .acta\nplan_depth: minimal\nbuild_executor: inline\n", User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "English", BuildExecutor: "inline", PlanDepth: "minimal"}, []string{"build_executor", "plan_depth"}},
		{"plan_depth: \"\"\n", global, nil},
		{"plan_depth: 3\n", global, nil},
		{"plan_depth:\n  - minimal\n", global, nil},
		{"root: .acta\n", global, nil},
	}
	for _, c := range cases {
		got, from, err := MergeRepo(global, writeRepoYAML(t, c.yaml))
		if err != nil {
			t.Fatalf("%q: %v", c.yaml, err)
		}
		if got != c.want {
			t.Errorf("%q: got %+v, want %+v", c.yaml, got, c.want)
		}
		if len(from) != len(c.from) {
			t.Errorf("%q: from %v, want %v", c.yaml, from, c.from)
		}
		for _, k := range c.from {
			if !from[k] {
				t.Errorf("%q: from lacks %s: %v", c.yaml, k, from)
			}
		}
	}
}

// No .acta.yaml means the global config is the whole answer.
func TestMergeRepoMissingFile(t *testing.T) {
	global := User{ChatLanguage: "English", Style: "adhd", RepoLanguage: "English"}
	got, from, err := MergeRepo(global, t.TempDir())
	if err != nil || got != global || len(from) != 0 {
		t.Errorf("got %+v %v %v, want the global config and nothing from the repo", got, from, err)
	}
}

// Personal settings follow the person, not the repo, so a repo file that
// tries to set one is refused, and the message names the key.
func TestMergeRepoRefusesPersonalKeys(t *testing.T) {
	for _, k := range []string{"chat_language", "style", "tone", "theme", "subagent_models"} {
		_, _, err := MergeRepo(UserDefault(), writeRepoYAML(t, k+": x\n"))
		if err == nil || !strings.Contains(err.Error(), k) {
			t.Errorf("%s in .acta.yaml: err %v, want one naming the key", k, err)
		}
	}
}

// A bad value fails the same way in either file.
func TestPlanDepthBadValue(t *testing.T) {
	for _, body := range []string{"plan_depth: deep\n", "build_executor: robot\n"} {
		if _, _, err := MergeRepo(UserDefault(), writeRepoYAML(t, body)); !errors.Is(err, ErrBadUser) {
			t.Errorf(".acta.yaml %q: err %v, want ErrBadUser", body, err)
		}
		path := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, _, err := LoadUser(path); !errors.Is(err, ErrBadUser) {
			t.Errorf("global %q: err %v, want ErrBadUser", body, err)
		}
	}
	for _, ok := range []string{"", "minimal", "full"} {
		v := UserDefault()
		v.PlanDepth = ok
		if err := v.Validate(); err != nil {
			t.Errorf("plan_depth %q: %v", ok, err)
		}
	}
}

// A coding_guide in .acta.yaml beats the user value both ways, and the set says
// the repo gave it. A repo that says nothing, or says it the wrong way, leaves
// the user value alone.
func TestMergeRepoCodingGuide(t *testing.T) {
	user := func(guide string) User {
		return User{ChatLanguage: "English", Style: "adhd", RepoLanguage: "English", CodingGuide: guide}
	}
	cases := []struct {
		name, global, yaml, want string
		fromRepo                 bool
	}{
		{"repo off beats user lean", "lean", "coding_guide: off\n", "off", true},
		{"repo lean beats user off", "off", "coding_guide: lean\n", "lean", true},
		{"repo off beats unset user", "", "coding_guide: off\n", "off", true},
		{"spaces are trimmed", "", "coding_guide: \" off \"\n", "off", true},
		{"repo silent, user off stays", "off", "root: .acta\n", "off", false},
		{"repo silent, unset stays empty", "", "root: .acta\n", "", false},
		{"empty text is ignored", "off", "coding_guide: \"\"\n", "off", false},
		{"number is ignored", "off", "coding_guide: 3\n", "off", false},
		{"yes or no is ignored", "off", "coding_guide: false\n", "off", false},
		{"list is ignored", "off", "coding_guide:\n  - lean\n", "off", false},
		{"plain off is the text off", "lean", "coding_guide: off\n", "off", true},
		{"quoted off is the text off", "lean", "coding_guide: \"off\"\n", "off", true},
	}
	for _, c := range cases {
		got, from, err := MergeRepo(user(c.global), writeRepoYAML(t, c.yaml))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got != user(c.want) {
			t.Errorf("%s: got %+v, want %+v", c.name, got, user(c.want))
		}
		if from["coding_guide"] != c.fromRepo || len(from) > 1 {
			t.Errorf("%s: from %v, want coding_guide=%v and nothing else", c.name, from, c.fromRepo)
		}
	}
}

// A bad coding_guide in .acta.yaml is refused by name, so it is never used.
func TestMergeRepoCodingGuideBadValue(t *testing.T) {
	for _, bad := range []string{"deep", "Lean", "on", "minimal", "lean off"} {
		_, _, err := MergeRepo(UserDefault(), writeRepoYAML(t, "coding_guide: "+bad+"\n"))
		if !errors.Is(err, ErrBadUser) || !strings.Contains(err.Error(), "coding_guide") {
			t.Errorf("coding_guide %q in .acta.yaml: err %v, want ErrBadUser naming the key", bad, err)
		}
	}
}

// coding_guide can be saved per repo. Comments and other keys stay, a second
// save replaces the value, and a new file is made when there is none.
func TestSaveRepoUserCodingGuide(t *testing.T) {
	dir := writeRepoYAML(t, "# keep me\nroot: docs/acta\nplan_depth: minimal\n")
	path, err := SaveRepoUser(dir, map[string]string{"coding_guide": "off"})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	for _, want := range []string{"# keep me", "root: docs/acta", "plan_depth: minimal", "coding_guide: off"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("file lacks %q:\n%s", want, raw)
		}
	}
	if _, err := SaveRepoUser(dir, map[string]string{"coding_guide": "lean"}); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(path)
	if strings.Count(string(raw), "coding_guide") != 1 || !strings.Contains(string(raw), "coding_guide: lean") {
		t.Errorf("second save did not replace the value:\n%s", raw)
	}
	fresh := t.TempDir()
	if _, err := SaveRepoUser(fresh, map[string]string{"coding_guide": "off"}); err != nil {
		t.Fatal(err)
	}
	if v, from, err := MergeRepo(UserDefault(), fresh); err != nil || v.CodingGuide != "off" || !from["coding_guide"] {
		t.Errorf("fresh file: %+v %v %v", v, from, err)
	}
}

// Saving keeps what the user already wrote in .acta.yaml, comments too.
func TestSaveRepoUserKeepsOtherKeys(t *testing.T) {
	dir := writeRepoYAML(t, "# where the planning files live\nroot: docs/acta\nplan_depth: full\n")
	path, err := SaveRepoUser(dir, map[string]string{"plan_depth": "minimal", "build_executor": "inline"})
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, ".acta.yaml") {
		t.Errorf("wrote %s", path)
	}
	raw, _ := os.ReadFile(path)
	got := string(raw)
	for _, want := range []string{"# where the planning files live", "root: docs/acta", "plan_depth: minimal", "build_executor: inline"} {
		if !strings.Contains(got, want) {
			t.Errorf("file lacks %q:\n%s", want, got)
		}
	}
	if strings.Count(got, "plan_depth") != 1 {
		t.Errorf("plan_depth written twice:\n%s", got)
	}
	// A new file is made when there is none.
	fresh := t.TempDir()
	if _, err := SaveRepoUser(fresh, map[string]string{"plan_depth": "minimal"}); err != nil {
		t.Fatal(err)
	}
	if v, from, err := MergeRepo(UserDefault(), fresh); err != nil || v.PlanDepth != "minimal" || !from["plan_depth"] {
		t.Errorf("fresh file: %+v %v %v", v, from, err)
	}
	// A personal key is never written.
	if _, err := SaveRepoUser(fresh, map[string]string{"chat_language": "Korean"}); err == nil {
		t.Error("SaveRepoUser wrote chat_language")
	}
}

// A commit_history in .acta.yaml beats the user value both ways, and the set
// says the repo gave it. A repo that says nothing, or says it the wrong way,
// leaves the user value alone.
func TestMergeRepoCommitHistory(t *testing.T) {
	user := func(history string) User {
		return User{ChatLanguage: "English", Style: "adhd", RepoLanguage: "English", CommitHistory: history}
	}
	cases := []struct {
		name, global, yaml, want string
		fromRepo                 bool
	}{
		{"repo full beats user tidy", "tidy", "commit_history: full\n", "full", true},
		{"repo tidy beats user full", "full", "commit_history: tidy\n", "tidy", true},
		{"repo full beats unset user", "", "commit_history: full\n", "full", true},
		{"spaces are trimmed", "", "commit_history: \" full \"\n", "full", true},
		{"repo silent, user full stays", "full", "root: .acta\n", "full", false},
		{"repo silent, unset stays empty", "", "root: .acta\n", "", false},
		{"empty text is ignored", "full", "commit_history: \"\"\n", "full", false},
		{"number is ignored", "full", "commit_history: 3\n", "full", false},
		{"list is ignored", "full", "commit_history:\n  - tidy\n", "full", false},
	}
	for _, c := range cases {
		got, from, err := MergeRepo(user(c.global), writeRepoYAML(t, c.yaml))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got != user(c.want) {
			t.Errorf("%s: got %+v, want %+v", c.name, got, user(c.want))
		}
		if from["commit_history"] != c.fromRepo || len(from) > 1 {
			t.Errorf("%s: from %v, want commit_history=%v and nothing else", c.name, from, c.fromRepo)
		}
	}
}

// A bad commit_history in .acta.yaml is refused by name, so it is never used.
func TestMergeRepoCommitHistoryBadValue(t *testing.T) {
	for _, bad := range []string{"deep", "Tidy", "minimal", "tidy full"} {
		_, _, err := MergeRepo(UserDefault(), writeRepoYAML(t, "commit_history: "+bad+"\n"))
		if !errors.Is(err, ErrBadUser) || !strings.Contains(err.Error(), "commit_history") {
			t.Errorf("commit_history %q in .acta.yaml: err %v, want ErrBadUser naming the key", bad, err)
		}
	}
}

// A repo key is an override only when its value differs from what the user
// would get anyway. A key the user never set counts as its default.
func TestOverrides(t *testing.T) {
	user := User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "English", BuildExecutor: "dispatch", PlanDepth: "minimal"}
	cases := []struct {
		name string
		yaml string
		user User
		want []Override
	}{
		{"same value", "build_executor: dispatch\nplan_depth: minimal\n", user, nil},
		{"different value", "build_executor: subagent\n", user, []Override{{"build_executor", "subagent", "dispatch"}}},
		{"user unset, default differs", "plan_depth: minimal\n", UserDefault(), []Override{{"plan_depth", "minimal", "full"}}},
		{"user unset, default equals", "plan_depth: full\ncommit_history: tidy\ncoding_guide: lean\n", UserDefault(), nil},
		{"user unset commit_history, repo full", "commit_history: full\n", UserDefault(), []Override{{"commit_history", "full", "tidy"}}},
		{"user unset coding_guide, repo off", "coding_guide: off\n", UserDefault(), []Override{{"coding_guide", "off", "lean"}}},
		{"user has no executor", "build_executor: inline\n", UserDefault(), []Override{{"build_executor", "inline", ""}}},
		{"repo language differs", "repo_language: Korean\n", UserDefault(), []Override{{"repo_language", "Korean", "English"}}},
		{"only differing keys, in RepoKeys order",
			"coding_guide: off\nplan_depth: full\nrepo_language: Korean\nbuild_executor: dispatch\n", user,
			[]Override{{"repo_language", "Korean", "English"}, {"plan_depth", "full", "minimal"}, {"coding_guide", "off", "lean"}}},
		{"other keys in the file are ignored", "root: x\nplan_depth: minimal\n", user, nil},
		{"empty value is not set", "plan_depth: \"\"\n", user, nil},
	}
	for _, c := range cases {
		got, err := Overrides(c.user, writeRepoYAML(t, c.yaml))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

// No .acta.yaml at all means nothing overrides.
func TestOverridesNoFile(t *testing.T) {
	got, err := Overrides(UserDefault(), t.TempDir())
	if err != nil || len(got) != 0 {
		t.Errorf("no file: got %+v, %v, want an empty list and no error", got, err)
	}
}

// Overrides fails the way MergeRepo does: bad yaml, a personal key, a bad
// value, and a file that cannot be read.
func TestOverridesErrors(t *testing.T) {
	for _, body := range []string{"plan_depth: [\n", "tone: x\n", "style: plain\n", "plan_depth: deep\n"} {
		got, err := Overrides(UserDefault(), writeRepoYAML(t, body))
		if err == nil || len(got) != 0 {
			t.Errorf("%q: got %+v, %v, want an error and no list", body, got, err)
		}
	}
	// A folder named .acta.yaml cannot be read as a file.
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".acta.yaml"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Overrides(UserDefault(), dir); err == nil {
		t.Error("unreadable .acta.yaml: want an error")
	}
	if _, err := Overrides(UserDefault(), writeRepoYAML(t, "chat_language: x\n")); !errors.Is(err, ErrBadUser) {
		t.Errorf("personal key: err %v, want ErrBadUser", err)
	}
}

func TestUnsetRepoUser(t *testing.T) {
	cases := []struct {
		name string
		body string
		keys []string
		want string
	}{
		{"removes only the named key and keeps the rest in order",
			"root: x\nplan_depth: minimal\n# keep me\nbuild_executor: dispatch\ndirs: y\n", []string{"plan_depth"},
			"root: x\n# keep me\nbuild_executor: dispatch\ndirs: y\n"},
		{"removes two keys", "plan_depth: minimal\ncoding_guide: off\nroot: x\n", []string{"coding_guide", "plan_depth"}, "root: x\n"},
		{"leaves an empty file when nothing is left", "plan_depth: minimal\n", []string{"plan_depth"}, ""},
		{"key not in the file is fine", "root: x\n", []string{"plan_depth"}, "root: x\n"},
		{"no keys given", "plan_depth: minimal\n", nil, "plan_depth: minimal\n"},
		{"repeated key", "plan_depth: minimal\nroot: x\n", []string{"plan_depth", "plan_depth"}, "root: x\n"},
	}
	for _, c := range cases {
		dir := writeRepoYAML(t, c.body)
		path, err := UnsetRepoUser(dir, c.keys)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if path != filepath.Join(dir, ".acta.yaml") {
			t.Errorf("%s: path %q", c.name, path)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if string(raw) != c.want {
			t.Errorf("%s: file %q, want %q", c.name, raw, c.want)
		}
	}
}

// yaml ties a comment to the key under it. Removing the key must not eat a
// comment the person wrote for the file, so it moves to the next key left.
func TestUnsetRepoUserKeepsComments(t *testing.T) {
	cases := []struct {
		name string
		body string
		keys []string
		want string
	}{
		{"header right above the removed first key",
			"# acta settings for this repo\nbuild_executor: dispatch # omp for everyone\nroot: .acta\n", []string{"build_executor"},
			"# acta settings for this repo\nroot: .acta\n"},
		{"header above a removed middle key",
			"root: x\n# why the guide is off\ncoding_guide: off\ndirs: y\n", []string{"coding_guide"},
			"root: x\n# why the guide is off\ndirs: y\n"},
		{"header above the only key stays in the file",
			"# acta settings for this repo\nbuild_executor: dispatch\n", []string{"build_executor"},
			"# acta settings for this repo\n"},
		{"two removed keys in a row keep both headers",
			"# one\nplan_depth: minimal\n# two\ncoding_guide: off\nroot: x\n", []string{"plan_depth", "coding_guide"},
			"# one\n# two\nroot: x\n"},
		{"header above the removed last key stays",
			"root: x\n# trailing note\nplan_depth: minimal\n", []string{"plan_depth"},
			"root: x\n# trailing note\n"},
		{"note after the removed last key stays",
			"root: x\nplan_depth: minimal\n\n# note at the end\n", []string{"plan_depth"},
			"root: x\n\n# note at the end\n"},
	}
	for _, c := range cases {
		dir := writeRepoYAML(t, c.body)
		path, err := UnsetRepoUser(dir, c.keys)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if string(raw) != c.want {
			t.Errorf("%s: file %q, want %q", c.name, raw, c.want)
		}
	}
}

// With no file there is nothing to remove, and no file is made.
func TestUnsetRepoUserNoFile(t *testing.T) {
	dir := t.TempDir()
	path, err := UnsetRepoUser(dir, []string{"plan_depth"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a file appeared: %v", err)
	}
}

// Any key outside the five repo keys is refused, and the file stays as it was.
func TestUnsetRepoUserRefusesOtherKeys(t *testing.T) {
	for _, k := range []string{"chat_language", "tone", "root", "nope"} {
		body := "plan_depth: minimal\n"
		dir := writeRepoYAML(t, body)
		_, err := UnsetRepoUser(dir, []string{"plan_depth", k})
		if !errors.Is(err, ErrBadUser) {
			t.Fatalf("%s: err %v, want ErrBadUser", k, err)
		}
		for _, rk := range RepoKeys {
			if !strings.Contains(err.Error(), rk) {
				t.Errorf("%s: error %q does not name %s", k, err, rk)
			}
		}
		raw, _ := os.ReadFile(filepath.Join(dir, ".acta.yaml"))
		if string(raw) != body {
			t.Errorf("%s: file changed to %q", k, raw)
		}
	}
}
