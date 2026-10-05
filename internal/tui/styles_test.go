package tui

import (
	"math"
	"reflect"
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

func TestNoWorkBrushIsLeft(t *testing.T) {
	t.Parallel()

	if _, ok := reflect.TypeOf(styles{}).FieldByName("work"); ok {
		t.Error("styles still has a work brush; work under way is plain now")
	}
}

func TestTerminalThemeUsesANSISlots(t *testing.T) {
	th, _ := theme.Builtin("terminal")
	s := newStyles(th, true)
	if s.accentColor != lipgloss.Color("12") {
		t.Fatalf("accent = %v", s.accentColor)
	}
	withTrueColor(func() {
		out := s.accent.Render("x") + s.dim.Render("x") + s.selected.Render("x")
		if strings.Contains(out, "38;2;") || strings.Contains(out, "48;2;") {
			t.Fatalf("24-bit color in terminal mode: %q", out)
		}
		if out == "" {
			t.Fatal("terminal mode drew no color at all")
		}
	})
	if s.paintFrame("a\nb", 1) != "a\nb" {
		t.Fatalf("terminal mode painted the frame: %q", s.paintFrame("a\nb", 1))
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
	if s.pulse[0].GetForeground() != lipgloss.Color(th.ANSI[slotGreen]) {
		t.Fatalf("pulse frame 0 = %v", s.pulse[0].GetForeground())
	}
	if s.dim.GetForeground() != lipgloss.Color("#414868") {
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
		out := s.paintFrame(s.accent.Render("a")+"b\n"+"c", 10)
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

// TestPaintFrameErasesOnlyShortLines guards the right wall. Erase-line right
// after a full line clears its last cell, because the cursor is still sitting
// in the last column. So only a short line may get it.
func TestPaintFrameErasesOnlyShortLines(t *testing.T) {
	th, _ := theme.Builtin("tokyo-night")
	s := newStyles(th, true)
	withTrueColor(func() {
		out := strings.Split(s.paintFrame("abcde\nab\n"+s.accent.Render("abcd")+"┐", 5), "\n")
		if strings.HasSuffix(out[0], "\x1b[K") {
			t.Errorf("full line got an erase: %q", out[0])
		}
		if !strings.HasSuffix(out[1], "\x1b[K") {
			t.Errorf("short line lost its erase, so the background stops short: %q", out[1])
		}
		if strings.HasSuffix(out[2], "\x1b[K") {
			t.Errorf("full line with colors and a wide rune edge got an erase: %q", out[2])
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

// contrast is the WCAG contrast ratio of two #rrggbb colors. Terminals like
// Ghostty swap a foreground for white or black when this is under their
// floor, so a dim color that sits too close to the background turns bright.
func contrast(t *testing.T, a, b string) float64 {
	t.Helper()
	lum := func(h string) float64 {
		n, err := strconv.ParseUint(strings.TrimPrefix(h, "#"), 16, 32)
		if err != nil || len(h) != 7 {
			t.Fatalf("not a #rrggbb color: %q", h)
		}
		var out float64
		for i, w := range []float64{0.2126, 0.7152, 0.0722} {
			c := float64(n>>(16-8*i)&0xff) / 255
			if c <= 0.03928 {
				c /= 12.92
			} else {
				c = math.Pow((c+0.055)/1.055, 2.4)
			}
			out += w * c
		}
		return out
	}
	la, lb := lum(a), lum(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// TestDimKeepsContrastOverTheBackground checks every built-in theme with its
// own colors. The dim color must stay far enough from the background that a
// terminal with a minimum contrast of 1.5 leaves it alone.
func TestDimKeepsContrastOverTheBackground(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, ok := theme.Builtin(name)
		if !ok || th.BG == "" {
			continue
		}
		s := newStyles(th, true)
		got := s.dim.GetForeground()
		if got != lipgloss.Color(th.ANSI[slotDim]) {
			t.Errorf("%s: dim %v, want slot 8 %s", name, got, th.ANSI[slotDim])
			continue
		}
		if r := contrast(t, string(got.(lipgloss.Color)), th.BG); r < 1.6 {
			t.Errorf("%s: dim %v on %s has contrast %.2f, want at least 1.6", name, got, th.BG, r)
		}
	}
}

func TestDimIsNeverFaint(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, _ := theme.Builtin(name)
		if newStyles(th, true).dim.GetFaint() {
			t.Errorf("%s: dim is faint, which lowers its contrast again", name)
		}
	}
	term, _ := theme.Builtin("terminal")
	if got := newStyles(term, true).dim.GetForeground(); got != lipgloss.Color("8") {
		t.Errorf("terminal theme: dim %v, want slot 8", got)
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
			board.KindBug: 1, board.KindDebt: 3, board.KindDebtItem: 3,
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
			t.Errorf("%s: Activity tab = %v, want slot 6", c.name, got)
		}
		for name, r := range map[string]struct {
			brush lipgloss.Style
			slot  int
		}{
			"label": {s.label, 6}, "footLabel": {s.footLabel, 5}, "done": {s.done, 2},
			"waiting": {s.waiting, 8}, "problem": {s.problem, 1}, "live": {s.live, 2},
			// The dot of work under way is the pulse, so its first frame is green.
			"pulseDot": {s.pulse[0], 2},
		} {
			if got := r.brush.GetForeground(); got != c.at(r.slot) {
				t.Errorf("%s: %s = %v, want slot %d", c.name, name, got, r.slot)
			}
		}
	}
}

// TestScratchIsPink checks every built-in theme. A scratch id and the
// Scratchpad tab must be pink, so a scratch never looks like finished work.
// Done keeps green, so nothing else may take the pink.
func TestScratchIsPink(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, _ := theme.Builtin(name)
		for _, dark := range []bool{true, false} {
			want := lipgloss.Color("#c2185b")
			switch {
			case th.BG == "":
				want = "212"
			case th.Dark(dark):
				want = "#ff79c6"
			}
			s := newStyles(th, dark)
			if got := s.kind(board.KindScratch).GetForeground(); got != want {
				t.Errorf("%s dark=%v: scratch id = %v, want %v", name, dark, got, want)
			}
			if got := s.tabColor(board.KindScratch); got != want {
				t.Errorf("%s dark=%v: Scratchpad tab = %v, want %v", name, dark, got, want)
			}
			if got := s.done.GetForeground(); got == want {
				t.Errorf("%s dark=%v: done took the scratch pink", name, dark)
			}
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

// TestPulseFramesStayReadable walks every built-in theme. A frame too close
// to the background would trip the terminal's minimum contrast and flash
// white, so none may drop under 1.6.
func TestPulseFramesStayReadable(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, _ := theme.Builtin(name)
		s := newStyles(th, true)
		if len(s.pulse) != pulseSteps {
			t.Fatalf("%s: %d pulse frames, want %d", name, len(s.pulse), pulseSteps)
		}
		fg := func(i int) string { return string(s.pulse[i].GetForeground().(lipgloss.Color)) }
		if th.BG == "" {
			for i := range s.pulse {
				want := "2"
				if i >= pulseSteps/2 {
					want = "10"
				}
				if fg(i) != want {
					t.Errorf("%s frame %d = %s, want %s", name, i, fg(i), want)
				}
			}
			continue
		}
		if fg(0) != th.ANSI[slotGreen] {
			t.Errorf("%s frame 0 = %s, want green %s", name, fg(0), th.ANSI[slotGreen])
		}
		for i := 1; i < pulseSteps/2; i++ {
			if fg(i) != fg(pulseSteps-i) {
				t.Errorf("%s: frame %d %s and frame %d %s differ, the pulse is not symmetric", name, i, fg(i), pulseSteps-i, fg(pulseSteps-i))
			}
		}
		prev := contrast(t, fg(0), th.BG)
		for i := range s.pulse {
			c := contrast(t, fg(i), th.BG)
			if c < minPulseContrast {
				t.Errorf("%s frame %d %s has contrast %.2f, want at least %.1f", name, i, fg(i), c, minPulseContrast)
			}
			if i > 0 && i <= pulseSteps/2 && c > prev+1e-9 {
				t.Errorf("%s frame %d gets brighter on the way down: %.3f after %.3f", name, i, c, prev)
			}
			prev = c
		}
	}
}

func TestTokyoNightPulseMoves(t *testing.T) {
	t.Parallel()

	th, _ := theme.Builtin("tokyo-night")
	s := newStyles(th, true)
	if s.pulse[0].GetForeground() == s.pulse[pulseSteps/2].GetForeground() {
		t.Fatal("frame 4 is still green, so the dot would not pulse")
	}
}

func TestGoingDotIsFrameZero(t *testing.T) {
	withTrueColor(func() {
		for _, name := range []string{"tokyo-night", "terminal"} {
			th, _ := theme.Builtin(name)
			s := newStyles(th, true)
			if s.dot(dotGoing).GetForeground() != s.pulse[0].GetForeground() {
				t.Errorf("%s: dot of work under way is %v, want frame 0 %v", name, s.dot(dotGoing).GetForeground(), s.pulse[0].GetForeground())
			}
			if s.goingDot != s.pulse[0].Render(dotGoing) {
				t.Errorf("%s: goingDot %q, want %q", name, s.goingDot, s.pulse[0].Render(dotGoing))
			}
		}
	})
}

func TestFarthestMixKeepsTheFloor(t *testing.T) {
	t.Parallel()

	if f := farthestMix("#1a1b26", "#1a1b26", 1.6); f != 0 {
		t.Errorf("a green already on the floor mixed %.2f, want 0", f)
	}
	f := farthestMix("#9ece6a", "#1a1b26", 1.6)
	if f <= 0 || f >= 1 {
		t.Fatalf("farthestMix = %.2f, want between 0 and 1", f)
	}
	if c := contrastRatio(mixHex("#9ece6a", "#1a1b26", f), "#1a1b26"); c < 1.6 {
		t.Errorf("the farthest mix has contrast %.2f, want at least 1.6", c)
	}
	if got := mixHex("#000000", "#ffffff", 0.5); got != "#808080" && got != "#7f7f7f" {
		t.Errorf("mixHex halfway = %s", got)
	}
	if got := mixHex("nope", "#ffffff", 0.5); got != "nope" {
		t.Errorf("mixHex on a bad color = %s, want it back as it came", got)
	}
}
