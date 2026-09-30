package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/board"
)

// helpHint is the one hint the status line never drops, so the full key map
// is always one key away.
const helpHint = "Help: ?"

// hints lists the keys the focused pane can use right now, most useful first,
// the way lazygit shows them on its bottom line.
func (m Model) hints() []string {
	it := m.Selected()
	// The status popup refuses tasks and debt lines, so those rows offer the
	// tick keys in its place.
	tick := it != nil && (it.Kind == board.KindTask || it.Kind == board.KindDebtItem)
	// The popup and the editor also turn the group row, an item shown from
	// another worktree, a legacy file and a branch that is not checked out
	// away, so those rows get only the keys that work anywhere.
	own := it != nil && it.OnDisk && it.Worktree == "" && !it.Legacy
	var out []string
	if tick && own {
		out = append(out, "Tick: +", "Untick: -")
	}
	if m.focus == paneDetail {
		out = append(out, "Scroll: j k")
		if own {
			if !tick {
				out = append(out, "Status: s")
			}
			out = append(out, "Edit: e", "Copy id: y")
		}
		return append(out, "Back: esc")
	}
	out = append(out, "Detail: enter")
	if own {
		if !tick {
			out = append(out, "Status: s", "Type: t")
		}
		out = append(out, "Edit: e", "Copy id: y")
	}
	out = append(out, "New bug: n", "Sort: o")
	if m.foldsRow() {
		out = append(out, "Fold: space")
	}
	if m.focus == paneDone {
		out = append(out, "Done tab: [ ]")
	}
	return out
}

// foldsRow says whether the space key opens or shuts the row under the
// cursor. Only the head of a tree row does; the tasks under it stay as they
// are, and the key does nothing there.
func (m Model) foldsRow() bool {
	rows := m.listOf()
	i := m.cursor()
	return i >= 0 && i < len(rows) && rows[i].tree && rows[i].depth == 0
}

// hintLine joins the hints that fit in w cells. It drops them from the right,
// but Help always stays.
func (m Model) hintLine(w int) string {
	hs := m.hints()
	for n := len(hs); n > 0; n-- {
		line := strings.Join(append(hs[:n:n], helpHint), " | ")
		if lipgloss.Width(line) <= w {
			return line
		}
	}
	return helpHint
}
