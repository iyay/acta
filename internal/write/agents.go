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
	Agent string `json:"agent"`
	At    string `json:"at"`
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

// RecordAgent writes which agent works on taskID into the root folder's
// .agents.json. The tick lock already keeps two ticks on one plan apart, so
// this only has to keep the file itself readable: it writes a temp file next
// to it and renames, which no reader ever sees half-written. A file that is
// not valid JSON is replaced, because a broken record must not stop a tick.
func RecordAgent(root, taskID, agent string, now time.Time) error {
	if agent == "" {
		return nil
	}
	path := filepath.Join(root, ".agents.json")
	recs := map[string]agentRec{}
	if b, err := os.ReadFile(path); err == nil {
		var old map[string]agentRec
		if json.Unmarshal(b, &old) == nil && old != nil {
			recs = old
		}
	}
	recs[taskID] = agentRec{Agent: agent, At: now.Format(time.RFC3339)}
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
