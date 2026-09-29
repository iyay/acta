package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/iyay/acta/internal/theme"
)

// withTrueColor lets a style actually draw, because lipgloss drops every
// color code when the profile is plain ASCII.
func withTrueColor(f func()) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	f()
}

func TestTerminalThemeUsesANSISlots(t *testing.T) {
	th, _ := theme.Builtin("terminal")
	s := newStyles(th, true)
	if s.accentColor != lipgloss.Color("12") {
		t.Fatalf("accent = %v", s.accentColor)
	}
	withTrueColor(func() {
		out := s.accent.Render("x") + s.dim.Render("x") + s.selected.Render("x") + s.work.Render("x")
		if strings.Contains(out, "38;2;") || strings.Contains(out, "48;2;") {
			t.Fatalf("24-bit color in terminal mode: %q", out)
		}
		if out == "" {
			t.Fatal("terminal mode drew no color at all")
		}
	})
	if s.paintFrame("a\nb") != "a\nb" {
		t.Fatalf("terminal mode painted the frame: %q", s.paintFrame("a\nb"))
	}
}

func TestTerminalSelectedFollowsDarkFlag(t *testing.T) {
	t.Parallel()

	th, _ := theme.Builtin("terminal")
	dark := newStyles(th, true).selected
	light := newStyles(th, false).selected
	if dark.GetForeground() != lipgloss.Color("15") || dark.GetBackground() != lipgloss.Color("0") {
		t.Fatalf("dark selected %v on %v", dark.GetForeground(), dark.GetBackground())
	}
	if light.GetForeground() != lipgloss.Color("0") || light.GetBackground() != lipgloss.Color("7") {
		t.Fatalf("light selected %v on %v", light.GetForeground(), light.GetBackground())
	}
}

func TestHexThemeRoles(t *testing.T) {
	t.Parallel()

	th, _ := theme.Builtin("tokyo-night")
	s := newStyles(th, false)
	if s.accentColor != lipgloss.Color("#7aa2f7") {
		t.Fatalf("accent = %v", s.accentColor)
	}
	if s.work.GetForeground() != lipgloss.Color("#7aa2f7") {
		t.Fatalf("work = %v", s.work.GetForeground())
	}
	if s.dim.GetForeground() != lipgloss.Color("#2d3147") {
		t.Fatalf("dim = %v", s.dim.GetForeground())
	}
	if !s.faint.GetFaint() {
		t.Fatal("faint does not faint")
	}
	if s.selected.GetBackground() != lipgloss.Color("#283457") || s.selected.GetForeground() != lipgloss.Color("#c0caf5") {
		t.Fatalf("selected does not use the selection colors: %v on %v", s.selected.GetForeground(), s.selected.GetBackground())
	}
	if s.bg != th.BG || s.fg != th.FG {
		t.Fatalf("frame colors are %q on %q", s.fg, s.bg)
	}
}

func TestHexSelectedFallsBackToSlots(t *testing.T) {
	t.Parallel()

	th, _ := theme.Builtin("tokyo-night")
	th.SelectionBG, th.SelectionFG = "", ""
	s := newStyles(th, true)
	if s.selected.GetBackground() != lipgloss.Color(th.ANSI[0]) || s.selected.GetForeground() != lipgloss.Color(th.ANSI[15]) {
		t.Fatalf("dark fallback is %v on %v", s.selected.GetForeground(), s.selected.GetBackground())
	}
	day, _ := theme.Builtin("tokyo-night-day")
	day.SelectionBG, day.SelectionFG = "", ""
	s = newStyles(day, true) // the theme itself is light, so the flag is ignored
	if s.selected.GetBackground() != lipgloss.Color(day.ANSI[7]) || s.selected.GetForeground() != lipgloss.Color(day.ANSI[0]) {
		t.Fatalf("light fallback is %v on %v", s.selected.GetForeground(), s.selected.GetBackground())
	}
}

// Every inner reset would drop the theme background. The frame must carry
// it again right after each one, and on every line.
func TestPaintFrameKeepsBackgroundAfterResets(t *testing.T) {
	th, _ := theme.Builtin("tokyo-night")
	s := newStyles(th, true)
	withTrueColor(func() {
		out := s.paintFrame(s.accent.Render("a") + "b\n" + "c")
		bg := "48;2;26;27;38"
		lines := strings.Split(out, "\n")
		if len(lines) != 2 {
			t.Fatalf("paintFrame wrote %d lines, want 2", len(lines))
		}
		for i, ln := range lines {
			if !strings.HasPrefix(ln, "\x1b[") || !strings.Contains(ln, bg) {
				t.Fatalf("line %d has no background: %q", i, ln)
			}
			// Each reset divides the line. Every piece after the first one
			// has to carry the background again, or the text behind the
			// reset is left on the terminal's own color. A line that holds
			// no reset needs no repair.
			segs := strings.Split(ln, "\x1b[0m")
			for j, seg := range segs[1:] {
				if !strings.Contains(seg, bg) {
					t.Fatalf("line %d piece %d after a reset has no background: %q", i, j+1, seg)
				}
			}
		}
	})
}

func TestWithThemeFallsBack(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, b := fixture(t)
	m := New(cfg, b, true).WithTheme("nope", true)
	if m.styles.bg != "#1a1b26" {
		t.Fatalf("bg = %q", m.styles.bg)
	}
	if !strings.Contains(m.status, `theme "nope"`) {
		t.Fatalf("status = %q", m.status)
	}
	_ = sized(m, 80, 20).View() // must not panic
}

func TestWithThemeLoadsAKnownTheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, b := fixture(t)
	m := New(cfg, b, true).WithTheme("gruvbox-dark", false)
	if m.styles.bg != "#282828" || m.styles.accentColor != lipgloss.Color("#83a598") {
		t.Fatalf("bg = %q accent = %v", m.styles.bg, m.styles.accentColor)
	}
	if m.status != "" {
		t.Fatalf("a theme that loads leaves no status, got %q", m.status)
	}
}

func TestNewDefaultsToTokyoNight(t *testing.T) {
	t.Parallel()

	cfg, b := fixture(t)
	if New(cfg, b, true).styles.bg != "#1a1b26" {
		t.Fatal("New did not default to tokyo-night")
	}
}

// The whole screen a person sees has to carry the theme background, so the
// check runs on what View draws and not only on paintFrame alone.
func TestViewCarriesTheThemeBackground(t *testing.T) {
	cfg, b := fixture(t)
	withTrueColor(func() {
		for _, name := range []string{"tokyo-night", "terminal"} {
			v := sized(New(cfg, b, true).WithTheme(name, true), 120, 40).View()
			for i, ln := range strings.Split(v, "\n") {
				if name == "terminal" {
					if strings.Contains(ln, "38;2;") || strings.Contains(ln, "48;2;") {
						t.Fatalf("theme %q line %d holds 24-bit color: %q", name, i, ln)
					}
					continue
				}
				if !strings.HasPrefix(ln, "\x1b[") || !strings.Contains(ln, "48;2;26;27;38") {
					t.Fatalf("theme %q line %d has no background: %q", name, i, ln)
				}
			}
		}
	})
}

// TestDimFadesTowardTheBackground checks every built-in theme with its own
// colors: the dim color sits halfway between the background and slot 8, so
// it is never a color the panes already wear.
func TestDimFadesTowardTheBackground(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, ok := theme.Builtin(name)
		if !ok || th.BG == "" {
			continue
		}
		s := newStyles(th, true)
		got := s.dim.GetForeground()
		if want := lipgloss.Color(mixHex(th.BG, th.ANSI[slotDim])); got != want {
			t.Errorf("%s: dim %v, want %v", name, got, want)
		}
		for _, used := range []string{th.FG, th.ANSI[slotDim], th.ANSI[slotAccent], th.ANSI[slotWork]} {
			if th.BG != th.ANSI[slotDim] && got == lipgloss.Color(used) {
				t.Errorf("%s: dim %v is a color the panes already wear", name, got)
			}
		}
	}
	term, _ := theme.Builtin("terminal")
	if got := newStyles(term, true).dim.GetForeground(); got != lipgloss.Color("8") {
		t.Errorf("terminal theme: dim %v, want slot 8", got)
	}
}

func TestMixHex(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ a, b, want string }{
		{"#1a1b26", "#414868", "#2d3147"},
		{"#000000", "#ffffff", "#7f7f7f"},
		{"#ABCDEF", "#abcdef", "#abcdef"},
		{"", "#414868", "#414868"},
		{"#12345", "#414868", "#414868"},
		{"#zzzzzz", "#414868", "#414868"},
	} {
		if got := mixHex(c.a, c.b); got != c.want {
			t.Errorf("mixHex(%q, %q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}
