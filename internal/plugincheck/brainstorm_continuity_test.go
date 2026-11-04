package plugincheck

import (
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
)

// TestBrainstormContinuity checks that the case rules live in the shape skill
// only. The skill must say: stay by default, name the case, let only the user
// pick a move, and list every way to move. No text may bring back a count,
// because the count was the old rule. The session note keeps one short rule
// about moving a brainstorm: real id, a way this harness has, launch nothing.
func TestBrainstormContinuity(t *testing.T) {
	session := func(herdr, codex bool) string {
		return hook.SessionStart(hook.Input{Voice: config.UserDefault(), VoiceExists: true, Herdr: herdr, Codex: codex})
	}
	shape := readFile(t, "skills", "shape", "SKILL.md")
	for _, want := range []string{
		"Related:", "Unrelated:", "Resumed:", "Mixed:", "never a count", "launch nothing unasked", "real id",
		"a new session typing `brainstorm SCR-0001`",
		"Claude Code only: `claude --bg 'brainstorm SCR-0001'`",
		"a herdr tab only when the session text has the `herdr:` line",
	} {
		if !strings.Contains(shape, want) {
			t.Errorf("shape skill is missing %q", want)
		}
	}
	texts := map[string]string{
		"shape skill":         shape,
		"default rules":       readFile(t, "hooks", "default-rules.md"),
		"session start":       session(false, false),
		"herdr session start": session(true, false),
		"codex session start": session(false, true),
	}
	for name, text := range texts {
		for _, old := range []string{"One Architectural brainstorm per session", "one per session", "cannot start in this session", "already brainstormed"} {
			if strings.Contains(text, old) {
				t.Errorf("%s still limits brainstorms: %q", name, old)
			}
		}
		if name != "shape skill" {
			for _, long := range []string{"Several Architectural brainstorms", "Related:", "Unrelated:", "Resumed:", "Mixed:", "stay here by default", "never a count"} {
				if strings.Contains(text, long) {
					t.Errorf("%s carries the long brainstorm rule (%q); it lives in shape only", name, long)
				}
			}
			// The short move rule stays in every session text.
			for _, want := range []string{"real id, like SCR-0001", "a new session typing brainstorm SCR-0001", "Claude Code only: claude --bg 'brainstorm SCR-0001'", "Launch nothing yourself."} {
				if !strings.Contains(text, want) {
					t.Errorf("%s is missing the move rule part %q", name, want)
				}
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
	// Only the rules count: the skill index above them may name herdr.
	plain := texts["session start"]
	if i := strings.Index(plain, "Core rules:"); i < 0 {
		t.Error("session start has no Core rules: heading")
	} else if strings.Contains(strings.ToLower(plain[i:]), "herdr") {
		t.Error("session start outside herdr names herdr")
	}
	// Frame leaves the item raw for its own reason, not to save a count.
	if frame := readFile(t, "skills", "frame", "SKILL.md"); strings.Contains(strings.ToLower(frame), "one architectural brainstorm") {
		t.Error("frame skill still counts brainstorms")
	}
}
