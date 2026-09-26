package write

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const taskID = "plans/2026-09-26-short-ids#task-3"

func TestAgentName(t *testing.T) {
	// A flag wins over the environment; otherwise AI_AGENT is read.
	vars := func(m map[string]string) func(string) string {
		return func(k string) string { return m[k] }
	}
	cases := []struct {
		name string
		flag string
		vars map[string]string
		want string
	}{
		{"flag wins", "omp", map[string]string{"AI_AGENT": "claude-code_2-1-283_agent"}, "omp"},
		{"claude code becomes claude", "", map[string]string{"AI_AGENT": "claude-code_2-1-283_agent"}, "claude"},
		{"other value keeps the first part", "", map[string]string{"AI_AGENT": "OpenCode_1-2"}, "opencode"},
		{"plain name is kept", "", map[string]string{"AI_AGENT": "omp"}, "omp"},
		{"blank flag falls back to the env", "  ", map[string]string{"AI_AGENT": "claude-code_2-1-283_agent"}, "claude"},
		{"nothing set records nothing", "", nil, ""},
		{"blank everywhere records nothing", " ", nil, ""},
	}
	for _, c := range cases {
		if got := AgentName(c.flag, vars(c.vars)); got != c.want {
			t.Errorf("%s: AgentName(%q) = %q, want %q", c.name, c.flag, got, c.want)
		}
	}
}

func TestRecordAgentWritesTheRecord(t *testing.T) {
	root := t.TempDir()
	at := time.Date(2026, 9, 26, 20, 46, 0, 0, time.FixedZone("WIB", 7*3600))
	if err := RecordAgent(root, taskID, "omp", at); err != nil {
		t.Fatal(err)
	}
	recs := readAgents(t, root)
	r, ok := recs[taskID]
	if !ok {
		t.Fatalf("%s missing from %v", taskID, recs)
	}
	if r.Agent != "omp" || r.At != at.Format(time.RFC3339) {
		t.Fatalf("record = %+v, want agent omp at %s", r, at.Format(time.RFC3339))
	}
	// The temp file used for the safe write must not stay behind.
	names, _ := filepath.Glob(filepath.Join(root, "*"))
	if len(names) != 1 {
		t.Errorf("root holds %v, want only .agents.json", names)
	}
}

func TestRecordAgentUpdatesOneKeyAndKeepsTheRest(t *testing.T) {
	root := t.TempDir()
	first := time.Date(2026, 9, 26, 9, 0, 0, 0, time.UTC)
	later := first.Add(time.Hour)
	if err := RecordAgent(root, "plans/p#task-1", "claude", first); err != nil {
		t.Fatal(err)
	}
	if err := RecordAgent(root, taskID, "omp", later); err != nil {
		t.Fatal(err)
	}
	recs := readAgents(t, root)
	if got := recs["plans/p#task-1"]; got.Agent != "claude" || got.At != first.Format(time.RFC3339) {
		t.Errorf("other key changed: %+v", got)
	}
	if got := recs[taskID]; got.At != later.Format(time.RFC3339) {
		t.Errorf("at = %q, want the newer %q", got.At, later.Format(time.RFC3339))
	}
}

func TestRecordAgentReplacesBrokenJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".agents.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RecordAgent(root, taskID, "omp", time.Now()); err != nil {
		t.Fatalf("broken file must be replaced, not crash: %v", err)
	}
	if recs := readAgents(t, root); recs[taskID].Agent != "omp" {
		t.Errorf("record after broken file = %v", recs)
	}
}

func TestRecordAgentWithoutANameWritesNothing(t *testing.T) {
	root := t.TempDir()
	if err := RecordAgent(root, taskID, "", time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents.json")); !os.IsNotExist(err) {
		t.Errorf("an empty agent name wrote the file: %v", err)
	}
}

func readAgents(t *testing.T, root string) map[string]agentRec {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".agents.json"))
	if err != nil {
		t.Fatal(err)
	}
	var recs map[string]agentRec
	if err := json.Unmarshal(b, &recs); err != nil {
		t.Fatalf("%s: %v (%s)", filepath.Join(root, ".agents.json"), err, b)
	}
	if strings.TrimSpace(string(b)) == "" {
		t.Fatal("empty file")
	}
	return recs
}
