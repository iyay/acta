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

// Skills is the index printed every session: names only, because Claude Code
// and omp already show each skill's description. The plugin's skill folders
// must match these names exactly.
var Skills = []string{"shape", "slice", "build", "tdd", "lean", "debug", "review", "land", "bug", "scratch", "setup", "migrate"}

// Input is everything the hook text depends on.
type Input struct {
	Voice       config.User
	VoiceExists bool
	VoiceErr    error
	RepoErr     error    // .acta.yaml could not be read; its repo keys are left out
	Conflicts   []string // enabled plugins that overlap acta
	Herdr       bool     // this session runs in a herdr tab; the hook reads the environment, the agent often cannot
}

const coreRules = `
Core rules:
1. No action without an ask. Reading, answering and planning are the default.
2. No code before an approved design and an approved plan. A plan with depth: minimal in its frontmatter needs no plan yes.
3. Every code change happens in a worktree and starts with a failing test.
4. Review once, at the close, three rounds at most. Land without asking when clean.
5. Never push. Never run a destructive command without a full-sentence warning and a yes.
6. When the user's own CLAUDE.md or AGENTS.md says otherwise, follow it.
7. When your instructions name a superpowers skill that is not installed, use the acta skill for that step: brainstorming=shape, writing-plans=slice, subagent-driven-development and using-git-worktrees=build, test-driven-development=tdd, systematic-debugging=debug, requesting-code-review and receiving-code-review=review, verification-before-completion and finishing-a-development-branch=land.
8. One Architectural brainstorm per session. A second one cannot start in this session. File the scratch item first, with one acta scratch new call whose body is stdin: acta scratch new <slug> --title <title> < body.md: no Skill tool, no acta scratch add, and "written, not committed" still counts as filed. In the same reply, name the two ways to open it elsewhere: put the id the command printed in place of SCRATCH-n, an id like SCR-0001, never a shortened one. A background agent (claude --bg 'brainstorm SCRATCH-n') and a new session where the user types brainstorm SCRATCH-n. When the user picks one, load acta:shape for that way. Do not design it here.
`

// herdrExtra is the one extra choice, added only inside a herdr session. The
// agent often cannot read the environment, so the text decides this, not the
// agent: a session outside herdr never hears the word.
const herdrExtra = `
herdr: this session runs in a herdr tab. For a second brainstorm, offer the user one more way to run it: a new herdr tab.
`

// leanSummary is the short form of the coding guide. The skill acta:lean holds
// the full text. It is left out when the config says coding_guide: off.
const leanSummary = `
Lean coding guide (full text: acta:lean):
- Understand the task and the code it touches before choosing.
- Then take the first rung that works: skip it, reuse code here, stdlib, a native platform feature, an installed dependency, the fewest lines.
- No abstraction with one user, no config for a fixed value, no scaffolding for later.
- Fix a bug where every caller passes through, not only the reported path.
- Never cut checks at trust boundaries, error handling that prevents data loss, security or accessibility.
`

const firstRun = `
Voice: not set up yet. If the user's CLAUDE.md or AGENTS.md already names a chat language or style, do not ask again: offer once to save those values with /acta:setup and wait for a yes.
Otherwise, before other work in this session, run /acta:setup. It runs acta doctor, then asks for the chat language, the style and, if you want, a tone, and saves the answers.
Until then, write in English (or the language CLAUDE.md names), adhd style.
`

// SessionStart is the text for the start of a session and after compaction.
func SessionStart(in Input) string {
	var b strings.Builder
	fmt.Fprintf(&b, "acta plugin is active. Before each workflow step, load the matching acta skill with the Skill tool and follow it: %s. The rules live in the skills.\n", strings.Join(Skills, ", "))
	b.WriteString(coreRules)
	if in.Herdr {
		b.WriteString(herdrExtra)
	}
	switch {
	case !in.VoiceExists:
		b.WriteString(firstRun)
		b.WriteString("- Style: adhd.\n")
	case in.VoiceErr != nil:
		fmt.Fprintf(&b, "\nVoice: the voice file could not be read (%v). Write in English, adhd style, until it is fixed; acta:setup rewrites it.\n", in.VoiceErr)
		b.WriteString("- Style: adhd.\n")
	default:
		v := in.Voice
		fmt.Fprintf(&b, "\nVoice:\n- Write every chat message to the user in %s.\n", v.ChatLanguage)
		fmt.Fprintf(&b, "- Write everything that goes into the repo (code, comments, commits, specs, plans) in %s. Comments use short, plain words and say why.\n", v.RepoLanguage)
		fmt.Fprintf(&b, "- Style: %s.\n", v.Style)
		if v.Tone != "" {
			b.WriteString("- Tone, in the user's words:\n")
			for _, ln := range strings.Split(v.Tone, "\n") {
				fmt.Fprintf(&b, "  %s\n", ln)
			}
		}
		if in.RepoErr != nil {
			fmt.Fprintf(&b, "- The repo settings could not be read (%v). The ones above come from your own config until .acta.yaml is fixed.\n", in.RepoErr)
		}
	}
	if in.Voice.CodingGuide != "off" {
		b.WriteString(leanSummary)
	}
	if len(in.Conflicts) > 0 {
		off := map[string]bool{}
		for _, c := range in.Conflicts {
			off[c] = false
		}
		// json.Marshal sorts map keys, so the snippet is stable.
		snippet, _ := json.Marshal(map[string]any{"enabledPlugins": off})
		fmt.Fprintf(&b, "\nA plugin that overlaps acta is enabled here: %s.\n", strings.Join(in.Conflicts, ", "))
		b.WriteString("Tell the user once, at the start: two plugins that do the same job pull the agent two ways. To turn it off in this repo only, add this to .claude/settings.local.json:\n")
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

// LoadKnown reads the list of plugins that overlap acta: one name per line,
// blank lines and # comments skipped. No file means no list.
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
