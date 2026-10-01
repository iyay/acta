package tui

import (
	"strconv"
	"strings"
	"testing"

	glamour "github.com/charmbracelet/glamour/ansi"
	preset "github.com/charmbracelet/glamour/styles"

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
		// body is the body color. Empty means the body gets no color at all,
		// so whatever the terminal already paints shows through.
		body string
	}{
		{"tokyo-night", hex, func(i int) string { return hex.ANSI[i] }, hex.FG},
		{"terminal", plain, func(i int) string { return strconv.Itoa(i) }, ""},
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

			// Code blocks take their colors from a chroma of our own, because
			// the preset's is shared with every other renderer in the process.
			if cfg.CodeBlock.Chroma == nil {
				t.Fatal("code block has no chroma")
			}
			if cfg.CodeBlock.Chroma == preset.DarkStyleConfig.CodeBlock.Chroma ||
				cfg.CodeBlock.Chroma == preset.LightStyleConfig.CodeBlock.Chroma {
				t.Error("code block shares the preset chroma instead of its own")
			}
			for _, tok := range []struct {
				name string
				got  glamour.StylePrimitive
				want string
			}{
				{"chroma comment", cfg.CodeBlock.Chroma.Comment, tc.slot(slotDim)},
				{"chroma keyword", cfg.CodeBlock.Chroma.Keyword, tc.slot(slotMagenta)},
				{"chroma string", cfg.CodeBlock.Chroma.LiteralString, tc.slot(slotGreen)},
				{"chroma number", cfg.CodeBlock.Chroma.LiteralNumber, tc.slot(slotYellow)},
				{"chroma function", cfg.CodeBlock.Chroma.NameFunction, tc.slot(slotBlue)},
				{"chroma other text", cfg.CodeBlock.Chroma.Text, tc.body},
			} {
				wantColor(tok.name, tok.got.Color, tok.want)
			}
		})
	}
}
