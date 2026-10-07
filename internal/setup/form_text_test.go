package setup_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/setup"
)

// pinHome keeps the test away from the real home and config.
func pinHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", home+"/config.yaml")
	t.Setenv("TMPDIR", home)
}

// TestFormOptionNotes pins the note on every option of every select, with
// and without herdr.
func TestFormOptionNotes(t *testing.T) {
	pinHome(t)
	want := map[int][]string{
		1: {"adhd: short, one next step at the end", "plain: normal paragraphs"},
		4: {"subagent: helper agents in the same session", "inline: the main agent writes the code"},
		5: {"default: your own config decides", "split: sonnet writes code, the rest use a stronger model"},
		6: {"full: real code in every step, waits for your yes", "minimal: short steps, build starts at once"},
		7: {"one: one question at a time", "probe: a batch, each with the agent's pick"},
	}
	for _, herdr := range []string{"", "1"} {
		t.Setenv("HERDR_ENV", herdr)
		content := formGroupContent(t, setup.Env{Current: config.User{}})
		for i, notes := range want {
			for _, n := range notes {
				if !strings.Contains(content[i], n) {
					t.Errorf("herdr=%q group %d: missing option %q in %q", herdr, i+1, n, content[i])
				}
			}
		}
		dispatch := "dispatch: an omp agent in its own herdr tab"
		if got := strings.Contains(content[4], dispatch); got != (herdr == "1") {
			t.Errorf("herdr=%q: dispatch note present = %v", herdr, got)
		}
	}
}

// TestFormNoFirstPerson checks no wizard text says I, me or my.
func TestFormNoFirstPerson(t *testing.T) {
	pinHome(t)
	re := regexp.MustCompile(`(?i)\b(i|me|my)\b`)
	t.Setenv("HERDR_ENV", "1")
	env := setup.Env{Harnesses: []string{"claude"}, Current: config.User{}}
	all := append(formGroupText(t, env), formGroupContent(t, env)...)
	for _, txt := range all {
		if m := re.FindString(txt); m != "" {
			t.Errorf("first person word %q in %q", m, txt)
		}
	}
}

// TestFormAnsweredLine checks the answered line shows the bare value, not
// the option note.
func TestFormAnsweredLine(t *testing.T) {
	pinHome(t)
	hold := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	defer lipgloss.SetColorProfile(hold)

	got := setup.AnsweredLines(setup.Env{Current: config.User{}})
	if got[0] != "◇  Chat language\n│  English\n│\n" {
		t.Errorf("chat language line = %q", got[0])
	}
	if got[1] != "◇  Reply style\n│  adhd\n│\n" {
		t.Errorf("reply style line = %q", got[1])
	}
	if got[4] != "◇  Build executor\n│  subagent\n│\n" {
		t.Errorf("executor line = %q", got[4])
	}
}
