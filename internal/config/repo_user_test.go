package config

import (
	"errors"
	"os"
	"path/filepath"
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
