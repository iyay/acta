package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
	chromastyles "github.com/alecthomas/chroma/v2/styles"
	glamour "github.com/charmbracelet/glamour/ansi"

	"github.com/muesli/termenv"

	"github.com/iyay/acta/internal/theme"
)

// TestRendererPaintsThemeColors checks what the screen actually gets, not
// what the style config says: the colors must come from the theme and the
// markdown markers must be gone.
func TestRendererPaintsThemeColors(t *testing.T) {
	t.Parallel()

	const md = "# H\n\n**b** and `c`"

	hex, ok := theme.Builtin("tokyo-night")
	if !ok {
		t.Fatal("tokyo-night theme is missing")
	}
	plain, ok := theme.Builtin("terminal")
	if !ok {
		t.Fatal("terminal theme is missing")
	}

	t.Run("hex theme", func(t *testing.T) {
		t.Parallel()

		out := newRenderer(hex, true)(md, 80)

		// The inline code wears the theme green as a truecolor sequence.
		// The last step is 105, not the 106 of #9ece6a, because termenv turns
		// the hex back into numbers and cuts the fraction off.
		if !strings.Contains(out, "\x1b[38;2;158;206;105m") {
			t.Errorf("output has no theme green:\n%q", out)
		}
		// A 38;5; sequence would be glamour's own 256-color palette, which
		// is no color a theme holds.
		if strings.Contains(out, "38;5;") {
			t.Errorf("output still carries a 256-color sequence:\n%q", out)
		}
		wantNoMarkers(t, out)
	})

	t.Run("heading keeps the theme yellow", func(t *testing.T) {
		t.Parallel()

		out := newRenderer(hex, true)("# H\n\nplain body\n\n## Two", 80)

		// The heading wears the theme yellow as a truecolor sequence. The
		// last step is 104, not a rounded value, because the theme hex
		// #e0af68 turns into those three numbers whole.
		for _, word := range []string{"H", "Two"} {
			want := "\x1b[38;2;224;175;104;1m" + word + "\x1b[0m"
			if !strings.Contains(out, want) {
				t.Errorf("heading %q is not painted %q:\n%q", word, want, out)
			}
		}
		// Plain body text still wears the theme foreground, the same
		// #c0caf5 the rest of the pane uses.
		if !strings.Contains(out, "\x1b[38;2;192;202;245mplain") {
			t.Errorf("body text is not painted the theme foreground:\n%q", out)
		}
	})

	t.Run("terminal heading is plain yellow", func(t *testing.T) {
		t.Parallel()

		out := newRenderer(plain, true)("# H\n\nplain body\n\n## Two", 80)

		// The terminal theme has no hex of its own, so the heading is the
		// plain ANSI yellow and the terminal picks the shade.
		for _, word := range []string{"H", "Two"} {
			want := "\x1b[33;1m" + word + "\x1b[0m"
			if !strings.Contains(out, want) {
				t.Errorf("heading %q is not painted %q:\n%q", word, want, out)
			}
		}
		// The terminal theme has no foreground either, so the body line
		// carries no color of its own.
		if line := lineWith(t, strings.Split(out, "\n"), "plain body"); strings.Contains(line, "\x1b[") {
			t.Errorf("body text carries a color the terminal theme has no hex for: %q", line)
		}
	})

	t.Run("terminal theme", func(t *testing.T) {
		t.Parallel()

		out := newRenderer(plain, true)(md, 80)

		// The terminal theme hands over an ANSI number, so the inline code
		// is plain green and the terminal picks the shade.
		if !strings.Contains(out, "\x1b[32m") {
			t.Errorf("output has no plain green:\n%q", out)
		}
		if strings.Contains(out, "38;2;") {
			t.Errorf("output carries a truecolor sequence the terminal cannot use:\n%q", out)
		}
		wantNoMarkers(t, out)
	})
}

// wantNoMarkers fails when the raw markdown reached the screen, because the
// color of a thing already says what it is.
func wantNoMarkers(t *testing.T, out string) {
	t.Helper()
	for _, mark := range []string{"**", "`", "# "} {
		if strings.Contains(out, mark) {
			t.Errorf("output still shows the marker %q:\n%q", mark, out)
		}
	}
}

// goBlock is a fenced Go block with one token of every kind the spec maps: a
// comment, a keyword, a string, a number and a function name.
const goBlock = "```go\n// note\nfunc main() { s := \"hi\"; n := 42 }\n```"

// codeTokens names one word of the code block for each kind of token the
// spec paints, and the slot that paints it. A test reads the color code in
// front of the word, so it can say which token it is looking at.
var codeTokens = []struct {
	word string
	slot int
}{
	{"// note", slotDim},
	{"func", slotMagenta},
	{`"hi"`, slotGreen},
	{"42", slotYellow},
	{"main", slotBlue},
}

// codeBefore gives the numbers inside the color code that sits right in
// front of a word, so a test reads the color of that word and not another
// color the same line happens to hold. A word can also turn up inside a
// color code, as the 42 in 38;2;248;248;242 does, so the code and the word
// have to sit next to each other for the pair to count. It is empty when
// nothing colors the word.
func codeBefore(t *testing.T, out, word string) string {
	t.Helper()
	found := regexp.MustCompile(`\x1b\[([0-9;]*)m`+regexp.QuoteMeta(word)).FindAllStringSubmatch(out, -1)
	if len(found) == 0 {
		return ""
	}
	return found[len(found)-1][1]
}

// chromaCode gives the numbers chroma writes for a code token. chroma reads a
// theme hex as three bytes and writes those bytes as they are, so the numbers
// are the hex digits themselves.
func chromaCode(hex string) string {
	return fmt.Sprintf("38;2;%d;%d;%d", hexByte(hex, 1), hexByte(hex, 3), hexByte(hex, 5))
}

// hexByte reads one pair of digits out of a hex color.
func hexByte(hex string, at int) int {
	n, _ := strconv.ParseUint(hex[at:at+2], 16, 8)
	return int(n)
}

// glamourCode gives the numbers glamour writes for an element it paints
// itself, such as an image. It asks termenv, which is what glamour asks, and
// termenv turns the hex back into numbers and cuts the fraction off, so the
// last of the three can be one lower than the hex digit. Code tokens are not
// like this: chroma writes the hex bytes whole, which is what chromaCode
// reads.
func glamourCode(color string) string {
	return termenv.TrueColor.Color(color).Sequence(false)
}

// plainCode is the plain code the terminal theme paints a slot with, because
// the terminal has no hex of its own and picks the shade. The first eight
// slots are the dim ones and the rest are the bright ones.
func plainCode(slot int) string {
	if slot < 8 {
		return strconv.Itoa(30 + slot)
	}
	return strconv.Itoa(90 + slot - 8)
}

// ansiToken is the code chroma hands back for the name of an ANSI color, and
// the shade is the normal one, because the rest of the terminal theme paints
// its slots in the normal shades too. The map is keyed by the word of the code
// block, because that is what a test looks the color up by.
var ansiToken = map[string]string{
	"// note": "90",
	"func":    "35",
	`"hi"`:    "32",
	"42":      "33",
	"main":    "34",
}

// TestEachThemeKeepsItsOwnCodeColors walks every theme that ships with acta
// and asks each one for the same code block, one after another in this one
// process. It runs without t.Parallel, because chroma reads its style list
// with no lock of its own while it paints, and the race detector calls that a
// race the moment a second theme draws beside it.
func TestEachThemeKeepsItsOwnCodeColors(t *testing.T) {
	for _, name := range theme.Names() {
		t.Run(name, func(t *testing.T) {
			th, ok := theme.Builtin(name)
			if !ok {
				t.Fatalf("%s theme is missing", name)
			}
			out := newRenderer(th, true)(goBlock, 80)
			// A 38;5; sequence is glamour's own 256-color palette, which is
			// no color a theme holds.
			if strings.Contains(out, "38;5;") {
				t.Errorf("code block still carries a 256-color sequence:\n%q", out)
			}
			for _, tok := range codeTokens {
				// The terminal theme has no hex of its own, so the token
				// has to ask for a plain code instead of a truecolor one.
				want, plain := ansiToken[tok.word], true
				if th.BG != "" {
					want, plain = chromaCode(th.ANSI[tok.slot]), false
				}
				got := codeBefore(t, out, tok.word)
				if !strings.HasPrefix(got, want) {
					t.Errorf("code token %q is painted %q, want %q:\n%q", tok.word, got, want, out)
				}
				if plain && strings.Contains(got, ";") {
					t.Errorf("code token %q is painted %q, which is not one plain code:\n%q", tok.word, got, out)
				}
			}
		})
	}
}

// TestNoGlamourPresetColorSurvives feeds the elements the spec does not map
// through the renderer and reads the screen, because glamour's preset still
// holds colors of its own for them and every color it paints is a color
// reaching the pane. The dark preset and the light preset are both checked,
// since they do not hold the same numbers, and the terminal theme with them.
func TestNoGlamourPresetColorSurvives(t *testing.T) {
	for _, tc := range []struct {
		md string
		// word is a piece of the output to read, and slot is the theme slot
		// it has to wear. A word is only read when it is there, so an
		// element glamour drops does not fail the test.
		word string
		slot int
	}{
		// The words that name an image are dim and the image itself is
		// blue, both from the theme.
		{"![alt](http://x)", "Image:", slotDim},
		{"![alt](http://x)", "http://x", slotBlue},
		// A block of raw HTML is dim. An inline tag is not readable here:
		// glamour strips the tag before it draws anything, so the words
		// inside it come out as ordinary text, and the test below says so.
		{"<div>\nhi\n</div>\n", "hi", slotDim},
	} {
		for _, name := range []string{"tokyo-night", "catppuccin-latte", "terminal"} {
			t.Run(name+"/"+tc.md, func(t *testing.T) {
				th, ok := theme.Builtin(name)
				if !ok {
					t.Fatalf("%s theme is missing", name)
				}
				out := newRenderer(th, true)(tc.md, 80)
				// The numbers below are glamour's own, out of its dark and
				// its light preset: 212 and 205 for an image, 243 for the
				// words that name it.
				for _, presetColor := range []string{"38;5;212", "38;5;205", "38;5;243"} {
					if strings.Contains(out, presetColor) {
						t.Errorf("output still carries the glamour preset color %q:\n%q", presetColor, out)
					}
				}
				// The terminal theme has no hex of its own, so the terminal
				// picks the shade. Any other theme is drawn by glamour
				// through termenv, which turns the hex back into numbers
				// and cuts the fraction off.
				want := plainCode(tc.slot)
				if th.BG != "" {
					want = glamourCode(th.ANSI[tc.slot])
				}
				got := codeBefore(t, out, tc.word)
				if !strings.HasPrefix(got, want) {
					t.Errorf("%q is painted %q, want the theme slot %d as %q:\n%q", tc.word, got, tc.slot, want, out)
				}
			})
		}
	}
}

// TestAnInlineTagLosesItsTagsAndKeepsItsWords says what actually happens to
// an inline HTML tag, so nobody reads the dim color of an HTML span as
// something the test proved. glamour strips the tag before it draws, so what
// is left is ordinary text and it wears the body color, not a color of its
// own and not a color out of a glamour preset.
func TestAnInlineTagLosesItsTagsAndKeepsItsWords(t *testing.T) {
	th, ok := theme.Builtin("tokyo-night")
	if !ok {
		t.Fatal("tokyo-night theme is missing")
	}
	out := newRenderer(th, true)("a <span class=\"x\">bold</span> c", 80)
	for _, gone := range []string{"<span", "</span>", "class"} {
		if strings.Contains(out, gone) {
			t.Errorf("the tag %q reached the screen:\n%q", gone, out)
		}
	}
	// The pane draws a margin around the words, so it is cut off here.
	if got := strings.TrimSpace(plain(lineWith(t, strings.Split(out, "\n"), "bold"))); got != "a bold c" {
		t.Errorf("the words around the tag are %q, want %q", got, "a bold c")
	}
	// The words are body text, so they wear the theme foreground and
	// nothing else.
	if got := codeBefore(t, out, "bold"); got != glamourCode(th.FG) {
		t.Errorf("the words inside the tag are painted %q, want the body color %q:\n%q", got, glamourCode(th.FG), out)
	}
}

// TestCodeBlockTakesThemeColors reads what a fenced code block paints for one
// theme of each kind. It runs without t.Parallel, because chroma reads its
// style list with no lock of its own while it paints, and the race detector
// calls that a race the moment another test draws a theme beside it.
func TestCodeBlockTakesThemeColors(t *testing.T) {
	hex, ok := theme.Builtin("tokyo-night")
	if !ok {
		t.Fatal("tokyo-night theme is missing")
	}
	plain, ok := theme.Builtin("terminal")
	if !ok {
		t.Fatal("terminal theme is missing")
	}

	t.Run("hex theme", func(t *testing.T) {
		out := newRenderer(hex, true)(goBlock, 80)
		// A 38;5; sequence is glamour's own 256-color palette, which is no
		// color the theme holds.
		if strings.Contains(out, "38;5;") {
			t.Errorf("code block still carries a 256-color sequence:\n%q", out)
		}
		// The comment wears the dim slot and the keyword the magenta slot,
		// each as a truecolor sequence. chroma writes the hex bytes whole,
		// so the digits here are the hex digits of the theme itself.
		for _, want := range []string{
			"\x1b[38;2;65;72;104m// note",
			"\x1b[38;2;187;154;247mfunc",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("code block is missing %q:\n%q", want, out)
			}
		}
		// The keyword and the function name on the same line wear different
		// colors, so the block was token by token and not flattened to one.
		line := lineWith(t, strings.Split(out, "\n"), "main")
		if !strings.Contains(line, "\x1b[38;2;187;154;247mfunc") ||
			!strings.Contains(line, "\x1b[38;2;122;162;247mmain") {
			t.Errorf("keyword and function name are not two colors of their own: %q", line)
		}
	})

	t.Run("terminal theme", func(t *testing.T) {
		out := newRenderer(plain, true)(goBlock, 80)
		// The terminal theme has no hex, so the block must ask for plain
		// sixteen-color codes and nothing else.
		for _, unwanted := range []string{"38;2;", "38;5;"} {
			if strings.Contains(out, unwanted) {
				t.Errorf("code block carries %q, which the terminal theme cannot use:\n%q", unwanted, out)
			}
		}
		// The comment is dim and the keyword magenta, as plain codes.
		for _, want := range []string{"\x1b[90m// note", "\x1b[35mfunc"} {
			if !strings.Contains(out, want) {
				t.Errorf("code block is missing %q:\n%q", want, out)
			}
		}
		// A bright code is a shade no slot names, so no code in the block may
		// be one. Tokens the table above does not name count too, which is
		// why the whole block is read and not only those five words.
		for _, code := range regexp.MustCompile(`\x1b\[([0-9;]*)m`).FindAllStringSubmatch(out, -1) {
			for _, part := range strings.Split(code[1], ";") {
				n, _ := strconv.Atoi(part)
				if n >= 91 && n <= 97 {
					t.Errorf("code block carries the bright code %d, which no slot names:\n%q", n, out)
				}
			}
		}
	})
}

// wantHex is the hex chroma ends up with for a color a theme asked for by the
// name of an ANSI color, and the color itself when it asked by hex. An empty
// color stays empty, because a token with no color of its own has none.
func wantHex(color string) string {
	if color == "" {
		return ""
	}
	return chroma.ParseColour(color).String()
}

func TestMarkdownStyleUsesThemeSlots(t *testing.T) {
	t.Parallel()

	hex, ok := theme.Builtin("tokyo-night")
	if !ok {
		t.Fatal("tokyo-night theme is missing")
	}
	plain, ok := theme.Builtin("terminal")
	if !ok {
		t.Fatal("terminal theme is missing")
	}

	for _, tc := range []struct {
		name string
		th   theme.Theme
		// slot is the color the theme paints a slot with: its own hex, or the
		// plain ANSI number when the theme has no hex of its own.
		slot func(int) string
		// chromaSlot is the same slot as chroma takes it. chroma only reads
		// hex values, so the terminal theme hands over the name of the ANSI
		// color instead of its number.
		chromaSlot func(int) string
		// body is the body color. Empty means the body gets no color at all,
		// so whatever the terminal already paints shows through.
		body string
	}{
		{"tokyo-night", hex, func(i int) string { return hex.ANSI[i] }, func(i int) string { return hex.ANSI[i] }, hex.FG},
		{"terminal", plain, func(i int) string { return strconv.Itoa(i) }, func(i int) string { return ansiChroma[i] }, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := markdownStyle(tc.th, true)

			// wantColor reads a glamour color, which is a pointer so that no
			// color at all is different from an empty one.
			wantColor := func(what string, got *string, want string) {
				t.Helper()
				if want == "" {
					if got != nil {
						t.Errorf("%s color is %q, want no color", what, *got)
					}
					return
				}
				if got == nil {
					t.Errorf("%s has no color, want %q", what, want)
					return
				}
				if *got != want {
					t.Errorf("%s color is %q, want %q", what, *got, want)
				}
			}

			// Body text wears the theme foreground, or nothing at all on the
			// terminal theme, and it wears the color of the block it sits in.
			wantColor("document", cfg.Document.Color, tc.body)
			wantColor("code block", cfg.CodeBlock.Color, tc.body)
			// A text node must have no color of its own, because glamour lets
			// the node's color win over the block around it. A body color
			// here would paint every heading the color of a paragraph.
			if cfg.Text.Color != nil {
				t.Errorf("text color is %q, want none so a text node wears its block", *cfg.Text.Color)
			}

			// Every heading is yellow, bold and bare: no bar behind it and no
			// # marker in front, because the color already says it is one.
			for _, h := range []struct {
				name  string
				block glamour.StyleBlock
			}{
				{"heading", cfg.Heading},
				{"h1", cfg.H1},
				{"h2", cfg.H2},
				{"h3", cfg.H3},
				{"h4", cfg.H4},
				{"h5", cfg.H5},
				{"h6", cfg.H6},
			} {
				wantColor(h.name, h.block.Color, tc.slot(slotYellow))
				if h.block.BackgroundColor != nil {
					t.Errorf("%s background is %q, want none", h.name, *h.block.BackgroundColor)
				}
				if h.block.Prefix != "" {
					t.Errorf("%s prefix is %q, want empty", h.name, h.block.Prefix)
				}
				if h.block.Suffix != "" {
					t.Errorf("%s suffix is %q, want empty", h.name, h.block.Suffix)
				}
				if h.block.Bold == nil || !*h.block.Bold {
					t.Errorf("%s is not bold", h.name)
				}
			}

			// Bold is red, the slot the rest of the TUI uses to say something
			// matters.
			wantColor("strong", cfg.Strong.Color, tc.slot(slotRed))
			if cfg.Strong.Bold == nil || !*cfg.Strong.Bold {
				t.Error("strong is not bold")
			}

			// Italic stays italic and wears the body color, so leaning on a
			// word does not change what it means.
			wantColor("emph", cfg.Emph.Color, tc.body)
			if cfg.Emph.Italic == nil || !*cfg.Emph.Italic {
				t.Error("emph is not italic")
			}

			// Inline code is green with no box and no padding spaces.
			wantColor("code", cfg.Code.Color, tc.slot(slotGreen))
			if cfg.Code.BackgroundColor != nil {
				t.Errorf("code background is %q, want none", *cfg.Code.BackgroundColor)
			}
			if cfg.Code.Prefix != "" || cfg.Code.Suffix != "" {
				t.Errorf("code padding is %q%q, want empty", cfg.Code.Prefix, cfg.Code.Suffix)
			}

			// A link and the words inside it are both blue.
			wantColor("link", cfg.Link.Color, tc.slot(slotBlue))
			wantColor("link text", cfg.LinkText.Color, tc.slot(slotBlue))

			// A quote and a rule are dim, so they sit back behind the text.
			wantColor("block quote", cfg.BlockQuote.Color, tc.slot(slotDim))
			wantColor("horizontal rule", cfg.HorizontalRule.Color, tc.slot(slotDim))

			// A code block names a chroma style of its own instead of
			// carrying a chroma, because glamour files a chroma under one
			// name it picks itself and keeps the first set for the whole
			// process. Every token the spec names has to be in that style,
			// in the theme's own slots.
			if cfg.CodeBlock.Chroma != nil {
				t.Errorf("code block carries a chroma, which glamour would file under one name for the whole process")
			}
			if cfg.CodeBlock.Theme != codeChromaName(tc.th) {
				t.Errorf("code block names the chroma style %q, want one of its own theme %q", cfg.CodeBlock.Theme, codeChromaName(tc.th))
			}
			style, ok := chromastyles.Registry[cfg.CodeBlock.Theme]
			if !ok {
				t.Fatalf("the chroma style %q is not registered", cfg.CodeBlock.Theme)
			}
			for _, tok := range []struct {
				name  string
				token chroma.TokenType
				want  string
			}{
				{"chroma comment", chroma.Comment, tc.chromaSlot(slotDim)},
				{"chroma keyword", chroma.Keyword, tc.chromaSlot(slotMagenta)},
				{"chroma string", chroma.LiteralString, tc.chromaSlot(slotGreen)},
				{"chroma number", chroma.LiteralNumber, tc.chromaSlot(slotYellow)},
				{"chroma function", chroma.NameFunction, tc.chromaSlot(slotBlue)},
				{"chroma other text", chroma.Text, tc.body},
			} {
				// chroma turns the name of an ANSI color into the hex of
				// that color, so a token the terminal theme paints by name
				// reads back as hex. A token with no color of its own has
				// no color at all.
				var got *string
				if c := style.Get(tok.token).Colour; c.IsSet() {
					hex := c.String()
					got = &hex
				}
				wantColor(tok.name, got, wantHex(tok.want))
			}
		})
	}
}
