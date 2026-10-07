package setup_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/setup"
)

// formGroupText renders each wizard group's header, so the test reads the
// title, count and description the user would see. It never runs the form:
// nobody is asked anything and no harness command runs.
func formGroupText(t *testing.T, e setup.Env) []string {
	t.Helper()
	hold := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	defer lipgloss.SetColorProfile(hold)

	var out []string
	for _, g := range setup.FormGroups(e) {
		out = append(out, g.Header())
	}
	return out
}

// formGroupContent renders each group's fields plainly, so the test reads
// the options and harness names on each screen.
func formGroupContent(t *testing.T, e setup.Env) []string {
	t.Helper()
	hold := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	defer lipgloss.SetColorProfile(hold)

	var out []string
	for _, g := range setup.FormGroups(e) {
		out = append(out, g.Content())
	}
	return out
}

func TestFormGroups(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", home+"/config.yaml")
	t.Setenv("TMPDIR", home)

	env := setup.Env{
		Harnesses: []string{"claude", "omp"},
		Current:   config.User{},
	}
	groups := formGroupText(t, env)

	// Eight question screens, then one screen with a yes/no per harness.
	if len(groups) != 9 {
		t.Fatalf("got %d groups, want 9", len(groups))
	}

	// Fixed order: every screen says which question it is, with its own
	// step count and a one-line plain description.
	want := []struct {
		count string
		desc  string
	}{
		{"1/9", "The language I use when I talk with you."},
		{"2/9", "Short replies for speed, or full sentences."},
		{"3/9", "Anything about tone, in your own words. Optional."},
		{"4/9", "The language for files written to the repo."},
		{"5/9", "Who writes the code when a plan runs."},
		{"6/9", "Who picks the model for background work."},
		{"7/9", "How much a plan spells out before it runs."},
		{"8/9", "How I ask you things while working."},
		{"9/9", "Install the acta plugin into each tool found."},
	}
	for i, w := range want {
		if !strings.Contains(groups[i], "acta setup") {
			t.Errorf("group %d: missing title %q", i+1, "acta setup")
		}
		if !strings.Contains(groups[i], w.count) {
			t.Errorf("group %d: missing step count %q in %q", i+1, w.count, groups[i])
		}
		if !strings.Contains(groups[i], w.desc) {
			t.Errorf("group %d: missing description %q", i+1, w.desc)
		}
	}

	// The harness screen lists every harness found.
	content := formGroupContent(t, env)
	if !strings.Contains(content[8], "claude") || !strings.Contains(content[8], "omp") {
		t.Errorf("harness group lists every harness, got %q", content[8])
	}
}

func TestFormGroupsNoHarness(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", home+"/config.yaml")
	t.Setenv("TMPDIR", home)

	env := setup.Env{Current: config.User{}}
	groups := formGroupText(t, env)
	if len(groups) != 8 {
		t.Fatalf("got %d groups, want 8", len(groups))
	}
	content := formGroupContent(t, env)
	if strings.Contains(content[7], "claude") || strings.Contains(content[7], "omp") {
		t.Errorf("last group must not name a harness, got %q", content[7])
	}
}

func TestFormUsesSetupTheme(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", home+"/config.yaml")
	t.Setenv("TMPDIR", home)

	// The setup look marks the focused option with ›, while plain huh marks
	// it with >. The style question must show ›, or the form ignores
	// Theme().
	env := setup.Env{Current: config.User{}}
	got := formGroupContent(t, env)[1]
	if !strings.Contains(got, "›") {
		t.Fatalf("style question must mark the focused option with ›, got %q", got)
	}

	plain := huh.NewGroup(
		huh.NewSelect[string]().Title("Style: adhd or plain?").
			Options(huh.NewOptions("adhd", "plain")...).
			Value(new(string)),
	).WithTheme(huh.ThemeBase())
	hold := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	defer lipgloss.SetColorProfile(hold)
	if strings.Contains(plain.Content(), "›") {
		t.Fatalf("plain huh theme must not use ›, got %q", plain.Content())
	}
}

func TestFormExecutorOptionsFollowHerdr(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", home+"/config.yaml")
	t.Setenv("TMPDIR", home)

	// Outside a herdr pane only subagent and inline are offered; inside one
	// (HERDR_ENV=1) dispatch joins them.
	t.Setenv("HERDR_ENV", "")
	plain := formGroupContent(t, setup.Env{Current: config.User{}})[4]
	if strings.Contains(plain, "dispatch") {
		t.Errorf("executor screen must not offer dispatch outside herdr, got %q", plain)
	}

	t.Setenv("HERDR_ENV", "1")
	herdr := formGroupContent(t, setup.Env{Current: config.User{}})[4]
	if !strings.Contains(herdr, "dispatch") {
		t.Errorf("executor screen must offer dispatch inside herdr, got %q", herdr)
	}
}
