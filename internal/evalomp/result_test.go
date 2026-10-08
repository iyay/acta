package evalomp

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseStream(t *testing.T) {
	f, err := os.Open("testdata/stream.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	res, err := ParseStream(f)
	if err != nil {
		t.Fatal(err)
	}
	if res.Reply != "Filed as SCR-0001.\nNext: brainstorm it." {
		t.Errorf("reply = %q", res.Reply)
	}
	if len(res.Calls) != 2 || res.Calls[0].Tool != "bash" || res.Calls[1].Tool != "read" {
		t.Fatalf("calls = %+v", res.Calls)
	}
	if res.Calls[0].Input != `{"command":"acta scratch new idea --title idea"}` {
		t.Errorf("input = %q", res.Calls[0].Input)
	}
}

func TestParseStreamErrors(t *testing.T) {
	for name, tc := range map[string]struct{ text, want string }{
		"no agent_end": {`{"type":"agent_start"}` + "\n", "agent_end"},
		"not json":     {"hello\n", "omp json stream"},
		"empty":        {"", "agent_end"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ParseStream(strings.NewReader(tc.text))
			if err == nil {
				t.Fatal("want an error")
			}
			// The message says which of the two failures happened, so a run
			// that printed junk is not read as a run that just stopped.
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestParseStreamOneLine(t *testing.T) {
	// omp sometimes prints several events on one line, so the reader is a json
	// decoder and not a line splitter.
	text := `{"type":"tool_execution_start","toolName":"grep","args":{"q":"x"}} {"type":"agent_end","messages":[{"role":"assistant","content":[{"type":"text","text":"done"}]}]}` + "\n"
	res, err := ParseStream(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	if res.Reply != "done" {
		t.Errorf("reply = %q", res.Reply)
	}
	if len(res.Calls) != 1 || res.Calls[0].Tool != "grep" {
		t.Fatalf("calls = %+v", res.Calls)
	}
}

func TestParseStreamLastAssistantWins(t *testing.T) {
	// Two assistant messages carry text. The user read the later one, so that
	// is the reply, not the first one.
	text := `{"type":"agent_end","messages":[{"role":"assistant","content":[{"type":"text","text":"first"}]},{"role":"assistant","content":[{"type":"text","text":"second"}]}]}` + "\n"
	res, err := ParseStream(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	if res.Reply != "second" {
		t.Errorf("reply = %q", res.Reply)
	}
}

func TestParseStreamOnlyCustomMessage(t *testing.T) {
	// The plugin notice carries plain text, not content parts. It is skipped,
	// so a run whose last message is that notice has no reply but no error.
	text := `{"type":"agent_end","messages":[{"role":"custom","customType":"acta","content":"acta plugin is active."}]}` + "\n"
	res, err := ParseStream(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	if res.Reply != "" {
		t.Errorf("reply = %q", res.Reply)
	}
}

func TestParseStreamSkipsUnusableMessages(t *testing.T) {
	for name, text := range map[string]string{
		// A plugin notice after the reply text must not end the walk.
		"custom message last": `{"type":"agent_end","messages":[{"role":"assistant","content":[{"type":"text","text":"done"}]},{"role":"custom","customType":"acta","content":"acta plugin is active."}]}` + "\n",
		// An assistant message whose content is plain text has no parts to
		// join, so the walk keeps going instead of failing.
		"assistant string content": `{"type":"agent_end","messages":[{"role":"assistant","content":"done"},{"role":"assistant","content":[{"type":"text","text":"really done"}]}]}` + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			res, err := ParseStream(strings.NewReader(text))
			if err != nil {
				t.Fatal(err)
			}
			if res.Reply == "" {
				t.Error("want the earlier reply, not an empty one")
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	dir := t.TempDir()
	// A local helper, so this task does not wait for case_test.go's write.
	put := func(rel string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	put("a.txt")
	put(".acta/scratch/one.md")
	put(".git/HEAD")
	got, err := ListFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !got["a.txt"] || !got[".acta/scratch/one.md"] {
		t.Errorf("files = %v", got)
	}
	if got[".git/HEAD"] {
		t.Error("git's own files must stay out of the list")
	}
}

// messageEnd builds the line omp prints when one assistant message is done.
func messageEnd(stop, text string) string {
	return `{"type":"message_end","message":{"role":"assistant","content":[{"type":"text","text":"` + text +
		`"}],"stopReason":"` + stop + `"}}` + "\n"
}

// The first assistant message that stops is the reply. Whatever omp does
// after it, a later tool call, reply or agent_end, must not leak in.
func TestParseStreamStopsAtFirstFinalReply(t *testing.T) {
	text := `{"type":"tool_execution_start","toolName":"bash","args":{"command":"a"}}` + "\n" +
		messageEnd("toolUse", "working on it") +
		messageEnd("stop", "done") +
		`{"type":"agent_start"}` + "\n" +
		`{"type":"tool_execution_start","toolName":"read","args":{}}` + "\n" +
		messageEnd("stop", "second reply") +
		`{"type":"agent_end","messages":[{"role":"assistant","content":[{"type":"text","text":"late"}]}]}` + "\n"
	res, err := ParseStream(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	if res.Reply != "done" {
		t.Errorf("reply = %q, want the first final reply", res.Reply)
	}
	if len(res.Calls) != 1 || res.Calls[0].Tool != "bash" {
		t.Errorf("calls = %+v, want only the call before the final reply", res.Calls)
	}
}

// The reader must hand back at the final reply while omp is still running,
// or RunCase would wait for an exit that never comes.
func TestParseStreamReturnsWithoutEndOfInput(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()
	go io.WriteString(pw, messageEnd("stop", "done"))
	got := make(chan Result, 1)
	go func() {
		res, _ := ParseStream(pr)
		got <- res
	}()
	select {
	case res := <-got:
		if res.Reply != "done" {
			t.Errorf("reply = %q", res.Reply)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("ParseStream kept reading after the final reply")
	}
}

// A reply that stops for another reason is not a final reply: an aborted or
// tool-use message must not end the run, so a stream with only those errors.
func TestParseStreamOtherStopReasonsAreNotFinal(t *testing.T) {
	for _, reason := range []string{"toolUse", "aborted", "error", ""} {
		t.Run(reason, func(t *testing.T) {
			if _, err := ParseStream(strings.NewReader(messageEnd(reason, "x"))); err == nil {
				t.Error("want an error: no final reply and no agent_end")
			}
		})
	}
}

// A final reply with no text parts is still the end. It gives an empty reply
// and no error, like the old agent_end with no text did.
func TestParseStreamFinalReplyWithoutText(t *testing.T) {
	text := `{"type":"message_end","message":{"role":"assistant","content":[{"type":"toolCall","id":"c"}],"stopReason":"stop"}}` + "\n"
	res, err := ParseStream(strings.NewReader(text))
	if err != nil || res.Reply != "" {
		t.Errorf("res = %+v, err = %v", res, err)
	}
}
