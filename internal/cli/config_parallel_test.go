package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The two parallel plan flags save to the user file, so slice can read them
// back from config show.
func TestConfigSetParallelPlanKeys(t *testing.T) {
	_, globalPath := repoWithGlobal(t, "chat_language: English\nstyle: adhd\n", "")
	mustRun(t, "config", "set", "--parallel-plans", "subagent")
	mustRun(t, "config", "set", "--subagent-offer", "declined")
	raw, _ := os.ReadFile(globalPath)
	for _, want := range []string{"parallel_plans: subagent", "subagent_offer: declined"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("user file lacks %q:\n%s", want, raw)
		}
	}
}

// A bad value stops the call before the save, so the file keeps what it had.
func TestConfigSetParallelPlanKeysBadValue(t *testing.T) {
	before := "chat_language: English\nstyle: adhd\n"
	_, globalPath := repoWithGlobal(t, before, "")
	for _, args := range [][]string{
		{"config", "set", "--parallel-plans", "inline"},
		{"config", "set", "--parallel-plans", "dispatch"},
		{"config", "set", "--subagent-offer", "yes"},
	} {
		if code, _, errs := runCodeOut(args...); code != exitBadInput || errs == "" {
			t.Errorf("%v: exit %d, stderr %q, want bad input with a message", args, code, errs)
		}
		if raw, _ := os.ReadFile(globalPath); string(raw) != before {
			t.Errorf("%v changed the user file:\n%s", args, raw)
		}
	}
}

// The keys are personal. --repo refuses them and never writes .acta.yaml.
func TestConfigSetParallelPlanKeysRefuseRepo(t *testing.T) {
	dir, _ := repoWithGlobal(t, "chat_language: English\nstyle: adhd\n", "")
	for _, args := range [][]string{
		{"config", "set", "--repo", "--parallel-plans", "subagent"},
		{"config", "set", "--repo", "--subagent-offer", "declined"},
	} {
		if code, _, _ := runCodeOut(args...); code != exitBadInput {
			t.Errorf("%v: exit %d, want bad input", args, code)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); !os.IsNotExist(err) {
		t.Errorf(".acta.yaml was written: %v", err)
	}
}

// show prints all three keys on both outputs. Unset values carry no
// (default) mark, since setup asks about every setting with that mark and
// these are not setup questions.
func TestConfigShowParallelPlanKeys(t *testing.T) {
	repoWithGlobal(t, "chat_language: English\nstyle: adhd\n", "")
	out := mustRun(t, "config", "show")
	for _, want := range []string{"parallel_plans: ask\n", "subagent_picks: 0\n", "subagent_offer: open\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("show lacks %q:\n%s", want, out)
		}
	}
	mustRun(t, "config", "set", "--parallel-plans", "subagent", "--subagent-offer", "declined")
	out = mustRun(t, "config", "show")
	for _, want := range []string{"parallel_plans: subagent\n", "subagent_offer: declined\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("show lacks %q:\n%s", want, out)
		}
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if got["parallel_plans"] != "subagent" || got["subagent_picks"] != float64(0) || got["subagent_offer"] != "declined" {
		t.Errorf("json %v %v %v", got["parallel_plans"], got["subagent_picks"], got["subagent_offer"])
	}
}
