package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

// colorDiff paints the lines of `git show --stat -p` output and cuts each one
// to w cells. It gives back one line for each line of text, so a caller can
// scroll by line number. A newline at the very end adds no empty line.
//
// The colors come from the styles the rest of the TUI already paints with, so
// a new theme changes the diff too: green for added, red for removed, cyan for
// a hunk head, bold for the commit and file headers, the rest plain.
func colorDiff(text string, w int, st styles) []string {
	if text == "" {
		return nil
	}
	bold := lipgloss.NewStyle().Bold(true)
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	out := make([]string, len(lines))
	// A "---" or "+++" line is a file header only before the first hunk of a
	// file. Inside a hunk it is a removed or added line whose text starts with
	// dashes or pluses, so it keeps the red or green.
	inHunk := false
	for i, ln := range lines {
		ln = strings.TrimSuffix(ln, "\r")
		brush, nextInHunk := diffBrush(ln, inHunk, st, bold)
		inHunk = nextInHunk
		out[i] = paintDiffLine(ln, w, brush)
	}
	return out
}

// diffBrush picks the one style of a line, or nil for a plain line. It also
// says whether the next line is inside a hunk.
func diffBrush(ln string, inHunk bool, st styles, bold lipgloss.Style) (brush *lipgloss.Style, nowInHunk bool) {
	switch {
	case strings.HasPrefix(ln, "diff --git"):
		return &bold, false
	case strings.HasPrefix(ln, "commit "):
		return &bold, false
	case strings.HasPrefix(ln, "@@"):
		return &st.label, true
	case !inHunk && (strings.HasPrefix(ln, "--- ") || strings.HasPrefix(ln, "+++ ")):
		return &bold, inHunk
	case strings.HasPrefix(ln, "+"):
		return &st.done, inHunk
	case strings.HasPrefix(ln, "-"):
		return &st.problem, inHunk
	}
	return nil, inHunk
}

// paintDiffLine expands tabs, cuts the line to w cells and only then paints
// it. Cutting the plain text first means a color code is never split, and a
// wide letter that does not fit is dropped whole instead of being half drawn.
func paintDiffLine(ln string, w int, brush *lipgloss.Style) string {
	if w <= 0 {
		return ""
	}
	cut := xansi.Truncate(expandTabs(ln), w, "")
	if brush == nil || cut == "" {
		return cut
	}
	return brush.Render(cut)
}
