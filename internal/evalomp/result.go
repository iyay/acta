package evalomp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

// Call is one tool call the agent made. Input is the call's arguments as raw
// JSON, which is what a tool_used grader matches against.
type Call struct {
	Tool  string
	Input string
}

// Result is what one omp run said and did.
type Result struct {
	Reply string
	Calls []Call
}

// Workspace is what one case run left behind. Before lists the files that were
// there before the agent started, so graders can tell what the agent made.
type Workspace struct {
	Dir    string
	Before map[string]bool
	Result Result
}

type streamEvent struct {
	Type     string          `json:"type"`
	ToolName string          `json:"toolName"`
	Args     json.RawMessage `json:"args"`
	Messages []streamMessage `json:"messages"`
	Message  streamMessage   `json:"message"`
}

// streamMessage is one chat message. StopReason is only set on the assistant
// message a message_end event carries.
type streamMessage struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content"`
	StopReason string          `json:"stopReason"`
}

// ParseStream reads the events omp prints with --mode json and stops at the
// first assistant message that ends with stopReason "stop", or at agent_end,
// whichever comes first. omp can keep working after that reply, so anything
// later is not read and never counts. A run that reaches neither did not
// finish, so it is an error, not an empty reply.
func ParseStream(r io.Reader) (Result, error) {
	var res Result
	dec := json.NewDecoder(r)
	for {
		var ev streamEvent
		err := dec.Decode(&ev)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return res, fmt.Errorf("omp json stream: %w", err)
		}
		switch ev.Type {
		case "tool_execution_start":
			res.Calls = append(res.Calls, Call{Tool: ev.ToolName, Input: string(ev.Args)})
		case "message_end":
			if ev.Message.Role == "assistant" && ev.Message.StopReason == "stop" {
				res.Reply = replyText(ev.Message)
				return res, nil
			}
		case "agent_end":
			res.Reply = lastReply(ev)
			return res, nil
		}
	}
	return res, errors.New("omp stopped before its final reply or agent_end event")
}

// lastReply is the text of the last assistant message that has any. The very
// last message can be only a tool call, and that is not what the user saw.
func lastReply(ev streamEvent) string {
	for i := len(ev.Messages) - 1; i >= 0; i-- {
		m := ev.Messages[i]
		if m.Role != "assistant" {
			continue
		}
		if text := replyText(m); text != "" {
			return text
		}
	}
	return ""
}

// replyText joins the text parts of one message. A message whose content is
// not a list of parts has none.
func replyText(m streamMessage) string {
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(m.Content, &parts) != nil {
		return ""
	}
	var texts []string
	for _, p := range parts {
		if p.Type == "text" {
			texts = append(texts, p.Text)
		}
	}
	return strings.Join(texts, "\n")
}

// ListFiles gives every file under dir as a slash path relative to dir. The
// .git folder is left out: git's own files are never what a grader asks about.
func ListFiles(dir string) (map[string]bool, error) {
	files := map[string]bool{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = true
		return nil
	})
	return files, err
}
