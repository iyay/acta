package tui

import (
	"fmt"
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
	t.Parallel()

	m := toPlans(planModel(t))
	shut := []string{"plans/2026-09-21-alpha", "plans/2026-09-23-lonely"}
	if got := rowIDs(m); !slices.Equal(got, shut) {
		t.Fatalf("plans list %q, want every plan shut %q", got, shut)
	}
	open := []string{"plans/2026-09-21-alpha", "plans/2026-09-21-alpha#task-1", "plans/2026-09-21-alpha#task-2", "plans/2026-09-23-lonely"}
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
	t.Parallel()

	m := press(toPlans(planModel(t)), " ")
	if it := m.Selected(); it == nil || it.Kind != board.KindPlan {
		t.Fatalf("a plan row selects %v, want the plan", it)
	}
	m = press(m, "j")
	it := m.Selected()
	if it == nil || it.ID != "plans/2026-09-21-alpha#task-1" || it.Kind != board.KindTask {
		t.Fatalf("a task row selects %v, want the task", it)
	}
	before := rowIDs(m)
	// space leaves a task row alone, and enter opens its own detail, because
	// only a head folds.
	if got := rowIDs(press(m, " ")); !slices.Equal(got, before) {
		t.Errorf("space on a task row changed the list to %q", got)
	}
	if after := press(m, "enter"); after.focus != paneDetail {
		t.Error("enter on a task row did not move the focus to the detail")
	} else if got := after.Selected(); got == nil || got.ID != "plans/2026-09-21-alpha#task-1" {
		t.Errorf("the detail shows %v, want the task the cursor was on", got)
	}
}

func TestTreeRowsWearTheirMarks(t *testing.T) {
	t.Parallel()

	m := toPlans(planModel(t))
	rows := m.rowsOf(m.listPane())
	if head := m.rowText(rows[1], m.board.Get(rows[1].id), 80); !strings.HasPrefix(head, "+ ") {
		t.Errorf("a shut plan row reads %q, want it to start with +", head)
	}
	m = press(m, "g", " ")
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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	m := toPlans(planModel(t))
	_ = press(m, " ")
	if got := len(rowIDs(m)); got != 2 {
		t.Errorf("the model before space now shows %d rows, want its 2 shut plans", got)
	}
	// Open a second plan from an already open model, so the older model holds
	// a map of its own that a shared one would change behind its back.
	first := press(m, " ")
	shut := []string{"plans/2026-09-21-alpha", "plans/2026-09-21-alpha#task-1", "plans/2026-09-21-alpha#task-2", "plans/2026-09-23-lonely"}
	if got := rowIDs(first); !slices.Equal(got, shut) {
		t.Fatalf("the first open plan gave %q, want %q", got, shut)
	}
	second := press(first, "j", "j", "j", " ")
	if got := rowIDs(first); !slices.Equal(got, shut) {
		t.Errorf("opening a second plan changed the older model to %q, want %q", got, shut)
	}
	if got := rowIDs(second); len(got) != 5 {
		t.Errorf("two open plans gave %q, want 5 rows", got)
	}
}

// foldState says which heads the model holds open and shut, so a test can see
// a fold that the list of rows does not show.
func foldState(m Model) string { return fmt.Sprint(m.openPlans, m.shutActs) }

func TestHAndLFoldThePlanTree(t *testing.T) {
	t.Parallel()

	shut := []string{"plans/2026-09-21-alpha", "plans/2026-09-23-lonely"}
	open := []string{"plans/2026-09-21-alpha", "plans/2026-09-21-alpha#task-1", "plans/2026-09-21-alpha#task-2", "plans/2026-09-23-lonely"}
	m := press(toPlans(planModel(t)), "l")
	if got := rowIDs(m); !slices.Equal(got, open) {
		t.Fatalf("l on a shut plan: %q, want %q", got, open)
	}
	if got := rowIDs(press(m, "l")); !slices.Equal(got, open) {
		t.Errorf("l on an open plan changed the list: %q", got)
	}
	onTask := press(m, "j", "j")
	if got := rowIDs(press(onTask, "l")); !slices.Equal(got, open) {
		t.Errorf("l on a task row changed the list: %q", got)
	}
	if got, want := foldState(press(onTask, "l")), foldState(onTask); got != want {
		t.Errorf("l on a task row folded a head: %s, want %s", got, want)
	}
	back := press(onTask, "h")
	if got := rowIDs(back); !slices.Equal(got, shut) {
		t.Errorf("h on a task row: %q, want %q", got, shut)
	}
	if it := back.Selected(); it == nil || it.ID != "plans/2026-09-21-alpha" {
		t.Errorf("h on a task row left the cursor on %v, want the plan", it)
	}
	if got := rowIDs(press(m, "h")); !slices.Equal(got, shut) {
		t.Errorf("h on an open plan: %q, want %q", got, shut)
	}
	if got := rowIDs(press(m, "h", "h")); !slices.Equal(got, shut) {
		t.Errorf("h on a shut plan opened it: %q", got)
	}
	if got := press(m, "right").top; got != (tabPlans+1)%len(topTabs) {
		t.Errorf("right no longer switches tabs, top %d", got)
	}
}

// The Done box of a tree tab lists a tree too, so h and l fold it the same
// way as the List box above it.
func TestHAndLFoldTheDoneTree(t *testing.T) {
	t.Parallel()

	m := toPlansDone(planModel(t))
	rows := doneRowIDs(m)
	if len(rows) == 0 {
		t.Fatal("the fixture has no finished plan, so this test proves nothing")
	}
	first := m.board.Get(rows[0])
	opened := append(append([]string{rows[0]}, first.Children...), rows[1:]...)
	if got := doneRowIDs(press(m, "l")); !slices.Equal(got, opened) {
		t.Errorf("l on a shut Done plan: %q, want %q", got, opened)
	}
	shut := append([]string{rows[0]}, opened[1+len(first.Children):]...)
	if got := doneRowIDs(press(m, "l", "h")); !slices.Equal(got, shut) {
		t.Errorf("l then h on a Done plan: %q, want %q", got, shut)
	}
}

// The detail box has no tree of its own, so h and l do nothing while it has
// the focus.
func TestHAndLDoNothingWhileTheDetailHasTheFocus(t *testing.T) {
	t.Parallel()

	open := []string{"plans/2026-09-21-alpha", "plans/2026-09-21-alpha#task-1", "plans/2026-09-21-alpha#task-2", "plans/2026-09-23-lonely"}
	m := press(toPlans(planModel(t)), " ", "j", "enter")
	if m.focus != paneDetail {
		t.Fatalf("enter on a task row left the focus on pane %d, want the detail", m.focus)
	}
	// Each key is tried on its own, so one cannot undo what the other did.
	for _, k := range []string{"h", "l"} {
		after := press(m, k)
		if got, want := foldState(after), foldState(m); got != want {
			t.Errorf("%s in the detail folded a head: %s, want %s", k, got, want)
		}
		if got := rowIDs(press(after, "esc")); !slices.Equal(got, open) {
			t.Errorf("the list behind the detail is now %q, want %q", got, open)
		}
	}
}
