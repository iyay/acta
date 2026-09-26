// Package hook builds the text the pm plugin puts in front of the agent: the
// session rules at the start and a one-line voice reminder on every message.
package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pm-board/internal/voice"
)

// Skill is one pm skill and when to use it.
type Skill struct{ Name, When string }

// Skills is the index printed every session. The plugin's skill folders must
// match these names exactly.
var Skills = []Skill{
	{"brainstorm", "before any new feature, fix or behaviour change; design first, then wait for a yes"},
	{"plan", "after the design is approved; tasks with verify lines and waves, then wait for a yes"},
	{"build", "run an approved plan in a worktree; executor subagent (default), dispatch or inline"},
	{"tdd", "every code change; a failing test first"},
	{"debug", "any bug, error, red test or wrong output, before touching code"},
	{"review", "when every task is done; two reviewers, BLOCKER or NOTE, three rounds at most"},
	{"land", "after a clean review; gates, merge --no-ff, clean up, never push"},
	{"bug", "record a confirmed bug with pmb bug new"},
	{"dispatch", "run build through an omp agent in its own herdr tab"},
	{"setup", "change the chat language, style or tone"},
	{"migrate", "move docs from another workflow plugin into .pm/"},
}

// Input is everything the hook text depends on.
type Input struct {
	Voice       voice.Voice
	VoiceExists bool
	VoiceErr    error
	Conflicts   []string // enabled workflow plugins that clash with pm
}

const coreRules = `
Core rules:
1. No action without an ask. Reading, answering and planning are the default.
2. No code before an approved design and an approved plan.
3. Every code change happens in a worktree and starts with a failing test.
4. Review once, at the close, three rounds at most. Land without asking when clean.
5. Never push. Never run a destructive command without a full-sentence warning and a yes.
6. When the user's own CLAUDE.md or AGENTS.md says otherwise, follow it.
`

// adhdRules is a short form of the i-have-adhd plugin's rules (MIT).
const adhdRules = `
Style (ADHD reader):
- The first line is the answer or the next action. No preamble.
- Multi-step work gets numbered steps, one action each, as few as work.
- End with one next action the reader can do in under two minutes.
- Restate where the work stands every turn.
- Give time estimates in concrete units.
- Show what now works and how to see it.
- Errors: cause and fix, plainly.
- Lists: five items at most.
- No recap, no closing pleasantries.
`

const firstRun = `
Voice: not set up yet. If the user's CLAUDE.md or AGENTS.md already names a chat language or style, do not ask the questions below: offer once to save those values with the pmb voice set command below, and wait for a yes.
Otherwise, before other work in this session, ask the user once, in English:
1. Which language should chat use? (default English)
2. Style: adhd (answer first, short steps) or plain? (default adhd)
3. Anything about tone, in their own words? (optional)
Then save it: pmb voice set --language <full language name> --style <adhd|plain> [--tone "<text>"]
Until then, write in English (or the language CLAUDE.md names), adhd style.
`

// SessionStart is the text for the start of a session and after compaction.
func SessionStart(in Input) string {
	var b strings.Builder
	b.WriteString("pm plugin is active. Use its skills for every workflow step:\n")
	for _, s := range Skills {
		fmt.Fprintf(&b, "- pm:%s: %s\n", s.Name, s.When)
	}
	b.WriteString(coreRules)
	switch {
	case !in.VoiceExists:
		b.WriteString(firstRun)
		b.WriteString(adhdRules)
	case in.VoiceErr != nil:
		fmt.Fprintf(&b, "\nVoice: the voice file could not be read (%v). Write in English, adhd style, until it is fixed; pm:setup rewrites it.\n", in.VoiceErr)
		b.WriteString(adhdRules)
	default:
		v := in.Voice
		fmt.Fprintf(&b, "\nVoice:\n- Write every chat message to the user in %s.\n", v.ChatLanguage)
		fmt.Fprintf(&b, "- Write everything that goes into the repo (code, comments, commits, specs, plans) in %s. Comments use short, plain words and say why.\n", v.RepoLanguage)
		b.WriteString("- Warnings before a destructive command, and security findings, are always full, clear sentences.\n")
		if v.Tone != "" {
			b.WriteString("- Tone, in the user's words:\n")
			for _, ln := range strings.Split(v.Tone, "\n") {
				fmt.Fprintf(&b, "  %s\n", ln)
			}
		}
		if v.Style == "adhd" {
			b.WriteString(adhdRules)
		}
	}
	if len(in.Conflicts) > 0 {
		off := map[string]bool{}
		for _, c := range in.Conflicts {
			off[c] = false
		}
		// json.Marshal sorts map keys, so the snippet is stable.
		snippet, _ := json.Marshal(map[string]any{"enabledPlugins": off})
		fmt.Fprintf(&b, "\nAnother workflow plugin is enabled here: %s.\n", strings.Join(in.Conflicts, ", "))
		b.WriteString("Tell the user once, at the start: two workflow plugins pull the agent two ways. To turn it off in this repo only, add this to .claude/settings.local.json:\n")
		fmt.Fprintf(&b, "%s\n", snippet)
	}
	return b.String()
}

// Prompt is the one-line reminder added to every user message.
func Prompt(in Input) string {
	switch {
	case !in.VoiceExists:
		return "pm voice: not set up yet; reply in English, adhd style, and ask the user once (see the session rules)."
	case in.VoiceErr != nil:
		return "pm voice: the voice file could not be read; reply in English, adhd style."
	default:
		return fmt.Sprintf("pm voice: reply in %s, %s style.", in.Voice.ChatLanguage, in.Voice.Style)
	}
}

// ClaudeDir is CLAUDE_CONFIG_DIR when set, else ~/.claude.
func ClaudeDir() string {
	if d := os.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

// EnabledPlugins reads enabledPlugins from the user, project and local
// settings, in that order, so a later file can turn a plugin off. It only
// reads; a missing or broken file is skipped because a hook must not fail.
func EnabledPlugins(claudeDir, repoRoot string) []string {
	merged := map[string]bool{}
	for _, p := range []string{
		filepath.Join(claudeDir, "settings.json"),
		filepath.Join(repoRoot, ".claude", "settings.json"),
		filepath.Join(repoRoot, ".claude", "settings.local.json"),
	} {
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var s struct {
			EnabledPlugins map[string]bool `json:"enabledPlugins"`
		}
		if json.Unmarshal(raw, &s) != nil {
			continue
		}
		for k, v := range s.EnabledPlugins {
			merged[k] = v
		}
	}
	out := []string{}
	for k, v := range merged {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// LoadKnown reads the list of workflow plugins that clash with pm: one name
// per line, blank lines and # comments skipped. No file means no list.
func LoadKnown(path string) []string {
	if path == "" {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []string
	for _, ln := range strings.Split(string(raw), "\n") {
		ln = strings.TrimSpace(ln)
		if ln != "" && !strings.HasPrefix(ln, "#") {
			out = append(out, ln)
		}
	}
	return out
}

// Conflicts keeps enabled plugins whose plugin name or marketplace name is on
// the known list. pm itself never counts.
func Conflicts(enabled, known []string) []string {
	var out []string
	for _, e := range enabled {
		name, market, ok := strings.Cut(e, "@")
		if !ok || strings.EqualFold(name, "pm") {
			continue
		}
		for _, k := range known {
			if strings.EqualFold(k, name) || strings.EqualFold(k, market) {
				out = append(out, e)
				break
			}
		}
	}
	return out
}
