package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHintsFollowTheFocusedPane(t *testing.T) {
	t.Parallel()

	list := press(sized(newModel(t), 200, 40), tabKey(tabSpecs))
	for _, h := range []string{"Detail: enter", "Status: s", "Type: t", "Edit: e", "Copy id: y", "New bug: n", "Sort: o"} {
		if !slices.Contains(list.hints(), h) {
			t.Errorf("the list pane is missing %q: %v", h, list.hints())
		}
	}
	if slices.Contains(list.hints(), "Done tab: [ ]") || slices.Contains(list.hints(), "Fold: space") {
		t.Errorf("the Specs list shows a key it cannot use: %v", list.hints())
	}
	done := press(list, "tab")
	if done.focus != paneDone || !slices.Contains(done.hints(), "Done tab: [ ]") {
		t.Errorf("the Done pane is missing Done tab: %v", done.hints())
	}
	detail := press(list, "enter")
	if detail.focus != paneDetail {
		t.Fatalf("enter gave focus %d", detail.focus)
	}
	for _, h := range []string{"Scroll: j k", "Status: s", "Edit: e", "Copy id: y", "Back: esc"} {
		if !slices.Contains(detail.hints(), h) {
			t.Errorf("the detail pane is missing %q: %v", h, detail.hints())
		}
	}
	if slices.Contains(detail.hints(), "Detail: enter") {
		t.Errorf("the detail pane offers enter: %v", detail.hints())
	}
	plans := press(list, tabKey(tabPlans))
	if !slices.Contains(plans.hints(), "Fold: space") {
		t.Errorf("the Plans list is missing Fold: %v", plans.hints())
	}
}

func TestHintsOnATaskRowOfferTickNotStatus(t *testing.T) {
	t.Parallel()

	m := press(sized(newModel(t), 200, 40), tabKey(tabPlans))
	// Open the first plan and step onto its first task.
	m = press(m, "l", "j")
	it := m.Selected()
	if it == nil || it.Kind != "task" {
		t.Fatalf("the row under the cursor is %+v, want a task", it)
	}
	hs := m.hints()
	if hs[0] != "Tick: +" || hs[1] != "Untick: -" {
		t.Errorf("a task row should start with Tick and Untick: %v", hs)
	}
	if slices.Contains(hs, "Status: s") || slices.Contains(hs, "Type: t") {
		t.Errorf("a task row offers a key the popup refuses: %v", hs)
	}
}

func TestHintLineDropsFromTheRightAndKeepsHelp(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 200, 40)
	full := m.hintLine(1000)
	if !strings.HasSuffix(full, " | "+helpHint) || !strings.HasPrefix(full, m.hints()[0]) {
		t.Fatalf("the full hint line is %q", full)
	}
	for w := range lipgloss.Width(full) {
		got := m.hintLine(w)
		if !strings.HasSuffix(got, helpHint) {
			t.Fatalf("at %d cells Help is gone: %q", w, got)
		}
		if lipgloss.Width(got) > w && got != helpHint {
			t.Fatalf("at %d cells the line %q does not fit", w, got)
		}
	}
	for w := 30; w <= 200; w++ {
		last := plain(lastLine(sized(m, w, 20).View()))
		if lipgloss.Width(last) != w {
			t.Fatalf("at %d columns the status line is %d cells", w, lipgloss.Width(last))
		}
	}
}

func TestHintsSkipTheKeysARowRefuses(t *testing.T) {
	t.Parallel()

	// The group row holds no item, so the popup and the editor have nothing
	// to work on and say so when they are pressed.
	specs := sized(press(newModel(t), tabKey(tabSpecs)), 200, 40)
	rows := specs.rowsOf(paneList)
	i := slices.IndexFunc(rows, func(r row) bool { return r.group })
	if i < 0 {
		t.Fatal("the fixture has no group row to stand on")
	}
	group := specs
	group.sel[paneList], group.idx[paneList] = "", i
	for _, h := range []string{"Status: s", "Type: t", "Edit: e", "Copy id: y"} {
		if slices.Contains(group.hints(), h) {
			t.Errorf("the group row shows %q, which does nothing: %v", h, group.hints())
		}
	}
	if !slices.Contains(group.hints(), "Detail: enter") {
		t.Errorf("the group row should still open or shut itself: %v", group.hints())
	}

	// A task sits under its plan, so space cannot fold it. The plan head
	// above it is the row space folds.
	task := press(sized(newModel(t), 200, 40), tabKey(tabPlans), "l", "j")
	if slices.Contains(task.hints(), "Fold: space") {
		t.Errorf("a task row shows Fold: space, which does nothing: %v", task.hints())
	}
	if !slices.Contains(task.hints(), "Tick: +") {
		t.Errorf("a task row should still offer the tick keys: %v", task.hints())
	}
}
