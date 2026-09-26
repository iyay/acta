package tui

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"pm-board/internal/board"
	"pm-board/internal/config"
	"pm-board/internal/write"
)

func fixture(t *testing.T) (config.Config, *board.Board) {
	t.Helper()
	dir, err := filepath.Abs("../board/testdata/basic")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default(dir)
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return cfg, b
}

func newModel(t *testing.T) Model {
	t.Helper()
	cfg, b := fixture(t)
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return m
}

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		next, _ := m.Update(key(k))
		m = next.(Model)
	}
	return m
}

func rowIDs(m Model) []string {
	var out []string
	for _, r := range m.rows() {
		out = append(out, r.id)
	}
	return out
}

func TestTabsAndDefaultRows(t *testing.T) {
	m := newModel(t)
	if got := strings.Join(rowIDs(m), " "); got != "plans/2026-09-23-lonely specs/2026-09-22-beta specs/2026-09-20-alpha specs/2026-09-18-weird specs/2026-09-17-broken "+groupRowID {
		t.Fatalf("stories rows %q", got)
	}
	m = press(m, "2")
	if got := rowIDs(m); len(got) != 2 || got[0] != "plans/2026-09-23-lonely#task-1" {
		t.Fatalf("tasks rows %v", got)
	}
	m = press(m, "tab")
	if m.tab != tabBugs || len(m.rows()) != 2 {
		t.Fatalf("tab %v rows %v", m.tab, rowIDs(m))
	}
	m = press(m, "tab")
	if m.tab != tabStories {
		t.Fatalf("tab did not wrap: %v", m.tab)
	}
}

func TestMoveKeys(t *testing.T) {
	m := newModel(t)
	if m.Selected().ID != "plans/2026-09-23-lonely" {
		t.Fatalf("first selection %s", m.Selected().ID)
	}
	m = press(m, "j", "j")
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("after jj: %s", m.Selected().ID)
	}
	m = press(m, "k")
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("after k: %s", m.Selected().ID)
	}
	m = press(m, "G")
	if m.Selected() != nil || m.rows()[m.cursor()].id != groupRowID {
		t.Fatal("G should land on the untyped group row")
	}
	m = press(m, "j")
	if m.rows()[m.cursor()].id != groupRowID {
		t.Fatal("j past the end should stay on the last row")
	}
	m = press(m, "g", "k")
	if m.Selected().ID != "plans/2026-09-23-lonely" {
		t.Fatal("k at the top should stay on the first row")
	}
}

func TestSelectionIsPerTab(t *testing.T) {
	m := press(newModel(t), "j", "2", "1")
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("stories selection lost: %s", m.Selected().ID)
	}
}

func TestShowAllAndGroup(t *testing.T) {
	m := newModel(t)
	before := len(m.rows())
	m = press(m, "a")
	if !m.showAll || len(m.rows()) <= before {
		t.Fatalf("a did not show more rows: %d -> %d", before, len(m.rows()))
	}
	m = press(m, "a", "G", "enter")
	if !m.groupOpen || rowIDs(m)[len(m.rows())-1] != "docs/superpowers/specs/2026-01-01-old" {
		t.Fatalf("group did not open: %v", rowIDs(m))
	}
	m = press(m, "enter") // the cursor is still on the group row
	if m.groupOpen {
		t.Fatal("enter on the group row should close it")
	}
}

func TestSearch(t *testing.T) {
	m := press(newModel(t), "/", "c", "r", "a", "s", "h")
	if !m.searching || m.query != "crash" {
		t.Fatalf("searching %v query %q", m.searching, m.query)
	}
	for _, id := range rowIDs(m) {
		it := m.board.Get(id)
		if !strings.Contains(strings.ToLower(it.Title+it.Slug+it.Ref), "crash") {
			t.Errorf("row %s does not match", id)
		}
	}
	m = press(m, "backspace")
	if m.query != "cras" {
		t.Fatalf("backspace: %q", m.query)
	}
	m = press(m, "enter")
	if m.searching || m.query != "cras" {
		t.Fatal("enter should stop typing and keep the query")
	}
	m = press(m, "esc")
	if m.query != "" {
		t.Fatal("esc should clear the query")
	}
	m = press(m, "/", "x", "esc")
	if m.searching || m.query != "" {
		t.Fatal("esc while typing should stop and clear")
	}
}

func TestPopupRefusals(t *testing.T) {
	m := press(newModel(t), "2", "s")
	if m.popup != nil || !strings.Contains(m.status, "checkboxes") {
		t.Fatalf("task popup: %v %q", m.popup, m.status)
	}
	m = press(newModel(t), "a", "G", "enter", "j", "t")
	if m.popup != nil || !strings.Contains(m.status, "legacy") {
		t.Fatalf("legacy popup: %v %q", m.popup, m.status)
	}
}

func TestStatusPopupSetsValue(t *testing.T) {
	m := newModel(t)
	var got []string
	m.setValue = func(id, field, value string) (write.Outcome, error) {
		got = []string{id, field, value}
		return write.Outcome{Committed: true}, nil
	}
	m = press(m, "3", "s")
	if m.popup == nil || m.popup.field != "status" || strings.Join(m.popup.options, ",") != "open,fixing,fixed,wontfix" || m.popup.idx != 0 {
		t.Fatalf("popup %+v", m.popup)
	}
	m = press(m, "j", "j", "enter")
	if strings.Join(got, " ") != "bugs/2026-09-26-open status fixed" {
		t.Fatalf("setValue got %v", got)
	}
	if m.popup != nil || !strings.Contains(m.status, "committed") {
		t.Fatalf("popup %v status %q", m.popup, m.status)
	}
}

func TestPopupEscAndOutcomes(t *testing.T) {
	m := press(newModel(t), "3", "t", "esc")
	if m.popup != nil {
		t.Fatal("esc should close the popup")
	}
	m.setValue = func(id, field, value string) (write.Outcome, error) {
		return write.Outcome{Skipped: true, Reason: "HEAD is detached"}, nil
	}
	m = press(m, "t", "enter")
	if !strings.Contains(m.status, "not committed: HEAD is detached") {
		t.Fatalf("status %q", m.status)
	}
	m.setValue = func(id, field, value string) (write.Outcome, error) {
		return write.Outcome{}, errors.New("boom")
	}
	m = press(m, "t", "enter")
	if !strings.Contains(m.status, "error: boom") {
		t.Fatalf("status %q", m.status)
	}
}

func TestReloadKeepsSelection(t *testing.T) {
	cfg, b := fixture(t)
	m := press(newModel(t), "j", "j") // alpha
	next, _ := m.Update(reloadMsg{b: b})
	m = next.(Model)
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("selection moved on reload: %s", m.Selected().ID)
	}

	// Drop alpha from the board: the cursor stays at the same row number.
	var kept []*board.Item
	for _, it := range b.Items {
		if it.ID != "specs/2026-09-20-alpha" {
			kept = append(kept, it)
		}
	}
	smaller, _ := board.Load(cfg)
	smaller.Items = kept
	next, _ = m.Update(reloadMsg{b: smaller})
	m = next.(Model)
	if m.Selected() == nil || m.Selected().ID != "specs/2026-09-18-weird" {
		t.Fatalf("after alpha vanished: %v", m.Selected())
	}

	next, _ = m.Update(reloadMsg{err: errors.New("disk")})
	m = next.(Model)
	if !strings.Contains(m.status, "reload failed") {
		t.Fatalf("status %q", m.status)
	}
}

func TestWatchFailedGoesManual(t *testing.T) {
	next, _ := newModel(t).Update(WatchFailed(errors.New("too many files")))
	m := next.(Model)
	if !m.manual || !strings.Contains(m.status, "press r") {
		t.Fatalf("manual %v status %q", m.manual, m.status)
	}
}

func TestNewBugSlugInput(t *testing.T) {
	m := press(newModel(t), "n", "a", "B", "-", "1", " ", "backspace")
	if m.slug == nil || *m.slug != "a-" {
		t.Fatalf("slug %v", m.slug)
	}
	m = press(m, "esc")
	if m.slug != nil {
		t.Fatal("esc should cancel the slug prompt")
	}
}

func TestScrollAndQuit(t *testing.T) {
	m := press(newModel(t), "ctrl+d", "ctrl+d", "ctrl+u")
	if m.scroll != 10 {
		t.Fatalf("scroll %d", m.scroll)
	}
	m = press(m, "ctrl+u", "ctrl+u")
	if m.scroll != 0 {
		t.Fatalf("scroll went below 0: %d", m.scroll)
	}
	if _, cmd := m.Update(key("q")); cmd == nil {
		t.Fatal("q should return a quit command")
	}
}
