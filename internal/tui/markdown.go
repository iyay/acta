package tui

import (
	"strconv"

	glamour "github.com/charmbracelet/glamour/ansi"
	preset "github.com/charmbracelet/glamour/styles"

	"github.com/iyay/acta/internal/theme"
)

// markdownStyle builds the glamour style of the detail pane out of the theme
// the rest of the screen already uses, so the markdown reads as part of the
// same window instead of a guest with colors of its own. It copies the glamour
// preset for the margins, the bullets and the list layout, and changes only
// the colors and the markers: a color says what a thing is, so the markers
// that repeat it are dropped.
func markdownStyle(t theme.Theme, dark bool) glamour.StyleConfig {
	cfg := preset.DarkStyleConfig
	if !t.Dark(dark) {
		cfg = preset.LightStyleConfig
	}

	// slot gives the theme's own hex, or the plain ANSI number for the
	// terminal theme so the terminal picks the color.
	slot := func(i int) *string {
		s := strconv.Itoa(i)
		if t.BG != "" {
			s = t.ANSI[i]
		}
		return &s
	}
	// body is the theme foreground, or no color at all for the terminal theme,
	// because it has no foreground of its own either.
	var body *string
	if t.FG != "" {
		fg := t.FG
		body = &fg
	}
	yes := true
	// A heading is yellow and bold, with no bar behind it and no # in front.
	// The margins and spacing the preset chose stay as they are.
	for _, h := range []*glamour.StyleBlock{
		&cfg.Heading, &cfg.H1, &cfg.H2, &cfg.H3, &cfg.H4, &cfg.H5, &cfg.H6,
	} {
		h.Color = slot(slotYellow)
		h.BackgroundColor = nil
		h.Prefix, h.Suffix = "", ""
		h.Bold = &yes
	}
	cfg.Strong.Color = slot(slotRed)
	cfg.Strong.Bold = &yes
	cfg.Emph = glamour.StylePrimitive{Color: body, Italic: &yes}

	// Inline code is green, with no box and no padding spaces around it.
	cfg.Code = glamour.StyleBlock{
		StylePrimitive: glamour.StylePrimitive{Color: slot(slotGreen)},
	}

	// A link and the words inside it are both blue.
	cfg.Link.Color = slot(slotBlue)
	cfg.LinkText.Color = slot(slotBlue)

	// A quote and a rule are dim, so they sit back behind the text.
	cfg.BlockQuote.Color = slot(slotDim)
	cfg.HorizontalRule.Color = slot(slotDim)

	// The whole document and the code block around it take the body color, so
	// no preset gray from the glamour palette can reach the screen. A text
	// node gets no color of its own: it wears the color of the block it sits
	// in, so a heading keeps its yellow while a paragraph keeps the
	// foreground. Giving a text node the body color instead would paint
	// every heading in it the color of the paragraph.
	cfg.Document.Color = body
	cfg.CodeBlock.Color = body
	// The preset's chroma is a pointer every renderer in the process shares, so
	// a chroma of our own is the only way to give code blocks theme colors
	// without recoloring somebody else's code block.
	cfg.CodeBlock.Chroma = &glamour.Chroma{
		Comment:       glamour.StylePrimitive{Color: slot(slotDim)},
		Keyword:       glamour.StylePrimitive{Color: slot(slotMagenta)},
		LiteralString: glamour.StylePrimitive{Color: slot(slotGreen)},
		LiteralNumber: glamour.StylePrimitive{Color: slot(slotYellow)},
		NameFunction:  glamour.StylePrimitive{Color: slot(slotBlue)},
		Text:          glamour.StylePrimitive{Color: body},
	}
	return cfg
}
