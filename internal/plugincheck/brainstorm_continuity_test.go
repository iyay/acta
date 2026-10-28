package plugincheck

import (
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
)

// TestBrainstormContinuity reads every text that tells an agent what to do
// when another topic comes up in the same session. They must all say the same
// thing: stay by default, name the case, and let only the user pick a move. No
// text may bring back a count, because the count was the old rule.
func TestBrainstormContinuity(t *testing.T) {
	session := func(herdr, codex bool) string {
		return hook.SessionStart(hook.Input{Voice: config.UserDefault(), VoiceExists: true, Herdr: herdr, Codex: codex})
	}
	texts := map[string]string{
		"shape skill":         readFile(t, "skills", "shape", "SKILL.md"),
		"default rules":       readFile(t, "hooks", "default-rules.md"),
		"session start":       session(false, false),
		"herdr session start": session(true, false),
		"codex session start": session(false, true),
	}
	for name, text := range texts {
		for _, want := range []string{"Related:", "Unrelated:", "Resumed:", "Mixed:", "never a count", "launch nothing unasked", "real id"} {
			if !strings.Contains(text, want) {
				t.Errorf("%s is missing %q", name, want)
			}
		}
		for _, old := range []string{"One Architectural brainstorm per session", "one per session", "cannot start in this session", "already brainstormed"} {
			if strings.Contains(text, old) {
				t.Errorf("%s still limits brainstorms: %q", name, old)
			}
		}
		// claude --bg runs only in Claude Code. Each place that names it must
		// say so, so an agent in Codex or omp never offers it.
		for rest := text; ; {
			i := strings.Index(rest, "claude --bg")
			if i < 0 {
				break
			}
			if !strings.HasSuffix(strings.TrimRight(rest[:i], "` "), "Claude Code only:") {
				t.Errorf("%s names claude --bg without saying it is for Claude Code only", name)
			}
			rest = rest[i+1:]
		}
	}
	// A herdr tab is a way to move only inside herdr. Outside it the session
	// text must not name herdr at all.
	if !strings.Contains(texts["herdr session start"], "a new herdr tab is one more way") {
		t.Error("herdr session start does not offer a herdr tab")
	}
	if rules := texts["session start"]; strings.Contains(strings.ToLower(rules[strings.Index(rules, "Core rules:"):]), "herdr") {
		t.Error("session start outside herdr names herdr")
	}
	// Frame leaves the item raw for its own reason, not to save a count.
	if frame := readFile(t, "skills", "frame", "SKILL.md"); strings.Contains(strings.ToLower(frame), "one architectural brainstorm") {
		t.Error("frame skill still counts brainstorms")
	}
}
