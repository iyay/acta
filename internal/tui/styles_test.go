package tui

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/iyay/acta/internal/board"
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

// sgr finds each color code a style writes.
var sgr = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

// sgrHas says whether s turns on the plain code, like "1" for bold or "2"
// for faint. The numbers inside a color, like the 2 in "38;2;r;g;b", are
// skipped, so they never pass for faint.
func sgrHas(s, code string) bool {
	for _, m := range sgr.FindAllStringSubmatch(s, -1) {
		ps := strings.Split(m[1], ";")
		for i := 0; i < len(ps); i++ {
			switch ps[i] {
			case "38", "48":
				if i+1 < len(ps) && ps[i+1] == "5" {
					i += 2
				} else {
					i += 4
				}
				continue
			}
			if ps[i] == code {
				return true
			}
		}
	}
	return false
}

func TestKindAndRoleSlots(t *testing.T) {
	t.Parallel()

	hex, _ := theme.Builtin("tokyo-night")
	term, _ := theme.Builtin("terminal")
	for _, c := range []struct {
		name string
		th   theme.Theme
		at   func(int) lipgloss.Color
	}{
		{"tokyo-night", hex, func(i int) lipgloss.Color { return lipgloss.Color(hex.ANSI[i]) }},
		{"terminal", term, func(i int) lipgloss.Color { return lipgloss.Color(strconv.Itoa(i)) }},
	} {
		s := newStyles(c.th, true)
		for k, slot := range map[board.Kind]int{
			board.KindScratch: 2, board.KindBug: 1, board.KindDebt: 3, board.KindDebtItem: 3,
			board.KindStory: 5, board.KindPlan: 4, board.KindTask: 4,
		} {
			if got := s.kind(k).GetForeground(); got != c.at(slot) {
				t.Errorf("%s: kind %s = %v, want slot %d", c.name, k, got, slot)
			}
			if got := s.tabColor(k); got != c.at(slot) {
				t.Errorf("%s: tab %s = %v, want slot %d", c.name, k, got, slot)
			}
		}
		if got := s.tabColor(""); got != c.at(6) {
			t.Errorf("%s: Activities tab = %v, want slot 6", c.name, got)
		}
		for name, r := range map[string]struct {
			brush lipgloss.Style
			slot  int
		}{
			"label": {s.label, 6}, "footLabel": {s.footLabel, 5}, "done": {s.done, 2},
			"waiting": {s.waiting, 8}, "problem": {s.problem, 1}, "live": {s.live, 2},
		} {
			if got := r.brush.GetForeground(); got != c.at(r.slot) {
				t.Errorf("%s: %s = %v, want slot %d", c.name, name, got, r.slot)
			}
		}
		if s.work.GetFaint() {
			t.Errorf("%s: work is still faint", c.name)
		}
	}
}

func TestKindWithNoColorIsPlain(t *testing.T) {
	t.Parallel()

	th, _ := theme.Builtin("tokyo-night")
	s := newStyles(th, true)
	if _, ok := s.kind("").GetForeground().(lipgloss.NoColor); !ok {
		t.Fatalf("empty kind has color %v", s.kind("").GetForeground())
	}
	if _, ok := s.kind("nope").GetForeground().(lipgloss.NoColor); !ok {
		t.Fatalf("unknown kind has color %v", s.kind("nope").GetForeground())
	}
}

func TestBandTextUsesTheBackground(t *testing.T) {
	t.Parallel()

	hex, _ := theme.Builtin("tokyo-night")
	if got := newStyles(hex, true).bandFG; got != lipgloss.Color(hex.BG) {
		t.Fatalf("tokyo-night band text = %v, want %v", got, hex.BG)
	}
	term, _ := theme.Builtin("terminal")
	if got := newStyles(term, true).bandFG; got != lipgloss.Color("0") {
		t.Fatalf("terminal band text = %v, want slot 0", got)
	}
}

func TestSgrHasSkipsColorNumbers(t *testing.T) {
	t.Parallel()

	if sgrHas("\x1b[38;2;2;2;2mx\x1b[0m", "2") {
		t.Fatal("a 24-bit color passed for faint")
	}
	if !sgrHas("\x1b[2;38;2;1;1;1mx\x1b[0m", "2") {
		t.Fatal("faint next to a color was missed")
	}
	if !sgrHas("\x1b[1mx\x1b[0m", "1") {
		t.Fatal("plain bold was missed")
	}
}
