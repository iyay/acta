package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
)

// toPlans puts the focus on the open plans list, the way a reader gets there.
func toPlans(m Model) Model { return press(m, tabKey(tabPlans)) }

// toPlansDone puts the focus on the finished plans list.
func toPlansDone(m Model) Model { return press(toPlans(m), "tab") }

func planModel(t *testing.T) Model {
	t.Helper()
	return sized(newModel(t), 160, 50)
}

func TestPlanRowsStartShutAndSpaceOrEnterFlipThem(t *testing.T) {
	m := toPlans(planModel(t))
	shut := []string{"plans/2026-09-23-lonely", "plans/2026-09-21-alpha"}
	if got := rowIDs(m); !slices.Equal(got, shut) {
		t.Fatalf("plans list %q, want every plan shut %q", got, shut)
	}
	open := []string{"plans/2026-09-23-lonely", "plans/2026-09-23-lonely#task-1", "plans/2026-09-21-alpha"}
	for _, k := range []string{" ", "enter"} {
		opened := press(m, k)
		if got := rowIDs(opened); !slices.Equal(got, open) {
			t.Errorf("after %q the list is %q, want %q", k, got, open)
		}
		if opened.focus == paneDetail {
			t.Errorf("%q on a plan row moved the focus to the detail", k)
		}
		if got := rowIDs(press(opened, k)); !slices.Equal(got, shut) {
			t.Errorf("a second %q left %q, want the plan shut again", k, got)
		}
	}
}

func TestTreeRowsSelectTheirOwnDetail(t *testing.T) {
	m := press(toPlans(planModel(t)), " ")
	if it := m.Selected(); it == nil || it.Kind != board.KindPlan {
		t.Fatalf("a plan row selects %v, want the plan", it)
	}
	m = press(m, "j")
	it := m.Selected()
	if it == nil || it.ID != "plans/2026-09-23-lonely#task-1" || it.Kind != board.KindTask {
		t.Fatalf("a task row selects %v, want the task", it)
	}
	before := rowIDs(m)
	for _, k := range []string{" ", "enter"} {
		after := press(m, k)
		if got := rowIDs(after); !slices.Equal(got, before) {
			t.Errorf("%q on a task row changed the list to %q", k, got)
		}
		if after.focus == paneDetail {
			t.Errorf("%q on a task row moved the focus to the detail", k)
		}
	}
}

func TestTreeRowsWearTheirMarks(t *testing.T) {
	m := toPlans(planModel(t))
	rows := m.rowsOf(m.listPane())
	if head := m.rowText(rows[1], m.board.Get(rows[1].id), 80); !strings.HasPrefix(head, "+ ") {
		t.Errorf("a shut plan row reads %q, want it to start with +", head)
	}
	m = press(m, "j", " ")
	rows = m.rowsOf(m.listPane())
	want := map[string]string{
		"plans/2026-09-21-alpha":        "- ",
		"plans/2026-09-21-alpha#task-1": "  " + dotDone + " ",
		"plans/2026-09-21-alpha#task-2": "  " + dotGoing + " ",
	}
	for _, r := range rows {
		mark, ok := want[r.id]
		if !ok {
			continue
		}
		if head := m.rowText(r, m.board.Get(r.id), 80); !strings.HasPrefix(head, mark) {
			t.Errorf("row %s reads %q, want it to start with %q", r.id, head, mark)
		}
		delete(want, r.id)
	}
	if len(want) != 0 {
		t.Errorf("rows never drawn: %v", want)
	}
}

func TestAPlanWithNoTasksAddsNoRows(t *testing.T) {
	cfg := treeCfg(t, map[string]string{".acta/plans/2026-09-20-empty.md": "# Empty plan\n"})
	m := toPlans(sized(detailModel(t, cfg), 160, 50))
	if got := rowIDs(m); !slices.Equal(got, []string{"plans/2026-09-20-empty"}) {
		t.Fatalf("plans list %q", got)
	}
	r := m.rowsOf(m.listPane())[0]
	if head := m.rowText(r, m.board.Get(r.id), 80); !strings.HasPrefix(head, "+ ") {
		t.Errorf("a plan with no tasks reads %q, want it to start with +", head)
	}
	if got := rowIDs(press(m, " ")); !slices.Equal(got, []string{"plans/2026-09-20-empty"}) {
		t.Errorf("opening a plan with no tasks gave %q, want no new row", got)
	}
}

func TestTheDonePlansTreeBehavesTheSame(t *testing.T) {
	m := toPlansDone(planModel(t))
	rows := doneRowIDs(m)
	if len(rows) == 0 {
		t.Fatal("the fixture has no finished plan, so this test proves nothing")
	}
	for _, id := range rows {
		if it := m.board.Get(id); it == nil || it.Kind != board.KindPlan {
			t.Fatalf("a shut Done tree holds %s, want plans only", id)
		}
	}
	first := m.board.Get(rows[0])
	opened := doneRowIDs(press(m, " "))
	want := append(append([]string{rows[0]}, first.Children...), rows[1:]...)
	if !slices.Equal(opened, want) {
		t.Errorf("opening %s in Done gave %q, want %q", rows[0], opened, want)
	}
}

func TestOpeningAPlanLeavesOlderModelsAlone(t *testing.T) {
	m := toPlans(planModel(t))
	_ = press(m, " ")
	if got := len(rowIDs(m)); got != 2 {
		t.Errorf("the model before space now shows %d rows, want its 2 shut plans", got)
	}
	// Open a second plan from an already open model, so the older model holds
	// a map of its own that a shared one would change behind its back.
	first := press(m, " ")
	shut := []string{"plans/2026-09-23-lonely", "plans/2026-09-23-lonely#task-1", "plans/2026-09-21-alpha"}
	if got := rowIDs(first); !slices.Equal(got, shut) {
		t.Fatalf("the first open plan gave %q, want %q", got, shut)
	}
	second := press(first, "j", "j", " ")
	if got := rowIDs(first); !slices.Equal(got, shut) {
		t.Errorf("opening a second plan changed the older model to %q, want %q", got, shut)
	}
	if got := rowIDs(second); len(got) != 5 {
		t.Errorf("two open plans gave %q, want 5 rows", got)
	}
}
