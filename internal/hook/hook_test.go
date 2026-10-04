package hook

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
)

var update = flag.Bool("update", false, "rewrite plugin/hooks/default-rules.md")

func korean() Input {
	return Input{Voice: config.User{ChatLanguage: "Korean", Style: "adhd", RepoLanguage: "English"}, VoiceExists: true}
}

// skillNames are the skills the index must name. A skill added later joins the
// index too, so this list only has to be a part of it.
var skillNames = []string{"shape", "slice", "build", "tdd", "debug", "review", "land", "bug", "scratch", "setup", "migrate"}

// leanHeading is the first line of the lean summary.
const leanHeading = "Lean coding guide (full text: acta:lean):"

// indexNames reads the skill names from the first line of the text.
func indexNames(t *testing.T, out string) []string {
	t.Helper()
	first, _, _ := strings.Cut(out, "\n")
	_, list, ok := strings.Cut(first, "follow it: ")
	list, tail, ok2 := strings.Cut(list, ". The rules live in the skills.")
	if !ok || !ok2 || tail != "" {
		t.Fatalf("the first line is not the one-line skill index: %q", first)
	}
	return strings.Split(list, ", ")
}

// rule7 reads rule 7 into a map from each superpowers name to its acta skill.
func rule7(t *testing.T, out string) map[string]string {
	t.Helper()
	_, rest, ok := strings.Cut(out, "\n7. ")
	line, _, _ := strings.Cut(rest, "\n")
	_, list, ok2 := strings.Cut(line, "use the acta skill for that step: ")
	if !ok || !ok2 {
		t.Fatalf("rule 7 is missing or has a new shape:\n%s", out)
	}
	got := map[string]string{}
	for _, group := range strings.Split(strings.TrimSuffix(list, "."), ", ") {
		names, skill, _ := strings.Cut(group, "=")
		for _, name := range strings.Split(names, " and ") {
			got[name] = skill
		}
	}
	return got
}

func TestSessionStartListsSkillsAndRules(t *testing.T) {
	out := SessionStart(korean())
	names := indexNames(t, out)
	for _, want := range skillNames {
		if !slices.Contains(names, want) {
			t.Errorf("skill %s missing from the index", want)
		}
	}
	// The index is one line of names. A line per skill with its hint is the
	// old shape: the skill list of the harness already shows the hints.
	if strings.Contains(out, "\n- acta:") {
		t.Error("the index still has a line per skill")
	}
	for _, want := range []string{
		"No code before an approved design and an approved plan. A plan with depth: minimal in its frontmatter needs no plan yes.",
		"Never push.",
		"CLAUDE.md or AGENTS.md",
		"Write every chat message to the user in Korean.",
		"(code, comments, commits, specs, plans) in English.",
		"- Style: adhd.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestSessionStartPlainAndTone(t *testing.T) {
	in := korean()
	in.Voice.Style = "plain"
	in.Voice.Tone = "Casual.\nNo jokes."
	out := SessionStart(in)
	if !strings.Contains(out, "\n- Style: plain.\n") || strings.Contains(out, "Style: adhd") {
		t.Errorf("plain style line wrong:\n%s", out)
	}
	if !strings.Contains(out, "- Tone, in the user's words:\n  Casual.\n  No jokes.\n") {
		t.Errorf("tone block wrong:\n%s", out)
	}
}

func TestSessionStartFirstRun(t *testing.T) {
	out := SessionStart(Input{Voice: config.UserDefault()})
	for _, want := range []string{"Voice: not set up yet.", "/acta:setup", "acta doctor", "\n- Style: adhd.\n",
		"already names a chat language or style", "or the language CLAUDE.md names"} {
		if !strings.Contains(out, want) {
			t.Errorf("first run missing %q", want)
		}
	}
	if strings.Contains(out, "1. Which language should chat use?") {
		t.Error("first run still asks the three voice questions instead of pointing at /acta:setup")
	}
}

func TestSessionStartBrokenVoice(t *testing.T) {
	out := SessionStart(Input{Voice: config.UserDefault(), VoiceExists: true, VoiceErr: errors.New("style must be adhd")})
	if !strings.Contains(out, "could not be read (style must be adhd)") || !strings.Contains(out, "\n- Style: adhd.\n") {
		t.Errorf("broken voice text wrong:\n%s", out)
	}
}

// The session note tells shape that the user picked probe, and only that value
// does. Any other voice path must stay as it was, byte for byte.
func TestSessionStartQuestionsProbeLine(t *testing.T) {
	probe := korean()
	probe.Voice.Questions = "probe"
	out := SessionStart(probe)
	if !strings.Contains(out, "\n- Style: adhd.\n- Questions: probe.\n") {
		t.Errorf("probe line must sit right after the style line:\n%s", out)
	}
	one, broken := korean(), Input{Voice: config.User{Questions: "probe"}, VoiceExists: true, VoiceErr: errors.New("x")}
	one.Voice.Questions = "one"
	for name, in := range map[string]Input{
		"one":           one,
		"key absent":    korean(),
		"no voice file": {Voice: config.User{Questions: "probe"}},
		"broken file":   broken,
	} {
		if got := SessionStart(in); strings.Contains(got, "Questions: probe") {
			t.Errorf("%s: the probe line must be absent:\n%s", name, got)
		}
	}
}

// Not probe adds no text at all, so the default session start keeps its size.
func TestSessionStartQuestionsAddsNoTextWhenNotProbe(t *testing.T) {
	unset, one := korean(), korean()
	one.Voice.Questions = "one"
	if got, want := len(SessionStart(one)), len(SessionStart(unset)); got != want {
		t.Errorf("questions: one made the session start %d bytes, unset is %d", got, want)
	}
	if got := SessionStart(Input{Voice: config.UserDefault(), VoiceExists: true}); strings.Contains(got, "Questions:") {
		t.Errorf("the default session start now mentions Questions:\n%s", got)
	}
}

func TestSessionStartConflicts(t *testing.T) {
	in := korean()
	in.Conflicts = []string{"superpowers@superpowers-dev"}
	out := SessionStart(in)
	if !strings.Contains(out, "A plugin that overlaps acta is enabled here: superpowers@superpowers-dev.") ||
		!strings.Contains(out, "two plugins that do the same job pull the agent two ways") ||
		!strings.Contains(out, `{"enabledPlugins":{"superpowers@superpowers-dev":false}}`) {
		t.Errorf("conflict text wrong:\n%s", out)
	}
	for _, old := range []string{"Another workflow plugin", "two workflow plugins"} {
		if strings.Contains(out, old) {
			t.Errorf("conflict text still says %q", old)
		}
	}
}

// Every input shape gets the same checks: one style line with the right word,
// no ADHD block (the output style holds it), no destructive-warning line (the
// output style holds that too), and the lean summary unless coding_guide is off.
func TestSessionStartEveryShape(t *testing.T) {
	plain, lean, off, herdr, conflicts := korean(), korean(), korean(), korean(), korean()
	plain.Voice.Style = "plain"
	lean.Voice.CodingGuide = "lean"
	off.Voice.CodingGuide = "off"
	herdr.Herdr = true
	conflicts.Conflicts = []string{"superpowers@superpowers-dev"}
	cases := map[string]struct {
		in    Input
		style string // the word after "- Style: "
		lean  bool   // whether the lean summary is printed
	}{
		"first run": {Input{Voice: config.UserDefault()}, "adhd", true},
		"broken":    {Input{Voice: config.UserDefault(), VoiceExists: true, VoiceErr: errors.New("x")}, "adhd", true},
		"adhd":      {korean(), "adhd", true},
		"plain":     {plain, "plain", true},
		"lean":      {lean, "adhd", true},
		"off":       {off, "adhd", false},
		"herdr":     {herdr, "adhd", true},
		"conflicts": {conflicts, "adhd", true},
	}
	for name, c := range cases {
		out := SessionStart(c.in)
		if n := strings.Count(out, "\n- Style: "); n != 1 {
			t.Errorf("%s: %d style lines, want 1", name, n)
		}
		other := map[string]string{"adhd": "plain", "plain": "adhd"}[c.style]
		if !strings.Contains(out, "\n- Style: "+c.style+".\n") || strings.Contains(out, "Style: "+other) {
			t.Errorf("%s: the style line is not %q:\n%s", name, c.style, out)
		}
		for _, gone := range []string{"Style (ADHD reader)", "The first line is the answer or the next action",
			"Warnings before a destructive command", "full, clear sentences"} {
			if strings.Contains(out, gone) {
				t.Errorf("%s: still prints %q", name, gone)
			}
		}
		if got := strings.Contains(out, leanHeading); got != c.lean {
			t.Errorf("%s: lean summary printed = %v, want %v", name, got, c.lean)
		}
		if !c.lean && strings.Contains(out, "acta:lean") {
			t.Errorf("%s: coding_guide is off but the text still names acta:lean", name)
		}
	}
}

// The lean summary sits after the voice lines and before the plugin note.
func TestSessionStartLeanSummary(t *testing.T) {
	in := korean()
	in.Conflicts = []string{"superpowers@superpowers-dev"}
	out := SessionStart(in)
	voice, lean, note := strings.Index(out, "\nVoice:\n"), strings.Index(out, leanHeading), strings.Index(out, "A plugin that overlaps acta")
	if voice < 0 || lean < 0 || note < 0 || !(voice < lean && lean < note) {
		t.Fatalf("want voice < lean summary < plugin note, got %d %d %d:\n%s", voice, lean, note, out)
	}
	for _, want := range []string{
		"- Understand the task and the code it touches before choosing.\n",
		"- Then take the first rung that works: skip it, reuse code here, stdlib, a native platform feature, an installed dependency, the fewest lines.\n",
		"- No abstraction with one user, no config for a fixed value, no scaffolding for later.\n",
		"- Fix a bug where every caller passes through, not only the reported path.\n",
		"- Never cut checks at trust boundaries, error handling that prevents data loss, security or accessibility.\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("lean summary missing %q", want)
		}
	}
}

func TestSessionStartStaysShort(t *testing.T) {
	worst := korean()
	// The worst case is everything at once: a long tone, three clashing
	// plugins and the herdr line. Leave one out and the cap stops proving
	// the cap.
	worst.Voice.Tone = strings.TrimSpace(strings.Repeat("A tone line.\n", 8))
	worst.Conflicts = []string{"superpowers@a", "gstack@b", "x@mattpocock"}
	worst.Herdr = true // the herdr line counts against the cap too
	for name, in := range map[string]Input{
		"worst":     worst,
		"first run": {Voice: config.UserDefault(), Conflicts: worst.Conflicts},
		"broken":    {Voice: config.UserDefault(), VoiceExists: true, VoiceErr: errors.New("x"), Conflicts: worst.Conflicts},
	} {
		if n := strings.Count(SessionStart(in), "\n"); n > 60 {
			t.Errorf("%s: %d lines, cap is 60", name, n)
		}
	}
}

// An agent can answer a second big brainstorm from this rule alone, so the
// rule must name the real choices. Whether a herdr tab is one of them is the
// hook's call, not the agent's: the hook reads the environment, the agent
// often cannot, so a session outside herdr must say nothing about herdr.
func TestSessionStartNamesSecondBrainstormChoices(t *testing.T) {
	for name, in := range map[string]Input{
		"normal":    korean(),
		"first run": {Voice: config.UserDefault()},
		"broken":    {Voice: config.UserDefault(), VoiceExists: true, VoiceErr: errors.New("x")},
	} {
		out := SessionStart(in)
		for _, want := range []string{
			"acta:shape",
			"claude --bg 'brainstorm SCRATCH-n'",
			"File the scratch item first",
			"one acta scratch new call",
			"whose body is stdin: acta scratch new <slug> --title <title> < body.md",
			"put the id the command printed in place of SCRATCH-n, an id like SCR-0001, never a shortened one",
			"written, not committed",
			"new session",
			"load acta:shape for that way",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: session start missing %q", name, want)
			}
		}
		if !slices.Contains(indexNames(t, out), "scratch") {
			t.Errorf("%s: the skill index does not name scratch", name)
		}
		// Case-insensitive, so this also rules out HERDR_ENV. Only the rules
		// and the voice lines count: the header above them is an index, not an
		// offer of a tab.
		rules := out[strings.Index(out, "Core rules:"):]
		if strings.Contains(strings.ToLower(rules), "herdr") {
			t.Errorf("%s: session start outside herdr names herdr:\n%s", name, rules)
		}
	}
	in := korean()
	in.Herdr = true
	out := SessionStart(in)
	if !strings.Contains(strings.ToLower(out), "herdr tab") {
		t.Errorf("herdr session start missing the herdr tab sentence:\n%s", out)
	}
}

// The rule must not forbid a choice another text offers: herdrExtra offers a
// new herdr tab inside a herdr session, and SKILL.md offers the same tab. So
// no clause may send the user away to another terminal or tab, and the skill
// may still be loaded for the way the user picked. Both herdr states are
// checked, because the words must not change with the environment.
func TestSessionStartNoChoiceIsForbidden(t *testing.T) {
	for _, herdr := range []bool{false, true} {
		in := korean()
		in.Herdr = herdr
		out := SessionStart(in)
		for _, banned := range []string{
			"another terminal or tab",
			"never offer to load acta:shape",
		} {
			if strings.Contains(out, banned) {
				t.Errorf("herdr=%v: session start forbids a choice another text offers, has %q", herdr, banned)
			}
		}
	}
}

func TestPrompt(t *testing.T) {
	cases := map[string]struct {
		in   Input
		want string
	}{
		"set":     {korean(), "acta config: reply in Korean, adhd style."},
		"missing": {Input{Voice: config.UserDefault()}, "acta config: not set up yet; reply in English, adhd style, and run /acta:setup once (see the session rules)."},
		"broken":  {Input{Voice: config.UserDefault(), VoiceExists: true, VoiceErr: errors.New("x")}, "acta config: the config file could not be read; reply in English, adhd style."},
	}
	for name, c := range cases {
		if got := Prompt(c.in); got != c.want {
			t.Errorf("%s: got %q", name, got)
		}
	}
}

func writeJSON(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEnabledPluginsMergesScopes(t *testing.T) {
	claude, repo := t.TempDir(), t.TempDir()
	writeJSON(t, filepath.Join(claude, "settings.json"),
		`{"enabledPlugins": {"superpowers@superpowers-dev": true, "caveman@caveman": true, "old@x": false}}`)
	writeJSON(t, filepath.Join(repo, ".claude", "settings.json"), `{"enabledPlugins": {"gstack@g": true}}`)
	writeJSON(t, filepath.Join(repo, ".claude", "settings.local.json"),
		`{"enabledPlugins": {"superpowers@superpowers-dev": false}}`)
	got := EnabledPlugins(claude, repo)
	if !reflect.DeepEqual(got, []string{"caveman@caveman", "gstack@g"}) {
		t.Fatalf("got %v", got)
	}
	writeJSON(t, filepath.Join(repo, ".claude", "settings.json"), `{not json`)
	if got := EnabledPlugins(claude, repo); !reflect.DeepEqual(got, []string{"caveman@caveman"}) {
		t.Fatalf("broken project file should be skipped, got %v", got)
	}
	if got := EnabledPlugins(t.TempDir(), t.TempDir()); len(got) != 0 {
		t.Fatalf("no files: got %v", got)
	}
}

func TestLoadKnownAndConflicts(t *testing.T) {
	p := filepath.Join(t.TempDir(), "known.txt")
	writeJSON(t, p, "# comment\n\nsuperpowers\n  GStack  \nmattpocock\n")
	known := LoadKnown(p)
	if !reflect.DeepEqual(known, []string{"superpowers", "GStack", "mattpocock"}) {
		t.Fatalf("known = %v", known)
	}
	if LoadKnown("") != nil || LoadKnown(filepath.Join(t.TempDir(), "none")) != nil {
		t.Fatal("missing file should give nil")
	}
	enabled := []string{"superpowers@superpowers-dev", "gstack@x", "skills@mattpocock", "caveman@caveman", "acta@acta-local", "noat"}
	got := Conflicts(enabled, known)
	if !reflect.DeepEqual(got, []string{"superpowers@superpowers-dev", "gstack@x", "skills@mattpocock"}) {
		t.Fatalf("conflicts = %v", got)
	}
	if got := Conflicts([]string{"acta@acta-local"}, []string{"acta", "acta-local"}); len(got) != 0 {
		t.Fatalf("acta must never match itself: %v", got)
	}
}

func TestClaudeDir(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", "/x/claude")
	if ClaudeDir() != "/x/claude" {
		t.Fatal("CLAUDE_CONFIG_DIR ignored")
	}
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	home, _ := os.UserHomeDir()
	if ClaudeDir() != filepath.Join(home, ".claude") {
		t.Fatal("default dir wrong")
	}
}

// The fallback file the hook script prints when acta is missing must match
// what acta would print for a default voice.
func TestDefaultRulesFile(t *testing.T) {
	path := filepath.Join("..", "..", "plugin", "hooks", "default-rules.md")
	want := SessionStart(Input{Voice: config.UserDefault(), VoiceExists: true})
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run: go test ./internal/hook -run TestDefaultRulesFile -update)", err)
	}
	if string(got) != want {
		t.Fatal("plugin/hooks/default-rules.md is stale; run: go test ./internal/hook -run TestDefaultRulesFile -update")
	}
}

func TestSessionStartMakesAgentsLoadSkills(t *testing.T) {
	worst := korean()
	worst.Conflicts = []string{"superpowers@a"}
	mappings := map[string]string{
		"brainstorming": "shape", "writing-plans": "slice",
		"subagent-driven-development": "build", "using-git-worktrees": "build",
		"test-driven-development": "tdd", "systematic-debugging": "debug",
		"requesting-code-review": "review", "receiving-code-review": "review",
		"verification-before-completion": "land", "finishing-a-development-branch": "land",
	}
	for name, in := range map[string]Input{
		"voice set": korean(),
		"conflicts": worst,
		"first run": {Voice: config.UserDefault()},
		"broken":    {Voice: config.UserDefault(), VoiceExists: true, VoiceErr: errors.New("x")},
	} {
		out := SessionStart(in)
		for _, want := range []string{
			"load the matching acta skill with the Skill tool and follow it: ", "The rules live in the skills.",
			"name a superpowers skill that is not installed",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: missing %q", name, want)
			}
		}
		// Each pair is checked on its own, so a name that falls out of the
		// shorter rule cannot hide behind the others.
		got := rule7(t, out)
		for superpowers, skill := range mappings {
			if got[superpowers] != skill {
				t.Errorf("%s: rule 7 maps %q to %q, want %q", name, superpowers, got[superpowers], skill)
			}
		}
		if len(got) != len(mappings) {
			t.Errorf("%s: rule 7 holds %d names, want %d", name, len(got), len(mappings))
		}
	}
}
