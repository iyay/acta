package tui

import (
	"fmt"
	"math"
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
	cyan                                           lipgloss.Color   // the Activities tab, which holds no one kind
	bandFG                                         lipgloss.Color   // text on a colored band
	pulse                                          []lipgloss.Style // the frames of the dot of work under way
	goingDot                                       string           // that dot drawn in frame 0, as every view draws it
	bg, fg                                         string           // empty for the terminal theme
}

// Each role always takes the same slot, so any theme with 16 colors works.
const (
	slotAccent      = 12
	slotWork        = 4
	slotDim         = 8
	slotRed         = 1
	slotGreen       = 2
	slotYellow      = 3
	slotBlue        = 4
	slotMagenta     = 5
	slotCyan        = 6
	slotBrightGreen = 10
)

// The pulse of a dot whose work is under way: eight frames, from green toward
// the background and back. Its darkest frame still keeps this much contrast,
// or a terminal with a minimum contrast would draw it white.
const (
	pulseSteps       = 8
	minPulseContrast = 1.6
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
	pulse := pulseBrushes(t, slot)
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
		pulse:     pulse,
		goingDot:  pulse[0].Render(dotGoing),
	}
}

// pulseBrushes gives the frames of the pulse. A theme with its own colors
// fades green toward its background, but only as far as keeps it readable.
// The terminal theme has no hex to fade, so it swaps green for bright green.
func pulseBrushes(t theme.Theme, slot func(int) lipgloss.Color) []lipgloss.Style {
	out := make([]lipgloss.Style, pulseSteps)
	if t.BG == "" {
		for i := range out {
			c := slotGreen
			if i >= pulseSteps/2 {
				c = slotBrightGreen
			}
			out[i] = lipgloss.NewStyle().Foreground(slot(c))
		}
		return out
	}
	green := t.ANSI[slotGreen]
	far := farthestMix(green, t.BG, minPulseContrast)
	half := pulseSteps / 2
	for i := range out {
		step := i
		if step > half {
			step = pulseSteps - i
		}
		out[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(mixHex(green, t.BG, far*float64(step)/float64(half))))
	}
	return out
}

// farthestMix is how far fg can move toward bg, from 0 to 1, and still keep
// the floor contrast. A color already under the floor cannot move at all.
func farthestMix(fg, bg string, floor float64) float64 {
	best := 0.0
	for f := 0.01; f <= 1; f += 0.01 {
		if contrastRatio(mixHex(fg, bg, f), bg) < floor {
			break
		}
		best = f
	}
	return best
}

// mixHex moves the #rrggbb color a the part f of the way toward b. When
// either one is not a #rrggbb color there is nothing to mix, so a comes back.
func mixHex(a, b string, f float64) string {
	ca, okA := hexRGB(a)
	cb, okB := hexRGB(b)
	if !okA || !okB {
		return a
	}
	var out [3]int
	for i := range out {
		out[i] = int(math.Round(float64(ca[i]) + (float64(cb[i])-float64(ca[i]))*f))
	}
	return fmt.Sprintf("#%02x%02x%02x", out[0], out[1], out[2])
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

// contrastRatio is the WCAG contrast of two #rrggbb colors, the same measure
// terminals use for their minimum contrast. A color that cannot be read
// gives 1, the lowest there is.
func contrastRatio(a, b string) float64 {
	lum := func(s string) (float64, bool) {
		c, ok := hexRGB(s)
		if !ok {
			return 0, false
		}
		var out float64
		for i, w := range []float64{0.2126, 0.7152, 0.0722} {
			v := float64(c[i]) / 255
			if v <= 0.03928 {
				v /= 12.92
			} else {
				v = math.Pow((v+0.055)/1.055, 2.4)
			}
			out += w * v
		}
		return out, true
	}
	la, okA := lum(a)
	lb, okB := lum(b)
	if !okA || !okB {
		return 1
	}
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
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
// on right after each reset and at the start of each line. A short line gets
// an erase to the end, so the background reaches the right edge. A full line
// gets none: the cursor still sits in its last column, and an erase there
// would clear the last cell, which is the right wall.
func (s styles) paintFrame(out string, width int) string {
	if s.bg == "" {
		return out
	}
	p := termenv.TrueColor
	set := termenv.CSI + p.Color(s.bg).Sequence(true) + "m" + termenv.CSI + p.Color(s.fg).Sequence(false) + "m"
	lines := strings.Split(out, "\n")
	for i, ln := range lines {
		painted := set + strings.ReplaceAll(ln, "\x1b[0m", "\x1b[0m"+set)
		if lipgloss.Width(ln) < width {
			painted += "\x1b[K"
		}
		lines[i] = painted
	}
	return strings.Join(lines, "\n")
}
