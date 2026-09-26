// Package tui is the lazygit-style screen over a Board.
package tui

import (
	"fmt"
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

// titlePieces lays out the title of pane p: the corner with the pane number,
// then the tab names joined by " ─ ".
func titlePieces(p pane, names []string) []titlePiece {
	out := make([]titlePiece, 0, 2*len(names)+1)
	out = append(out, titlePiece{text: fmt.Sprintf("─[%d]─", p+1), tab: -1})
	for i, name := range names {
		if i > 0 {
			out = append(out, titlePiece{text: " ─ ", tab: -1, sep: true})
		}
		out = append(out, titlePiece{text: name, tab: i})
	}
	return out
}

// tabX gives where each tab name starts in the title of the pane whose left
// edge is at x, walked over the same pieces the view draws, so a click and the
// drawn name always come from the same numbers.
func tabX(p pane, x int, names []string) []tabBox {
	x++ // the corner of the border sits before every piece
	out := make([]tabBox, 0, len(names))
	for _, piece := range titlePieces(p, names) {
		if piece.tab >= 0 {
			out = append(out, tabBox{x: x, w: lipgloss.Width(piece.text)})
		}
		x += lipgloss.Width(piece.text)
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
