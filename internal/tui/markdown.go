package tui

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"

	"github.com/alecthomas/chroma/v2"
	chromastyles "github.com/alecthomas/chroma/v2/styles"
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

	// An image is blue and the words that name it are dim, so nothing out of
	// a glamour preset can reach the screen through them. Raw HTML is dim as
	// well, because it is not text the pane knows how to paint.
	cfg.Image.Color = slot(slotBlue)
	cfg.ImageText.Color = slot(slotDim)
	cfg.HTMLBlock = glamour.StyleBlock{StylePrimitive: glamour.StylePrimitive{Color: slot(slotDim)}}
	cfg.HTMLSpan = glamour.StyleBlock{StylePrimitive: glamour.StylePrimitive{Color: slot(slotDim)}}

	// The whole document and the code block around it take the body color, so
	// no preset gray from the glamour palette can reach the screen. A text
	// node gets no color of its own: it wears the color of the block it sits
	// in, so a heading keeps its yellow while a paragraph keeps the
	// foreground. Giving a text node the body color instead would paint
	// every heading in it the color of the paragraph.
	cfg.Document.Color = body
	cfg.CodeBlock.Color = body
	// The code colors go into a chroma style of this theme's own, so nothing
	// out of a glamour preset can reach the screen through a code block. The
	// chroma the preset brought along is cleared, because a chroma left on
	// the style makes glamour pick one name for these colors on its own, keep
	// the first set it is given for the whole process and hand the same set
	// to every theme drawn after it. The code block points at the name of the
	// style instead, and that name says which theme the colors belong to.
	cfg.CodeBlock.Chroma = nil
	cfg.CodeBlock.Theme = codeStyle(t, body)
	return cfg
}

// codeChromaName is the name a theme's code colors are filed under in chroma's
// style list. chroma keeps one list for the whole process, so the name has to
// say which theme the colors belong to, or two themes would end up sharing
// one set of them. The colors go into the name too, so a theme file that
// changed between two picks of it gets a name of its own and never wears the
// colors the file had before.
func codeChromaName(t theme.Theme) string {
	h := fnv.New32a()
	fmt.Fprintf(h, "%s %s %v", t.BG, t.FG, t.ANSI)
	return "acta-" + t.Name + "-" + strconv.FormatUint(uint64(h.Sum32()), 16)
}

// codeStyle registers the code colors of a theme as a chroma style of their
// own and gives the name back. Registering the same name twice would be
// harmless, because a name holds one set of colors, but the check keeps a
// second pick of a theme from writing over a style another renderer is
// reading right now.
func codeStyle(t theme.Theme, body *string) string {
	name := codeChromaName(t)
	chromaLock.Lock()
	defer chromaLock.Unlock()
	if _, ok := chromastyles.Registry[name]; !ok {
		// Every token the spec does not name is left out of the style, and
		// so wears the text color, which is the color of the body.
		chromastyles.Register(chroma.MustNewStyle(name, chroma.StyleEntries{
			chroma.Text:          chromaColor(body),
			chroma.Comment:       chromaSlot(t, slotDim),
			chroma.Keyword:       chromaSlot(t, slotMagenta),
			chroma.LiteralString: chromaSlot(t, slotGreen),
			chroma.LiteralNumber: chromaSlot(t, slotYellow),
			chroma.NameFunction:  chromaSlot(t, slotBlue),
		}))
	}
	return name
}

// chromaLock guards acta's own call that adds a theme to chroma's style list,
// so that two themes drawn for the first time in the same moment cannot both
// write that list at once. glamour reads the same list with no lock of its
// own, so this lock does not cover that side. The TUI draws on one goroutine,
// so nothing draws beside it today.
var chromaLock sync.Mutex

// chromaColor is a color as chroma reads it, and nothing at all when there is
// none, which is the case for the body of the terminal theme.
func chromaColor(c *string) string {
	if c == nil {
		return ""
	}
	return *c
}

// chromaSlot is a slot as chroma reads a color. chroma takes a hex value, so
// the plain ANSI number the terminal theme hands over would be read as a hex
// digit and turned down. The ANSI color names say the same sixteen colors in
// a way chroma understands, so a theme without a hex of its own still gets a
// code block in the slots the rest of the pane uses.
func chromaSlot(t theme.Theme, i int) string {
	if s := t.ANSI[i]; s != "" {
		return s
	}
	return ansiChroma[i]
}

// ansiChroma names the ANSI colors for chroma, which only takes hex values.
// The names must be the normal shade of the slot, because the terminal theme
// paints the rest of the pane in the normal shades too, and a fenced block
// that sits next to them may not pick the bright ones on its own.
var ansiChroma = map[int]string{
	slotDim:     "#ansidarkgray",
	slotGreen:   "#ansidarkgreen",
	slotYellow:  "#ansibrown",
	slotBlue:    "#ansidarkblue",
	slotMagenta: "#ansipurple",
}

// chromaFormatter picks the formatter that can carry the colors of this
// theme. A theme with a hex of its own gets truecolor. The terminal theme
// hands over ANSI names, and those come back as plain sixteen-color codes,
// because a truecolor sequence is exactly what it has no colors for.
func chromaFormatter(t theme.Theme) string {
	if t.BG == "" {
		return "terminal16"
	}
	return "terminal16m"
}
