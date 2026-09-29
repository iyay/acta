package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/iyay/acta/internal/theme"
)

// styles are the brushes the TUI paints with. They come from one theme, so
// changing the theme changes every color at once.
type styles struct {
	accent, work, faint, dim, selected lipgloss.Style
	accentColor                        lipgloss.TerminalColor
	bg, fg                             string // empty for the terminal theme
}

// Each role always takes the same slot, so any theme with 16 colors works.
const (
	slotAccent = 12
	slotWork   = 4
	slotDim    = 8
)

func newStyles(t theme.Theme, dark bool) styles {
	// slot gives the theme's hex, or the plain ANSI number for the terminal
	// theme so the terminal picks the color.
	slot := func(i int) lipgloss.Color {
		if t.BG == "" {
			return lipgloss.Color(strconv.Itoa(i))
		}
		return lipgloss.Color(t.ANSI[i])
	}
	selFG, selBG := slot(15), slot(0)
	if !t.Dark(dark) {
		selFG, selBG = slot(0), slot(7)
	}
	if t.SelectionBG != "" && t.SelectionFG != "" {
		selFG, selBG = lipgloss.Color(t.SelectionFG), lipgloss.Color(t.SelectionBG)
	}
	accent := slot(slotAccent)
	return styles{
		accentColor: accent,
		accent:      lipgloss.NewStyle().Foreground(accent),
		// work marks a row whose work has begun. Dimmed, because the selected
		// row is the one that should catch the eye.
		work: lipgloss.NewStyle().Foreground(slot(slotWork)).Faint(true),
		// faint paints every row the cursor is not on.
		faint: lipgloss.NewStyle().Faint(true),
		// dim paints the screen behind a popup, so the box on top is the only
		// thing left with a color of its own. The rows behind are already
		// faint, so dim has to fade further: halfway to the background.
		dim: lipgloss.NewStyle().Faint(true).Foreground(dimColor(t, slot(slotDim))),
		// selected is a band across the row with bright text on it, never
		// reversed video, so the words stay readable wherever it falls.
		selected: lipgloss.NewStyle().Bold(true).Foreground(selFG).Background(selBG),
		bg:       t.BG,
		fg:       t.FG,
	}
}

// dimColor is the color behind a popup. A theme with its own colors mixes
// slot 8 halfway toward its background. The terminal theme has no hex, so
// mixHex gives slot 8 straight back and the terminal keeps picking the color.
func dimColor(t theme.Theme, grey lipgloss.Color) lipgloss.Color {
	return lipgloss.Color(mixHex(t.BG, string(grey)))
}

// mixHex gives the color halfway between two #rrggbb colors. When either one
// is not a #rrggbb color there is nothing to mix, so b comes back as it is.
func mixHex(a, b string) string {
	ca, okA := hexRGB(a)
	cb, okB := hexRGB(b)
	if !okA || !okB {
		return b
	}
	return fmt.Sprintf("#%02x%02x%02x", (ca[0]+cb[0])/2, (ca[1]+cb[1])/2, (ca[2]+cb[2])/2)
}

// hexRGB reads #rrggbb into its red, green and blue parts.
func hexRGB(s string) ([3]int, bool) {
	s, ok := strings.CutPrefix(s, "#")
	if !ok || len(s) != 6 {
		return [3]int{}, false
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return [3]int{}, false
	}
	return [3]int{int(n >> 16 & 0xff), int(n >> 8 & 0xff), int(n & 0xff)}, true
}

// paintFrame lays the theme background under the whole screen. Every style
// ends with a reset that drops the background, so the theme colors go back
// on right after each reset and at the start of each line.
func (s styles) paintFrame(out string) string {
	if s.bg == "" {
		return out
	}
	p := termenv.TrueColor
	set := termenv.CSI + p.Color(s.bg).Sequence(true) + "m" + termenv.CSI + p.Color(s.fg).Sequence(false) + "m"
	lines := strings.Split(out, "\n")
	for i, ln := range lines {
		lines[i] = set + strings.ReplaceAll(ln, "\x1b[0m", "\x1b[0m"+set) + "\x1b[K"
	}
	return strings.Join(lines, "\n")
}
