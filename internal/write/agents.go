package write

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// agentRec is what the board reads for one task: who works on it, and when
// they last said so.
type agentRec struct {
	Agent   string `json:"agent"`
	At      string `json:"at"`
	Started bool   `json:"started,omitempty"` // true once someone ran tick --start
}

// AgentName works out the name to record. The flag wins; otherwise the
// environment speaks, which is how a harness names itself. Claude Code sends
// a long version string, so only the product name is kept.
func AgentName(flag string, env func(string) string) string {
	name := strings.TrimSpace(flag)
	if name == "" {
		name = strings.TrimSpace(env("AI_AGENT"))
	}
	switch {
	case name == "":
		return ""
	case strings.HasPrefix(name, "claude-code_"):
		return "claude"
	default:
		head, _, _ := strings.Cut(name, "_")
		return strings.ToLower(head)
	}
}

// defaultAgentKey is the reserved .agents.json key holding the name a
// dispatch sent the worktree to. No task id can ever be "*", so readers
// skip it and a tick with no name of its own falls back to it.
const defaultAgentKey = "*"

// loadAgents reads the records file. A missing file or one that is not
// valid JSON gives an empty map, because a broken record must not stop
// a tick.
func loadAgents(path string) map[string]agentRec {
	recs := map[string]agentRec{}
	if b, err := os.ReadFile(path); err == nil {
		var old map[string]agentRec
		if json.Unmarshal(b, &old) == nil && old != nil {
			recs = old
		}
	}
	return recs
}

// storeAgents writes the records through a temp file next to the real one
// and renames, which no reader ever sees half-written.
func storeAgents(root, path string, recs map[string]agentRec) error {
	b, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	tmp, err := os.CreateTemp(root, ".agents.json.*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // a no-op once the rename below succeeds
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// RecordAgent writes which agent works on taskID into the root folder's
// .agents.json. The file lock keeps two ticks apart even when they work on
// different plans, because the whole read, change and write happens under
// it, so no name is lost. A file that is not valid JSON is replaced,
// because a broken record must not stop a tick.
func RecordAgent(root, taskID, agent string, now time.Time, started bool) error {
	path := filepath.Join(root, ".agents.json")
	unlock, err := lock(path)
	if err != nil {
		return err
	}
	defer unlock()
	recs := loadAgents(path)
	prev := recs[taskID]
	name := agent
	if name == "" {
		// A start with no name must not wipe the name an earlier tick
		// left, or the board shows nobody on a task someone is working.
		name = prev.Agent
	}
	if name == "" {
		// A dispatched worktree names its agent under "*", so a tick the
		// harness left nameless still says who it was.
		name = recs[defaultAgentKey].Agent
	}
	// A plain tick that still names nobody has nothing to write, so it
	// leaves the file alone.
	if name == "" && !started {
		return nil
	}
	// A later tick keeps an earlier start: losing it would drop the task
	// back to todo while someone is still on it.
	recs[taskID] = agentRec{Agent: name, At: now.Format(time.RFC3339), Started: started || prev.Started}
	return storeAgents(root, path, recs)
}

// SetDefaultAgent records the name a dispatch sent the worktree to under
// the "*" key, so later ticks with no name of their own fall back to it.
// It holds the same lock and writes through the same temp-file rename as
// RecordAgent: the two must never interleave.
func SetDefaultAgent(root, agent string) error {
	if agent == "" {
		return nil
	}
	path := filepath.Join(root, ".agents.json")
	unlock, err := lock(path)
	if err != nil {
		return err
	}
	defer unlock()
	recs := loadAgents(path)
	recs[defaultAgentKey] = agentRec{Agent: agent, At: time.Now().Format(time.RFC3339)}
	return storeAgents(root, path, recs)
}
