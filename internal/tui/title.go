// Package tui is the lazygit-style screen over a Board.
package tui

import (
	"os/exec"
	"runtime"

	"github.com/charmbracelet/lipgloss"
)

// titlePiece is one piece of a pane title: the pane number, a tab name, or the
// dashes between two names. tab is the tab the piece names, or -1 for a piece
// that names none; sep marks the dashes between two names. The view paints
// these pieces and the mouse counts their cells, so the click boxes and the
// drawn names cannot drift apart.
type titlePiece struct {
	text string
	tab  int
	sep  bool
}

// namedPieces lays out the title of pane p: the corner with the pane number,
// then the kept tab names joined by sep. keep nil means every name is kept; a
// dropped name leaves no piece behind at all, so it takes no room and draws
// no click box.
func namedPieces(p pane, names []string, keep []bool, sep string) []titlePiece {
	out := make([]titlePiece, 0, 2*len(names)+1)
	out = append(out, titlePiece{text: paneKey(p), tab: -1})
	first := true
	for i, name := range names {
		if keep != nil && !keep[i] {
			continue
		}
		if !first {
			out = append(out, titlePiece{text: sep, tab: -1, sep: true})
		}
		out = append(out, titlePiece{text: name, tab: i})
		first = false
	}
	return out
}

// titleWidth adds up the screen cells every piece takes.
func titleWidth(pieces []titlePiece) int {
	w := 0
	for _, p := range pieces {
		w += lipgloss.Width(p.text)
	}
	return w
}

// dropOrder gives the tab indices to drop one at a time when a title still
// does not fit: starting from the far (right) end, but skipping the tab that
// is open, since that one must never be the one that goes missing.
func dropOrder(n, on int) []int {
	out := make([]int, 0, max(0, n-1))
	for i := n - 1; i >= 0; i-- {
		if i != on {
			out = append(out, i)
		}
	}
	return out
}

// titlePieces works out which tab names actually fit a title of inner cells:
// first the full names with the usual " ─ " dashes, then the same names with
// a tighter single-space separator, and only then by dropping names (farthest
// from the open one first, which is never dropped) until what is left fits.
// paneTop and tabX both call this, so the letters the screen draws and the
// boxes the mouse checks can never drift apart.
func titlePieces(p pane, names []string, on, inner int) []titlePiece {
	for _, sep := range []string{" ─ ", " "} {
		pieces := namedPieces(p, names, nil, sep)
		if titleWidth(pieces) <= inner {
			return pieces
		}
	}
	keep := make([]bool, len(names))
	for i := range keep {
		keep[i] = true
	}
	pieces := namedPieces(p, names, keep, " ")
	for _, i := range dropOrder(len(names), on) {
		if titleWidth(pieces) <= inner {
			break
		}
		keep[i] = false
		pieces = namedPieces(p, names, keep, " ")
	}
	return pieces
}

// tabX gives where each tab name starts in the title of the pane whose left
// edge is at x and whose border is w cells wide, walked over the same pieces
// the view draws, so a click and the drawn name always come from the same
// numbers. A name the title drops keeps its slot in the result but gets a
// zero-width box, so no click can ever land on it.
func tabX(p pane, x, w int, names []string, on int) []tabBox {
	out := make([]tabBox, len(names))
	pos := x + 1 // the corner of the border sits before every piece
	for _, piece := range titlePieces(p, names, on, max(0, w-2)) {
		if piece.tab >= 0 {
			out[piece.tab] = tabBox{x: pos, w: lipgloss.Width(piece.text)}
		}
		pos += lipgloss.Width(piece.text)
	}
	return out
}

// WithVersion sets the build version the bottom line shows. Empty and
// "(devel)" both read as dev, the local-build word.
func (m Model) WithVersion(v string) Model {
	m.version = normalizeVersion(v)
	return m
}

// normalizeVersion keeps a tagged build number and reads every local build
// as dev, so the footer never shows a blank or a raw "(devel)".
func normalizeVersion(v string) string {
	if v == "" || v == "(devel)" {
		return "dev"
	}
	return v
}

// defaultOpen hands a link to the platform browser: open on macOS, xdg-open
// everywhere else.
func defaultOpen(url string) error {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", url).Run()
	}
	return exec.Command("xdg-open", url).Run()
}

// linkAt gives the url under a bottom-line cell, reading the boxes off the
// line the view drew, so a click opens only the link it lands on.
func (m Model) linkAt(x int) (string, bool) {
	for _, b := range m.statusLineBoxes() {
		if x >= b.x && x < b.x+b.w && b.url != "" {
			return b.url, true
		}
	}
	return "", false
}
