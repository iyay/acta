package tui

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"pm-board/internal/board"
	"pm-board/internal/config"
)

func sized(m Model, w, h int) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

func TestViewShowsListAndDetail(t *testing.T) {
	m := sized(press(newModel(t), "j", "j"), 120, 40)
	v := m.View()
	for _, want := range []string{
		"pmb · basic · active · live",
		"[1] Stories 5", "[2] Tasks 2", "[3] Bugs 2",
		"Alpha story", "A-1", "in-progress", "1/2",
		"First step", "Second step",
		"specs/2026-09-20-alpha",
		"Some text.",
	} {
		if !strings.Contains(v, want) {
			t.Errorf("view is missing %q", want)
		}
	}
	for _, line := range strings.Split(v, "\n") {
		if w := len([]rune(line)); w > 120 {
			t.Fatalf("line wider than the terminal (%d): %q", w, line)
		}
	}
}

func TestViewTaskShowsParentAndOwnSection(t *testing.T) {
	m := sized(press(newModel(t), "2", "j"), 120, 40)
	v := m.View()
	if !strings.Contains(v, "Second step · Alpha story") {
		t.Error("task row does not show its parent")
	}
	if strings.Contains(v, "Task 1: First step") {
		t.Error("task detail shows other sections of the plan")
	}
}

func TestViewProblemsGroupAndStatus(t *testing.T) {
	m := sized(press(newModel(t), "a"), 120, 60)
	v := m.View()
	if !strings.Contains(v, "! ") || !strings.Contains(v, "untyped (1)") {
		t.Error("problem marker or untyped group missing")
	}
	m.status = "pm: x status done ✓ committed"
	m.manual = true
	v = m.View()
	if !strings.Contains(v, "✓ committed") || !strings.Contains(v, "manual") {
		t.Error("status bar or manual mode missing")
	}
}

func TestViewEmptyRepo(t *testing.T) {
	b, err := board.Load(config.Default(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	m := sized(New(config.Default(t.TempDir()), b, true), 100, 30)
	if !strings.Contains(m.View(), "no .pm/ yet") {
		t.Error("empty repo message missing")
	}
}

func TestViewHelpAndPopup(t *testing.T) {
	m := sized(press(newModel(t), "?"), 100, 30)
	if !strings.Contains(m.View(), "new bug") {
		t.Error("help screen missing")
	}
	m = sized(press(newModel(t), "3", "s"), 100, 30)
	if !strings.Contains(m.View(), "wontfix") {
		t.Error("popup options not shown")
	}
	m = sized(press(newModel(t), "n", "a", "b"), 100, 30)
	if !strings.Contains(m.View(), "new bug slug: ab") {
		t.Error("slug prompt not shown")
	}
}

func TestViewNarrowNeverOverflows(t *testing.T) {
	for _, wh := range [][2]int{{80, 24}, {40, 12}, {30, 10}} {
		w, h := wh[0], wh[1]
		m := sized(newModel(t), w, h)
		for name, sm := range map[string]Model{
			"base": m, "tasks": press(m, "2"), "popup": press(m, "3", "s"),
			"slug": press(m, "n"), "help": press(m, "?"),
		} {
			for _, line := range strings.Split(sized(sm, w, h).View(), "\n") {
				if got := lipgloss.Width(line); got > w {
					t.Errorf("%dx%d %s: line width %d > %d: %q", w, h, name, got, w, line)
				}
			}
		}
	}
}
