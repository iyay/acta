package write

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const taskID = "plans/2026-09-26-short-ids#task-3"

func TestAgentName(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	root := t.TempDir()
	at := time.Date(2026, 9, 26, 20, 46, 0, 0, time.FixedZone("WIB", 7*3600))
	if err := RecordAgent(root, taskID, "omp", at, false); err != nil {
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
	t.Parallel()

	root := t.TempDir()
	first := time.Date(2026, 9, 26, 9, 0, 0, 0, time.UTC)
	later := first.Add(time.Hour)
	if err := RecordAgent(root, "plans/p#task-1", "claude", first, false); err != nil {
		t.Fatal(err)
	}
	if err := RecordAgent(root, taskID, "omp", later, false); err != nil {
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
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".agents.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RecordAgent(root, taskID, "omp", time.Now(), false); err != nil {
		t.Fatalf("broken file must be replaced, not crash: %v", err)
	}
	if recs := readAgents(t, root); recs[taskID].Agent != "omp" {
		t.Errorf("record after broken file = %v", recs)
	}
}

func TestRecordAgentWithoutANameWritesNothing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := RecordAgent(root, taskID, "", time.Now(), false); err != nil {
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

// A start is kept: the next normal tick updates name and time but must not
// clear the started flag, or the board would drop back to todo.
func TestRecordAgentKeepsStartedOnALaterTick(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := RecordAgent(root, taskID, "", time.Now(), true); err != nil {
		t.Fatal(err)
	}
	if err := RecordAgent(root, taskID, "omp", time.Now().Add(time.Hour), false); err != nil {
		t.Fatal(err)
	}
	r := readAgents(t, root)[taskID]
	if !r.Started || r.Agent != "omp" {
		t.Fatalf("record = %+v, want agent omp with started kept", r)
	}
}

// A start with no agent name still writes: the board must see the start
// even when the harness names nobody.
func TestRecordAgentStartWritesWithoutAName(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := RecordAgent(root, taskID, "", time.Now(), true); err != nil {
		t.Fatal(err)
	}
	r, ok := readAgents(t, root)[taskID]
	if !ok || !r.Started {
		t.Fatalf("record = %+v (present %t), want started true", r, ok)
	}
}

// Two ticks can land in the same .agents.json at the same moment, one for
// each plan. The lock keeps them apart, so both names stay in the file.
func TestRecordAgentTwoTasksAtOnceKeepsBoth(t *testing.T) {
	useLockBase(t)
	ids := []string{"plans/p#task-1", "plans/p#task-2"}
	at := time.Date(2026, 9, 26, 10, 0, 0, 0, time.UTC)
	for round := range 50 {
		root := t.TempDir()
		var wg sync.WaitGroup
		for _, id := range ids {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := RecordAgent(root, id, "omp", at, false); err != nil {
					t.Error(err)
				}
			}()
		}
		wg.Wait()
		recs := readAgents(t, root)
		for _, id := range ids {
			if got := recs[id]; got.Agent != "omp" || got.At != at.Format(time.RFC3339) {
				t.Fatalf("round %d: %s = %+v, want omp at %s", round, id, got, at.Format(time.RFC3339))
			}
		}
	}
}

// A start with no agent name must not wipe the name an earlier tick left.
func TestRecordAgentNamelessStartKeepsTheName(t *testing.T) {
	useLockBase(t)
	root := t.TempDir()
	first := time.Date(2026, 9, 26, 9, 0, 0, 0, time.UTC)
	later := first.Add(time.Hour)
	if err := RecordAgent(root, taskID, "omp", first, false); err != nil {
		t.Fatal(err)
	}
	if err := RecordAgent(root, taskID, "", later, true); err != nil {
		t.Fatal(err)
	}
	r := readAgents(t, root)[taskID]
	if r.Agent != "omp" || !r.Started || r.At != later.Format(time.RFC3339) {
		t.Fatalf("record = %+v, want agent omp, started, at %s", r, later.Format(time.RFC3339))
	}
}

// A tick that has a name of its own still replaces the older one, so the
// keep-the-name rule does not freeze a name forever.
func TestRecordAgentNamedTickReplacesTheOlderName(t *testing.T) {
	useLockBase(t)
	root := t.TempDir()
	first := time.Date(2026, 9, 26, 9, 0, 0, 0, time.UTC)
	later := first.Add(time.Hour)
	if err := RecordAgent(root, taskID, "claude", first, false); err != nil {
		t.Fatal(err)
	}
	if err := RecordAgent(root, taskID, "omp", later, false); err != nil {
		t.Fatal(err)
	}
	r := readAgents(t, root)[taskID]
	if r.Agent != "omp" || r.At != later.Format(time.RFC3339) {
		t.Fatalf("record = %+v, want agent omp at %s", r, later.Format(time.RFC3339))
	}
}

// A tick with no name falls back to the worktree default a dispatch left
// under "*", so a dispatched tick still names its agent.
func TestRecordAgentFallsBackToDefaultAgent(t *testing.T) {
	useLockBase(t)
	root := t.TempDir()
	def := `{"*": {"agent": "omp", "at": "2026-10-06T00:00:00Z"}}`
	if err := os.WriteFile(filepath.Join(root, ".agents.json"), []byte(def), 0o644); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 10, 6, 9, 0, 0, 0, time.UTC)
	if err := RecordAgent(root, taskID, "", at, true); err != nil {
		t.Fatal(err)
	}
	r := readAgents(t, root)[taskID]
	if r.Agent != "omp" || !r.Started || r.At != at.Format(time.RFC3339) {
		t.Fatalf("record = %+v, want agent omp, started, at %s", r, at.Format(time.RFC3339))
	}
}

// A plain tick with no name falls back to the default too, not just a start.
func TestRecordAgentPlainTickFallsBackToDefault(t *testing.T) {
	useLockBase(t)
	root := t.TempDir()
	def := `{"*": {"agent": "omp", "at": "2026-10-06T00:00:00Z"}}`
	if err := os.WriteFile(filepath.Join(root, ".agents.json"), []byte(def), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RecordAgent(root, taskID, "", time.Now(), false); err != nil {
		t.Fatal(err)
	}
	if r := readAgents(t, root)[taskID]; r.Agent != "omp" {
		t.Fatalf("record = %+v, want agent omp from the default", r)
	}
}

// The task's own name beats the default: an earlier tick is more exact
// than the whole worktree's name.
func TestRecordAgentKeepsTaskNameOverDefault(t *testing.T) {
	useLockBase(t)
	root := t.TempDir()
	def := "{\"*\": {\"agent\": \"omp\", \"at\": \"2026-10-06T00:00:00Z\"}, \"" + taskID + "\": {\"agent\": \"claude\", \"at\": \"2026-10-06T08:00:00Z\"}}"
	if err := os.WriteFile(filepath.Join(root, ".agents.json"), []byte(def), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RecordAgent(root, taskID, "", time.Now(), true); err != nil {
		t.Fatal(err)
	}
	if r := readAgents(t, root)[taskID]; r.Agent != "claude" {
		t.Fatalf("record = %+v, want agent claude kept over the default", r)
	}
}

// A tick that names someone beats the default outright.
func TestRecordAgentNamedTickBeatsTheDefault(t *testing.T) {
	useLockBase(t)
	root := t.TempDir()
	def := `{"*": {"agent": "omp", "at": "2026-10-06T00:00:00Z"}}`
	if err := os.WriteFile(filepath.Join(root, ".agents.json"), []byte(def), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RecordAgent(root, taskID, "claude", time.Now(), false); err != nil {
		t.Fatal(err)
	}
	if r := readAgents(t, root)[taskID]; r.Agent != "claude" {
		t.Fatalf("record = %+v, want agent claude, not the default", r)
	}
}
