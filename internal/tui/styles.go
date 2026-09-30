package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/theme"
)

// styles are the brushes the TUI paints with. They come from one theme, so
// changing the theme changes every color at once.
type styles struct {
	accent, work, faint, dim, selected             lipgloss.Style
	label, footLabel, done, waiting, problem, live lipgloss.Style
	accentColor                                    lipgloss.TerminalColor
	kinds                                          map[board.Kind]lipgloss.Color
	cyan                                           lipgloss.Color // the Activities tab, which holds no one kind
	bandFG                                         lipgloss.Color // text on a colored band
	bg, fg                                         string         // empty for the terminal theme
}

// Each role always takes the same slot, so any theme with 16 colors works.
const (
	slotAccent  = 12
	slotWork    = 4
	slotDim     = 8
	slotRed     = 1
	slotGreen   = 2
	slotYellow  = 3
	slotBlue    = 4
	slotMagenta = 5
	slotCyan    = 6
)

// kindSlots gives each kind a color of its own, so an id tells what it is
// before anyone reads it. A task is part of a plan, so it wears the plan color.
var kindSlots = map[board.Kind]int{
	board.KindScratch: slotGreen, board.KindBug: slotRed,
	board.KindDebt: slotYellow, board.KindDebtItem: slotYellow,
	board.KindStory: slotMagenta, board.KindPlan: slotBlue, board.KindTask: slotBlue,
}

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
	kinds := make(map[board.Kind]lipgloss.Color, len(kindSlots))
	for k, i := range kindSlots {
		kinds[k] = slot(i)
	}
	// The terminal theme has no background color, so text on a band takes
	// slot 0, the dark end of its colors.
	bandFG := slot(0)
	if t.BG != "" {
		bandFG = lipgloss.Color(t.BG)
	}
	return styles{
		accentColor: accent,
		accent:      lipgloss.NewStyle().Foreground(accent),
		// work marks a row whose work has begun.
		work: lipgloss.NewStyle().Foreground(slot(slotWork)),
		// faint paints every row the cursor is not on.
		faint: lipgloss.NewStyle().Faint(true),
		// dim paints the screen behind a popup, so the box on top is the only
		// thing left with a color of its own. It is slot 8 as it is, with no
		// faint: a color closer to the background trips the minimum contrast
		// some terminals keep, and they then draw it bright.
		dim: lipgloss.NewStyle().Foreground(slot(slotDim)),
		// selected is a band across the row with bright text on it, never
		// reversed video, so the words stay readable wherever it falls.
		selected:  lipgloss.NewStyle().Bold(true).Foreground(selFG).Background(selBG),
		label:     lipgloss.NewStyle().Foreground(slot(slotCyan)),
		footLabel: lipgloss.NewStyle().Foreground(slot(slotMagenta)),
		done:      lipgloss.NewStyle().Foreground(slot(slotGreen)),
		waiting:   lipgloss.NewStyle().Foreground(slot(slotDim)),
		problem:   lipgloss.NewStyle().Foreground(slot(slotRed)),
		live:      lipgloss.NewStyle().Foreground(slot(slotGreen)),
		kinds:     kinds,
		cyan:      slot(slotCyan),
		bandFG:    bandFG,
		bg:        t.BG,
		fg:        t.FG,
	}
}

// kind is the brush of an id of kind k. A kind with no color of its own
// gets a plain brush, so its id reads in the theme foreground.
func (s styles) kind(k board.Kind) lipgloss.Style {
	c, ok := s.kinds[k]
	if !ok {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(c)
}

// tabColor is the color of a top tab. The Activities tab has no kind, so it
// takes cyan.
func (s styles) tabColor(k board.Kind) lipgloss.Color {
	if c, ok := s.kinds[k]; ok {
		return c
	}
	return s.cyan
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
