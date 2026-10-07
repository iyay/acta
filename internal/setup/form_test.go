package setup_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
		PluginDir: "/p",
		Current:   config.User{},
	}
	groups := formGroupText(t, env)

	// Eight question screens, then one screen with a yes/no per harness.
	if len(groups) != 9 {
		t.Fatalf("got %d groups, want 9", len(groups))
	}

	// Fixed order: every screen opens with the blue-diamond title line,
	// its own step count, and a dim one-line description on the rail.
	want := []struct {
		title string
		count string
		desc  string
	}{
		{"Chat language", "(1/9)", "The language the agent chats in. Code and files use the repo language."},
		{"Reply style", "(2/9)", ""},
		{"Tone", "(3/9)", "Optional. Your own words, like \"casual, no jargon\"."},
		{"Repo language", "(4/9)", "Code, comments, commits, specs and plans."},
		{"Build executor", "(5/9)", ""},
		{"Subagent models", "(6/9)", "Claude Code only."},
		{"Plan detail", "(7/9)", ""},
		{"Questions", "(8/9)", ""},
		{"Plugin install", "(9/9)", "One row per tool found."},
	}
	for i, w := range want {
		lines := strings.Split(groups[i], "\n")
		if w.desc == "" {
			if len(lines) != 1 {
				t.Fatalf("group %d: header must be a title line only, got %q", i+1, groups[i])
			}
		} else if len(lines) != 2 {
			t.Fatalf("group %d: header must be a title line and a description line, got %q", i+1, groups[i])
		}
		if !strings.HasPrefix(lines[0], "◆  "+w.title+"  "+w.count) {
			t.Errorf("group %d: title line = %q, want ◆, %q, %q", i+1, lines[0], w.title, w.count)
		}
		if w.desc != "" && lines[1] != "│  "+w.desc {
			t.Errorf("group %d: description line = %q, want a rail line with %q", i+1, lines[1], w.desc)
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

// TestFormPlanDepthTitle keeps the plan-depth question worded exactly: the
// screen title tells the user what a plan spells out.
func TestFormPlanDepthTitle(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", home+"/config.yaml")
	t.Setenv("TMPDIR", home)

	header := formGroupText(t, setup.Env{Current: config.User{}})[6]
	if !strings.Contains(header, "Plan detail") {
		t.Errorf("plan-depth question must read %q, got %q", "Plan detail", header)
	}
}

// TestFormRail checks every content line of every screen sits on the rail,
// and that select options use the clack marks: ● chosen, ○ the rest, › on
// the focused row.
func TestFormRail(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", home+"/config.yaml")
	t.Setenv("TMPDIR", home)

	env := setup.Env{Harnesses: []string{"claude", "omp"}, PluginDir: "/p", Current: config.User{}}
	content := formGroupContent(t, env)
	for i, c := range content {
		for _, line := range strings.Split(c, "\n") {
			if !strings.HasPrefix(line, "│  ") {
				t.Errorf("group %d: line %q is off the rail", i+1, line)
			}
		}
	}
	style := content[1]
	for _, want := range []string{"› ● adhd", "  ○ plain"} {
		if !strings.Contains(style, want) {
			t.Errorf("style screen must show %q, got %q", want, style)
		}
	}
	// Yes/no sits on the rail as `● Yes  ○ No` for each harness.
	if !strings.Contains(content[8], "claude") || !strings.Contains(content[8], "● Yes  ○ No") {
		t.Errorf("harness screen must read `● Yes  ○ No`, got %q", content[8])
	}
}

// TestFormInstallRows checks who gets a Yes/No row: only a tool that still
// needs the plugin, and only when a plugin dir lets the wizard install it.
func TestFormInstallRows(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", home+"/config.yaml")
	t.Setenv("TMPDIR", home)

	cases := []struct {
		name      string
		env       setup.Env
		wantGroup bool
		has, not  []string
	}{
		{"not installed with dir", setup.Env{Harnesses: []string{"claude", "omp"}, PluginDir: "/p"},
			true, []string{"claude", "omp"}, nil},
		{"not installed without dir", setup.Env{Harnesses: []string{"claude", "omp"}},
			false, nil, nil},
		{"all installed with dir", setup.Env{Harnesses: []string{"claude", "omp"}, PluginDir: "/p",
			Installed: map[string]bool{"claude": true, "omp": true}}, false, nil, nil},
		{"mix with dir", setup.Env{Harnesses: []string{"claude", "omp"}, PluginDir: "/p",
			Installed: map[string]bool{"claude": true}}, true, []string{"omp"}, []string{"claude"}},
		{"mix without dir", setup.Env{Harnesses: []string{"claude", "omp"},
			Installed: map[string]bool{"claude": true}}, false, nil, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			content := formGroupContent(t, c.env)
			if c.wantGroup {
				if len(content) != 9 {
					t.Fatalf("got %d groups, want 9", len(content))
				}
				last := content[8]
				for _, h := range c.has {
					if !strings.Contains(last, h) {
						t.Errorf("install screen misses %q: %q", h, last)
					}
				}
				for _, h := range c.not {
					if strings.Contains(last, h) {
						t.Errorf("install screen asks about installed %q: %q", h, last)
					}
				}
				return
			}
			if len(content) != 8 {
				t.Fatalf("got %d groups, want 8 (no install screen)", len(content))
			}
		})
	}
}

// pressEnter drives one question like the user would: Init, then Enter, and
// returns what is on screen.
func pressEnter(t *testing.T, g *huh.Group) string {
	t.Helper()
	hold := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	defer lipgloss.SetColorProfile(hold)
	f := huh.NewForm(g)
	f.Init()
	var m tea.Model = f
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return m.View()
}

// TestFormFirstRunEmpty checks a run with no config: language inputs start
// empty and refuse an empty answer with `required`; tone may stay empty;
// selects start on their first option.
func TestFormFirstRunEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", home+"/config.yaml")
	t.Setenv("TMPDIR", home)

	d := setup.FormDefaults(setup.Env{})
	if d.Language != "" || d.RepoLanguage != "" || d.Tone != "" {
		t.Fatalf("text defaults = %q/%q/%q, want all empty", d.Language, d.RepoLanguage, d.Tone)
	}
	if d.Style != "adhd" || d.BuildExecutor != "subagent" || d.SubagentModels != "default" ||
		d.PlanDepth != "full" || d.Questions != "one" {
		t.Fatalf("select defaults = %+v, want the first options", d)
	}

	content := formGroupContent(t, setup.Env{})
	if !strings.Contains(content[0], "> E") {
		t.Errorf("chat language must show the placeholder, got %q", content[0])
	}
	for i, name := range map[int]string{0: "chat language", 3: "repo language"} {
		view := pressEnter(t, setup.FormGroups(setup.Env{})[i])
		if !strings.Contains(view, "▲ required") {
			t.Errorf("%s: empty answer must show required, got %q", name, view)
		}
		if !strings.Contains(view, "│  ") {
			t.Errorf("%s: question must stay open, got %q", name, view)
		}
	}
	// Tone is optional: Enter on empty moves on, so no error shows.
	if view := pressEnter(t, setup.FormGroups(setup.Env{})[2]); strings.Contains(view, "required") {
		t.Errorf("tone must be optional, got %q", view)
	}
}

// TestFormWithConfigValues checks a config file's values are the starting
// text, not the placeholder.
func TestFormWithConfigValues(t *testing.T) {
	d := setup.FormDefaults(setup.Env{Current: config.User{ChatLanguage: "Korean", RepoLanguage: "French"}})
	if d.Language != "Korean" || d.RepoLanguage != "French" {
		t.Fatalf("defaults = %+v, want the config values", d)
	}
}
