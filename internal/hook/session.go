package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ToolEvent is the part of a Claude Code tool hook payload this package needs.
type ToolEvent struct {
	SessionID string `json:"session_id"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

// stateFile holds one scratch item stem per session, under the acta root.
const stateFile = "state/sessions.json"

const (
	blockText = "acta: this session already brainstormed %s. One Architectural brainstorm per session: " +
		"file this one as a scratch item and offer the user the choices rule 8 names."
	reminderText = "acta: this session already brainstormed %s; a second brainstorm gets the choices rule 8 names."
)

var (
	// A command is only a brainstorm when it starts a chain piece, and it
	// must end it, so trailing words mean the agent is doing something else.
	brainCmd = regexp.MustCompile(`(?:^|&&|;|\|\|)\s*acta set scratch/([A-Za-z0-9._-]+) status brainstorming\s*$`)

	// A chain point is where the shell starts its next command. Cutting the
	// command at those points is what keeps a quoted acta set inside an echo
	// from looking like a command of its own.
	chainCut = regexp.MustCompile(`&&|\|\||;`)
)

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

// BrainstormStem gives the stem when the command really runs
// `acta set scratch/<stem> status brainstorming`, and nothing otherwise.
func BrainstormStem(command string) (string, bool) {
	for _, piece := range chainCut.Split(command, -1) {
		if m := brainCmd.FindStringSubmatch(piece); m != nil {
			return m[1], true
		}
	}
	return "", false
}

// RecordBrainstorm notes the item this session brainstormed. A command that is
// not a brainstorm is not a mistake, so it records nothing and returns no
// error: the post-tool hook runs on every single Bash call.
func RecordBrainstorm(root string, ev ToolEvent) error {
	stem, ok := BrainstormStem(ev.ToolInput.Command)
	if !ok || ev.SessionID == "" {
		return nil
	}
	state := readState(root)
	state[ev.SessionID] = stem
	return writeState(root, state)
}

// PreTool says whether to stop this command. Only a second, different
// brainstorm in a session that already has one is stopped, because a hook that
// guesses wrong stops real work the user asked for.
func PreTool(root string, ev ToolEvent) (bool, string) {
	stem, ok := BrainstormStem(ev.ToolInput.Command)
	if !ok {
		return false, ""
	}
	done, ok := readState(root)[ev.SessionID]
	if !ok || done == stem {
		return false, ""
	}
	return true, fmt.Sprintf(blockText, itemLabel(root, done))
}

// Reminder is the line added to the prompt of a session that has already
// brainstormed, and nothing at all for a session that has not.
func Reminder(root, sessionID string) string {
	stem, ok := readState(root)[sessionID]
	if !ok {
		return ""
	}
	return fmt.Sprintf(reminderText, itemLabel(root, stem))
}

// readState maps session id to the item that session brainstormed. A file that
// is missing, unreadable or broken gives an empty map, so a state file that
// went wrong looks like a session that has not brainstormed yet.
func readState(root string) map[string]string {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(stateFile)))
	if err != nil {
		return map[string]string{}
	}
	var state map[string]string
	if err := json.Unmarshal(data, &state); err != nil || state == nil {
		return map[string]string{}
	}
	return state
}

// writeState keeps the whole map in one file. The new file is written next to
// the old one and then moved over it, so a hook that dies mid-write cannot
// leave half a file behind for the next hook to choke on.
func writeState(root string, state map[string]string) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	path := filepath.Join(root, filepath.FromSlash(stateFile))
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "sessions-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// itemLabel names the item the way the user already knows it: the SCRATCH-n
// from the item's own frontmatter, or the stem when the file is gone or has no
// id, because a rough name beats no name at all.
func itemLabel(root, stem string) string {
	data, err := os.ReadFile(filepath.Join(root, "scratch", stem+".md"))
	if err != nil {
		return stem
	}
	for n, line := range strings.Split(string(data), "\n") {
		if n > 0 && strings.TrimSpace(line) == "---" {
			break
		}
		if id, ok := strings.CutPrefix(strings.TrimSpace(line), "id:"); ok {
			if id = strings.TrimSpace(id); id != "" {
				return id
			}
		}
	}
	return stem
}
