package hook

import (
	"encoding/json"
	"io"
	"os"
)

// ToolEvent is the part of a Claude Code tool hook payload this package needs.
type ToolEvent struct {
	SessionID string `json:"session_id"`
	// AgentID is there only when the hook fires inside a subagent, so the main
	// thread has none.
	AgentID   string `json:"agent_id"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command  string `json:"command"`
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

// ParseEvent reads a tool hook payload. It says no on anything it cannot use,
// because a hook that cannot read its input must stay quiet and let the tool
// run rather than stop a session on one bad byte.
func ParseEvent(r io.Reader) (ToolEvent, bool) {
	var e ToolEvent
	data, err := io.ReadAll(r)
	if err != nil {
		return e, false
	}
	return e, json.Unmarshal(data, &e) == nil && e.SessionID != ""
}

// IsCodex says the hook runs under Codex. Codex sets PLUGIN_ROOT for plugin
// hooks next to CLAUDE_PLUGIN_ROOT, and Claude Code sets only the second one.
// See plugin/codex/FACTS.md.
func IsCodex() bool {
	return os.Getenv("PLUGIN_ROOT") != ""
}
