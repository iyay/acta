package hook

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/voice"
)

var update = flag.Bool("update", false, "rewrite plugin/hooks/default-rules.md")

func korean() Input {
	return Input{Voice: voice.Voice{ChatLanguage: "Korean", Style: "adhd", RepoLanguage: "English"}, VoiceExists: true}
}

func TestSessionStartListsSkillsAndRules(t *testing.T) {
	out := SessionStart(korean())
	for _, s := range Skills {
		if !strings.Contains(out, "- acta:"+s.Name+": ") {
			t.Errorf("skill %s missing", s.Name)
		}
	}
	for _, want := range []string{
		"No code before an approved design and an approved plan.",
		"Never push.",
		"CLAUDE.md or AGENTS.md",
		"Write every chat message to the user in Korean.",
		"(code, comments, commits, specs, plans) in English.",
		"full, clear sentences",
		"Style (ADHD reader):",
		`- acta:scratch: raw ideas ("catet", "nanti", side ideas); file with acta scratch new, never memory`,
		"- acta:setup: first-run setup and later changes: doctor, voice, build executor, subagent models, CLAUDE.md block",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	if len(Skills) != 12 {
		t.Fatalf("%d skills, want 12", len(Skills))
	}
}

func TestSessionStartPlainAndTone(t *testing.T) {
	in := korean()
	in.Voice.Style = "plain"
	in.Voice.Tone = "Casual.\nNo jokes."
	out := SessionStart(in)
	if strings.Contains(out, "Style (ADHD reader):") {
		t.Error("plain style still prints the ADHD rules")
	}
	if !strings.Contains(out, "- Tone, in the user's words:\n  Casual.\n  No jokes.\n") {
		t.Errorf("tone block wrong:\n%s", out)
	}
}

func TestSessionStartFirstRun(t *testing.T) {
	out := SessionStart(Input{Voice: voice.Default()})
	for _, want := range []string{"Voice: not set up yet.", "/acta:setup", "acta doctor", "Style (ADHD reader):",
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
	out := SessionStart(Input{Voice: voice.Default(), VoiceExists: true, VoiceErr: errors.New("style must be adhd")})
	if !strings.Contains(out, "could not be read (style must be adhd)") || !strings.Contains(out, "Style (ADHD reader):") {
		t.Errorf("broken voice text wrong:\n%s", out)
	}
}

func TestSessionStartConflicts(t *testing.T) {
	in := korean()
	in.Conflicts = []string{"superpowers@superpowers-dev"}
	out := SessionStart(in)
	if !strings.Contains(out, "Another workflow plugin is enabled here: superpowers@superpowers-dev.") ||
		!strings.Contains(out, `{"enabledPlugins":{"superpowers@superpowers-dev":false}}`) {
		t.Errorf("conflict text wrong:\n%s", out)
	}
}

func TestSessionStartStaysShort(t *testing.T) {
	worst := korean()
	worst.Voice.Tone = strings.TrimSpace(strings.Repeat("A tone line.\n", 8))
	worst.Conflicts = []string{"superpowers@a", "gstack@b", "x@mattpocock"}
	for name, in := range map[string]Input{
		"worst":     worst,
		"first run": {Voice: voice.Default(), Conflicts: worst.Conflicts},
		"broken":    {Voice: voice.Default(), VoiceExists: true, VoiceErr: errors.New("x"), Conflicts: worst.Conflicts},
	} {
		if n := strings.Count(SessionStart(in), "\n"); n > 60 {
			t.Errorf("%s: %d lines, cap is 60", name, n)
		}
	}
}

// An agent can answer a second big brainstorm from this rule alone,
// without loading the skill. So the rule must name the real choices.
func TestSessionStartNamesSecondBrainstormChoices(t *testing.T) {
	for name, in := range map[string]Input{
		"normal":    korean(),
		"first run": {Voice: voice.Default()},
		"broken":    {Voice: voice.Default(), VoiceExists: true, VoiceErr: errors.New("x")},
	} {
		out := SessionStart(in)
		for _, want := range []string{
			"acta:brainstorm",
			"claude --bg 'brainstorm SCRATCH-n'",
			"new session",
			"HERDR_ENV=1",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: session start missing %q", name, want)
			}
		}
	}
}

func TestPrompt(t *testing.T) {
	cases := map[string]struct {
		in   Input
		want string
	}{
		"set":     {korean(), "acta voice: reply in Korean, adhd style."},
		"missing": {Input{Voice: voice.Default()}, "acta voice: not set up yet; reply in English, adhd style, and run /acta:setup once (see the session rules)."},
		"broken":  {Input{Voice: voice.Default(), VoiceExists: true, VoiceErr: errors.New("x")}, "acta voice: the voice file could not be read; reply in English, adhd style."},
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
	want := SessionStart(Input{Voice: voice.Default(), VoiceExists: true})
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
	mappings := []string{
		"brainstorming→acta:brainstorm", "writing-plans→acta:plan",
		"subagent-driven-development→acta:build", "using-git-worktrees→acta:build",
		"test-driven-development→acta:tdd", "systematic-debugging→acta:debug",
		"requesting-code-review→acta:review", "receiving-code-review→acta:review",
		"verification-before-completion→acta:land", "finishing-a-development-branch→acta:land",
	}
	for name, in := range map[string]Input{
		"voice set": korean(),
		"conflicts": worst,
		"first run": {Voice: voice.Default()},
		"broken":    {Voice: voice.Default(), VoiceExists: true, VoiceErr: errors.New("x")},
	} {
		out := SessionStart(in)
		for _, want := range append([]string{
			"load its acta skill with the Skill tool", "This list is only an index",
			"skill from the superpowers plugin that is not installed",
		}, mappings...) {
			if !strings.Contains(out, want) {
				t.Errorf("%s: missing %q", name, want)
			}
		}
	}
}
