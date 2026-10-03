package plugincheck

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
	"gopkg.in/yaml.v3"
)

// styleWordCap is the most words the style body may hold. Every reply pays for
// these words, so the cap stops the text from growing a little at a time.
const styleWordCap = 300

// styleHeadings are the two blocks of the style body, word for word. Other
// text points at these titles to say which block applies.
var styleHeadings = []string{
	"## Every reply",
	"## ADHD reader (only when the acta session note says Style: adhd)",
}

// styleFrontmatter holds the keys of the style file that Claude Code acts on.
type styleFrontmatter struct {
	Name  string `yaml:"name"`
	Keep  bool   `yaml:"keep-coding-instructions"`
	Force bool   `yaml:"force-for-plugin"`
}

// styleParts reads plugin/output-styles/acta.md and splits it into its
// frontmatter keys and the body under them.
func styleParts(t *testing.T) (styleFrontmatter, string) {
	t.Helper()
	text := strings.ReplaceAll(readFile(t, "output-styles", "acta.md"), "\r\n", "\n")
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		t.Fatal("output-styles/acta.md has no frontmatter block")
	}
	head, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		t.Fatal("output-styles/acta.md frontmatter is never closed")
	}
	var fm styleFrontmatter
	if err := yaml.Unmarshal([]byte(head), &fm); err != nil {
		t.Fatal(err)
	}
	return fm, body
}

// TestStyleFrontmatter checks the three keys. Without force-for-plugin the
// style is not on when the plugin is on. Without keep-coding-instructions
// Claude Code drops its own coding rules.
func TestStyleFrontmatter(t *testing.T) {
	fm, _ := styleParts(t)
	if fm.Name != "acta" {
		t.Errorf("style name is %q, want acta", fm.Name)
	}
	if !fm.Keep {
		t.Error("keep-coding-instructions must be true, or Claude Code drops its coding rules")
	}
	if !fm.Force {
		t.Error("force-for-plugin must be true, or the style is not on with the plugin")
	}
}

// TestStyleHeadings checks that both blocks are there, each as a whole line.
func TestStyleHeadings(t *testing.T) {
	_, body := styleParts(t)
	lines := strings.Split(body, "\n")
	for _, h := range styleHeadings {
		if !slices.Contains(lines, h) {
			t.Errorf("style has no heading line %q", h)
		}
	}
}

// TestStyleBodyWordCap keeps the body short, because every reply pays for it.
func TestStyleBodyWordCap(t *testing.T) {
	_, body := styleParts(t)
	if n := len(strings.Fields(body)); n > styleWordCap {
		t.Errorf("style body is %d words, cap %d", n, styleWordCap)
	}
}

// adhdPhrase reads the words the session note must hold for the ADHD block to
// apply: the text after "says" in the ADHD heading, up to its closing bracket.
func adhdPhrase(t *testing.T, body string) string {
	t.Helper()
	for _, ln := range strings.Split(body, "\n") {
		if !strings.HasPrefix(ln, "## ADHD") {
			continue
		}
		_, rest, ok := strings.Cut(ln, "says ")
		phrase, _, closed := strings.Cut(rest, ")")
		if ok && closed && phrase != "" {
			return phrase
		}
	}
	t.Fatal("the ADHD heading does not say which words the session note holds")
	return ""
}

// TestStyleHeadingMatchesHookLine keeps the hook and the style file in step.
// The ADHD block applies only when the session note says the words in its
// heading. If the hook line and the heading drift apart, the block silently
// never applies, or applies to everyone.
func TestStyleHeadingMatchesHookLine(t *testing.T) {
	_, body := styleParts(t)
	phrase := adhdPhrase(t, body)
	adhd := config.User{ChatLanguage: "English", Style: "adhd", RepoLanguage: "English"}
	plain := adhd
	plain.Style = "plain"
	// First run and a broken config both fall back to adhd, so they must
	// print the line the heading names too.
	for name, in := range map[string]hook.Input{
		"adhd":      {Voice: adhd, VoiceExists: true},
		"first run": {Voice: config.UserDefault()},
		"broken":    {Voice: config.UserDefault(), VoiceExists: true, VoiceErr: errors.New("x")},
	} {
		out := hook.SessionStart(in)
		if n := strings.Count(out, phrase); n != 1 {
			t.Errorf("%s: the session text holds %q %d times, want once", name, phrase, n)
		}
		if !strings.Contains(out, "\n- "+phrase+".\n") {
			t.Errorf("%s: the session text has no line %q", name, "- "+phrase+".")
		}
	}
	out := hook.SessionStart(hook.Input{Voice: plain, VoiceExists: true})
	if !strings.Contains(out, "\n- Style: plain.\n") || strings.Contains(out, phrase) {
		t.Errorf("plain: the session text must say Style: plain and never %q:\n%s", phrase, out)
	}
}
