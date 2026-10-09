package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"

	"github.com/iyay/acta/internal/theme"
)

// sampleDiff has one line of every kind `git show --stat -p` can print. The
// "-- gone" line is a removed line whose text starts with two dashes, so it
// reads like a file header but sits inside a hunk.
const sampleDiff = "commit 0123456789abcdef\n" +
	"Author: A <a@b.c>\n" +
	"\n" +
	"    feat: a thing\n" +
	"\n" +
	" a.go | 3 ++-\n" +
	" 1 file changed, 2 insertions(+), 1 deletion(-)\n" +
	"\n" +
	"diff --git a/a.go b/a.go\n" +
	"index 111..222 100644\n" +
	"--- a/a.go\n" +
	"+++ b/a.go\n" +
	"@@ -1,3 +1,4 @@ func main() {\n" +
	" context\n" +
	"-old\n" +
	"--- gone\n" +
	"+new\n" +
	"++ added\n" +
	"\n" +
	"+\tTabbed\n"

func diffTestStyles(t *testing.T) styles {
	t.Helper()
	th, ok := theme.Builtin("terminal")
	if !ok {
		t.Fatal("no terminal theme")
	}
	return newStyles(th, true)
}

func TestColorDiffStylesEveryLineKind(t *testing.T) {
	st := diffTestStyles(t)
	withTrueColor(func() {
		const (
			red   = "\x1b[31m"
			green = "\x1b[32m"
			cyan  = "\x1b[36m"
			bold  = "\x1b[1m"
		)
		want := map[string]string{
			"commit 0123456789abcdef": bold,
			"Author: A <a@b.c>":       "",
			"    feat: a thing":       "",
			" a.go | 3 ++-":           "",
			" 1 file changed, 2 insertions(+), 1 deletion(-)": "",
			"diff --git a/a.go b/a.go":                        bold,
			"index 111..222 100644":                           "",
			"--- a/a.go":                                      bold,
			"+++ b/a.go":                                      bold,
			"@@ -1,3 +1,4 @@ func main() {":                   cyan,
			" context":                                        "",
			"-old":                                            red,
			"--- gone":                                        red,
			"+new":                                            green,
			"++ added":                                        green,
			"":                                                "",
		}
		out := colorDiff(sampleDiff, 80, st)
		if len(out) != strings.Count(sampleDiff, "\n") {
			t.Fatalf("got %d lines, want %d", len(out), strings.Count(sampleDiff, "\n"))
		}
		for _, ln := range out {
			plain := xansi.Strip(ln)
			code, known := want[plain]
			if !known {
				if plain != "+       Tabbed" {
					t.Errorf("unexpected line %q", plain)
				}
				code = green
			}
			hasCode := strings.Contains(ln, "\x1b[")
			if code == "" {
				if hasCode {
					t.Errorf("plain line %q got a style: %q", plain, ln)
				}
				continue
			}
			if !strings.HasPrefix(ln, code) {
				t.Errorf("line %q = %q, want prefix %q", plain, ln, code)
			}
			// Exactly one style: one opening code, so no other color joins it.
			if n := strings.Count(ln, "\x1b["); n != 2 {
				t.Errorf("line %q has %d escape codes, want 2 (open and reset): %q", plain, n, ln)
			}
		}
	})
}

func TestColorDiffNeverWiderThanWidth(t *testing.T) {
	st := diffTestStyles(t)
	long := strings.Repeat("x", 200)
	text := "commit abc\n" +
		"diff --git a/f b/f\n--- a/f\n+++ b/f\n@@ -1 +1 @@\n" +
		"+" + long + "\n-" + long + "\n " + long + "\n" +
		"+日本語日本語日本語日本語\n-😀😀😀😀😀😀😀😀\n 日本語 and 😀 mixed\n" +
		"+\t\tdeep tab\n"
	withTrueColor(func() {
		for w := 1; w <= 40; w++ {
			out := colorDiff(text, w, st)
			if len(out) != strings.Count(text, "\n") {
				t.Fatalf("w=%d: got %d lines", w, len(out))
			}
			for _, ln := range out {
				if got := lipgloss.Width(ln); got > w {
					t.Fatalf("w=%d: line is %d cells wide: %q", w, got, ln)
				}
			}
		}
	})
}

func TestColorDiffCutKeepsStyleOnWideRunes(t *testing.T) {
	st := diffTestStyles(t)
	withTrueColor(func() {
		out := colorDiff("diff --git a/f b/f\n@@ -1 +1 @@\n+日本語\n", 4, st)
		// "+" is one cell, each rune two, so only one rune fits in 4 cells.
		if got := xansi.Strip(out[2]); got != "+日" {
			t.Fatalf("cut line = %q, want %q", got, "+日")
		}
		if !strings.HasPrefix(out[2], "\x1b[32m") {
			t.Fatalf("cut line lost its green: %q", out[2])
		}
	})
}

func TestColorDiffTrailingNewlineAddsNoLine(t *testing.T) {
	st := diffTestStyles(t)
	if got := colorDiff("a\nb\n", 10, st); len(got) != 2 {
		t.Fatalf("with trailing newline: %d lines, want 2", len(got))
	}
	if got := colorDiff("a\nb", 10, st); len(got) != 2 {
		t.Fatalf("without trailing newline: %d lines, want 2", len(got))
	}
	if got := colorDiff("", 10, st); len(got) != 0 {
		t.Fatalf("empty text: %d lines, want 0", len(got))
	}
}

func TestColorDiffExpandsTabs(t *testing.T) {
	st := diffTestStyles(t)
	out := colorDiff("+\tx\n", 80, st)
	if got := xansi.Strip(out[0]); got != "+       x" {
		t.Fatalf("tab expanded to %q, want %q", got, "+       x")
	}
}

func TestColorDiffWithThemeUsesThemeColors(t *testing.T) {
	th, ok := theme.Builtin("dracula")
	if !ok {
		t.Skip("no dracula theme")
	}
	st := newStyles(th, true)
	withTrueColor(func() {
		out := colorDiff("@@ -1 +1 @@\n-a\n+b\n", 20, st)
		for i, slot := range []int{6, 1, 2} {
			want := lipgloss.NewStyle().Foreground(lipgloss.Color(th.ANSI[slot])).Render(xansi.Strip(out[i]))
			if out[i] != want {
				t.Errorf("line %d = %q, want theme slot %d: %q", i, out[i], slot, want)
			}
		}
	})
}
