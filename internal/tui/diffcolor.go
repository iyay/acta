package tui

import (
	"fmt"
	"regexp"
	"strconv"
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
//
// Lines inside a hunk get a dim gutter with the old and the new line number.
// Every other line gets a blank gutter of the same width, so the text lines up.
func colorDiff(text string, w int, st styles) []string {
	if text == "" {
		return nil
	}
	bold := lipgloss.NewStyle().Bold(true)
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	brushes := make([]*lipgloss.Style, len(lines))
	nums := make([]lineNums, len(lines))
	// A "---" or "+++" line is a file header only before the first hunk of a
	// file. Inside a hunk it is a removed or added line whose text starts with
	// dashes or pluses, so it keeps the red or green.
	inHunk := false
	oldNo, newNo, widest := 0, 0, 0
	for i, ln := range lines {
		ln = strings.TrimSuffix(ln, "\r")
		lines[i] = ln
		if m := hunkHead.FindStringSubmatch(ln); m != nil {
			oldNo, _ = strconv.Atoi(m[1])
			newNo, _ = strconv.Atoi(m[2])
		} else if inHunk {
			nums[i], oldNo, newNo = numberLine(ln, oldNo, newNo)
			widest = max(widest, nums[i].old, nums[i].new)
		}
		brushes[i], inHunk = diffBrush(ln, inHunk, st, bold)
	}
	// The gutter is as wide as the biggest number in the whole diff. When w
	// cannot hold it and one cell of text, the numbers are left out.
	digits := len(strconv.Itoa(widest))
	gutterW := 0
	if widest > 0 && w >= 2*digits+3 {
		gutterW = 2*digits + 2
	}
	out := make([]string, len(lines))
	for i, ln := range lines {
		gutter := ""
		if gutterW > 0 {
			gutter = renderGutter(nums[i], digits, st)
		}
		out[i] = gutter + paintDiffLine(ln, w-gutterW, brushes[i])
	}
	return out
}

// lineNums holds the old and new line number of one diff line. Zero means the
// line has no number on that side.
type lineNums struct{ old, new int }

// hunkHead reads the two start lines out of a hunk head such as
// "@@ -5,3 +7 @@". A missing count after the comma does not matter here.
var hunkHead = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)`)

// numberLine gives the numbers of one line inside a hunk and moves the two
// counters on. A context line counts on both sides, "-" on the old side and
// "+" on the new side. Anything else, such as "\ No newline at end of file" or
// an empty line, has no number and moves nothing.
func numberLine(ln string, oldNo, newNo int) (lineNums, int, int) {
	if ln == "" {
		return lineNums{}, oldNo, newNo
	}
	switch ln[0] {
	case ' ':
		return lineNums{oldNo, newNo}, oldNo + 1, newNo + 1
	case '-':
		return lineNums{old: oldNo}, oldNo + 1, newNo
	case '+':
		return lineNums{new: newNo}, oldNo, newNo + 1
	}
	return lineNums{}, oldNo, newNo
}

// renderGutter draws "old new " with both numbers right-aligned to digits
// cells. A side without a number stays blank, and a line with no number at
// all is plain spaces with no color.
func renderGutter(n lineNums, digits int, st styles) string {
	side := func(v int) string {
		if v == 0 {
			return strings.Repeat(" ", digits)
		}
		return fmt.Sprintf("%*d", digits, v)
	}
	g := side(n.old) + " " + side(n.new) + " "
	if n.old == 0 && n.new == 0 {
		return g
	}
	return st.faint.Render(g)
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
