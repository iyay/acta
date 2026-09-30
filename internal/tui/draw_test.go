package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// oldBody is the body the way draw built it before, with lipgloss joining the
// boxes. The new draw must give the same bytes.
func oldBody(m Model) string {
	g := m.geometry()
	var column []string
	for p := range g.side {
		if drawn := m.paneView(pane(p), g.side[p]); drawn != "" {
			column = append(column, drawn)
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, column...), m.paneView(paneDetail, g.detail))
}

var drawSizes = [][2]int{{60, 10}, {60, 40}, {80, 24}, {120, 40}, {200, 55}, {61, 5}, {300, 12}}

// The wide layout is joined by hand now. It must be the same screen the
// lipgloss join gave, on every size, with the detail box focused or not.
func TestDrawJoinMatchesTheOldJoin(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		for _, s := range drawSizes {
			m := sized(paneModel(t, p), s[0], s[1])
			if !m.geometry().wide {
				continue
			}
			want := m.frameFrom(strings.Split(oldBody(m), "\n"), false)
			if got := m.draw(); got != want {
				t.Errorf("pane %d, %dx%d: the frame changed", p, s[0], s[1])
			}
		}
	}
}

// Every line must fill the terminal exactly. A short line leaves old text on
// screen, and a long one wraps and breaks the whole layout.
func TestDrawEveryLineIsTheScreenWidth(t *testing.T) {
	t.Parallel()

	uni := press(boardModel(t, "# Plan ✓ ünïcödé 日本語 "+strings.Repeat("wide ", 40)+"\n\n"+strings.Repeat("日本語のテキスト ", 30)+"\n"), tabKey(tabPlans))
	for _, base := range []Model{paneModel(t, paneDetail), uni} {
		for _, s := range append(drawSizes, [2]int{40, 20}, [2]int{59, 20}) {
			m := sized(base, s[0], s[1])
			for i, ln := range strings.Split(m.draw(), "\n") {
				if got := lipgloss.Width(ln); got != s[0] {
					t.Errorf("%dx%d line %d is %d wide", s[0], s[1], i, got)
				}
			}
		}
	}
}
