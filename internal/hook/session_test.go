package hook

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// brokenReader stands in for a stdin that cannot be read.
type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("stdin is gone") }

func TestParseEvent(t *testing.T) {
	for _, in := range []string{"", "{bad", `{"tool_input":{"command":"x"}}`, "null", "[]", `{"session_id":""}`} {
		if _, ok := ParseEvent(strings.NewReader(in)); ok {
			t.Errorf("ParseEvent(%q) ok, want false", in)
		}
	}
	if _, ok := ParseEvent(brokenReader{}); ok {
		t.Error("a read error must not parse")
	}
	e, ok := ParseEvent(strings.NewReader(`{"session_id":"s1","tool_input":{"command":"ls"}}`))
	if !ok || e.SessionID != "s1" || e.ToolInput.Command != "ls" {
		t.Fatalf("got %+v %v", e, ok)
	}
	// A missing tool_input is still an event: it just has no command.
	e, ok = ParseEvent(strings.NewReader(`{"session_id":"s1"}`))
	if !ok || e.ToolInput.Command != "" {
		t.Fatalf("got %+v %v", e, ok)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Codex sets PLUGIN_ROOT for plugin hooks. Claude Code sets only
// CLAUDE_PLUGIN_ROOT, so that one alone must not count.
func TestIsCodex(t *testing.T) {
	t.Setenv("PLUGIN_ROOT", "")
	t.Setenv("CLAUDE_PLUGIN_ROOT", "/p")
	if IsCodex() {
		t.Error("IsCodex() = true with only CLAUDE_PLUGIN_ROOT set")
	}
	t.Setenv("PLUGIN_ROOT", "/p")
	if !IsCodex() {
		t.Error("IsCodex() = false with PLUGIN_ROOT set")
	}
}
