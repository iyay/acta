// Package hook builds the text the acta plugin puts in front of the agent: the
// session rules at the start and a one-line voice reminder on every message.
package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iyay/acta/internal/config"
)

// Skill is one acta skill and when to use it.
type Skill struct{ Name, When string }

// Skills is the index printed every session. The plugin's skill folders must
// match these names exactly.
var Skills = []Skill{
	{"brainstorm", "before any new feature, fix or behaviour change; design first, then wait for a yes"},
	{"plan", "after the design is approved; tasks with verify lines and waves, then wait for a yes"},
	{"build", "run an approved plan in a worktree; executor from `acta config show`, else ask: subagent, dispatch or inline"},
	{"tdd", "every code change; a failing test first"},
	{"debug", "any bug, error, red test or wrong output, before touching code"},
	{"review", "when every task is done; two reviewers, BLOCKER or NOTE, three rounds at most"},
	{"land", "after a clean review; gates, merge --no-ff, clean up, never push"},
	{"bug", "record a confirmed bug with acta bug new"},
	{"scratch", `raw ideas ("note this", "later", side ideas, any language); file with acta scratch new, never memory`},
	{"setup", "first-run setup and later changes: doctor, voice, build executor, subagent models, CLAUDE.md block"},
	{"migrate", "move docs from another workflow plugin into .acta/"},
}

// Input is everything the hook text depends on.
type Input struct {
	Voice       config.User
	VoiceExists bool
	VoiceErr    error
	Conflicts   []string // enabled workflow plugins that clash with acta
	Herdr       bool     // this session runs in a herdr tab; the hook reads the environment, the agent often cannot
}

const coreRules = `
Core rules:
1. No action without an ask. Reading, answering and planning are the default.
2. No code before an approved design and an approved plan.
3. Every code change happens in a worktree and starts with a failing test.
4. Review once, at the close, three rounds at most. Land without asking when clean.
5. Never push. Never run a destructive command without a full-sentence warning and a yes.
6. When the user's own CLAUDE.md or AGENTS.md says otherwise, follow it.
7. If your instructions name a skill from the superpowers plugin that is not installed, use the acta skill for that step: brainstorming→acta:brainstorm, writing-plans→acta:plan, subagent-driven-development→acta:build, using-git-worktrees→acta:build, test-driven-development→acta:tdd, systematic-debugging→acta:debug, requesting-code-review→acta:review, receiving-code-review→acta:review, verification-before-completion→acta:land, finishing-a-development-branch→acta:land.
8. One Architectural brainstorm per session. A second one cannot start in this session. File the scratch item first, with one acta scratch new call whose body is stdin: acta scratch new <slug> --title <title> < body.md: no Skill tool, no acta scratch add, and "written, not committed" still counts as filed. In the same reply, name the two ways to open it elsewhere: put the id the command printed in place of SCRATCH-n, an id like SCR-0001, never a shortened one. A background agent (claude --bg 'brainstorm SCRATCH-n') and a new session where the user types brainstorm SCRATCH-n. When the user picks one, load acta:brainstorm for that way. Do not design it here.
`

// herdrExtra is the one extra choice, added only inside a herdr session. The
// agent often cannot read the environment, so the text decides this, not the
// agent: a session outside herdr never hears the word.
const herdrExtra = `
herdr: this session runs in a herdr tab. For a second brainstorm, offer the user one more way to run it: a new herdr tab.
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
Voice: not set up yet. If the user's CLAUDE.md or AGENTS.md already names a chat language or style, do not ask again: offer once to save those values with /acta:setup and wait for a yes.
Otherwise, before other work in this session, run /acta:setup. It runs acta doctor, then asks for the chat language, the style and, if you want, a tone, and saves the answers.
Until then, write in English (or the language CLAUDE.md names), adhd style.
`

// SessionStart is the text for the start of a session and after compaction.
func SessionStart(in Input) string {
	var b strings.Builder
	b.WriteString("acta plugin is active. Before each workflow step, load its acta skill with the Skill tool and follow it. This list is only an index; the rules live in the skills:\n")
	for _, s := range Skills {
		fmt.Fprintf(&b, "- acta:%s: %s\n", s.Name, s.When)
	}
	b.WriteString(coreRules)
	if in.Herdr {
		b.WriteString(herdrExtra)
	}
	switch {
	case !in.VoiceExists:
		b.WriteString(firstRun)
		b.WriteString(adhdRules)
	case in.VoiceErr != nil:
		fmt.Fprintf(&b, "\nVoice: the voice file could not be read (%v). Write in English, adhd style, until it is fixed; acta:setup rewrites it.\n", in.VoiceErr)
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
		return "acta config: not set up yet; reply in English, adhd style, and run /acta:setup once (see the session rules)."
	case in.VoiceErr != nil:
		return "acta config: the config file could not be read; reply in English, adhd style."
	default:
		return fmt.Sprintf("acta config: reply in %s, %s style.", in.Voice.ChatLanguage, in.Voice.Style)
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

// LoadKnown reads the list of workflow plugins that clash with acta: one name
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
// the known list. acta itself never counts.
func Conflicts(enabled, known []string) []string {
	var out []string
	for _, e := range enabled {
		name, market, ok := strings.Cut(e, "@")
		if !ok || strings.EqualFold(name, "acta") {
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
