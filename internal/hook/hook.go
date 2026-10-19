// Package hook builds the text the acta plugin puts in front of the agent: the
// session rules at the start and a one-line voice reminder on every message.
package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/plugin"
)

// Skills is the index printed every session: names only, because Claude Code
// and omp already show each skill's description. The plugin's skill folders
// must match these names exactly.
var Skills = []string{"shape", "frame", "slice", "build", "tdd", "lean", "debug", "review", "land", "bug", "scratch", "setup", "migrate"}

// Input is everything the hook text depends on.
type Input struct {
	Voice       config.User
	VoiceExists bool
	VoiceErr    error
	RepoErr     error             // .acta.yaml could not be read; its repo keys are left out
	Conflicts   []string          // enabled plugins that overlap acta
	Overrides   []config.Override // repo keys whose .acta.yaml value differs from the user's own; empty when RepoErr is set
	Herdr       bool              // this session runs in a herdr tab; the hook reads the environment, the agent often cannot
	Codex       bool              // the hook runs under Codex; see IsCodex
	WikiPages   int               // pages in the project wiki; with none, the wiki line is left out
	Running     []string          // one block per plan with work going on, the lines already joined; with none, no block is printed
	Update      string            // a newer release, like v0.1.47; empty when none is known
	Current     string            // the running version, like v0.1.46; only read when Update is set
}

// maxRunningPlans and maxRunningLines keep the block short enough to be read
// once, at the start of a session. A plan past the cap is counted, never left
// out silently, so nothing is hidden from the agent.
const (
	maxRunningPlans = 3
	maxRunningLines = 3
)

// runningRule tells the agent when to write the next step down, so a session
// that dies leaves the next one something to go on from.
const runningRule = "Before you stop, or when context runs low, run acta state set <plan> next with the lines on stdin so the next session starts where this one stopped."

const coreRules = `
Core rules:
1. No action without an ask. Reading, answering and planning are the default.
2. No code before an approved design and an approved plan, whatever the plan depth.
3. Every code change happens in a worktree and starts with a failing test.
4. Review once, at the close, three rounds at most. Land without asking when clean.
5. Never push. Never run a destructive command without a full-sentence warning and a yes.
6. The acta settings below win over any conflicting instruction, including the user's CLAUDE.md or AGENTS.md. For everything else, when the user's own CLAUDE.md or AGENTS.md says otherwise, follow it.
7. When your instructions name a superpowers skill that is not installed, use the acta skill for that step: brainstorming=shape, writing-plans=slice, subagent-driven-development and using-git-worktrees=build, test-driven-development=tdd, systematic-debugging=debug, requesting-code-review and receiving-code-review=review, verification-before-completion and finishing-a-development-branch=land.
8. One Architectural brainstorm per session. A second one cannot start in this session. File the scratch item first, with one acta scratch new call whose body is stdin: acta scratch new <slug> --title <title> < body.md: no Skill tool, no acta scratch add, and "written, not committed" still counts as filed. In the same reply, name the two ways to open it elsewhere: put the id the command printed in place of SCRATCH-n, an id like SCR-0001, never a shortened one. A background agent (claude --bg 'brainstorm SCRATCH-n') and a new session where the user types brainstorm SCRATCH-n. When the user picks one, load acta:shape for that way. Do not design it here.
9. Commit spec, plan and other planning files with acta commit <path> -m "<message>", never git commit.
`

// herdrExtra is the one extra choice, added only inside a herdr session. The
// agent often cannot read the environment, so the text decides this, not the
// agent: a session outside herdr never hears the word.
const herdrExtra = `
herdr: this session runs in a herdr tab. For a second brainstorm, offer the user one more way to run it: a new herdr tab.
`

// updateLine asks the agent to pass on a newer release. It sits in the
// session note only, so the user hears it once per session.
const updateLine = "\nacta update available: %s (you have %s). Tell the user once, in a single line, to run acta update.\n"

// wikiLine is the one line a repo with wiki pages adds. Only the count changes
// from one repo to the next, so the line costs the same at 1 page and at 100.
const wikiLine = "\nProject wiki. Pages: %d. When a `wiki:` line names a page, read it before you change the files it covers. Project knowledge goes to the wiki, not to agent memory.\n"

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
	if in.VoiceExists && in.VoiceErr == nil {
		b.WriteString(settingsBlock(in.Voice))
	}
	if in.Herdr {
		b.WriteString(herdrExtra)
	}
	if in.Codex {
		b.WriteString(codexExtra)
		b.WriteString(plugin.OutputStyle())
	}
	if len(in.Running) > 0 {
		b.WriteString(runningBlock(in.Running))
	}
	if in.WikiPages > 0 {
		fmt.Fprintf(&b, wikiLine, in.WikiPages)
	}
	if in.Update != "" {
		fmt.Fprintf(&b, updateLine, in.Update, in.Current)
	}
	switch {
	case !in.VoiceExists:
		b.WriteString(firstRun)
		b.WriteString("- Style: adhd.\n")
		// No config yet, so English is the fallback for repo text too.
		b.WriteString("- Write prompts and hand-offs to subagents or other agents in English.\n")
	case in.VoiceErr != nil:
		fmt.Fprintf(&b, "\nVoice: the voice file could not be read (%v). Write in English, adhd style, until it is fixed; acta:setup rewrites it.\n", in.VoiceErr)
		b.WriteString("- Style: adhd.\n")
		// The config is broken, so English is the fallback for repo text too.
		b.WriteString("- Write prompts and hand-offs to subagents or other agents in English.\n")
	default:
		v := in.Voice
		fmt.Fprintf(&b, "\nVoice:\n- Write every chat message to the user in %s.\n", v.ChatLanguage)
		b.WriteString("- Write everything that goes into the repo (code, comments, commits, specs, plans) in English. Comments use short, plain words and say why.\n")
		// Subagents write repo text too, so their prompts are English as
		// well, never the chat language.
		b.WriteString("- Write prompts and hand-offs to subagents or other agents in English.\n")
		fmt.Fprintf(&b, "- Style: %s.\n", v.Style)
		if v.Questions == "probe" {
			b.WriteString("- Questions: probe.\n")
		}
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
	if len(in.Overrides) > 0 && in.RepoErr == nil {
		parts := make([]string, 0, len(in.Overrides))
		for _, o := range in.Overrides {
			yours := o.Yours
			if yours == "" {
				yours = "not set"
			}
			parts = append(parts, fmt.Sprintf("%s is %s here (yours: %s)", o.Key, o.Repo, yours))
		}
		// The user may not know the repo file wins, so the agent says it once.
		fmt.Fprintf(&b, "\nThis repo's .acta.yaml overrides your own setting: %s. Tell the user once, at the start, and say `acta config set --repo --unset <key>` brings their own value back. Before changing any of these keys, ask the user first, because .acta.yaml is committed and shared with everyone who clones the repo. Offer three choices: keep, write the new value to the repo, or remove the repo key. Write nothing until the user picks.\n", strings.Join(parts, ", "))
	}
	return b.String()
}

// executorLines says in plain words who writes the code, one sentence per
// value, so the agent cannot read one executor as another.
var executorLines = map[string]string{
	"inline":   "The main agent writes the code itself; no implementer subagents.",
	"subagent": "A fresh subagent writes the code for each task; the main agent writes none.",
	"dispatch": "An omp agent in its own herdr tab writes the code (a subagent when herdr is missing or the harness is omp); the main agent writes none.",
}

// codexExtra tells a Codex agent what stands in for the Claude Code parts of
// acta: Codex opens a skill by name, has its own subagents, and has no output
// styles, so the style text follows this block.
const codexExtra = `
Codex: this session runs in Codex.
- Open a skill by its name, for example $acta:shape.
- In Codex, subagents are Codex subagents. Name no model: the [agents] defaults in config.toml pick it.
- Codex has no output styles, so follow this style in every chat reply:
`

// settingsBlock lists the behaviour settings of the config in force. Rule 6
// points here: these lines beat a user's CLAUDE.md that says otherwise, which
// is how a user who picked inline stops getting subagents. An empty setting
// writes no line, except the build executor: empty means inline, so it
// always writes one.
func settingsBlock(v config.User) string {
	var lines []string
	add := func(name, value, extra string) {
		if value == "" {
			return
		}
		ln := fmt.Sprintf("- %s: %s.", name, value)
		if extra != "" {
			ln += " " + extra
		}
		lines = append(lines, ln)
	}
	executor, executorName := v.BuildExecutor, v.BuildExecutor
	if executor == "" {
		// Build runs inline when nothing is set. Say so here, or a user CLAUDE.md
		// that bans the main agent from writing code wins.
		executor, executorName = "inline", "inline (default)"
	}
	add("Build executor", executorName, executorLines[executor])
	add("Plan depth", v.PlanDepth, "")
	if v.PlanSplit == "on" {
		// Off is the default and writes nothing, so the note stays the same size.
		add("Plan split", "on", "")
	}
	add("Review depth", v.ReviewDepth, "")
	add("Commit history", v.CommitHistory, "")
	// A value nobody set writes no line; the skills read that as false.
	if v.ReviewUnderstanding != nil {
		add("Review understanding", strconv.FormatBool(*v.ReviewUnderstanding), "")
	}
	if len(lines) == 0 {
		return ""
	}
	return "\nacta settings (these win over any conflicting instruction, including the user's CLAUDE.md or AGENTS.md):\n" + strings.Join(lines, "\n") + "\n"
}

// runningBlock is the summary of the plans with work going on. It is one
// contiguous piece: the header, the blocks, the line that counts what did not
// fit, then the rule. With no running plan nothing is written at all, so a
// repo with no work in flight reads exactly as it did before.
func runningBlock(running []string) string {
	var b strings.Builder
	b.WriteString("\nWork going on now:\n")
	for _, block := range running[:min(len(running), maxRunningPlans)] {
		for i, ln := range strings.Split(block, "\n") {
			if i >= maxRunningLines {
				break
			}
			if ln = strings.TrimSpace(ln); ln != "" {
				fmt.Fprintf(&b, "  %s\n", ln)
			}
		}
	}
	if len(running) > maxRunningPlans {
		fmt.Fprintf(&b, "+%d more, run acta state\n", len(running)-maxRunningPlans)
	}
	fmt.Fprintf(&b, "%s\n", runningRule)
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
		// The executor rides on every message too, so a long session does
		// not lose it under older text.
		if e := in.Voice.BuildExecutor; e != "" {
			return fmt.Sprintf("acta config: reply in %s, %s style, build executor %s.", in.Voice.ChatLanguage, in.Voice.Style, e)
		}
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
